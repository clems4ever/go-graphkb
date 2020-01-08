package main

import (
	"fmt"
	"log"

	"github.com/clems4ever/go-graphkb/importer"
	"github.com/clems4ever/go-graphkb/internal/sources"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// ConfigPath string
var ConfigPath string

func onInit() {
	viper.SetConfigFile(ConfigPath)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("Cannot read configuration file from %s", ConfigPath))
	}
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	cobra.OnInitialize(onInit)

	rootCmd := &cobra.Command{
		Use: "source-csv [opts]",
		Run: func(cmd *cobra.Command, args []string) {
			if err := importer.Start(sources.NewCSVSource(), nil); err != nil {
				log.Fatal(err)
			}
		},
	}

	rootCmd.PersistentFlags().StringVar(&ConfigPath, "config", "config.yml",
		"Provide the path to the configuration file (required)")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
