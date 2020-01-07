package knowledge

import "errors"

// AssetType define an asset type
type AssetType string

// RelationKeyType define a relation type
type RelationKeyType string

// RelationType define a relation type
type RelationType struct {
	Type     RelationKeyType `json:"relation_type"`
	FromType AssetType       `json:"from_type"`
	ToType   AssetType       `json:"to_type"`
}

// SchemaRegistry represent a registry containing all defined types
type SchemaRegistry struct {
	assetTypes    map[AssetType]bool
	relationTypes map[RelationKeyType]bool
}

// SchemaRegistrySingleton is a singleton of schema registry
var SchemaRegistrySingleton SchemaRegistry

// ErrAssetTypeDoesNotExist error thrown when asset type does not exist in registry
var ErrAssetTypeDoesNotExist = errors.New("Asset type does not exist")

// ErrRelationTypeDoesNotExist error thrown when relation type does not exist in registry
var ErrRelationTypeDoesNotExist = errors.New("Relation type does not exist")

func init() {
	SchemaRegistrySingleton = *NewSchemaRegistry()
}

// NewSchemaRegistry create a schema registry
func NewSchemaRegistry() *SchemaRegistry {
	schemaRegistry := new(SchemaRegistry)
	schemaRegistry.assetTypes = make(map[AssetType]bool)
	schemaRegistry.relationTypes = make(map[RelationKeyType]bool)
	return schemaRegistry
}

// AddAssetType to registry
func (sr *SchemaRegistry) AddAssetType(assetType AssetType) AssetType {
	sr.assetTypes[assetType] = true
	return assetType
}

// AssetExists check if asset exists in registry
func (sr *SchemaRegistry) AssetExists(assetType AssetType) bool {
	_, ok := sr.assetTypes[assetType]
	return ok
}

// AddRelationType add a relation type to in the registry
func (sr *SchemaRegistry) AddRelationType(relationType RelationKeyType) RelationKeyType {
	sr.relationTypes[relationType] = true
	return relationType
}

// RelationExists check if relation exists in registry
func (sr *SchemaRegistry) RelationExists(relationType RelationKeyType) bool {
	_, ok := sr.relationTypes[relationType]
	return ok
}

// AssetTypes get the registered asset types
func (sr *SchemaRegistry) AssetTypes() []AssetType {
	assetTypes := make([]AssetType, 0)
	for t := range sr.assetTypes {
		assetTypes = append(assetTypes, t)
	}
	return assetTypes
}

// RelationTypes get the list of registered relation keys
func (sr *SchemaRegistry) RelationTypes() []RelationKeyType {
	relationTypesMap := make(map[RelationKeyType]bool)
	for t := range sr.relationTypes {
		relationTypesMap[t] = true
	}

	relationTypesSlice := make([]RelationKeyType, 0)
	for k := range relationTypesMap {
		relationTypesSlice = append(relationTypesSlice, k)
	}
	return relationTypesSlice
}
