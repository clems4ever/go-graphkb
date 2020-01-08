package knowledge

import "github.com/clems4ever/go-graphkb/internal/schema"

// SourceGraph represent a graph produce by a source
type SourceGraph struct {
	*Graph

	source string
}

var sourceAssetType schema.AssetType = "source"
var observedRelationType schema.RelationKeyType = "observed"

// NewSourceGraph create a new source graph
func NewSourceGraph(source string) *SourceGraph {
	sg := &SourceGraph{
		Graph:  NewGraph(),
		source: source,
	}
	sg.Graph.AddAsset(sourceAssetType, source)
	return sg
}

// AddAsset add one asset to the source graph
func (sg *SourceGraph) AddAsset(assetType schema.AssetType, assetKey string) AssetKey {
	ak := sg.Graph.AddAsset(assetType, assetKey)

	// Add an observation relation from source to asset
	sourceAsset := AssetKey{Type: sourceAssetType, Key: sg.source}
	sg.Graph.AddRelation(sourceAsset, observedRelationType, ak)
	return ak
}

// Merge merge two source graphs
func (sg *SourceGraph) Merge(otherGraph *SourceGraph) {
	sg.Graph.Merge(otherGraph.Graph)
}
