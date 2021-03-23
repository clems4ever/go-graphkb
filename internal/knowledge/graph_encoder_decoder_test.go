package knowledge

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	Asset1 = Asset{Type: "type1", Key: "value1"}
	Asset2 = Asset{Type: "type2", Key: "value2"}
	Asset3 = Asset{Type: "type3", Key: "value3"}
)

var (
	Relation1 = Relation{From: AssetKey(Asset1), Type: "is_linked_to", To: AssetKey(Asset2)}
	Relation2 = Relation{From: AssetKey(Asset1), Type: "has_relation_with", To: AssetKey(Asset2)}
	Relation3 = Relation{From: AssetKey(Asset1), Type: "has_weird_relation_with", To: AssetKey(Asset3)}
)

func TestEncodeAssets(t *testing.T) {
	buff := bytes.NewBuffer(nil)

	encoder := NewGraphEncoder(buff)

	assert.NoError(t, encoder.EncodeAsset(Asset1))
	assert.NoError(t, encoder.EncodeAsset(Asset2))

	assert.Equal(t,
		"A{\"type\":\"type1\",\"key\":\"value1\"}\nA{\"type\":\"type2\",\"key\":\"value2\"}\n",
		string(buff.Bytes()))
}

func TestEncodeAssetsAndRelations(t *testing.T) {
	buff := bytes.NewBuffer(nil)

	encoder := NewGraphEncoder(buff)

	assert.NoError(t, encoder.EncodeRelation(Relation1))
	assert.NoError(t, encoder.EncodeRelation(Relation2))

	assert.Equal(t,
		"R{\"type\":\"is_linked_to\",\"from\":{\"type\":\"type1\",\"key\":\"value1\"},\"to\":{\"type\":\"type2\",\"key\":\"value2\"}}\nR{\"type\":\"has_relation_with\",\"from\":{\"type\":\"type1\",\"key\":\"value1\"},\"to\":{\"type\":\"type2\",\"key\":\"value2\"}}\n",
		string(buff.Bytes()))
}

func TestEncodeAndDecode(t *testing.T) {
	buff := bytes.NewBuffer(nil)

	encoder := NewGraphEncoder(buff)

	assert.NoError(t, encoder.EncodeRelation(Relation1))
	assert.NoError(t, encoder.EncodeRelation(Relation2))
	assert.NoError(t, encoder.EncodeRelation(Relation3))

	decoder := NewGraphDecoder(buff)

	graph := NewGraph()

	assert.NoError(t, decoder.Decode(graph))

	assert.Len(t, graph.Assets(), 3)
	assert.Len(t, graph.Relations(), 3)

	assert.True(t, graph.HasAsset(Asset1))
	assert.True(t, graph.HasAsset(Asset2))
	assert.True(t, graph.HasAsset(Asset3))

	assert.True(t, graph.HasRelation(Relation1))
	assert.True(t, graph.HasRelation(Relation2))
	assert.True(t, graph.HasRelation(Relation3))
}
