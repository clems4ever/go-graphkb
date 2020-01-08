package schema

import mapset "github.com/deckarep/golang-set"

import "encoding/json"

// SchemaGraph represent the graph of a source
type SchemaGraph struct {
	Vertices mapset.Set
	Edges    mapset.Set
}

// SchemaGraphJSON is the json representation of a schema graph
type SchemaGraphJSON struct {
	Vertices []AssetType    `json:"vertices"`
	Edges    []RelationType `json:"edges"`
}

// NewSchemaGraph create a source graph
func NewSchemaGraph() SchemaGraph {
	return SchemaGraph{
		Vertices: mapset.NewSet(),
		Edges:    mapset.NewSet(),
	}
}

// AddAsset add an asset type as a vertex in the source graph
func (sg *SchemaGraph) AddAsset(assetType string) AssetType {
	t := AssetType(assetType)
	sg.Vertices.Add(t)
	return t
}

// Assets return all the assets in the graph
func (sg *SchemaGraph) Assets() []AssetType {
	assets := []AssetType{}
	for a := range sg.Vertices.Iter() {
		assets = append(assets, a.(AssetType))
	}
	return assets
}

// AddRelation add a relation between asset types
func (sg *SchemaGraph) AddRelation(fromType AssetType, relationType string, toType AssetType) RelationType {
	rt := RelationType{
		Type:     RelationKeyType(relationType),
		FromType: fromType,
		ToType:   toType,
	}
	sg.Edges.Add(rt)
	return rt
}

// Relations return all the relations in the graph
func (sg *SchemaGraph) Relations() []RelationType {
	relations := []RelationType{}
	for r := range sg.Edges.Iter() {
		relations = append(relations, r.(RelationType))
	}
	return relations
}

// Merge merge other graph into the current graph
func (sg *SchemaGraph) Merge(other SchemaGraph) {
	for vertex := range other.Vertices.Iter() {
		sg.Vertices.Add(vertex)
	}
	for edge := range other.Edges.Iter() {
		sg.Edges.Add(edge)
	}
}

// Equal check if two schema graphs are equal
func (sg *SchemaGraph) Equal(other SchemaGraph) bool {
	if !sg.Vertices.Equal(other.Vertices) {
		return false
	}

	if !sg.Edges.Equal(other.Edges) {
		return false
	}
	return true
}

func (sg *SchemaGraph) ToJSON() ([]byte, error) {
	schemaJson := new(SchemaGraphJSON)
	schemaJson.Vertices = []AssetType{}
	schemaJson.Edges = []RelationType{}

	for v := range sg.Vertices.Iter() {
		vertice := v.(AssetType)
		schemaJson.Vertices = append(schemaJson.Vertices, vertice)
	}

	for e := range sg.Edges.Iter() {
		edge := e.(RelationType)
		schemaJson.Edges = append(schemaJson.Edges, edge)
	}

	return json.Marshal(schemaJson)
}

func (sg *SchemaGraph) FromJSON(b []byte) error {
	j := SchemaGraphJSON{}
	if err := json.Unmarshal(b, &j); err != nil {
		return err
	}

	sg.Vertices = mapset.NewSet()
	sg.Edges = mapset.NewSet()

	for _, v := range j.Vertices {
		sg.Vertices.Add(v)
	}

	for _, e := range j.Edges {
		sg.Edges.Add(e)
	}
	return nil
}
