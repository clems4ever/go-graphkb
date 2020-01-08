package knowledge

import (
	"sync"

	"github.com/clems4ever/go-graphkb/internal/schema"
)

// GraphUpdateTransaction represent a transaction for updating graph.
type GraphUpdateTransaction struct {
	eventBus chan SourceSubGraphUpdates

	source  string
	Updates GraphUpdatesBulk
}

// Commit commit the transaction
func (gut *GraphUpdateTransaction) Commit() {
	gut.eventBus <- SourceSubGraphUpdates{
		Updates: gut.Updates,
		Source:  gut.source,
	}
}

// CompleteGraphTransaction represent a transaction made of a complete graph. When committed this graph
// will be diffed against the current version of the graph to generate updates that will be sent to the graph.
type CompleteGraphTransaction struct {
	GraphUpdateTransaction

	currentGraph *SourceGraph
	newGraph     *SourceGraph
	binder       *GraphBinder

	mutex sync.Mutex
}

// Relate create a relation between two assets
func (cgt *CompleteGraphTransaction) Relate(from string, relationType schema.RelationType, to string) {
	cgt.mutex.Lock()
	cgt.binder.Relate(from, relationType, to)
	cgt.mutex.Unlock()
}

// Bind bind one asset to an asset type from the schema
func (cgt *CompleteGraphTransaction) Bind(asset string, assetType schema.AssetType) {
	cgt.mutex.Lock()
	cgt.binder.Bind(asset, assetType)
	cgt.mutex.Unlock()
}

// Commit commit the transaction and gives ownership to the source for caching.
func (cgt *CompleteGraphTransaction) Commit() *SourceGraph {
	var currentGraph *Graph
	if cgt.currentGraph != nil {
		currentGraph = cgt.currentGraph.Graph
	}
	sg := cgt.newGraph.Graph.ExtractSchema()

	bulk := GenerateGraphUpdatesBulk(currentGraph, cgt.newGraph.Graph)

	cgt.eventBus <- SourceSubGraphUpdates{
		Updates: *bulk,
		Source:  cgt.source,
		Schema:  sg,
	}
	g := cgt.newGraph
	cgt.newGraph = NewSourceGraph(cgt.source)
	return g // give ownership of the transaction graph so that it can be cached if needed
}
