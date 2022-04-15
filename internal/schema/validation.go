package schema

import (
	"github.com/clems4ever/go-graphkb/internal/utils"
)

type AssetValidationFunc func(string) bool

type ValidationRegistry interface {
	Get(AssetType) ([]AssetValidationFunc, bool)
}

var (
	AssetValidationRegistry ValidationRegistry = utils.NewRegistry[AssetType, []AssetValidationFunc]()
)

func AddAssetValidator(asset AssetType, v AssetValidationFunc) {
	validators, _ := AssetValidationRegistry.Get(asset)
	validators = append(validators, v)
	AssetValidationRegistry.(*utils.Registry[AssetType, []AssetValidationFunc]).Set(asset, validators)
}
