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

// SourceListener represents the source listener waiting for source events
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

	for _, a := range updates.AssetUpserts {
		observedRelationsToAdd = append(observedRelationsToAdd, Relation{
			Type: "observed",
			From: AssetKey(assetsToAdd[0]),
			To:   AssetKey(a),
		})
	}

	for _, a := range updates.AssetRemovals {
		observedRelationsToRemove = append(observedRelationsToRemove, Relation{
			Type: "observed",
			From: AssetKey(assetsToAdd[0]),
			To:   AssetKey(a),
		})
	}

	updates.AssetUpserts = append(updates.AssetUpserts, assetsToAdd...)
	updates.RelationUpserts = append(updates.RelationUpserts, observedRelationsToAdd...)
	updates.RelationRemovals = append(updates.RelationRemovals, observedRelationsToRemove...)
}

func (sl *GraphUpdater) updateSchema(source string, sg *schema.SchemaGraph) error {
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

	schemaEqual := previousSchema.Equal(*sg)

	if !schemaEqual {
		fmt.Println("The schema needs an update")
		if err := sl.schemaPersistor.SaveSchema(context.Background(), source, *sg); err != nil {
			fmt.Printf("[ERROR] Unable to write schema in DB: %v.\n", err)
			fmt.Println("[WARNING] The graph has not been updated.")
			return err
		}
	}
	return nil
}

func (sl *GraphUpdater) doUpdate(updates SourceSubGraphUpdates) error {
	if err := sl.updateSchema(updates.Source, &updates.Schema); err != nil {
		return err
	}

	sl.appendObservedRelations(updates.Source, &updates.Updates)

	if err := sl.graphDB.UpdateGraph(updates.Source, &updates.Updates); err != nil {
		fmt.Printf("[ERROR] Unable to write schema in graph DB: %v\n", err)
		return err
	}
	return nil
}

// Listen events coming from the event bus
func (sl *GraphUpdater) Listen(updatesC chan SourceSubGraphUpdates) chan struct{} {
	closeC := make(chan struct{})

	go func() {
		for updates := range updatesC {
			if err := sl.doUpdate(updates); err != nil {
				fmt.Println(err)
			}
		}
		close(closeC)
	}()
	return closeC
}
