package knowledge

import (
	"fmt"
	"log"
)

// AssetKey represent the key of the asset
type AssetKey struct {
	Type AssetType `json:"type"`
	Key  string    `json:"key"`
}

// Asset represent the asset with details
type Asset AssetKey

// NewAsset create a new asset from type and key
func NewAsset(assetType AssetType, assetKey string) Asset {
	return Asset{
		Type: assetType,
		Key:  assetKey,
	}
}

// RelationKey a relation key of the KB
type RelationKey struct {
	Type RelationKeyType `json:"type"`
	From AssetKey        `json:"from"`
	To   AssetKey        `json:"to"`
}

// Relation represent the relation with details
type Relation RelationKey

// Graph represent a Graph
type Graph struct {
	assets    map[Asset]bool
	relations map[Relation]bool
}

// NewGraph create a graph
func NewGraph() *Graph {
	return &Graph{
		assets:    make(map[Asset]bool),
		relations: make(map[Relation]bool),
	}
}

// AddAsset add an asset to the graph
func (g *Graph) AddAsset(assetType AssetType, assetKey string) AssetKey {
	if !SchemaRegistrySingleton.AssetExists(assetType) {
		log.Fatal(fmt.Errorf("Asset type '%s' does not exist to add asset %s", assetType, assetKey))
	}
	asset := Asset{Type: assetType, Key: assetKey}
	g.assets[asset] = true
	return AssetKey(asset)
}

// AddRelation add a relation to the graph
func (g *Graph) AddRelation(from AssetKey, relationType RelationKeyType, to AssetKey) Relation {
	if !SchemaRegistrySingleton.RelationExists(relationType) {
		log.Fatal(fmt.Errorf("Relation type '%s' does not exist", relationType))
	}
	relation := Relation{
		Type: relationType,
		From: from,
		To:   to,
	}
	g.relations[relation] = true
	return relation
}

// Assets return the assets in the graph
func (g *Graph) Assets() []Asset {
	assets := make([]Asset, 0)
	for a := range g.assets {
		assets = append(assets, a)
	}
	return assets
}

// Relations return the relations in the graph
func (g *Graph) Relations() []Relation {
	relations := make([]Relation, 0)
	for r := range g.relations {
		relations = append(relations, r)
	}
	return relations
}

// HasAsset return true if the asset is in the graph, false otherwise.
func (g *Graph) HasAsset(asset Asset) bool {
	_, ok := g.assets[asset]
	return ok
}

// HasRelation return true if the relation is in the graph, false otherwise.
func (g *Graph) HasRelation(relation Relation) bool {
	_, ok := g.relations[relation]
	return ok
}

// Merge merge other graph into the current graph
func (g *Graph) Merge(other *Graph) {
	for a := range other.assets {
		g.assets[a] = true
	}
	for r := range other.relations {
		g.relations[r] = true
	}
}

// Copy perform a deep copy of the graph
func (g *Graph) Copy() *Graph {
	graph := NewGraph()
	for k, v := range g.assets {
		graph.assets[k] = v
	}
	for k, v := range g.relations {
		graph.relations[k] = v
	}
	return graph
}

// Equal return true if graphs are equal, otherwise return false
func (g *Graph) Equal(other *Graph) bool {
	if len(g.assets) != len(other.assets) {
		return false
	}

	if len(g.relations) != len(other.relations) {
		return false
	}

	for k, v := range g.assets {
		v2, found := other.assets[k]
		if !found {
			return false
		}
		if v != v2 {
			return false
		}
	}

	for k, v := range g.relations {
		v2, found := other.relations[k]
		if !found {
			return false
		}
		if v != v2 {
			return false
		}
	}
	return true
}
