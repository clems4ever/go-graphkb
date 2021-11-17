package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/VividCortex/mysqlerr"
	"github.com/clems4ever/go-graphkb/internal/knowledge"
	"github.com/clems4ever/go-graphkb/internal/schema"
	"github.com/clems4ever/go-graphkb/internal/utils"
	mysql "github.com/go-sql-driver/mysql"
	"github.com/golang-collections/go-datastructures/queue"
	"github.com/sirupsen/logrus"
)

var zeroBytes = []byte{0}

// MariaDB mariadb as graph storage backend
type MariaDB struct {
	db *sql.DB

	sourcesCache map[string]int
}

// NewMariaDB create an instance of mariadb
func NewMariaDB(username string, password string, host string, databaseName string, allowCleartextPassword bool) *MariaDB {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@(%s)/%s?allowCleartextPasswords=%s", username, password,
		host, databaseName, strconv.FormatBool(allowCleartextPassword)))
	if err != nil {
		logrus.Fatal(err)
	}
	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(10)
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

			UNIQUE unique_source (name, auth_token)
		)`)
	if err != nil {
		return fmt.Errorf("Unable to create sources table: %v", err)
	}

	// type must be part of the primary key to be a partition key
	_, err = m.db.ExecContext(context.Background(), `
		CREATE TABLE IF NOT EXISTS assets (
			id BIGINT UNSIGNED NOT NULL,
			value VARCHAR(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL,
			type VARCHAR(255) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,

			CONSTRAINT pk_asset PRIMARY KEY (id),

			INDEX value_idx (value),
			INDEX type_idx (type))`)
	if err != nil {
		return fmt.Errorf("Unable to create assets table: %v", err)
	}

	_, err = m.db.ExecContext(context.Background(), `
		CREATE TABLE IF NOT EXISTS relations (
			id BIGINT UNSIGNED NOT NULL,
			from_id BIGINT UNSIGNED NOT NULL,
			to_id BIGINT UNSIGNED NOT NULL,
			type VARCHAR(255) CHARACTER SET ascii COLLATE ascii_bin NOT NULL,

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
		return fmt.Errorf("Unable to create relations table: %v", err)
	}

	_, err = m.db.ExecContext(context.Background(), `
		CREATE TABLE IF NOT EXISTS relations_by_source (
			source_id INT NOT NULL,
			relation_id BIGINT UNSIGNED NOT NULL,
			update_time TIMESTAMP,

			CONSTRAINT pk_relation_by_source PRIMARY KEY (source_id, relation_id),
			CONSTRAINT fk_relations_by_source_source_id FOREIGN KEY (source_id) REFERENCES sources (id) ON DELETE CASCADE,
			CONSTRAINT fk_relations_by_source_relation_id FOREIGN KEY (relation_id) REFERENCES relations (id) ON DELETE CASCADE,

			INDEX source_idx (source_id))`)
	if err != nil {
		return fmt.Errorf("Unable to create relations_by_source table: %v", err)
	}

	_, err = m.db.ExecContext(context.Background(), `
		CREATE TABLE IF NOT EXISTS assets_by_source (
			source_id INT NOT NULL,
			asset_id BIGINT UNSIGNED NOT NULL,
			update_time TIMESTAMP,

			CONSTRAINT pk_assets_by_source PRIMARY KEY (source_id, asset_id),
			CONSTRAINT fk_asset_by_source_source_id FOREIGN KEY (source_id) REFERENCES sources (id) ON DELETE CASCADE,
			CONSTRAINT fk_asset_by_source_asset_id FOREIGN KEY (asset_id) REFERENCES assets (id) ON DELETE CASCADE,

			INDEX source_idx (source_id))`)
	if err != nil {
		return fmt.Errorf("Unable to create assets_by_source tables: %v", err)
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
		return fmt.Errorf("Unable to create graph_schema tables: %v", err)
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
		return fmt.Errorf("Unable to create query_history tables: %v", err)
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

// resolveSourceIDFromDB resolve the source ID from the source name from the database
func (m *MariaDB) resolveSourceIDFromDB(ctx context.Context, sourceName string) (int, error) {
	r, err := m.db.QueryContext(ctx, "SELECT id FROM sources WHERE name = ? LIMIT 1", sourceName)
	if err != nil {
		return 0, err
	}
	defer r.Close()

	var sourceID int
	for r.Next() {
		err = r.Scan(&sourceID)
		if err != nil {
			return 0, err
		}
	}
	return sourceID, nil
}

// resolveSourceID resolve the source ID from the source name from the cache first and then from the DB
func (m *MariaDB) resolveSourceID(ctx context.Context, sourceName string) (int, error) {
	// TODO(c.michaud): invalidate cache after some time.
	if v, ok := m.sourcesCache[sourceName]; ok {
		return v, nil
	}

	return m.resolveSourceIDFromDB(ctx, sourceName)
}

func writeAsset(w io.Writer, asset knowledge.Asset) error {
	_, err := w.Write([]byte(asset.Type))
	if err != nil {
		return err
	}
	_, err = w.Write(zeroBytes)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(asset.Key))
	if err != nil {
		return err
	}
	return nil
}

func hashAsset(asset knowledge.Asset) uint64 {
	h := fnv.New64()
	writeAsset(h, asset)
	return h.Sum64()
}

func hashRelation(relation knowledge.Relation) uint64 {
	h := fnv.New64()

	rel := []byte(relation.Type)

	writeAsset(h, knowledge.Asset(relation.From))

	h.Write(zeroBytes)
	h.Write(rel)
	h.Write(zeroBytes)

	writeAsset(h, knowledge.Asset(relation.To))

	return h.Sum64()
}

// InTransaction make sure a function is properly using the transaction
func InTransaction(db *sql.DB, txFunc func(*sql.Tx) error) (err error) {
	tx, err := db.Begin()
	if err != nil {
		return
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p) // re-throw panic after Rollback
		} else if err != nil {
			tx.Rollback() // err is non-nil; don't change it
		} else {
			err = tx.Commit() // err is nil; if Commit returns error update err
		}
	}()
	err = txFunc(tx)
	return err
}

