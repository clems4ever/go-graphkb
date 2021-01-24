package client

import (
	"fmt"
	"sync"

	"github.com/clems4ever/go-graphkb/internal/knowledge"
	"github.com/clems4ever/go-graphkb/internal/schema"
)

// Transaction represent a transaction generating updates by diffing the provided graph against
// the previous version.
type Transaction struct {
	client *GraphClient

	currentGraph *knowledge.Graph

	// The graph being updated
	newGraph *knowledge.Graph
	binder   *knowledge.GraphBinder

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
func (cgt *Transaction) Commit() (*knowledge.Graph, error) {
	sg := cgt.newGraph.ExtractSchema()

	if err := cgt.client.UpdateSchema(sg); err != nil {
		return nil, fmt.Errorf("Unable to update the schema of the graph")
	}

	bulk := knowledge.GenerateGraphUpdatesBulk(cgt.currentGraph, cgt.newGraph)

	for _, r := range bulk.GetRelationRemovals() {
		if err := cgt.client.DeleteRelation(r); err != nil {
			return nil, fmt.Errorf("Unable to remove the relation %v: %v", r, err)
		}
	}
	for _, a := range bulk.GetAssetRemovals() {
		if err := cgt.client.DeleteAsset(a); err != nil {
			return nil, fmt.Errorf("Unable to remove the asset %v: %v", a, err)
		}
	}

	for _, a := range bulk.GetAssetUpserts() {
		if err := cgt.client.UpsertAsset(a); err != nil {
			return nil, fmt.Errorf("Unable to upsert the asset %v: %v", a, err)
		}
	}
	for _, r := range bulk.GetRelationUpserts() {
		if err := cgt.client.UpsertRelation(r); err != nil {
			return nil, fmt.Errorf("Unable to upsert the relation %v: %v", r, err)
		}
	}

	g := cgt.newGraph
	cgt.newGraph = knowledge.NewGraph()
	return g, nil // give ownership of the transaction graph so that it can be cached if needed
}
