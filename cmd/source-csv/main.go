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

// SourceName string
var SourceName string

func onInit() {
	viper.SetConfigFile(ConfigPath)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("Cannot read configuration file from %s", ConfigPath))
	}

	fmt.Println("Using config file:", viper.ConfigFileUsed())
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	cobra.OnInitialize(onInit)

	rootCmd := &cobra.Command{
		Use: "source-csv [opts]",
		Run: func(cmd *cobra.Command, args []string) {
			if err := importer.Start(SourceName, sources.NewCSVSource(), nil); err != nil {
				log.Fatal(err)
			}
		},
	}

	rootCmd.PersistentFlags().StringVar(&ConfigPath, "config", "config.yml", "Provide the path to the configuration file (required)")
	rootCmd.PersistentFlags().StringVar(&SourceName, "source-name", "", "Provide a unique source name")

	if err := cobra.MarkFlagRequired(rootCmd.PersistentFlags(), "source-name"); err != nil {
		log.Fatal(err)
	}

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
