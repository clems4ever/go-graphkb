package knowledge

import (
	"encoding/json"

	mapset "github.com/deckarep/golang-set"
)

// GraphUpdatesBulk represent a bulk of asset and relation updates to perform on the graph
type GraphUpdatesBulk struct {
	assetUpserts     mapset.Set
	assetRemovals    mapset.Set
	relationUpserts  mapset.Set
	relationRemovals mapset.Set
}

// GraphUpdatesBulkJSON represent a bulk in JSON form
type GraphUpdatesBulkJSON struct {
	AssetUpserts     []Asset    `json:"asset_upserts"`
	AssetRemovals    []Asset    `json:"asset_removals"`
	RelationUpserts  []Relation `json:"relation_upserts"`
	RelationRemovals []Relation `json:"relation_removals"`
}

// NewGraphUpdatesBulk create an instance of graph updates
func NewGraphUpdatesBulk() *GraphUpdatesBulk {
	return &GraphUpdatesBulk{
		assetUpserts:     mapset.NewSet(),
		assetRemovals:    mapset.NewSet(),
		relationUpserts:  mapset.NewSet(),
		relationRemovals: mapset.NewSet(),
	}
}

// Clear the bulk
func (gub *GraphUpdatesBulk) Clear() {
	gub.assetUpserts.Clear()
	gub.assetRemovals.Clear()
	gub.relationUpserts.Clear()
	gub.relationRemovals.Clear()
}

// GetAssetUpserts retrieve the list of all assets to be upsered
func (gub *GraphUpdatesBulk) GetAssetUpserts() []Asset {
	assets := []Asset{}
	for v := range gub.assetUpserts.Iter() {
		assets = append(assets, v.(Asset))
	}
	return assets
}

// HasAssetUpsert check whether asset needs to be upserted
func (gub *GraphUpdatesBulk) HasAssetUpsert(asset Asset) bool {
	return gub.assetUpserts.Contains(asset)
}

// UpsertAsset create an operation to upsert an asset
func (gub *GraphUpdatesBulk) UpsertAsset(asset Asset) {
	gub.assetUpserts.Add(asset)
}

// UpsertAssets append multiple assets to upsert
func (gub *GraphUpdatesBulk) UpsertAssets(assets ...Asset) {
	for _, a := range assets {
		gub.assetUpserts.Add(a)
	}
}

// GetAssetRemovals retrieve all assets to be removed
func (gub *GraphUpdatesBulk) GetAssetRemovals() []Asset {
	assets := []Asset{}
	for v := range gub.assetRemovals.Iter() {
		assets = append(assets, v.(Asset))
	}
	return assets
}

// HasAssetRemoval check whether this asset need to be removed
func (gub *GraphUpdatesBulk) HasAssetRemoval(asset Asset) bool {
	return gub.assetRemovals.Contains(asset)
}

// RemoveAsset create an operation to remove an asset
func (gub *GraphUpdatesBulk) RemoveAsset(asset Asset) {
	gub.assetRemovals.Add(asset)
}

// RemoveAssets create multiple asset removal operations
func (gub *GraphUpdatesBulk) RemoveAssets(assets ...Asset) {
	for _, a := range assets {
		gub.assetRemovals.Add(a)
	}
}

// GetRelationUpserts retrieve all relations to be upserted
func (gub *GraphUpdatesBulk) GetRelationUpserts() []Relation {
	relations := []Relation{}
	for v := range gub.relationUpserts.Iter() {
		relations = append(relations, v.(Relation))
	}
	return relations
}

// HasRelationUpsert check whether the relation needs to be upserted
func (gub *GraphUpdatesBulk) HasRelationUpsert(relation Relation) bool {
	return gub.relationUpserts.Contains(relation)
}

// UpsertRelation create an operation to upsert an relation
func (gub *GraphUpdatesBulk) UpsertRelation(relation Relation) {
	gub.relationUpserts.Add(relation)
}

// UpsertRelations create multiple relation upsert operations
func (gub *GraphUpdatesBulk) UpsertRelations(relations ...Relation) {
	for _, r := range relations {
		gub.relationUpserts.Add(r)
	}
}

// GetRelationRemovals retrieve all relations to be removed
func (gub *GraphUpdatesBulk) GetRelationRemovals() []Relation {
	relations := []Relation{}
	for v := range gub.relationRemovals.Iter() {
		relations = append(relations, v.(Relation))
	}
	return relations
}

// HasRelationRemoval check whether the bulk contains the removal of this relation
func (gub *GraphUpdatesBulk) HasRelationRemoval(relation Relation) bool {
	return gub.relationRemovals.Contains(relation)
}

// RemoveRelation create an operation to remove a relation
func (gub *GraphUpdatesBulk) RemoveRelation(relation Relation) {
	gub.relationRemovals.Add(relation)
}

// RemoveRelations create multiple relation removal operations
func (gub *GraphUpdatesBulk) RemoveRelations(relations ...Relation) {
	for _, r := range relations {
		gub.relationRemovals.Add(r)
	}
}

// MarshalJSON marshal a graph update bulk
func (gub *GraphUpdatesBulk) MarshalJSON() ([]byte, error) {
	j := &GraphUpdatesBulkJSON{}
	j.AssetUpserts = gub.GetAssetUpserts()
	j.AssetRemovals = gub.GetAssetRemovals()
	j.RelationUpserts = gub.GetRelationUpserts()
	j.RelationRemovals = gub.GetRelationRemovals()
	return json.Marshal(j)
}

// UnmarshalJSON unmarshal a graph updates bulk
func (gub *GraphUpdatesBulk) UnmarshalJSON(bytes []byte) error {
	j := GraphUpdatesBulkJSON{}
	if err := json.Unmarshal(bytes, &j); err != nil {
		return err
	}

	*gub = *NewGraphUpdatesBulk()
	gub.UpsertAssets(j.AssetUpserts...)
	gub.UpsertRelations(j.RelationUpserts...)
	gub.RemoveAssets(j.AssetRemovals...)
	gub.RemoveRelations(j.RelationRemovals...)
	return nil
}

// GenerateGraphUpdatesBulk generate a graph update bulk by taking the difference between new graph
// and previous graph. It means the updates would transform previous graph into new graph.
func GenerateGraphUpdatesBulk(newGraph *Graph) *GraphUpdatesBulk {
	if newGraph == nil {
		newGraph = NewGraph()
	}

	bulk := NewGraphUpdatesBulk()

	// upsert / delete assets
	for a, action := range newGraph.Assets() {
		switch action {
		case graphEntryAdd:
			bulk.UpsertAsset(a)
		case graphEntryRemove:
			bulk.RemoveAsset(a)
		}
	}

	// upsert / delete relations
	for r, action := range newGraph.Relations() {
		switch action {
		case graphEntryAdd:
			bulk.UpsertRelation(r)
		case graphEntryRemove:
			bulk.RemoveRelation(r)
		}
	}

	return bulk
}
