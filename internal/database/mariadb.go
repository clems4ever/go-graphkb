package database

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cheggaaa/pb"
	"github.com/clems4ever/go-graphkb/internal/knowledge"
	"github.com/clems4ever/go-graphkb/internal/query"
	"github.com/clems4ever/go-graphkb/internal/schema"
	"github.com/clems4ever/go-graphkb/internal/utils"
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
	_, err := m.db.QueryContext(context.Background(), `
CREATE TABLE IF NOT EXISTS assets (id VARCHAR(64) NOT NULL, value VARCHAR(255), type VARCHAR(64) NOT NULL,
CONSTRAINT pk_asset PRIMARY KEY (id, type),
INDEX value_idx (value),
INDEX type_idx (type))
PARTITION BY KEY (type) PARTITIONS 20`)
	if err != nil {
		return err
	}

	// type must be part of the primary key to be a partition key
	_, err = m.db.QueryContext(context.Background(), `
CREATE TABLE IF NOT EXISTS relations (
	id VARCHAR(64) NOT NULL,
	from_id VARCHAR(64) NOT NULL,
	to_id VARCHAR(64) NOT NULL,
	type VARCHAR(64) NOT NULL,
	source VARCHAR(64) NOT NULL,
CONSTRAINT pk_relation PRIMARY KEY (id, type),
INDEX type_idx (type),
INDEX from_idx (from_id),
INDEX to_idx (to_id),
INDEX left_relation_idx (from_id, type),
INDEX right_relation_idx (to_id, type),
INDEX full_relation_idx (type, from_id, to_id))
PARTITION BY KEY (type) PARTITIONS 20`)
	if err != nil {
		return err
	}

	// Create the table storing the schema graphs
	_, err = m.db.QueryContext(context.Background(), `
CREATE TABLE IF NOT EXISTS graph_schema (
	id INTEGER AUTO_INCREMENT NOT NULL,
	source_name VARCHAR(64) NOT NULL,
	graph TEXT NOT NULL,
	timestamp TIMESTAMP,
CONSTRAINT pk_schema PRIMARY KEY (id))`)
	if err != nil {
		return err
	}

	return nil
}

// Hasher is sha256 hasher with a cache for performance optimization
type Hasher struct {
	cache map[interface{}]string
	mutex sync.Mutex
}

// Hash a json-encodable valus with sha256
func (h *Hasher) Hash(v interface{}) string {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	if h, ok := h.cache[v]; ok {
		return h
	}
	b, _ := json.Marshal(v)
	bs := sha256.Sum256(b)
	hash := base64.StdEncoding.EncodeToString(bs[:])
	h.cache[v] = hash
	return hash
}

func isDuplicateEntryError(err error) bool {
	driverErr, ok := err.(*mysql.MySQLError)
	return ok && driverErr.Number == 1062
}

func isUnknownTableError(err error) bool {
	driverErr, ok := err.(*mysql.MySQLError)
	return ok && driverErr.Number == 1051
}

func (m *MariaDB) upsertAssets(assets []knowledge.Asset, hasher *Hasher) (int64, error) {
	if len(assets) == 0 {
		return 0, nil
	}
	fmt.Printf("Write bulk of %d assets\n", len(assets))
	bar := pb.StartNew(len(assets))
	insertedCount := int64(0)

	assetChunks := utils.ChunkSlice(assets, 1000).([][]interface{})

	for _, assetChunk := range assetChunks {
		tx, err := m.db.Begin()
		if err != nil {
			log.Fatal(err)
		}
		q, err := tx.PrepareContext(context.Background(), "INSERT INTO assets (id, type, value) VALUES (?, ?, ?)")
		if err != nil {
			log.Fatal(fmt.Errorf("Unable to prepare asset insertion query: %v", err))
		}
		for _, aC := range assetChunk {
			a := aC.(knowledge.Asset)
			_, err = q.ExecContext(context.Background(), hasher.Hash(knowledge.AssetKey(a)), a.Type, a.Key)
			if err != nil {
				if isDuplicateEntryError(err) {
					bar.Increment()
					continue
				}
				log.Fatal(fmt.Errorf("Unable to insert asset %v: %v", a, err))
			}
			atomic.AddInt64(&insertedCount, 1)
			bar.Increment()
		}

		err = tx.Commit()
		if err != nil {
			log.Fatal(err)
		}
	}

	bar.Finish()
	fmt.Println("Write bulk of assets done")
	return insertedCount, nil
}

