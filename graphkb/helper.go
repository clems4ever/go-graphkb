package graphkb

import (
	"regexp"

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

// CreateAssetOption is an option that can be used when creating an asset
type CreateAssetOption func(asset AssetType)

// WithRegexpValidation adds a regexp validation check to an asset
func WithRegexpValidation(r string) CreateAssetOption {
	reg := regexp.MustCompile(r)
	return func(asset AssetType) {
		schema.AddAssetValidator(asset, func(s string) bool {
			return reg.MatchString(s)
		})
	}
}

// WithValuesValidation adds a check ensuring an asset value is part of the given list
func WithValuesValidation(expected ...string) CreateAssetOption {
	ex := make(map[string]bool, len(expected))
	for _, e := range expected {
		ex[e] = true
	}
	return func(asset AssetType) {
		schema.AddAssetValidator(asset, func(s string) bool {
			return ex[s]
		})
	}
}

// CreateAsset helper function for creating an asset
func CreateAsset(fromType string, options ...CreateAssetOption) AssetType {
	asset := schema.AssetType(fromType)
	for _, o := range options {
		o(asset)
	}
	return asset
}
