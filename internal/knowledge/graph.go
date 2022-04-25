package knowledge

import (
	"encoding/json"
	"fmt"

	"github.com/clems4ever/go-graphkb/internal/schema"

	mapset "github.com/deckarep/golang-set"
)

// AssetKey represent the key of the asset
type AssetKey struct {
	Type schema.AssetType `json:"type"`
	Key  string           `json:"key"`
}

// Asset represent the asset with details
type Asset AssetKey

// NewAsset create a new asset from type and key
func NewAsset(assetType schema.AssetType, assetKey string) Asset {
	return Asset{
		Type: assetType,
		Key:  assetKey,
	}
}

// RelationKey a relation key of the KB
type RelationKey struct {
	Type schema.RelationKeyType `json:"type"`
	From AssetKey               `json:"from"`
	To   AssetKey               `json:"to"`
}

// Relation represent the relation with details
type Relation RelationKey

// Graph represent a Graph
type Graph struct {
	assets    mapset.Set
	relations mapset.Set
}

// GraphJSON is the json representation of a graph
type GraphJSON struct {
	Assets    []Asset    `json:"assets"`
	Relations []Relation `json:"relations"`
}

// NewGraph create a graph
func NewGraph() *Graph {
	return &Graph{
		assets:    mapset.NewSet(),
		relations: mapset.NewSet(),
	}
}

// AddAsset add an asset to the graph
func (g *Graph) AddAsset(assetType schema.AssetType, assetKey string) (AssetKey, error) {
	validators, _ := schema.AssetValidationRegistry.Get(assetType)
	for _, v := range validators {
		if !v(assetKey) {
			return AssetKey{}, fmt.Errorf("asset value %q does not match the type %q validators", assetKey, assetType)
		}
	}

	asset := Asset{Type: assetType, Key: assetKey}
	g.assets.Add(asset)
	return AssetKey(asset), nil
}

// AddRelation add a relation to the graph
func (g *Graph) AddRelation(from AssetKey, relationType schema.RelationKeyType, to AssetKey) Relation {
	relation := Relation{
		Type: relationType,
		From: from,
		To:   to,
	}
	g.relations.Add(relation)
	return relation
}

// Assets return the assets in the graph
func (g *Graph) Assets() []Asset {
	assets := make([]Asset, 0)
	for a := range g.assets.Iter() {
		assets = append(assets, a.(Asset))
	}
	return assets
}

// Relations return the relations in the graph
func (g *Graph) Relations() []Relation {
	relations := make([]Relation, 0)
	for r := range g.relations.Iter() {
		relations = append(relations, r.(Relation))
	}
	return relations
}

// HasAsset return true if the asset is in the graph, false otherwise.
func (g *Graph) HasAsset(asset Asset) bool {
	return g.assets.Contains(asset)
}

// HasRelation return true if the relation is in the graph, false otherwise.
func (g *Graph) HasRelation(relation Relation) bool {
	return g.relations.Contains(relation)
}

// Merge merge other graph into the current graph
func (g *Graph) Merge(other *Graph) {
	for a := range other.assets.Iter() {
		g.assets.Add(a)
	}
	for r := range other.relations.Iter() {
		g.relations.Add(r)
	}
}

// Copy perform a deep copy of the graph
func (g *Graph) Copy() *Graph {
	graph := NewGraph()
	for v := range g.assets.Iter() {
		graph.assets.Add(v)
	}
	for v := range g.relations.Iter() {
		graph.relations.Add(v)
	}
	return graph
}

// Equal return true if graphs are equal, otherwise return false
func (g *Graph) Equal(other *Graph) bool {
	if !g.assets.Equal(other.assets) {
		return false
	}

	if !g.relations.Equal(other.relations) {
		return false
	}
	return true
}

// ExtractSchema extract the schema from the graph
func (g *Graph) ExtractSchema() schema.SchemaGraph {
	sg := schema.NewSchemaGraph()

	for _, a := range g.Assets() {
		sg.AddAsset(string(a.Type))
	}

	for _, r := range g.Relations() {
		sg.AddRelation(r.From.Type, string(r.Type), r.To.Type)
	}

	return sg
}

// MarshalJSON marshall the graph in JSON
func (g *Graph) MarshalJSON() ([]byte, error) {
	schemaJSON := new(GraphJSON)
	schemaJSON.Assets = []Asset{}
	schemaJSON.Relations = []Relation{}

	for v := range g.assets.Iter() {
		schemaJSON.Assets = append(schemaJSON.Assets, v.(Asset))
	}

	for e := range g.relations.Iter() {
		schemaJSON.Relations = append(schemaJSON.Relations, e.(Relation))
	}

	return json.Marshal(schemaJSON)
}

// UnmarshalJSON unmarshal a graph from JSON
func (g *Graph) UnmarshalJSON(b []byte) error {
	j := GraphJSON{}
	if err := json.Unmarshal(b, &j); err != nil {
		return err
	}

	g.assets = mapset.NewSet()
	g.relations = mapset.NewSet()

	for _, v := range j.Assets {
		g.assets.Add(v)
	}

	for _, e := range j.Relations {
		g.relations.Add(e)
	}
	return nil
}
