package knowledge

import (
	"encoding/json"
	"fmt"

	"github.com/clems4ever/go-graphkb/internal/schema"
)

type GraphEntryAction uint8

const (
	GraphEntryRemove GraphEntryAction = iota
	GraphEntryAdd
	GraphEntryNone
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
	assets    map[Asset]GraphEntryAction
	relations map[Relation]GraphEntryAction
}

// GraphJSON is the json representation of a graph
type GraphJSON struct {
	Assets    []Asset    `json:"assets"`
	Relations []Relation `json:"relations"`
}

// NewGraph create a graph
func NewGraph() *Graph {
	return &Graph{
		assets:    map[Asset]GraphEntryAction{},
		relations: map[Relation]GraphEntryAction{},
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
	if action, ok := g.assets[asset]; ok && action != GraphEntryAdd {
		g.assets[asset] = GraphEntryNone
	} else {
		g.assets[asset] = GraphEntryAdd
	}
	return AssetKey(asset), nil
}

// AddRelation add a relation to the graph
func (g *Graph) AddRelation(from AssetKey, relationType schema.RelationKeyType, to AssetKey) Relation {
	relation := Relation{
		Type: relationType,
		From: from,
		To:   to,
	}
	if action, ok := g.relations[relation]; ok && action != GraphEntryAdd {
		g.relations[relation] = GraphEntryNone
	} else {
		g.relations[relation] = GraphEntryAdd
	}
	return relation
}

// Assets return the assets in the graph
func (g *Graph) Assets() map[Asset]GraphEntryAction {
	return g.assets
}

// Relations return the relations in the graph
func (g *Graph) Relations() map[Relation]GraphEntryAction {
	return g.relations
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

	for a := range g.assets {
		if _, ok := other.assets[a]; !ok {
			return false
		}
	}

	if len(g.relations) != len(other.relations) {
		return false
	}

	for r := range g.relations {
		if _, ok := other.relations[r]; !ok {
			return false
		}
	}

	return true
}

// ExtractSchema extract the schema from the graph
func (g *Graph) ExtractSchema() schema.SchemaGraph {
	sg := schema.NewSchemaGraph()

	for a := range g.Assets() {
		sg.AddAsset(string(a.Type))
	}

	for r := range g.Relations() {
		sg.AddRelation(r.From.Type, string(r.Type), r.To.Type)
	}

	return sg
}

func (g *Graph) Clean() {
	for k, v := range g.assets {
		if v == GraphEntryRemove {
			delete(g.assets, k)
		} else {
			g.assets[k] = GraphEntryRemove
		}
	}
	for k, v := range g.relations {
		if v == GraphEntryRemove {
			delete(g.relations, k)
		} else {
			g.relations[k] = GraphEntryRemove
		}
	}
}

// MarshalJSON marshall the graph in JSON
func (g *Graph) MarshalJSON() ([]byte, error) {
	schemaJSON := new(GraphJSON)
	schemaJSON.Assets = []Asset{}
	schemaJSON.Relations = []Relation{}

	for v := range g.assets {
		schemaJSON.Assets = append(schemaJSON.Assets, v)
	}

	for e := range g.relations {
		schemaJSON.Relations = append(schemaJSON.Relations, e)
	}

	return json.Marshal(schemaJSON)
}

// UnmarshalJSON unmarshal a graph from JSON
func (g *Graph) UnmarshalJSON(b []byte) error {
	j := GraphJSON{}
	if err := json.Unmarshal(b, &j); err != nil {
		return err
	}

	g.assets = map[Asset]GraphEntryAction{}
	g.relations = map[Relation]GraphEntryAction{}

	for _, v := range j.Assets {
		g.assets[v] = GraphEntryRemove
	}

	for _, e := range j.Relations {
		g.relations[e] = GraphEntryRemove
	}
	return nil
}