// InsertAssets insert multiple assets into the graph of the given source
func (m *MariaDB) InsertAssets(ctx context.Context, source string, assets []knowledge.Asset) error {
	sourceID, err := m.resolveSourceID(ctx, source)
	if err != nil {
		return fmt.Errorf("Unable to resolve source ID of source %s for inserting assets: %v", source, err)
	}

	return InTransaction(m.db, func(tx *sql.Tx) error {
		for _, asset := range assets {
			h := hashAsset(asset)

			_, err = tx.ExecContext(ctx,
				`INSERT INTO assets (id, type, value) VALUES (?, ?, ?)`,
				h, asset.Type, asset.Key)
			if err != nil {
				if driverErr, ok := err.(*mysql.MySQLError); ok && driverErr.Number == mysqlerr.ER_DUP_ENTRY {
					// If the entry is duplicated, it's fine but we still need insert a line into assets_by_source.
				} else {
					return fmt.Errorf("Unable to insert asset %v (%d) in DB from source %s: %v", asset, h, source, err)
				}
			}

			_, err = tx.ExecContext(ctx,
				`INSERT INTO assets_by_source (source_id, asset_id) VALUES (?, ?)`, sourceID, h)
			if err != nil {
				if driverErr, ok := err.(*mysql.MySQLError); ok && driverErr.Number == mysqlerr.ER_DUP_ENTRY {
					// TODO(c.michaud): update the update_time?
				} else {
					return fmt.Errorf("Unable to insert binding between asset %s (%d) and source %s: %v", asset, h, source, err)
				}
			}
		}
		return nil
	})
}

