package graphkb

import (
	"fmt"

	"github.com/clems4ever/go-graphkb/internal/knowledge"
	"github.com/clems4ever/go-graphkb/internal/schema"
	"github.com/clems4ever/go-graphkb/internal/sources"
)

// DataSourceOptions options for configuring the data source
type DataSourceOptions struct {
	URL        string
	AuthToken  string
	SkipVerify bool
}

// Start the data source with provided options
func Start(source sources.DataSource, options DataSourceOptions) error {
	if options.URL == "" {
		return fmt.Errorf("Please provide graphkb URL in configuration file")
	}
	if options.AuthToken == "" {
		return fmt.Errorf("Please provide a graphkb auth token to communicate with GraphKB")
	}

	api := knowledge.NewGraphAPI(options.URL, options.AuthToken, options.SkipVerify)
	dataSourceAPI := knowledge.NewDataSource(api)

	if err := source.Start(dataSourceAPI); err != nil {
		return fmt.Errorf("Unable to start data source: %v", err)
	}

	return nil
}

// CreateRelation helper function for creating a relation
func CreateRelation(fromType schema.AssetType, relation, toType schema.AssetType) RelationType {
	return schema.RelationType{
		FromType: fromType,
		Type:     RelationKeyType(relation),
		ToType:   toType,
	}
}

// CreateAsset helper function for creating an asset
func CreateAsset(fromType string) AssetType {
	return schema.AssetType(fromType)
}
