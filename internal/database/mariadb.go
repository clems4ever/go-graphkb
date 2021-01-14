package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/cheggaaa/pb"
	"github.com/clems4ever/go-graphkb/internal/knowledge"
	"github.com/clems4ever/go-graphkb/internal/schema"
	"github.com/clems4ever/go-graphkb/internal/utils"
	mapset "github.com/deckarep/golang-set"
	mysql "github.com/go-sql-driver/mysql"
	"github.com/golang-collections/go-datastructures/queue"
)

// MariaDB mariadb as graph storage backend
type MariaDB struct {
	db *sql.DB
}

// SourceRelation represent a relation coming from a source
type SourceRelation struct {
	knowledge.Relation `json:",inline"`
	Source             string `json:"source"`
}

// NewMariaDB create an instance of mariadb
func NewMariaDB(username string, password string, host string, databaseName string, allowCleartextPassword bool) *MariaDB {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@(%s)/%s?allowCleartextPasswords=%s", username, password,
		host, databaseName, strconv.FormatBool(allowCleartextPassword)))
	if err != nil {
		log.Fatal(err)
	}
	db.SetMaxIdleConns(0)
	return &MariaDB{db: db}
}

// InitializeSchema initialize the schema of the database
func (m *MariaDB) InitializeSchema() error {
	// Create the table storing data sources tokens
	_, err := m.db.ExecContext(context.Background(), `
		CREATE TABLE IF NOT EXISTS sources (
			id INTEGER AUTO_INCREMENT NOT NULL,
			name VARCHAR(64) NOT NULL,
			auth_token VARCHAR(64) NOT NULL,
		
			CONSTRAINT pk_source PRIMARY KEY (id),

			UNIQUE unique_source_idx (name, auth_token)
		)`)
	if err != nil {
		return err
	}

	// type must be part of the primary key to be a partition key
	_, err = m.db.ExecContext(context.Background(), `
CREATE TABLE IF NOT EXISTS assets (
	id INT NOT NULL AUTO_INCREMENT,
	source_id INT NOT NULL,
	value VARCHAR(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL,
	type VARCHAR(64) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,

	CONSTRAINT pk_asset PRIMARY KEY (id),
	CONSTRAINT fk_asset_source FOREIGN KEY (source_id) REFERENCES sources (id) ON DELETE CASCADE,

	UNIQUE unique_asset_idx (type, value, source_id),
	INDEX value_idx (value),
	INDEX type_idx (type))`)
	if err != nil {
		return err
	}

	// type must be part of the primary key to be a partition key
	_, err = m.db.ExecContext(context.Background(), `
CREATE TABLE IF NOT EXISTS relations (
	id INT NOT NULL AUTO_INCREMENT,
	from_id INT NOT NULL,
	to_id INT NOT NULL,
	type VARCHAR(64) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,
	source_id INT NOT NULL,

	CONSTRAINT pk_relation PRIMARY KEY (id),
	CONSTRAINT fk_from FOREIGN KEY (from_id) REFERENCES assets (id) ON DELETE CASCADE,
	CONSTRAINT fk_to FOREIGN KEY (to_id) REFERENCES assets (id) ON DELETE CASCADE,
	CONSTRAINT fk_relation_source FOREIGN KEY (source_id) REFERENCES sources (id),

	INDEX full_relation_type_from_to_idx (type, from_id, to_id),
    INDEX full_relation_type_to_from_idx (type, to_id, from_id),
    INDEX full_relation_from_type_to_idx (from_id, type, to_id),
    INDEX full_relation_from_to_type_idx (from_id, to_id, type),
    INDEX full_relation_to_from_type_idx (to_id, from_id, type),
    INDEX full_relation_to_type_from_idx (to_id, type, from_id))`)
	if err != nil {
		return err
	}

	// Create the table storing the schema graphs
	_, err = m.db.ExecContext(context.Background(), `
CREATE TABLE IF NOT EXISTS graph_schema (
	id INTEGER AUTO_INCREMENT NOT NULL,
	source_id INT NOT NULL,
	graph TEXT NOT NULL,
	timestamp TIMESTAMP,

	CONSTRAINT pk_schema PRIMARY KEY (id),
	CONSTRAINT fk_schema_source FOREIGN KEY (source_id) REFERENCES sources (id) ON DELETE CASCADE)`)
	if err != nil {
		return err
	}

	_, err = m.db.ExecContext(context.Background(), `
		CREATE TABLE IF NOT EXISTS query_history (
			id INTEGER AUTO_INCREMENT NOT NULL,
			timestamp TIMESTAMP,
			query_cypher TEXT NOT NULL,
			query_sql TEXT NOT NULL,
			execution_time_ms INT,
			status ENUM('SUCCESS', 'FAILURE'),
			error TEXT,
			CONSTRAINT pk_history PRIMARY KEY (id)
		)`)
	if err != nil {
		return err
	}

	return nil
}