// InsertRelations upsert one relation into the graph of the given source
func (m *MariaDB) InsertRelations(ctx context.Context, source string, relations []knowledge.Relation) error {
	sourceID, err := m.resolveSourceID(ctx, source)
	if err != nil {
		return fmt.Errorf("Unable to resolve source ID of source %s for inserting relations: %v", source, err)
	}

	return InTransaction(m.db, func(tx *sql.Tx) error {
		for _, relation := range relations {
			// TODO(c.michaud): make the source compute the hash directly to reduce the size of the payload.
			aFrom := hashAsset(knowledge.Asset(relation.From))
			aTo := hashAsset(knowledge.Asset(relation.To))
			rH := hashRelation(relation)

			_, err = tx.ExecContext(ctx,
				"INSERT INTO relations (id, from_id, to_id, type) VALUES (?, ?, ?, ?)",
				rH, aFrom, aTo, relation.Type)
			if err != nil {
				if driverErr, ok := err.(*mysql.MySQLError); ok && driverErr.Number == mysqlerr.ER_DUP_ENTRY {
					// If the entry is duplicated, it's fine but we still need insert a line into relations_by_source.
				} else {
					return fmt.Errorf("Unable insert relation %v (%d) in DB from source %s: %v", relation, rH, source, err)
				}
			}

			_, err = tx.ExecContext(ctx,
				`INSERT INTO relations_by_source (source_id, relation_id) VALUES (?, ?)`, sourceID, rH)
			if err != nil {
				if driverErr, ok := err.(*mysql.MySQLError); ok && driverErr.Number == mysqlerr.ER_DUP_ENTRY {
					// TODO(c.michaud): update the update_time?
				} else {
					return fmt.Errorf("Unable to insert binding between relation %v (%d) and source %s: %v", relation, rH, source, err)
				}
			}
		}
		return nil
	})
}

// RemoveAssets remove one asset from the graph of the given source
func (m *MariaDB) RemoveAssets(ctx context.Context, source string, assets []knowledge.Asset) error {
	sourceID, err := m.resolveSourceID(ctx, source)
	if err != nil {
		return fmt.Errorf("Unable to resolve source ID of source %s for removing assets: %v", source, err)
	}

	return InTransaction(m.db, func(tx *sql.Tx) error {
		for _, asset := range assets {
			h := hashAsset(asset)

			_, err = tx.ExecContext(ctx,
				`DELETE FROM assets_by_source WHERE asset_id = ? AND source_id = ?`,
				h, sourceID)
			if err != nil {
				return fmt.Errorf("Unable to remove binding between asset %v (%d) and source %s: %v", asset, h, source, err)
			}

			_, err = tx.ExecContext(ctx,
				`DELETE FROM assets WHERE id = ? AND NOT EXISTS (
			SELECT * FROM assets_by_source WHERE asset_id = ?
		)`,
				h, h)
			if err != nil {
				return fmt.Errorf("Unable to remove asset %v (%d) from source %s: %v", asset, h, source, err)
			}

		}
		return nil
	})
}

// RemoveRelations remove relations from the graph of the given source
func (m *MariaDB) RemoveRelations(ctx context.Context, source string, relations []knowledge.Relation) error {
	sourceID, err := m.resolveSourceID(ctx, source)
	if err != nil {
		return fmt.Errorf("Unable to resolve source ID of source %s for removing relations: %v", source, err)
	}
	return InTransaction(m.db, func(tx *sql.Tx) error {
		for _, relation := range relations {
			rH := hashRelation(relation)

			_, err = tx.ExecContext(ctx,
				`DELETE FROM relations_by_source WHERE relation_id = ? AND source_id = ?`,
				rH, sourceID)
			if err != nil {
				return fmt.Errorf("Unable to remove binding between relation %v (%d) and source %s: %v", relation, rH, source, err)
			}

			_, err = tx.ExecContext(ctx,
				`DELETE FROM relations WHERE id = ? AND NOT EXISTS (
			SELECT * FROM relations_by_source WHERE relation_id = ?
		)`, rH, rH)
			if err != nil {
				return fmt.Errorf("Unable to remove relation %v (%d) from source %s: %v", relation, rH, source, err)
			}
		}
		return nil
	})
}

