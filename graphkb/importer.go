package graphkb

import (
	"fmt"
	"github.com/clems4ever/go-graphkb/internal/knowledge"
	"github.com/clems4ever/go-graphkb/internal/sources"
	"github.com/spf13/viper"
)

type ImporterOptions struct {
	CacheGraph bool
}

func Start(source sources.Source, options *ImporterOptions) error {
	url := viper.GetString("graphkb.url")
	if url == "" {
		return fmt.Errorf("Please provide graphkb URL in configuration file")
	}

	authToken := viper.GetString("graphkb.auth_token")
	if authToken == "" {
		return fmt.Errorf("Please provide a graphkb auth token to communicate with GraphKB")
	}

	observableSource := sources.NewObservableSource(source)
	api := knowledge.NewGraphAPI(url, authToken)
	graphImporter := knowledge.NewGraphImporter(api)

	if err := observableSource.Start(graphImporter); err != nil {
		return fmt.Errorf("Unable to start importer: %v", err)
	}

	return nil
}