// AssetRegistry store ID of assets in a cache
type AssetRegistry struct {
	cache map[knowledge.AssetKey]int64
}

// Set id of an asset
func (ar *AssetRegistry) Set(a knowledge.AssetKey, idx int64) {
	ar.cache[a] = idx
}

// Get id of an asset
func (ar *AssetRegistry) Get(a knowledge.AssetKey) (int64, bool) {
	idx, ok := ar.cache[a]
	return idx, ok
}

func isDuplicateEntryError(err error) bool {
	driverErr, ok := err.(*mysql.MySQLError)
	return ok && driverErr.Number == 1062
}

func isUnknownTableError(err error) bool {
	driverErr, ok := err.(*mysql.MySQLError)
	return ok && driverErr.Number == 1051
}

func (m *MariaDB) resolveAssets(sourceID int, assets []knowledge.AssetKey, registry *AssetRegistry) error {
	bar := pb.StartNew(len(assets))
	defer bar.Finish()

	stmt, err := m.db.PrepareContext(context.Background(), "SELECT id FROM assets WHERE type = ? AND value = ? AND source_id = ?")
	if err != nil {
		return fmt.Errorf("Unable to prepare statement: %v", err)
	}

	for _, a := range assets {
		q, err := stmt.QueryContext(context.Background(), a.Type, a.Key, sourceID)
		if err != nil {
			return fmt.Errorf("Unable to query asset %v: %v", a, err)
		}
		defer q.Close()

		for q.Next() {
			var idx int64
			if err := q.Scan(&idx); err != nil {
				return fmt.Errorf("Unable to retrieve id for %v: %v", a, err)
			}
			registry.Set(knowledge.AssetKey(a), idx)
		}

		bar.Increment()
	}
	return nil
}

func (m *MariaDB) upsertAssets(sourceID int, assets []knowledge.Asset, registry *AssetRegistry) (int64, error) {
	if len(assets) == 0 {
		return 0, nil
	}

	unresolved := []knowledge.Asset{}
	for _, a := range assets {
		_, ok := registry.Get(knowledge.AssetKey(a))
		if !ok {
			unresolved = append(unresolved, a)
		}
	}

	bar := pb.StartNew(len(unresolved))
	defer bar.Finish()
	insertedCount := int64(0)

	assetChunks := utils.ChunkSlice(unresolved, 10000).([][]interface{})

	// A chunk of assets to store in a transaction
	for _, assetChunk := range assetChunks {
		insertQuery, err := m.db.PrepareContext(context.Background(), `
INSERT INTO assets (type, value, source_id) VALUES (?, ?, ?)`)
		if err != nil {
			return 0, fmt.Errorf("Unable to prepare asset insertion query: %v", err)
		}

		for _, aC := range assetChunk {
			a := aC.(knowledge.Asset)

			res, err := insertQuery.ExecContext(context.Background(), a.Type, a.Key, sourceID)
			if err != nil {
				return 0, fmt.Errorf("Unable to insert asset %v: %v", a, err)
			}
			idx, err := res.LastInsertId()
			if err != nil {
				return 0, fmt.Errorf("Unable to retrieve last inserted ID: %v", err)
			}
			registry.Set(knowledge.AssetKey(a), idx)

			atomic.AddInt64(&insertedCount, 1)
			bar.Increment()
		}
	}
	return insertedCount, nil
}

