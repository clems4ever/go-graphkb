package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"sync/atomic"
	"time"

	"github.com/cheggaaa/pb"
	"github.com/clems4ever/go-graphkb/internal/knowledge"
	"github.com/clems4ever/go-graphkb/internal/query"
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
func NewMariaDB(username string, password string, host string, databaseName string) *MariaDB {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@(%s)/%s", username, password, host, databaseName))
	if err != nil {
		log.Fatal(err)
	}
	db.SetMaxIdleConns(0)
	return &MariaDB{db: db}
}

// InitializeSchema initialize the schema of the database
func (m *MariaDB) InitializeSchema() error {
	// type must be part of the primary key to be a partition key
	q, err := m.db.QueryContext(context.Background(), `
CREATE TABLE IF NOT EXISTS assets (
	id INT NOT NULL AUTO_INCREMENT,
	value VARCHAR(255) NOT NULL,
	type VARCHAR(64) NOT NULL,

	CONSTRAINT pk_asset PRIMARY KEY (id),
	UNIQUE unique_asset_idx (type, value),
	INDEX value_idx (value),
	INDEX type_idx (type))`)
	if err != nil {
		return err
	}
	defer q.Close()

	// type must be part of the primary key to be a partition key
	q, err = m.db.QueryContext(context.Background(), `
CREATE TABLE IF NOT EXISTS relations (
	id INT NOT NULL AUTO_INCREMENT,
	from_id INT NOT NULL,
	to_id INT NOT NULL,
	type VARCHAR(64) NOT NULL,
	source VARCHAR(64) NOT NULL,

	CONSTRAINT pk_relation PRIMARY KEY (id),
	CONSTRAINT fk_from FOREIGN KEY (from_id) REFERENCES assets (id),
	CONSTRAINT fk_to FOREIGN KEY (to_id) REFERENCES assets (id),
	
	INDEX full_relation_type_from_to_idx (type, from_id, to_id),
    INDEX full_relation_type_to_from_idx (type, to_id, from_id),
    INDEX full_relation_from_type_to_idx (from_id, type, to_id),
    INDEX full_relation_from_to_type_idx (from_id, to_id, type),
    INDEX full_relation_to_from_type_idx (to_id, from_id, type),
    INDEX full_relation_to_type_from_idx (to_id, type, from_id))`)
	if err != nil {
		return err
	}
	defer q.Close()

	// Create the table storing the schema graphs
	q, err = m.db.QueryContext(context.Background(), `
CREATE TABLE IF NOT EXISTS graph_schema (
	id INTEGER AUTO_INCREMENT NOT NULL,
	source_name VARCHAR(64) NOT NULL,
	graph TEXT NOT NULL,
	timestamp TIMESTAMP,
CONSTRAINT pk_schema PRIMARY KEY (id))`)
	if err != nil {
		return err
	}
	defer q.Close()
	return nil
}

// AssetIDResolver store ID assets in a cache
type AssetRegistry struct {
	cache map[knowledge.AssetKey]int64
}

func (ar *AssetRegistry) Set(a knowledge.AssetKey, idx int64) {
	ar.cache[a] = idx
}

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

func (m *MariaDB) resolveAssets(assets []knowledge.AssetKey, registry *AssetRegistry) error {
	bar := pb.StartNew(len(assets))
	defer bar.Finish()

	tx, err := m.db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.PrepareContext(context.Background(), "SELECT id FROM assets WHERE type = ? AND value = ?")
	if err != nil {
		return err
	}

	for _, a := range assets {
		q, err := stmt.QueryContext(context.Background(), a.Type, a.Key)
		if err != nil {
			return err
		}
		defer q.Close()

		for q.Next() {
			var idx int64
			if err := q.Scan(&idx); err != nil {
				return err
			}
			registry.Set(knowledge.AssetKey(a), idx)
		}

		bar.Increment()
	}
	return tx.Commit()
}

func (m *MariaDB) upsertAssets(assets []knowledge.Asset, registry *AssetRegistry) (int64, error) {
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
		tx, err := m.db.Begin()
		if err != nil {
			log.Fatal(err)
		}

		insertQuery, err := tx.PrepareContext(context.Background(), `
INSERT INTO assets (type, value) VALUES (?, ?)`)
		if err != nil {
			log.Fatal(fmt.Errorf("Unable to prepare asset insertion query: %v", err))
		}

		for _, aC := range assetChunk {
			a := aC.(knowledge.Asset)

			res, err := insertQuery.ExecContext(context.Background(), a.Type, a.Key)
			if err != nil {
				return 0, fmt.Errorf("Unable to insert asset %v: %v", a, err)
			}
			idx, err := res.LastInsertId()
			if err != nil {
				return 0, err
			}
			registry.Set(knowledge.AssetKey(a), idx)

			atomic.AddInt64(&insertedCount, 1)
			bar.Increment()
		}

		err = tx.Commit()
		if err != nil {
			log.Fatal(err)
		}
	}
	return insertedCount, nil
}

