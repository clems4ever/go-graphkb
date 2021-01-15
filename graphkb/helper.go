package graphkb

import (
	"github.com/clems4ever/go-graphkb/internal/schema"
)

// CreateRelation helper function for creating a relation
func CreateRelation(fromType schema.AssetType, relation, toType schema.AssetType) RelationType {
	return schema.RelationType{
		FromType: fromType,
		Type:     RelationKeyType(relation),
		ToType:   toType,
	}
}

// CreateAsset helper function for creating an asset
func CreateAsset(fromType string) AssetType {
	return schema.AssetType(fromType)
}