func (m *MariaDB) upsertRelations(sourceID int, relations []knowledge.Relation, registry *AssetRegistry) (int64, error) {
	if len(relations) == 0 {
		return 0, nil
	}
	bar := pb.StartNew(len(relations))
	defer bar.Finish()
	insertedCount := int64(0)

	relationChunks := utils.ChunkSlice(relations, 1000).([][]interface{})

	for _, relationChunk := range relationChunks {
		q, err := m.db.PrepareContext(context.Background(),
			"INSERT INTO relations (from_id, to_id, type, source_id) VALUES (?, ?, ?, ?)")
		if err != nil {
			return 0, fmt.Errorf("Unable to prepare relation insertion query: %v", err)
		}
		defer q.Close()

		for _, rC := range relationChunk {
			r := rC.(knowledge.Relation)
			idxFrom, ok := registry.Get(r.From)
			if !ok {
				fmt.Printf("[WARNING] ID of asset %v (from) has not been found in cache\n", r.From)
				continue
			}

			idxTo, ok := registry.Get(r.To)
			if !ok {
				fmt.Printf("[WARNING] ID of asset %v (to) has not been found in cache\n", r.To)
				continue
			}

			_, err = q.ExecContext(context.Background(), idxFrom, idxTo, r.Type, sourceID)
			if err != nil {
				if isDuplicateEntryError(err) {
					bar.Increment()
					continue
				}
				return 0, fmt.Errorf("Unable to insert relation %v (%d -> %d): %v", r, idxFrom, idxTo, err)
			}
			bar.Increment()
			insertedCount++
		}
	}
	bar.Finish()
	return insertedCount, nil
}

func (m *MariaDB) removeRelations(sourceID int, relations []knowledge.Relation) (int64, int64, error) {
	if len(relations) == 0 {
		return 0, 0, nil
	}

	bar := pb.StartNew(len(relations))
	defer bar.Finish()
	removedCount := int64(0)

	stmt, err := m.db.PrepareContext(context.Background(), `
DELETE r FROM relations r
INNER JOIN assets a ON r.from_id = a.id
INNER JOIN assets b ON r.to_id = b.id
WHERE a.type = ? AND a.value = ? AND b.type = ? AND b.value = ? AND r.type = ? AND r.source_id = ?`)
	if err != nil {
		return 0, 0, err
	}
	defer stmt.Close()

	for _, r := range relations {
		res, err := stmt.ExecContext(context.Background(),
			r.From.Type, r.From.Key, r.To.Type, r.To.Key, r.Type, sourceID)
		if err != nil {
			return 0, 0, fmt.Errorf("Unable to detete relation %v: %v", r, err)
		}
		bar.Increment()
		rCount, err := res.RowsAffected()
		if err != nil {
			return 0, 0, fmt.Errorf("Unable to count rows affected by relations deletion: %v", err)
		}
		removedCount += rCount
	}

	res, err := m.db.ExecContext(context.Background(), `
DELETE a FROM assets a
WHERE id NOT IN (select from_id from relations)
AND id NOT IN (select to_id from relations)`)
	if err != nil {
		return 0, 0, fmt.Errorf("Unable to delete assets: %v", err)
	}

	removedAssetsCount, err := res.RowsAffected()
	if err != nil {
		return 0, 0, fmt.Errorf("Unable to count rows affected by assets deletion: %v", err)
	}
	return removedCount, removedAssetsCount, err
}

func (m *MariaDB) resolveSourceID(sourceName string) (int, error) {
	r, err := m.db.QueryContext(context.Background(), "SELECT id FROM sources WHERE name = ? LIMIT 1", sourceName)
	if err != nil {
		return 0, err
	}

	var sourceID int
	for r.Next() {
		err = r.Scan(&sourceID)
		if err != nil {
			return 0, err
		}
	}
	return sourceID, nil
}

// UpdateGraph update graph with bulk of operations
func (m *MariaDB) UpdateGraph(source string, bulk *knowledge.GraphUpdatesBulk) error {
	registry := AssetRegistry{cache: make(map[knowledge.AssetKey]int64)}
	now := time.Now()

	assetKeysSet := mapset.NewSet()
	for _, a := range bulk.GetAssetUpserts() {
		assetKeysSet.Add(knowledge.AssetKey(a))
	}
	for _, r := range bulk.GetRelationUpserts() {
		assetKeysSet.Add(r.From)
		assetKeysSet.Add(r.To)
	}

	assetKeys := []knowledge.AssetKey{}
	for a := range assetKeysSet.Iter() {
		assetKeys = append(assetKeys, a.(knowledge.AssetKey))
	}

	fmt.Println("Retrieve source ID from database")
	sourceID, err := m.resolveSourceID(source)
	if err != nil {
		return fmt.Errorf("Unable to resolve source ID from name %s: %v", source, err)
	}

	fmt.Println("Start resolving assets")
	err = m.resolveAssets(sourceID, assetKeys, &registry)
	if err != nil {
		return fmt.Errorf("Unable to resolve assets: %v", err)
	}

	fmt.Println("Start upserting assets")
	count, err := m.upsertAssets(sourceID, bulk.GetAssetUpserts(), &registry)
	if err != nil {
		return fmt.Errorf("Unable to upsert assets: %v", err)
	}

	nowAssetInsert := time.Now()
	fmt.Printf("%d assets inserted in %fs\n", count, nowAssetInsert.Sub(now).Seconds())

	fmt.Println("Start upserting relations")
	count, err = m.upsertRelations(sourceID, bulk.GetRelationUpserts(), &registry)
	if err != nil {
		return fmt.Errorf("Unable to upsert relations: %v", err)
	}
	nowRelationInsert := time.Now()
	fmt.Printf("%d relations inserted in %fs\n", count, nowRelationInsert.Sub(nowAssetInsert).Seconds())

	relCount, assetsCount, err := m.removeRelations(sourceID, bulk.GetRelationRemovals())
	if err != nil {
		return fmt.Errorf("Unable to remove relations: %v", err)
	}
	fmt.Printf("%d assets removed and %d relations removed in %fs\n",
		assetsCount,
		relCount,
		time.Since(nowRelationInsert).Seconds())
	return nil
}