// ReadGraph read source subgraph
func (m *MariaDB) ReadGraph(ctx context.Context, sourceName string, encoder *knowledge.GraphEncoder) error {
	logrus.Debugf("Start reading graph of data source with name %s", sourceName)
	sourceID, err := m.resolveSourceID(ctx, sourceName)
	if err != nil {
		return fmt.Errorf("Unable to resolve source ID from name %s: %v", sourceName, err)
	}

	now := time.Now()

	err = InTransaction(m.db, func(tx *sql.Tx) error {
		{
			// Select all relations produced by this source
			rows, err := tx.QueryContext(ctx, `
	SELECT a.type, a.value, b.type, b.value, r.type FROM relations_by_source rbs
	INNER JOIN relations r ON rbs.relation_id = r.id
	INNER JOIN assets a ON a.id=r.from_id
	INNER JOIN assets b ON b.id=r.to_id
	WHERE rbs.source_id = ?
		`, sourceID)

			if err != nil {
				return fmt.Errorf("Unable to retrieve relations: %v", err)
			}

			defer rows.Close()

			for rows.Next() {
				var FromType, ToType, FromKey, ToKey, Type string
				if err := rows.Scan(&FromType, &FromKey, &ToType, &ToKey, &Type); err != nil {
					return err
				}
				fromAsset := knowledge.AssetKey{
					Type: schema.AssetType(FromType),
					Key:  FromKey,
				}
				toAsset := knowledge.AssetKey{
					Type: schema.AssetType(ToType),
					Key:  ToKey,
				}

				relation := knowledge.Relation{
					Type: schema.RelationKeyType(Type),
					From: fromAsset,
					To:   toAsset,
				}

				err = encoder.EncodeRelation(relation)
				if err != nil {
					return fmt.Errorf("Unable to write relation %v: %v", relation, err)
				}
			}
		}

		{
			// Select all assets produced by this source. This is useful in case there are some standalone nodes in the graph of the source.
			// TODO(c.michaud): optimization could be done by only selecting assets without any relation since the others have already have been retrieved in the previous query.
			rows, err := tx.QueryContext(ctx, `
	SELECT a.type, a.value FROM assets_by_source abs
	INNER JOIN assets a ON a.id=abs.asset_id
	WHERE abs.source_id = ?
		`, sourceID)

			if err != nil {
				return fmt.Errorf("Unable to retrieve assets: %v", err)
			}

			defer rows.Close()

			for rows.Next() {
				var Key, Type string
				if err := rows.Scan(&Type, &Key); err != nil {
					return fmt.Errorf("Unable to read standalone asset: %v", err)
				}

				asset := knowledge.Asset{
					Type: schema.AssetType(Type),
					Key:  Key,
				}

				err := encoder.EncodeAsset(asset)
				if err != nil {
					return fmt.Errorf("Unable to write asset %v: %v", asset, err)
				}
			}
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("Unable to handle transaction: %v", err)
	}

	elapsed := time.Since(now)
	logrus.Debugf("Read graph of data source with name %s in %fs", sourceName, elapsed.Seconds())
	return nil
}

// FlushAll flush the database
func (m *MariaDB) FlushAll(ctx context.Context) error {
	return InTransaction(m.db, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, "DROP TABLE relations_by_source")
		if err != nil {
			if !isUnknownTableError(err) {
				return err
			}
		}

		_, err = tx.ExecContext(ctx, "DROP TABLE assets_by_source")
		if err != nil {
			if !isUnknownTableError(err) {
				return err
			}
		}

		_, err = tx.ExecContext(ctx, "DROP TABLE relations")
		if err != nil {
			if !isUnknownTableError(err) {
				return err
			}
		}

		_, err = tx.ExecContext(ctx, "DROP TABLE assets")
		if err != nil {
			if !isUnknownTableError(err) {
				return err
			}
		}

		_, err = tx.ExecContext(ctx, "DROP TABLE graph_schema")
		if err != nil {
			if !isUnknownTableError(err) {
				return err
			}
		}

		_, err = tx.ExecContext(ctx, "DROP TABLE query_history")
		if err != nil {
			if !isUnknownTableError(err) {
				return err
			}
		}
		return nil
	})
}

