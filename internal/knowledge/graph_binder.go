package knowledge

import "github.com/clems4ever/go-graphkb/internal/schema"

// GraphBinder represent a graph builder which bind assets and relations to graph schema provided by the source
type GraphBinder struct {
	graph *Graph
}

// NewGraphBinder create an instance of graph binder
func NewGraphBinder(graph *Graph) *GraphBinder {
	return &GraphBinder{
		graph: graph,
	}
}

// Relate relate one asset to another
func (gb *GraphBinder) Relate(from string, relationType schema.RelationType, to string) {
	fromAsset := gb.graph.AddAsset(relationType.FromType, from)
	toAsset := gb.graph.AddAsset(relationType.ToType, to)
	gb.graph.AddRelation(fromAsset, relationType.Type, toAsset)
}

// Bind bind one asset to a type
func (gb *GraphBinder) Bind(asset string, assetType schema.AssetType) {
	gb.graph.AddAsset(assetType, asset)
}