// ReadGraph read source subgraph
func (m *MariaDB) ReadGraph(sourceName string, graph *knowledge.Graph) error {
	fmt.Printf("Start reading graph of data source with name %s\n", sourceName)

	now := time.Now()
	// Select all assets for which there is an observed relation from the source
	rows, err := m.db.QueryContext(context.Background(), `
SELECT a.type, a.value, b.type, b.value, r.type FROM relations r
INNER JOIN assets a ON a.id=r.from_id
INNER JOIN assets b ON b.id=r.to_id
INNER JOIN sources s ON s.id=r.source_id
WHERE s.name = ? AND r.type <> 'observed' AND (a.type <> 'source' AND b.type <> 'source')
	`, sourceName)

	if err != nil {
		return err
	}

	for rows.Next() {
		var FromType, ToType, FromKey, ToKey, Type string
		if err := rows.Scan(&FromType, &FromKey, &ToType, &ToKey, &Type); err != nil {
			return err
		}
		fromAsset := knowledge.Asset{
			Type: schema.AssetType(FromType),
			Key:  FromKey,
		}
		toAsset := knowledge.Asset{
			Type: schema.AssetType(ToType),
			Key:  ToKey,
		}
		from := graph.AddAsset(fromAsset.Type, fromAsset.Key)
		to := graph.AddAsset(toAsset.Type, toAsset.Key)
		graph.AddRelation(from, schema.RelationKeyType(Type), to)
	}

	elapsed := time.Since(now)
	fmt.Printf("Read graph of data source with name %s in %fs\n", sourceName, elapsed.Seconds())
	return nil
}

// FlushAll flush the database
func (m *MariaDB) FlushAll() error {
	_, err := m.db.ExecContext(context.Background(), "DROP TABLE relations")
	if err != nil {
		if !isUnknownTableError(err) {
			return err
		}
	}

	_, err = m.db.ExecContext(context.Background(), "DROP TABLE assets")
	if err != nil {
		if !isUnknownTableError(err) {
			return err
		}
	}

	_, err = m.db.ExecContext(context.Background(), "DROP TABLE graph_schema")
	if err != nil {
		if !isUnknownTableError(err) {
			return err
		}
	}

	_, err = m.db.ExecContext(context.Background(), "DROP TABLE query_history")
	if err != nil {
		if !isUnknownTableError(err) {
			return err
		}
	}
	return nil
}

