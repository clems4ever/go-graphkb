package knowledge

import (
	"context"
	"fmt"

	"github.com/clems4ever/go-graphkb/internal/schema"
)

// SourceSubGraphUpdates represents the updates to perform on a source subgraph
type SourceSubGraphUpdates struct {
	Updates GraphUpdatesBulk
	Schema  schema.SchemaGraph
	Source  string
}

// GraphUpdater represents the updater of graph
type GraphUpdater struct {
	graphDB         GraphDB
	schemaPersistor schema.Persistor
}

// NewGraphUpdater create a new instance of graph updater
func NewGraphUpdater(graphDB GraphDB, schemaPersistor schema.Persistor) *GraphUpdater {
	return &GraphUpdater{graphDB, schemaPersistor}
}

// Augment the graph of the user with "observed" relation from the source to the each asset
func (sl *GraphUpdater) appendObservedRelations(source string, updates *GraphUpdatesBulk) {
	assetsToAdd := []Asset{Asset{Type: "source", Key: source}}
	observedRelationsToRemove := []Relation{}
	observedRelationsToAdd := []Relation{}

	for _, a := range updates.GetAssetUpserts() {
		observedRelationsToAdd = append(observedRelationsToAdd, Relation{
			Type: "observed",
			From: AssetKey(assetsToAdd[0]),
			To:   AssetKey(a),
		})
	}

	for _, a := range updates.GetAssetRemovals() {
		observedRelationsToRemove = append(observedRelationsToRemove, Relation{
			Type: "observed",
			From: AssetKey(assetsToAdd[0]),
			To:   AssetKey(a),
		})
	}

	updates.UpsertAssets(assetsToAdd...)
	updates.UpsertRelations(observedRelationsToAdd...)
	updates.RemoveRelations(observedRelationsToRemove...)
}

// UpdateSchema update the schema for the source with the one provided in the request
func (sl *GraphUpdater) UpdateSchema(source string, sg schema.SchemaGraph) error {
	for _, a := range sg.Assets() {
		sg.AddRelation(schema.AssetType("source"), "observed", a)
	}
	sg.AddAsset("source")

	previousSchema, err := sl.schemaPersistor.LoadSchema(context.Background(), source)
	if err != nil {
		fmt.Printf("[ERROR] Unable to read schema from DB: %v.\n", err)
		fmt.Println("[WARNING] The graph has not been updated.")
		return err
	}

	schemaEqual := previousSchema.Equal(sg)

	if !schemaEqual {
		fmt.Println("The schema needs an update")
		if err := sl.schemaPersistor.SaveSchema(context.Background(), source, sg); err != nil {
			fmt.Printf("[ERROR] Unable to write schema in DB: %v.\n", err)
			fmt.Println("[WARNING] The graph has not been updated.")
			return err
		}
	}
	return nil
}

// UpsertAsset upsert an asset in the graph of the data source
func (sl *GraphUpdater) UpsertAsset(source string, asset Asset) error {
	if err := sl.graphDB.UpsertAsset(source, asset); err != nil {
		return fmt.Errorf("Unable to upsert asset %v from source %s: %v", asset, source, err)
	}
	return nil
}

// UpsertRelation upsert a relation in the graph of the data source
func (sl *GraphUpdater) UpsertRelation(source string, relation Relation) error {
	if err := sl.graphDB.UpsertRelation(source, relation); err != nil {
		return fmt.Errorf("Unable to upsert relation %v from source %s: %v", relation, source, err)
	}
	return nil
}

// RemoveAsset upsert an asset in the graph of the data source
func (sl *GraphUpdater) RemoveAsset(source string, asset Asset) error {
	if err := sl.graphDB.RemoveAsset(source, asset); err != nil {
		return fmt.Errorf("Unable to remove asset %v from source %s: %v", asset, source, err)
	}
	return nil
}

// RemoveRelation upsert a relation in the graph of the data source
func (sl *GraphUpdater) RemoveRelation(source string, relation Relation) error {
	if err := sl.graphDB.RemoveRelation(source, relation); err != nil {
		return fmt.Errorf("Unable to remove relation %v from source %s: %v", relation, source, err)
	}
	return nil
}
