package importer

import (
	"fmt"
	"log"

	"github.com/clems4ever/go-graphkb/internal/database"
	"github.com/clems4ever/go-graphkb/internal/knowledge"
	"github.com/clems4ever/go-graphkb/internal/sources"
	"github.com/spf13/viper"
)

type ImporterOptions struct {
	CacheGraph bool
}

func Start(name string, source sources.Source, options *ImporterOptions) error {
	dbName := viper.GetString("mariadb.database")
	if dbName == "" {
		return fmt.Errorf("Please provide database_name option in your configuration file")
	}
	mariaDatabase := database.NewMariaDB(
		viper.GetString("mariadb.username"),
		viper.GetString("mariadb.password"),
		viper.GetString("mariadb.host"),
		dbName)

	observableSource := sources.NewObservableSource(source)

	eventBus := make(chan knowledge.SourceSubGraphUpdates)
	listener := knowledge.NewSourceListener(mariaDatabase, mariaDatabase)

	closeC := listener.Listen(eventBus)

	if err := mariaDatabase.InitializeSchema(); err != nil {
		log.Fatal(err)
	}

	emitter := knowledge.NewGraphEmitter(name, eventBus, mariaDatabase)
	if err := observableSource.Start(emitter); err != nil {
		return err
	}

	close(eventBus)
	<-closeC

	return nil
}
