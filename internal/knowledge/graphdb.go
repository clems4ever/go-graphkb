package knowledge

import (
	"context"
	"fmt"

	"github.com/clems4ever/go-graphkb/internal/schema"
)

// GraphQueryResult represent a result from the database with the projection model
type GraphQueryResult struct {
	Cursor      Cursor
	Projections []Projection
}

// GraphDB an interface to a graph DB such as Arango or neo4j
type GraphDB interface {
	Close() error

	InitializeSchema() error

	ReadGraph(ctx context.Context, sourceName string, encoder *GraphEncoder) error

	// Atomic operations on the graph
	InsertAssets(ctx context.Context, sourceName string, assets []Asset) error
	InsertRelations(ctx context.Context, sourceName string, relations []Relation) error
	RemoveAssets(ctx context.Context, sourceName string, assets []Asset) error
	RemoveRelations(ctx context.Context, sourceName string, relations []Relation) error

	GetAssetSources(ctx context.Context, ids []string) (map[string][]string, error)
	GetRelationSources(ctx context.Context, ids []string) (map[string][]string, error)

	FlushAll(ctx context.Context) error

	CountAssets(ctx context.Context) (int64, error)
	CountAssetsBySource(ctx context.Context) (map[string]int64, error)
	CountRelations(ctx context.Context) (int64, error)
	CountRelationsBySource(ctx context.Context) (map[string]int64, error)

	Query(ctx context.Context, query SQLTranslation) (*GraphQueryResult, error)

	// Collect some metrics about the database
	CollectMetrics(ctx context.Context) (map[string]int, error)
}

// Cursor is a cursor over the results
type Cursor interface {
	HasMore() bool
	Read(ctx context.Context, doc interface{}) error
	Close() error
}

// AssetWithID represent an asset with an ID from the database
type AssetWithID struct {
	ID    string `json:"_id"`
	Asset `json:",inline"`
}

func (a AssetWithID) String() string {
	return fmt.Sprintf("Asset{id:%s, type:%s, value:%s}", a.ID, a.Asset.Type, a.Asset.Key)
}

// RelationWithID represent a relation with an ID from the database
type RelationWithID struct {
	ID   string                 `json:"_id"`
	From string                 `json:"from_id"`
	To   string                 `json:"to_id"`
	Type schema.RelationKeyType `json:"type"`
}

func (r RelationWithID) String() string {
	return fmt.Sprintf("Relation{id:%s, from:%s, to:%s, type:%s", r.ID, r.From, r.To, r.Type)
}