// CountAssets count the total number of assets in db.
func (m *MariaDB) CountAssets(ctx context.Context) (int64, error) {
	var count int64
	row := m.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM assets")
	return count, row.Scan(&count)
}

// CountAssetsBySource count the total number of assets in db by source
func (m *MariaDB) CountAssetsBySource(ctx context.Context, sourceName string) (int64, error) {
	sourceID, err := m.resolveSourceID(ctx, sourceName)
	if err != nil {
		return 0, fmt.Errorf("Unable to resolve source ID from name %s: %w", sourceName, err)
	}

	var count int64
	row := m.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM assets_by_source WHERE source_id = ?", sourceID)

	return count, row.Scan(&count)
}

// CountRelations count the total number of relations in db.
func (m *MariaDB) CountRelations(ctx context.Context) (int64, error) {
	var count int64
	row := m.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM relations")
	return count, row.Scan(&count)
}

// CountRelationsBySource count the total number of relations in db by source.
func (m *MariaDB) CountRelationsBySource(ctx context.Context, sourceName string) (int64, error) {
	sourceID, err := m.resolveSourceID(ctx, sourceName)
	if err != nil {
		return 0, fmt.Errorf("Unable to resolve source ID from name %s: %w", sourceName, err)
	}

	var count int64
	row := m.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM relations_by_source WHERE source_id = ?", sourceID)
	return count, row.Scan(&count)
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
	logrus.Debug("Query to be executed: ", sql.Query)

	rows, err := m.db.QueryContext(ctx, sql.Query)
	if err != nil {
		return nil, err
	}

	res := new(knowledge.GraphQueryResult)
	res.Cursor = NewMariaDBCursor(rows, sql.ProjectionTypes)
	res.Projections = sql.ProjectionTypes
	return res, nil
}

func (m *MariaDB) GetAssetSources(ctx context.Context, ids []string) (map[string][]string, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	args := make([]interface{}, len(ids))
	for i, id := range ids {
		args[i] = id
	}

	idsSet := make(map[string][]string)
	argsSlices := utils.ChunkSlice(args, 500).([][]interface{})

	for _, argsSlice := range argsSlices {
		stmt, err := m.db.PrepareContext(ctx, `
SELECT asset_id, sources.name FROM sources
INNER JOIN assets_by_source ON sources.id = assets_by_source.source_id
WHERE asset_id IN (?`+strings.Repeat(",?", len(argsSlice)-1)+`)`)
		if err != nil {
			return nil, fmt.Errorf("Unable to prepare statement for retrieving asset sources: %w", err)
		}
		row, err := stmt.QueryContext(ctx, argsSlice...)
		if err != nil {
			return nil, fmt.Errorf("Unable to retrieve sources for assets: %w", err)
		}

		var source string
		var assetId uint64

		for row.Next() {
			err = row.Scan(&assetId, &source)
			if err != nil {
				return nil, fmt.Errorf("Unable to scan row of asset source: %w", err)
			}
			assetIdStr := fmt.Sprintf("%d", assetId)
			if _, ok := idsSet[assetIdStr]; !ok {
				idsSet[assetIdStr] = []string{}
			}
			idsSet[assetIdStr] = append(idsSet[assetIdStr], source)
		}
	}
	return idsSet, nil
}

func (m *MariaDB) GetRelationSources(ctx context.Context, ids []string) (map[string][]string, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		args[i] = id
	}
	idsSet := make(map[string][]string)

	argsSlices := utils.ChunkSlice(args, 500).([][]interface{})

	for _, argsSlice := range argsSlices {
		stmt, err := m.db.PrepareContext(ctx, `
		SELECT relation_id, sources.name FROM sources
		INNER JOIN relations_by_source ON sources.id = relations_by_source.source_id
		WHERE relation_id IN (?`+strings.Repeat(",?", len(argsSlice)-1)+`)`)
		if err != nil {
			return nil, fmt.Errorf("Unable to prepare statement for retrieving relation sources: %w", err)
		}
		row, err := stmt.QueryContext(ctx, argsSlice...)
		if err != nil {
			return nil, fmt.Errorf("Unable to retrieve sources for relations: %w", err)
		}

		var source string
		var relationId string
		for row.Next() {
			err = row.Scan(&relationId, &source)
			if err != nil {
				return nil, fmt.Errorf("Unable to scan row of relation source: %w", err)
			}
			if _, ok := idsSet[relationId]; !ok {
				idsSet[relationId] = []string{}
			}
			idsSet[relationId] = append(idsSet[relationId], source)
		}
	}
	return idsSet, nil
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

	sourceID, err := m.resolveSourceID(ctx, sourceName)
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