// CountAssets count the total number of assets in db.
func (m *MariaDB) CountAssets() (int64, error) {
	var count int64
	row := m.db.QueryRowContext(context.Background(), "SELECT COUNT(DISTINCT value, type) FROM assets")

	err := row.Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// CountRelations count the total number of relations in db.
func (m *MariaDB) CountRelations() (int64, error) {
	var count int64
	row := m.db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM relations")

	err := row.Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// Close close the connection to maria
func (m *MariaDB) Close() error {
	return m.db.Close()
}

// Query the database with provided intermediate query representation
func (m *MariaDB) Query(ctx context.Context, sql knowledge.SQLTranslation) (*knowledge.GraphQueryResult, error) {
	deadline, ok := ctx.Deadline()
	// If there is a deadline, we make sure the query stops right after it has been reached.
	if ok {
		// Query can take 35 seconds max before being aborted...
		sql.Query = fmt.Sprintf("SET STATEMENT max_statement_time=%f FOR %s", time.Until(deadline).Seconds()+5, sql.Query)
	}
	fmt.Println(sql.Query)

	rows, err := m.db.QueryContext(ctx, sql.Query)
	if err != nil {
		return nil, err
	}

	res := new(knowledge.GraphQueryResult)
	res.Cursor = NewMariaDBCursor(rows, sql.ProjectionTypes)
	res.Projections = sql.ProjectionTypes
	return res, nil
}

// SaveSuccessfulQuery log an entry to mark a successful query
func (m *MariaDB) SaveSuccessfulQuery(ctx context.Context, cypher, sql string, duration time.Duration) error {
	_, err := m.db.ExecContext(ctx, "INSERT INTO query_history (id, timestamp, query_cypher, query_sql, status, execution_time_ms) VALUES (NULL, CURRENT_TIMESTAMP(), ?, ?, 'SUCCESS', ?)",
		cypher, sql, duration)
	if err != nil {
		return err
	}
	return err
}

// SaveFailedQuery log an entry to mark a failed query
func (m *MariaDB) SaveFailedQuery(ctx context.Context, cypher, sql string, err error) error {
	_, inErr := m.db.ExecContext(ctx, "INSERT INTO query_history (id, timestamp, query_cypher, query_sql, status, error) VALUES (NULL, CURRENT_TIMESTAMP(), ?, ?, 'FAILURE', ?)",
		cypher, sql, err.Error())
	if inErr != nil {
		return inErr
	}
	return nil
}

// SaveSchema save the schema graph in database
func (m *MariaDB) SaveSchema(ctx context.Context, sourceName string, schema schema.SchemaGraph) error {
	b, err := json.Marshal(schema)
	if err != nil {
		return fmt.Errorf("Unable to json encode schema: %v", err)
	}

	sourceID, err := m.resolveSourceID(sourceName)
	if err != nil {
		return fmt.Errorf("Unable to resolve source ID for source name %s: %v", sourceName, err)
	}

	_, err = m.db.ExecContext(ctx, "INSERT INTO graph_schema (source_id, graph, timestamp) VALUES (?, ?, CURRENT_TIMESTAMP())",
		sourceID, string(b))
	if err != nil {
		return fmt.Errorf("Unable to save schema in DB: %v", err)
	}

	return nil
}

// LoadSchema load the schema graph of the source from DB
func (m *MariaDB) LoadSchema(ctx context.Context, sourceName string) (schema.SchemaGraph, error) {
	row := m.db.QueryRowContext(ctx, `
SELECT gs.graph FROM graph_schema gs
INNER JOIN sources s ON s.id = gs.source_id
WHERE s.name = ?
ORDER BY gs.id DESC LIMIT 1`,
		sourceName)
	var rawJSON string
	if err := row.Scan(&rawJSON); err != nil {
		if err == sql.ErrNoRows {
			return schema.NewSchemaGraph(), nil
		}
		return schema.NewSchemaGraph(), err
	}

	graph := schema.NewSchemaGraph()
	err := json.Unmarshal([]byte(rawJSON), &graph)
	if err != nil {
		return schema.NewSchemaGraph(), err
	}

	return graph, nil
}

// ListSources list sources with their authentication tokens
func (m *MariaDB) ListSources(ctx context.Context) (map[string]string, error) {
	rows, err := m.db.QueryContext(ctx, "SELECT name, auth_token FROM sources")

	if err != nil {
		return nil, fmt.Errorf("Unable to read sources from database: %v", err)
	}
	defer rows.Close()

	sources := make(map[string]string)
	for rows.Next() {
		var sourceName string
		var authToken string
		if err := rows.Scan(&sourceName, &authToken); err != nil {
			return nil, err
		}
		sources[sourceName] = authToken
	}
	return sources, nil
}

// MariaDBCursor is a cursor of data retrieved by MariaDB
type MariaDBCursor struct {
	*sql.Rows

	Projections          []knowledge.Projection
	temporaryIDGenerator *AssetTemporaryIDGenerator
}

// NewMariaDBCursor create a new instance of MariaDBCursor
func NewMariaDBCursor(rows *sql.Rows, projections []knowledge.Projection) *MariaDBCursor {
	return &MariaDBCursor{
		Rows:                 rows,
		Projections:          projections,
		temporaryIDGenerator: NewAssetTemporaryIDGenerator(),
	}
}

// HasMore tells whether there are more data to retrieve from the cursor
func (mc *MariaDBCursor) HasMore() bool {
	return mc.Rows.Next()
}

// Read read one more item from the cursor
func (mc *MariaDBCursor) Read(ctx context.Context, doc interface{}) error {
	type AssetWithID struct {
		ID int
		knowledge.Asset
	}

	type RelationWithID struct {
		ID   int
		From int
		To   int
		Type schema.RelationKeyType
	}

	var err error
	var fArr []string

	if fArr, err = mc.Rows.Columns(); err != nil {
		return err
	}

	values := make([]interface{}, len(fArr))
	valuesPtr := make([]interface{}, len(fArr))
	for i := range values {
		valuesPtr[i] = &values[i]
	}

	if err := mc.Rows.Scan(valuesPtr...); err != nil {
		return err
	}

	val := reflect.ValueOf(doc)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return fmt.Errorf("Output parameter should be a pointer")
	}

	q := queue.New(int64(len(fArr)))
	for _, v := range values {
		var err error
		b, ok := v.([]byte)
		if ok {
			err = q.Put(string(b))
		} else {
			err = q.Put(b)
		}
		if err != nil {
			return err
		}
	}

	output := make([]interface{}, len(mc.Projections))

	// This first pass, creates all temporary ids bound to nodes
	for i, pt := range mc.Projections {
		switch pt.ExpressionType {
		case knowledge.NodeExprType:
			items, err := q.Get(3)
			if err != nil {
				return nil
			}

			asset := knowledge.Asset{
				Type: schema.AssetType(fmt.Sprintf("%v", reflect.ValueOf(items[2]))),
				Key:  fmt.Sprintf("%v", reflect.ValueOf(items[1])),
			}

			dbID, err := strconv.ParseInt(reflect.ValueOf(items[0]).String(), 10, 32)
			if err != nil {
				return fmt.Errorf("Unable to parse DBID: %v", err)
			}

			awi := AssetWithID{
				ID:    int(dbID),
				Asset: asset,
			}
			output[i] = awi
		case knowledge.EdgeExprType:
			items, err := q.Get(3)
			if err != nil {
				return nil
			}
			dbIDFromStr := reflect.ValueOf(items[0]).String()

			dbIDFrom, err := strconv.ParseInt(dbIDFromStr, 10, 32)
			if err != nil {
				return fmt.Errorf("Unable to parse DBID %s (from): %v", dbIDFromStr, err)
			}

			dbIDToStr := reflect.ValueOf(items[1]).String()
			dbIDTo, err := strconv.ParseInt(dbIDToStr, 10, 32)
			if err != nil {
				return fmt.Errorf("Unable to parse DBID %s (to): %v", dbIDToStr, err)
			}

			r := RelationWithID{
				From: int(dbIDFrom),
				To:   int(dbIDTo),
				Type: schema.RelationKeyType(fmt.Sprintf("%v", reflect.ValueOf(items[2]))),
			}
			output[i] = r
		case knowledge.PropertyExprType:
			items, err := q.Get(1)
			if err != nil {
				return nil
			}
			output[i] = items[0]
		}
	}

	// Generate temporary IDs for each node
	for i, o := range output {
		switch o.(type) {
		case AssetWithID:
			a := o.(AssetWithID)

			tmpID, err := mc.temporaryIDGenerator.Push(a.Asset, a.ID)
			if err != nil {
				return fmt.Errorf("Unable to retrieve the temporary ID for DB ID %d: %v", a.ID, err)
			}

			output[i] = knowledge.AssetWithID{
				Asset: a.Asset,
				ID:    fmt.Sprintf("%v", tmpID),
			}
		}
	}

	// Replace DB IDs by temporary IDs to merge pivot points having one DB ID per source
	for i, o := range output {
		switch o.(type) {
		case RelationWithID:
			r := o.(RelationWithID)

			tmpIDFrom, err := mc.temporaryIDGenerator.Get(r.From)
			if err != nil {
				return fmt.Errorf("Unable to retrieve the temporary ID for DB ID %d (from): %v", r.From, err)
			}

			tmpIDTo, err := mc.temporaryIDGenerator.Get(r.To)
			if err != nil {
				return fmt.Errorf("Unable to retrieve the temporary ID for DB ID %d (to): %v", r.To, err)
			}

			output[i] = knowledge.RelationWithID{
				From: fmt.Sprintf("%v", tmpIDFrom),
				To:   fmt.Sprintf("%v", tmpIDTo),
				Type: r.Type,
			}
		}
	}

	val.Elem().Set(reflect.ValueOf(output))
	return nil
}

// Close the cursor
func (mc *MariaDBCursor) Close() error {
	return mc.Rows.Close()
}