func (m *MariaDB) upsertRelations(source string, relations []knowledge.Relation, hasher *Hasher) (int64, error) {
	if len(relations) == 0 {
		return 0, nil
	}

	fmt.Printf("Write bulk of %d relations\n", len(relations))
	bar := pb.StartNew(len(relations))
	insertedCount := int64(0)

	relationChunks := utils.ChunkSlice(relations, 1000).([][]interface{})

	for _, relationChunk := range relationChunks {
		tx, err := m.db.Begin()
		if err != nil {
			log.Fatal(err)
		}

		q, err := tx.PrepareContext(context.Background(),
			"INSERT INTO relations (id, from_id, to_id, type, source) VALUES (?, ?, ?, ?, ?)")
		if err != nil {
			log.Fatal(fmt.Errorf("Unable to prepare relation insertion query: %v", err))
		}

		for _, rC := range relationChunk {
			r := rC.(knowledge.Relation)
			rel := SourceRelation{
				Relation: r,
				Source:   source,
			}
			_, err = q.ExecContext(context.Background(), hasher.Hash(rel), hasher.Hash(r.From), hasher.Hash(r.To), r.Type, source)
			if err != nil {
				if isDuplicateEntryError(err) {
					bar.Increment()
					continue
				}
				log.Fatal(fmt.Errorf("Unable to insert relation %v: %v", r, err))
			}
			bar.Increment()
			insertedCount++
		}

		if err := tx.Commit(); err != nil {
			log.Fatal(err)
		}
	}
	bar.Finish()

	fmt.Println("Write bulk of relations done")
	return insertedCount, nil
}

func (m *MariaDB) removeRelations(source string, relations []knowledge.Relation, hasher *Hasher) (int64, error) {
	if len(relations) == 0 {
		return 0, nil
	}

	fmt.Printf("Remove bulk of %d relations\n", len(relations))
	tx, err := m.db.Begin()
	if err != nil {
		return 0, err
	}
	bar := pb.StartNew(len(relations))
	removedCount := int64(0)

	stmt, err := tx.PrepareContext(context.Background(), "DELETE FROM relations WHERE id = ?")
	if err != nil {
		return 0, err
	}

	for _, r := range relations {
		rel := SourceRelation{
			Relation: r,
			Source:   source,
		}
		res, err := stmt.ExecContext(context.Background(), hasher.Hash(rel))
		if err != nil {
			return 0, fmt.Errorf("Unable to detete relation %v: %v", r, err)
		}
		bar.Increment()
		rCount, err := res.RowsAffected()
		if err != nil {
			return 0, err
		}
		removedCount += rCount
	}
	bar.Finish()
	err = tx.Commit()
	if err != nil {
		return 0, err
	}
	fmt.Println("Remove bulk of relations done")
	return removedCount, err
}

func (m *MariaDB) removeAssets(source string, assets []knowledge.Asset, hasher *Hasher) (int64, error) {
	if len(assets) == 0 {
		return 0, nil
	}

	fmt.Printf("Remove bulk of %d assets\n", len(assets))
	tx, err := m.db.Begin()
	if err != nil {
		return 0, err
	}

	stmt, err := tx.PrepareContext(context.Background(),
		"DELETE FROM assets WHERE id = ? AND (SELECT COUNT(*) FROM relations WHERE (from_id = ? OR to_id = ?) AND source = ?)=0")
	if err != nil {
		return 0, err
	}

	bar := pb.StartNew(len(assets))
	removedCount := int64(0)
	for _, a := range assets {
		h := hasher.Hash(knowledge.AssetKey(a))
		_, err = stmt.ExecContext(context.Background(), h, h, h, source)
		if err != nil {
			return 0, fmt.Errorf("Unable to delete asset %v: %v", a, err)
		}
		bar.Increment()
		removedCount++
	}
	bar.Finish()
	err = tx.Commit()
	if err != nil {
		return 0, err
	}
	fmt.Println("Remove bulk of assets done")
	return removedCount, err
}

// UpdateGraph update graph with bulk of operations
func (m *MariaDB) UpdateGraph(source string, bulk *knowledge.GraphUpdatesBulk) error {
	hasher := Hasher{cache: make(map[interface{}]string)}
	count, err := m.upsertAssets(bulk.AssetUpserts, &hasher)
	if err != nil {
		return err
	}
	fmt.Printf("%d assets inserted\n", count)
	count, err = m.upsertRelations(source, bulk.RelationUpserts, &hasher)
	if err != nil {
		return err
	}
	fmt.Printf("%d relations inserted\n", count)

	count, err = m.removeRelations(source, bulk.RelationRemovals, &hasher)
	if err != nil {
		return err
	}
	fmt.Printf("%d relations removed\n", count)

	count, err = m.removeAssets(source, bulk.AssetRemovals, &hasher)
	if err != nil {
		return err
	}
	fmt.Printf("%d assets removed\n", count)
	return nil
}

// ReadGraph read source subgraph
func (m *MariaDB) ReadGraph(source string, graph *knowledge.Graph) error {
	fmt.Printf("Start reading graph of source %s\n", source)

	now := time.Now()
	// Select all assets for which there is an observed relation from the source
	rows, err := m.db.QueryContext(context.Background(),
		`SELECT from_assets.type AS from_type, from_assets.value AS from_value, to_assets.type AS to_type, to_assets.value AS to_value, relations.type FROM relations
JOIN assets from_assets ON from_assets.id=relations.from_id
JOIN assets to_assets ON to_assets.id=relations.to_id
WHERE relations.source = ?
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
	_, err := m.db.ExecContext(context.Background(), "DROP TABLE assets")
	if err != nil {
		if !isUnknownTableError(err) {
			return err
		}
	}

	_, err = m.db.ExecContext(context.Background(), "DROP TABLE relations")
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