func (m *MariaDB) upsertRelations(source string, relations []knowledge.Relation, registry *AssetRegistry) (int64, error) {
	if len(relations) == 0 {
		return 0, nil
	}
	bar := pb.StartNew(len(relations))
	defer bar.Finish()
	insertedCount := int64(0)

	relationChunks := utils.ChunkSlice(relations, 10000).([][]interface{})

	for _, relationChunk := range relationChunks {
		tx, err := m.db.Begin()
		if err != nil {
			log.Fatal(err)
		}

		q, err := tx.PrepareContext(context.Background(),
			"INSERT INTO relations (from_id, to_id, type, source) VALUES (?, ?, ?, ?)")
		if err != nil {
			log.Fatal(fmt.Errorf("Unable to prepare relation insertion query: %v", err))
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

			_, err = q.ExecContext(context.Background(), idxFrom, idxTo, r.Type, source)
			if err != nil {
				if isDuplicateEntryError(err) {
					bar.Increment()
					continue
				}
				log.Fatal(fmt.Errorf("Unable to insert relation %v (%d -> %d): %v", r, idxFrom, idxTo, err))
			}
			bar.Increment()
			insertedCount++
		}

		if err := tx.Commit(); err != nil {
			log.Fatal(err)
		}
	}
	bar.Finish()
	return insertedCount, nil
}

func (m *MariaDB) removeRelations(source string, relations []knowledge.Relation) (int64, int64, error) {
	if len(relations) == 0 {
		return 0, 0, nil
	}

	tx, err := m.db.Begin()
	if err != nil {
		return 0, 0, err
	}
	bar := pb.StartNew(len(relations))
	removedCount := int64(0)

	stmt, err := tx.PrepareContext(context.Background(), `
DELETE r FROM relations r
INNER JOIN assets a ON r.from_id = a.id
INNER JOIN assets b ON r.to_id = b.id
WHERE a.type = ? AND a.value = ? AND b.type = ? AND b.value = ? AND r.type = ?`)
	if err != nil {
		return 0, 0, err
	}
	defer stmt.Close()

	for _, r := range relations {
		rel := SourceRelation{
			Relation: r,
			Source:   source,
		}
		res, err := stmt.ExecContext(context.Background(),
			rel.From.Type, rel.From.Key, rel.To.Type, rel.To.Key, rel.Type)
		if err != nil {
			return 0, 0, fmt.Errorf("Unable to detete relation %v: %v", r, err)
		}
		bar.Increment()
		rCount, err := res.RowsAffected()
		if err != nil {
			return 0, 0, err
		}
		removedCount += rCount
	}
	bar.Finish()

	res, err := tx.ExecContext(context.Background(), `
DELETE a FROM assets a
WHERE id NOT IN (select from_id from relations)
AND id NOT IN (select to_id from relations)`)
	if err != nil {
		return 0, 0, err
	}

	removedAssetsCount, err := res.RowsAffected()
	if err != nil {
		return 0, 0, err
	}

	err = tx.Commit()
	if err != nil {
		return 0, 0, err
	}
	return removedCount, removedAssetsCount, err
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

	fmt.Println("Start resolving assets")
	err := m.resolveAssets(assetKeys, &registry)
	if err != nil {
		return err
	}

	fmt.Println("Start upserting assets")
	count, err := m.upsertAssets(bulk.GetAssetUpserts(), &registry)
	if err != nil {
		return err
	}

	nowAssetInsert := time.Now()
	fmt.Printf("%d assets inserted in %fs\n", count, nowAssetInsert.Sub(now).Seconds())

	fmt.Println("Start upserting relations")
	count, err = m.upsertRelations(source, bulk.GetRelationUpserts(), &registry)
	if err != nil {
		return err
	}
	nowRelationInsert := time.Now()
	fmt.Printf("%d relations inserted in %fs\n", count, nowRelationInsert.Sub(nowAssetInsert).Seconds())

	relCount, assetsCount, err := m.removeRelations(source, bulk.GetRelationRemovals())
	if err != nil {
		return err
	}
	fmt.Printf("%d assets removed and %d relations removed in %fs\n",
		assetsCount,
		relCount,
		time.Since(nowRelationInsert).Seconds())
	return nil
}

