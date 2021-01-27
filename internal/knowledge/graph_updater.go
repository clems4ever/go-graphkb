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

// UpdateSchema update the schema for the source with the one provided in the request
func (sl *GraphUpdater) UpdateSchema(source string, sg schema.SchemaGraph) error {
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

// InsertAssets insert multiple assets in the graph of the data source
func (sl *GraphUpdater) InsertAssets(source string, assets []Asset) error {
	if err := sl.graphDB.InsertAssets(source, assets); err != nil {
		return fmt.Errorf("Unable to insert assets %v from source %s: %v", assets, source, err)
	}
	return nil
}

// InsertRelations insert multiple relations in the graph of the data source
func (sl *GraphUpdater) InsertRelations(source string, relations []Relation) error {
	if err := sl.graphDB.InsertRelations(source, relations); err != nil {
		return fmt.Errorf("Unable to insert relations %v from source %s: %v", relations, source, err)
	}
	return nil
}

// RemoveAssets remove multiple assets from the graph of the data source
func (sl *GraphUpdater) RemoveAssets(source string, assets []Asset) error {
	if err := sl.graphDB.RemoveAssets(source, assets); err != nil {
		return fmt.Errorf("Unable to remove assets %v from source %s: %v", assets, source, err)
	}
	return nil
}

// RemoveRelations remove multiple relations from the graph of the data source
func (sl *GraphUpdater) RemoveRelations(source string, relations []Relation) error {
	if err := sl.graphDB.RemoveRelations(source, relations); err != nil {
		return fmt.Errorf("Unable to remove relation %v from source %s: %v", relations, source, err)
	}
	return nil
}
