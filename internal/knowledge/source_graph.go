package knowledge

// SourceGraph represent a graph produce by a source
type SourceGraph struct {
	*Graph

	source string
}

var sourceAssetType AssetType = "source"
var observedRelationType RelationKeyType = "observed"

func init() {
	SchemaRegistrySingleton.AddAssetType(sourceAssetType)
	SchemaRegistrySingleton.AddRelationType(observedRelationType)
}

// NewSourceGraph create a new source graph
func NewSourceGraph(source string) *SourceGraph {
	sg := &SourceGraph{
		Graph:  NewGraph(),
		source: source,
	}
	sg.Graph.AddAsset("source", source)
	return sg
}

// AddAsset add one asset to the source graph
func (sg *SourceGraph) AddAsset(assetType AssetType, assetKey string) AssetKey {
	ak := sg.Graph.AddAsset(assetType, assetKey)

	// Add an observation relation from source to asset
	sourceAsset := AssetKey{Type: sourceAssetType, Key: sg.source}
	sg.Graph.AddRelation(sourceAsset, "observed", ak)
	return ak
}

// Merge merge two source graphs
func (sg *SourceGraph) Merge(otherGraph *SourceGraph) {
	sg.Graph.Merge(otherGraph.Graph)
}
