package knowledge

import (
	"context"
	"fmt"

	"github.com/clems4ever/go-graphkb/internal/query"
)

type GraphQueryResult struct {
	Cursor      Cursor
	Projections []Projection
}

// GraphDB an interface to a graph DB such as Arango or neo4j
type GraphDB interface {
	Close() error

	InitializeSchema() error

	UpdateGraph(source string, bulk *GraphUpdatesBulk) error
	ReadGraph(source string, graph *Graph) error

	FlushAll() error

	CountAssets() (int64, error)
	CountRelations() (int64, error)

	Query(ctx context.Context, query *query.QueryIL) (*GraphQueryResult, error)
}

// Cursor is a cursor over the results
type Cursor interface {
	HasMore() bool
	Read(ctx context.Context, doc interface{}) error
	Close() error
}

// EmptyCursor represent a cursor with no result
type EmptyCursor struct{}

// Count always return 0 in case of empty cursor
func (ec *EmptyCursor) Count() int64 {
	return 0
}

// HasMore always return false in case of empty cursor
func (ec *EmptyCursor) HasMore() bool {
	return false
}

// Read read the cursor (should not be called in case of empty cursor)
func (ec *EmptyCursor) Read(ctx context.Context, doc interface{}) error {
	return fmt.Errorf("Empty cursor cannot be read")
}

// Close closes the cursor
func (ec *EmptyCursor) Close() error {
	return nil
}

type AssetWithID struct {
	ID    string `json:"_id"`
	Asset `json:",inline"`
}

func (a AssetWithID) String() string {
	return fmt.Sprintf("Asset{id:%s, type:%s, value:%s}", a.ID, a.Asset.Type, a.Asset.Key)
}

type RelationWithID struct {
	ID   string          `json:"_id"`
	From string          `json:"from_id"`
	To   string          `json:"to_id"`
	Type RelationKeyType `json:"type"`
}

func (r RelationWithID) String() string {
	return fmt.Sprintf("Relation{id:%s, from:%s, to:%s, type:%s", r.ID, r.From, r.To, r.Type)
}
