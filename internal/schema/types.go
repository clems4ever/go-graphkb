package schema

// AssetType define an asset type
type AssetType string

// RelationKeyType define a relation type
type RelationKeyType string

// RelationType define a relation type
type RelationType struct {
	FromType AssetType       `json:"from_type"`
	Type     RelationKeyType `json:"relation_type"`
	ToType   AssetType       `json:"to_type"`
}