func (m *MariaDB) CollectMetrics(ctx context.Context) (map[string]int, error) {
	rows, err := m.db.QueryContext(ctx, "show global status like 'Com_stmt%'")
	if err != nil {
		return nil, fmt.Errorf("Unable to collect metrics from database: %v", err)
	}

	var value int
	var variableName string
	metrics := make(map[string]int)

	for rows.Next() {
		err := rows.Scan(&variableName, &value)
		if err != nil {
			return nil, err
		}
		metrics[variableName] = value
	}
	return metrics, nil
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

	Projections []knowledge.Projection
}

// NewMariaDBCursor create a new instance of MariaDBCursor
func NewMariaDBCursor(rows *sql.Rows, projections []knowledge.Projection) *MariaDBCursor {
	return &MariaDBCursor{
		Rows:        rows,
		Projections: projections,
	}
}

// HasMore tells whether there are more data to retrieve from the cursor
func (mc *MariaDBCursor) HasMore() bool {
	return mc.Rows.Next()
}

// Read read one more item from the cursor
func (mc *MariaDBCursor) Read(ctx context.Context, doc interface{}) error {
	var err error
	var fArr []string

	if fArr, err = mc.Rows.Columns(); err != nil {
		return fmt.Errorf("Unable to retrieve row columns: %w", err)
	}

	values := make([]interface{}, len(fArr))
	valuesPtr := make([]interface{}, len(fArr))
	for i := range values {
		valuesPtr[i] = &values[i]
	}

	if err := mc.Rows.Scan(valuesPtr...); err != nil {
		return fmt.Errorf("Unable to scan row items: %w", err)
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
			return fmt.Errorf("Unable to enqueue item: %w", err)
		}
	}

	output := make([]interface{}, len(mc.Projections))

	// This first pass, creates all temporary ids bound to nodes
	for i, pt := range mc.Projections {
		switch pt.ExpressionType {
		case knowledge.NodeExprType:
			var itemCount int64 = 3
			items, err := q.Get(itemCount)
			if err != nil {
				return fmt.Errorf("Unable to get %d items to build a node: %v", itemCount, err)
			}

			asset := knowledge.Asset{
				Type: schema.AssetType(fmt.Sprintf("%v", reflect.ValueOf(items[2]))),
				Key:  fmt.Sprintf("%v", reflect.ValueOf(items[1])),
			}

			awi := knowledge.AssetWithID{
				ID:    reflect.ValueOf(items[0]).String(),
				Asset: asset,
			}
			output[i] = awi
		case knowledge.EdgeExprType:
			var itemCount int64 = 4
			items, err := q.Get(itemCount)
			if err != nil {
				return fmt.Errorf("Unable to get %d items to build an edge: %v", itemCount, err)
			}

			r := knowledge.RelationWithID{
				ID:   reflect.ValueOf(items[0]).String(),
				From: reflect.ValueOf(items[1]).String(),
				To:   reflect.ValueOf(items[2]).String(),
				Type: schema.RelationKeyType(fmt.Sprintf("%v", reflect.ValueOf(items[3]))),
			}
			output[i] = r
		case knowledge.PropertyExprType:
			items, err := q.Get(1)
			if err != nil {
				return fmt.Errorf("Unable to get 1 property item: %v", err)
			}
			output[i] = items[0]
		}
	}

	val.Elem().Set(reflect.ValueOf(output))
	return nil
}

// Close the cursor
func (mc *MariaDBCursor) Close() error {
	return mc.Rows.Close()
}
