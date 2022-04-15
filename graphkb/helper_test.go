package graphkb

import (
	"testing"

	"github.com/clems4ever/go-graphkb/internal/schema"
	"github.com/stretchr/testify/require"
)

func TestCreateAsset(t *testing.T) {
	asset := CreateAsset("foo")
	validators, _ := schema.AssetValidationRegistry.Get(asset)
	require.Empty(t, validators)

	asset = CreateAsset("foo", WithRegexpValidation(".*"))
	validators, _ = schema.AssetValidationRegistry.Get(asset)
	require.Equal(t, 1, len(validators))
}
