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

func (sl *GraphUpdater) doUpdate(updates SourceSubGraphUpdates) error {
	previousSchema, err := sl.schemaPersistor.LoadSchema(context.Background(), updates.Source)
	if err != nil {
		fmt.Printf("[ERROR] Unable to read schema from DB: %v.\n", err)
		fmt.Println("[WARNING] The graph has not been updated.")
		return err
	}

	schemaEqual := previousSchema.Equal(updates.Schema)

	if !schemaEqual {
		fmt.Println("The schema needs an update")
		if err := sl.schemaPersistor.SaveSchema(context.Background(), updates.Source, updates.Schema); err != nil {
			fmt.Printf("[ERROR] Unable to write schema in DB: %v.\n", err)
			fmt.Println("[WARNING] The graph has not been updated.")
			return err
		}
	}
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
