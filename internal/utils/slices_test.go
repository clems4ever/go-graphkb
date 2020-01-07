package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChunkSlice(t *testing.T) {
	assert.ElementsMatch(t, ChunkSlice(
		[]interface{}{10, 2, 3, 4, 5, 34, 34, 12}, 4),
		[][]interface{}{
			[]interface{}{10, 2, 3, 4},
			[]interface{}{5, 34, 34, 12},
		})

	assert.ElementsMatch(t, ChunkSlice(
		[]interface{}{10, 2, 3, 4, 5, 34, 34, 12, 13}, 4),
		[][]interface{}{
			[]interface{}{10, 2, 3, 4},
			[]interface{}{5, 34, 34, 12},
			[]interface{}{13},
		})

	assert.ElementsMatch(t, ChunkSlice(
		[]interface{}{}, 4),
		[][]interface{}{})

	assert.ElementsMatch(t, ChunkSlice(
		[]interface{}{10}, 4),
		[][]interface{}{
			[]interface{}{10},
		})
}
