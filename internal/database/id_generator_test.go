package database

import (
	"fmt"
	"testing"

	"github.com/clems4ever/go-graphkb/internal/knowledge"
	"github.com/clems4ever/go-graphkb/internal/schema"

	"github.com/stretchr/testify/assert"
)

func GenerateAssetHelper(idx int) knowledge.Asset {
	return knowledge.Asset{
		Key:  fmt.Sprintf("key%d", idx),
		Type: schema.AssetType(fmt.Sprintf("type%d", idx)),
	}
}

func TestGeneratedIDsAreDifferent(t *testing.T) {
	atig := NewAssetTemporaryIDGenerator()

	id1, _ := atig.Push(GenerateAssetHelper(1), 1)
	id2, _ := atig.Push(GenerateAssetHelper(2), 2)

	assert.Equal(t, 2, atig.Count())
	assert.NotEqual(t, id1, id2)
}

func TestPushMultipleTimesAssetWithDifferentDBID(t *testing.T) {
	atig := NewAssetTemporaryIDGenerator()

	id1, _ := atig.Push(GenerateAssetHelper(2), 2)
	id2, _ := atig.Push(GenerateAssetHelper(2), 4)

	assert.Equal(t, 1, atig.Count())
	assert.Equal(t, id1, id2)
}

func TestPushMultipleAssetsWithSameBID(t *testing.T) {
	atig := NewAssetTemporaryIDGenerator()

	atig.Push(GenerateAssetHelper(2), 2)
	_, err := atig.Push(GenerateAssetHelper(4), 2)

	assert.EqualError(t, err, "DBID 2 is already bound to another asset")
}

func TestGetExistingAssets(t *testing.T) {
	atig := NewAssetTemporaryIDGenerator()

	id1, _ := atig.Push(GenerateAssetHelper(2), 2)
	id2, _ := atig.Push(GenerateAssetHelper(4), 4)

	assert.NotEqual(t, id1, id2)

	idg, err := atig.Get(2)
	assert.NoError(t, err)
	assert.Equal(t, id1, idg)

	idg, err = atig.Get(4)
	assert.NoError(t, err)
	assert.Equal(t, id2, idg)

	idg, err = atig.Get(5)
	assert.EqualError(t, err, "DB ID 5 does not exist in generator")
}

func TestGetSameAsset(t *testing.T) {
	atig := NewAssetTemporaryIDGenerator()

	id1, _ := atig.Push(GenerateAssetHelper(2), 2)
	id2, _ := atig.Push(GenerateAssetHelper(2), 4)

	assert.Equal(t, id1, id2)

	idg, err := atig.Get(2)
	assert.NoError(t, err)
	assert.Equal(t, id1, idg)

	idg, err = atig.Get(4)
	assert.NoError(t, err)
	assert.Equal(t, id2, idg)
}