// ReadGraph read source subgraph
func (m *MariaDB) ReadGraph(source string, graph *knowledge.Graph) error {
	fmt.Printf("Start reading graph of source %s\n", source)

	now := time.Now()
	// Select all assets for which there is an observed relation from the source
	rows, err := m.db.QueryContext(context.Background(), `
SELECT a.type, a.value, b.type, b.value, r.type FROM relations r
INNER JOIN assets a ON a.id=r.from_id
INNER JOIN assets b ON b.id=r.to_id
WHERE r.source = ? AND r.type <> 'observed' AND (a.type <> 'source' AND b.type <> 'source')
	`, source)

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
	fmt.Printf("Read graph of source %s in %fs\n", source, elapsed.Seconds())
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
	return nil
}

// CountAssets count the total number of assets in db.
func (m *MariaDB) CountAssets() (int64, error) {
	var count int64
	row := m.db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM assets")

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
func (m *MariaDB) Query(ctx context.Context, query *query.QueryIL) (*knowledge.GraphQueryResult, error) {
	sql, err := knowledge.NewSQLQueryTranslator().Translate(query)
	if err != nil {
		return nil, err
	}

	fmt.Println(sql.Query)

	rows, err := m.db.QueryContext(ctx, sql.Query)
	if err != nil {
		return nil, err
	}

	res := new(knowledge.GraphQueryResult)
	res.Cursor = &MariaDBCursor{
		Rows:        rows,
		Projections: sql.ProjectionTypes,
	}
	res.Projections = sql.ProjectionTypes

	return res, nil
}

func (m *MariaDB) SaveSchema(ctx context.Context, sourceName string, schema schema.SchemaGraph) error {
	b, err := json.Marshal(schema)
	if err != nil {
		return fmt.Errorf("Unable to json encode schema: %v", err)
	}

	_, err = m.db.ExecContext(ctx, "INSERT INTO graph_schema (source_name, graph, timestamp) VALUES (?, ?, CURRENT_TIMESTAMP())",
		sourceName, string(b))
	if err != nil {
		return fmt.Errorf("Unable to save schema in DB: %v", err)
	}

	return nil
}

func (m *MariaDB) LoadSchema(ctx context.Context, sourceName string) (schema.SchemaGraph, error) {
	row := m.db.QueryRowContext(ctx, "SELECT graph FROM graph_schema WHERE source_name = ? ORDER BY id DESC LIMIT 1", sourceName)
	var rawJson string
	if err := row.Scan(&rawJson); err != nil {
		if err == sql.ErrNoRows {
			return schema.NewSchemaGraph(), nil
		} else {
			return schema.NewSchemaGraph(), err
		}
	}

	graph := schema.NewSchemaGraph()
	err := json.Unmarshal([]byte(rawJson), &graph)
	if err != nil {
		return schema.NewSchemaGraph(), err
	}

	return graph, nil
}

func (m *MariaDB) ListSources(ctx context.Context) ([]string, error) {
	rows, err := m.db.QueryContext(ctx, "SELECT DISTINCT source_name FROM graph_schema")

	if err != nil {
		return nil, fmt.Errorf("Unable to read sources from database: %v", err)
	}
	defer rows.Close()

	sources := make([]string, 0)
	for rows.Next() {
		var source string
		if err := rows.Scan(&source); err != nil {
			return nil, err
		}
		sources = append(sources, source)
	}
	return sources, nil
}

type MariaDBCursor struct {
	*sql.Rows

	Projections []knowledge.Projection
}

func (mc *MariaDBCursor) HasMore() bool {
	return mc.Rows.Next()
}

func (mc *MariaDBCursor) Read(ctx context.Context, doc interface{}) error {
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

	for i, pt := range mc.Projections {
		switch pt.ExpressionType {
		case knowledge.NodeExprType:
			items, err := q.Get(3)
			if err != nil {
				return nil
			}
			a := knowledge.AssetWithID{
				ID: fmt.Sprintf("%v", reflect.ValueOf(items[0])),
				Asset: knowledge.Asset{
					Type: schema.AssetType(fmt.Sprintf("%v", reflect.ValueOf(items[2]))),
					Key:  fmt.Sprintf("%v", reflect.ValueOf(items[1])),
				},
			}
			output[i] = a
		case knowledge.EdgeExprType:
			items, err := q.Get(5)
			if err != nil {
				return nil
			}
			r := knowledge.RelationWithID{
				ID:   fmt.Sprintf("%v", reflect.ValueOf(items[0])),
				From: fmt.Sprintf("%v", reflect.ValueOf(items[1])),
				To:   fmt.Sprintf("%v", reflect.ValueOf(items[2])),
				Type: schema.RelationKeyType(fmt.Sprintf("%v", reflect.ValueOf(items[3]))),
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
	val.Elem().Set(reflect.ValueOf(output))
	return nil
}

func (mc *MariaDBCursor) Close() error {
	return mc.Rows.Close()
}
