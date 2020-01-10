package graphkb

import (
	"fmt"

	"github.com/clems4ever/go-graphkb/internal/knowledge"
	"github.com/clems4ever/go-graphkb/internal/schema"
	"github.com/clems4ever/go-graphkb/internal/sources"
)

type ImporterOptions struct {
	URL        string
	AuthToken  string
	SkipVerify bool
}

func Start(source sources.Source, options ImporterOptions) error {
	if options.URL == "" {
		return fmt.Errorf("Please provide graphkb URL in configuration file")
	}
	if options.AuthToken == "" {
		return fmt.Errorf("Please provide a graphkb auth token to communicate with GraphKB")
	}

	api := knowledge.NewGraphAPI(options.URL, options.AuthToken, options.SkipVerify)
	graphImporter := knowledge.NewGraphImporter(api)

	if err := source.Start(graphImporter); err != nil {
		return fmt.Errorf("Unable to start importer: %v", err)
	}

	return nil
}

func CreateRelation(fromType schema.AssetType, relation, toType schema.AssetType) RelationType {
	return schema.RelationType{
		FromType: fromType,
		Type:     RelationKeyType(relation),
		ToType:   toType,
	}
}

func CreateAsset(fromType string) AssetType {
	return schema.AssetType(fromType)
}
