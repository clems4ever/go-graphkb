package database

import (
	"fmt"

	"github.com/clems4ever/go-graphkb/internal/knowledge"
)

// AssetTemporaryIDGenerator is a kind of cache associating IDs to assets so that pivot points are joined by those IDs
type AssetTemporaryIDGenerator struct {
	// map the DB ID with a temporarily generated ID.
	DBIDToTemporaryID map[int]int
	// map the Asset definition to the temporarily generated ID
	DefinitionToTemporaryID map[knowledge.Asset]int

	// The next ID to be used.
	generator int
}

// NewAssetTemporaryIDGenerator create an ID Generator
func NewAssetTemporaryIDGenerator() *AssetTemporaryIDGenerator {
	return &AssetTemporaryIDGenerator{
		generator:               0,
		DBIDToTemporaryID:       make(map[int]int),
		DefinitionToTemporaryID: make(map[knowledge.Asset]int),
	}
}

// Push an asset into the ID generator to insert a retrieve a temporary ID
func (atig *AssetTemporaryIDGenerator) Push(asset knowledge.Asset, DBID int) (int, error) {
	id, ok := atig.DefinitionToTemporaryID[asset]
	// If the asset is not in the DBID and Definition mappings, then we insert a new ID
	if !ok {
		_, ok2 := atig.DBIDToTemporaryID[DBID]
		// if the DBID has not been seen yet, assign the temporary ID
		if !ok2 {
			id = atig.generator
			atig.DefinitionToTemporaryID[asset] = id
			atig.DBIDToTemporaryID[DBID] = id
			atig.generator++
		} else {
			return 0, fmt.Errorf("DBID %d is already bound to another asset", DBID)
		}
	} else {
		id2, ok2 := atig.DBIDToTemporaryID[DBID]
		if ok2 && id != id2 {
			return 0, fmt.Errorf("DBID %d is already bound to another asset", DBID)
		}
		atig.DBIDToTemporaryID[DBID] = id
	}
	return id, nil
}

// Get the temporary ID related to the given DBID
func (atig *AssetTemporaryIDGenerator) Get(DBID int) (int, error) {
	id, ok := atig.DBIDToTemporaryID[DBID]
	if !ok {
		return 0, fmt.Errorf("DB ID %d does not exist in generator", DBID)
	}
	return id, nil
}

// Count the number of items in the generator
func (atig *AssetTemporaryIDGenerator) Count() int {
	return len(atig.DBIDToTemporaryID)
}
