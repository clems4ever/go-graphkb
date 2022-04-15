package knowledge

import (
	"fmt"

	"github.com/clems4ever/go-graphkb/internal/schema"
)

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
func (gb *GraphBinder) Relate(from string, relationType schema.RelationType, to string) error {
	fromAsset, err := gb.graph.AddAsset(relationType.FromType, from)
	if err != nil {
		return fmt.Errorf("relate: from asset: %w", err)
	}
	toAsset, err := gb.graph.AddAsset(relationType.ToType, to)
	if err != nil {
		return fmt.Errorf("relate: to asset: %w", err)
	}
	gb.graph.AddRelation(fromAsset, relationType.Type, toAsset)
	return nil
}

// Bind bind one asset to a type
func (gb *GraphBinder) Bind(asset string, assetType schema.AssetType) error {
	_, err := gb.graph.AddAsset(assetType, asset)
	return err
}
