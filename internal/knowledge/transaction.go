package knowledge

import (
	"sync"

	"github.com/clems4ever/go-graphkb/internal/schema"
)

type GraphUpdateRequestBody struct {
	Updates *GraphUpdatesBulk  `json:"updates"`
	Schema  schema.SchemaGraph `json:"schema"`
}

// Transaction represent a transaction generating updates by diffing the provided graph against
// the previous version.
type Transaction struct {
	api *GraphAPI

	currentGraph *Graph

	// The graph being updated
	newGraph *Graph
	binder   *GraphBinder

	// Lock used when binding or relating assets
	mutex sync.Mutex
}

// Relate create a relation between two assets
func (cgt *Transaction) Relate(from string, relationType schema.RelationType, to string) {
	cgt.mutex.Lock()
	cgt.binder.Relate(from, relationType, to)
	cgt.mutex.Unlock()
}

// Bind bind one asset to an asset type from the schema
func (cgt *Transaction) Bind(asset string, assetType schema.AssetType) {
	cgt.mutex.Lock()
	cgt.binder.Bind(asset, assetType)
	cgt.mutex.Unlock()
}

// Commit commit the transaction and gives ownership to the source for caching.
func (cgt *Transaction) Commit() (*Graph, error) {
	sg := cgt.newGraph.ExtractSchema()
	bulk := GenerateGraphUpdatesBulk(cgt.currentGraph, cgt.newGraph)

	if err := cgt.api.UpdateGraph(sg, *bulk); err != nil {
		return nil, err
	}

	g := cgt.newGraph
	cgt.newGraph = NewGraph()
	return g, nil // give ownership of the transaction graph so that it can be cached if needed
}
