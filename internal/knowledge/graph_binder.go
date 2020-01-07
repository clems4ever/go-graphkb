package knowledge

// GraphBinder represent a graph builder which bind assets and relations to graph schema provided by the source
type GraphBinder struct {
	graph *SourceGraph
}

// NewGraphBinder create an instance of graph binder
func NewGraphBinder(graph *SourceGraph) *GraphBinder {
	return &GraphBinder{
		graph: graph,
	}
}

// Relate relate one asset to another
func (gb *GraphBinder) Relate(from string, relationType RelationType, to string) {
	fromAsset := gb.graph.AddAsset(relationType.FromType, from)
	toAsset := gb.graph.AddAsset(relationType.ToType, to)
	gb.graph.AddRelation(fromAsset, relationType.Type, toAsset)
}

// Bind bind one asset to a type
func (gb *GraphBinder) Bind(asset string, assetType AssetType) {
	gb.graph.AddAsset(assetType, asset)
}
