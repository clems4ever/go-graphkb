package knowledge

// SchemaGraph represent the graph of a source
type SchemaGraph struct {
	vertices map[AssetType]bool
	edges    map[RelationType]bool
}

// NewSchemaGraph create a source graph
func NewSchemaGraph() SchemaGraph {
	return SchemaGraph{
		vertices: make(map[AssetType]bool),
		edges:    make(map[RelationType]bool),
	}
}

// AddAsset add an asset type as a vertex in the source graph
func (sg *SchemaGraph) AddAsset(assetType string) AssetType {
	t := AssetType(assetType)
	sg.vertices[t] = true
	return t
}

// Assets return all the assets in the graph
func (sg *SchemaGraph) Assets() []AssetType {
	assets := make([]AssetType, 0)
	for a := range sg.vertices {
		assets = append(assets, a)
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
	sg.edges[rt] = true
	return rt
}

// Relations return all the relations in the graph
func (sg *SchemaGraph) Relations() []RelationType {
	relations := make([]RelationType, 0)
	for r := range sg.edges {
		relations = append(relations, r)
	}
	return relations
}

// Merge merge other graph into the current graph
func (sg *SchemaGraph) Merge(other SchemaGraph) {
	for vertex := range other.vertices {
		sg.vertices[vertex] = true
	}
	for edge := range other.edges {
		sg.edges[edge] = true
	}
}
