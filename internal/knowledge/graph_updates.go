package knowledge

// GraphUpdatesBulk represent a bulk of asset and relation updates to perform on the graph
type GraphUpdatesBulk struct {
	AssetUpserts     []Asset
	AssetRemovals    []Asset
	RelationUpserts  []Relation
	RelationRemovals []Relation
}

// NewGraphUpdatesBulk create an instance of graph updates
func NewGraphUpdatesBulk() *GraphUpdatesBulk {
	return &GraphUpdatesBulk{
		AssetUpserts:     make([]Asset, 0),
		AssetRemovals:    make([]Asset, 0),
		RelationUpserts:  make([]Relation, 0),
		RelationRemovals: make([]Relation, 0),
	}
}

// UpsertAsset create an operation to upsert an asset
func (gub *GraphUpdatesBulk) UpsertAsset(asset Asset) {
	gub.AssetUpserts = append(gub.AssetUpserts, asset)
}

// UpsertAssets append multiple assets to upsert
func (gub *GraphUpdatesBulk) UpsertAssets(asset ...Asset) {
	gub.AssetUpserts = append(gub.AssetUpserts, asset...)
}

// RemoveAsset create an operation to remove an asset
func (gub *GraphUpdatesBulk) RemoveAsset(asset Asset) {
	gub.AssetRemovals = append(gub.AssetRemovals, asset)
}

// RemoveAssets create multiple asset removal operations
func (gub *GraphUpdatesBulk) RemoveAssets(asset ...Asset) {
	gub.AssetRemovals = append(gub.AssetRemovals, asset...)
}

// UpsertRelation create an operation to upsert an relation
func (gub *GraphUpdatesBulk) UpsertRelation(relation Relation) {
	gub.RelationUpserts = append(gub.RelationUpserts, relation)
}

// UpsertRelations create multiple relation upsert operations
func (gub *GraphUpdatesBulk) UpsertRelations(relation ...Relation) {
	gub.RelationUpserts = append(gub.RelationUpserts, relation...)
}

// RemoveRelation create an operation to remove a relation
func (gub *GraphUpdatesBulk) RemoveRelation(relation Relation) {
	gub.RelationRemovals = append(gub.RelationRemovals, relation)
}

// RemoveRelations create multiple relation removal operations
func (gub *GraphUpdatesBulk) RemoveRelations(relation ...Relation) {
	gub.RelationRemovals = append(gub.RelationRemovals, relation...)
}

// GenerateGraphUpdatesBulk generate a graph update bulk by taking the difference between new graph
// and previous graph. It means the updates would transform previous graph into new graph.
func GenerateGraphUpdatesBulk(previousGraph *Graph, newGraph *Graph) *GraphUpdatesBulk {
	if previousGraph == nil {
		previousGraph = NewGraph()
	}

	if newGraph == nil {
		newGraph = NewGraph()
	}

	bulk := NewGraphUpdatesBulk()

	// upsert new assets
	for _, a := range newGraph.Assets() {
		if found := previousGraph.HasAsset(a); !found {
			bulk.UpsertAsset(a)
		}
	}

	// Upsert new relations
	for _, r := range newGraph.Relations() {
		if found := previousGraph.HasRelation(r); !found {
			bulk.UpsertRelation(r)
		}
	}

	// Remove dead assets
	for _, a := range previousGraph.Assets() {
		if found := newGraph.HasAsset(a); !found {
			bulk.RemoveAsset(a)
		}
	}

	// Remove dead relations
	for _, r := range previousGraph.Relations() {
		if found := newGraph.HasRelation(r); !found {
			bulk.RemoveRelation(r)
		}
	}
	return bulk
}
