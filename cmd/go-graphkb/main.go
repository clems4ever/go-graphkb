package main

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/clems4ever/go-graphkb/internal/knowledge"
	"github.com/clems4ever/go-graphkb/internal/server"
	"github.com/clems4ever/go-graphkb/internal/sources"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Sources repository of sources
var Sources []sources.Source

// Database the selected database
var Database knowledge.GraphDB

// ConfigPath string
var ConfigPath string

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	Sources = []sources.Source{}

	for _, s := range Sources {
		sources.Registry.Add(s)
		g := s.Graph()

		for _, a := range g.Assets() {
			knowledge.SchemaRegistrySingleton.AddAssetType(a)
		}

		for _, r := range g.Relations() {
			knowledge.SchemaRegistrySingleton.AddRelationType(r.Type)
		}
	}
}

func main() {
	rootCmd := &cobra.Command{
		Use: "go-graphkb [opts]",
	}

	startCmd := &cobra.Command{
		Use:  "start",
		Run:  start,
		Args: cobra.MaximumNArgs(1),
	}

	listenCmd := &cobra.Command{
		Use: "listen",
		Run: listen,
	}

	cleanCmd := &cobra.Command{
		Use: "count",
		Run: count,
	}

	countCmd := &cobra.Command{
		Use: "flush",
		Run: flush,
	}

	readCmd := &cobra.Command{
		Use:  "read [source]",
		Run:  read,
		Args: cobra.ExactArgs(1),
	}

	sourceCmd := &cobra.Command{
		Use:  "source [source]",
		Run:  getSource,
		Args: cobra.ExactArgs(1),
	}

	queryCmd := &cobra.Command{
		Use:  "query [query]",
		Run:  queryFunc,
		Args: cobra.ExactArgs(1),
	}

	rootCmd.PersistentFlags().StringVar(&ConfigPath, "config", "config.yml", "Provide the path to the configuration file (required)")

	cobra.OnInitialize(onInit)

	rootCmd.AddCommand(startCmd, cleanCmd, sourceCmd, listenCmd, countCmd, readCmd, queryCmd)
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func onInit() {
	viper.SetConfigFile(ConfigPath)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("Cannot read configuration file from %s", ConfigPath))
	}

	fmt.Println("Using config file:", viper.ConfigFileUsed())

	dbName := viper.GetString("mariadb.database")
	if dbName == "" {
		log.Fatal("Please provide database_name option in your configuration file")
	}
	Database = knowledge.NewMariaDB(
		viper.GetString("mariadb.username"),
		viper.GetString("mariadb.password"),
		viper.GetString("mariadb.host"),
		dbName)
}

func start(cmd *cobra.Command, args []string) {
	var wg sync.WaitGroup
	wg.Add(1)

	eventBus := make(chan knowledge.SourceSubGraphUpdates)
	listener := knowledge.NewSourceListener(Database)

	if err := Database.InitializeSchema(); err != nil {
		log.Fatal(err)
	}

	// Start kb listening
	go listener.Listen(eventBus)

	var selectedSources []sources.Source

	// if argument is provided, we select the source
	if len(args) == 1 {
		for _, s := range Sources {
			if s.Name() == args[0] {
				selectedSources = []sources.Source{s}
				break
			}
		}
		if len(selectedSources) == 0 {
			log.Fatal(fmt.Sprintf("Unable to find source with name %s", args[0]))
		}

	} else {
		selectedSources = Sources
	}

	for _, source := range selectedSources {
		emitter := knowledge.NewGraphEmitter(source.Name(), eventBus, Database)
		err := source.Start(emitter)
		if err != nil {
			fmt.Printf("[ERROR] %s\n", err)
		}
	}

	wg.Wait()
}

func count(cmd *cobra.Command, args []string) {
	countAssets, err := Database.CountAssets()
	if err != nil {
		log.Fatal(err)
	}

	countRelations, err := Database.CountRelations()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%d assets\n%d relations\n", countAssets, countRelations)
}

func flush(cmd *cobra.Command, args []string) {
	if err := Database.FlushAll(); err != nil {
		log.Fatal(err)
	}
}

func getSource(cmd *cobra.Command, args []string) {
	sourceName := args[0]

	selectedSources := []sources.Source{}
	for _, s := range Sources {
		if s.Name() == sourceName {
			selectedSources = append(selectedSources, s)
		}
	}

	assets := make(map[string]bool)
	relations := make(map[string]bool)
	for _, s := range selectedSources {
		g := s.Graph()
		for _, a := range g.Assets() {
			assets[string(a)] = true
		}

		for _, r := range g.Relations() {
			t := fmt.Sprintf("%s_%s_%s", r.FromType, r.Type, r.ToType)
			relations[t] = true
		}
	}

	assetsSlice := make([]string, 0)
	relationsSlice := make([]string, 0)
	for a := range assets {
		assetsSlice = append(assetsSlice, fmt.Sprintf("\t%s", a))
	}
	for r := range relations {
		relationsSlice = append(relationsSlice, fmt.Sprintf("\t%s", r))
	}

	sort.Strings(assetsSlice)
	sort.Strings(relationsSlice)

	fmt.Printf("assets -> \n%s\nrelations -> \n%s\n",
		strings.Join(assetsSlice, "\n"),
		strings.Join(relationsSlice, "\n"))
}

func listen(cmd *cobra.Command, args []string) {
	server.StartServer(Database)
}

func read(cmd *cobra.Command, args []string) {
	g := knowledge.NewGraph()
	err := Database.ReadGraph(args[0], g)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("assets = %d\nrelations = %d\n", len(g.Assets()), len(g.Relations()))
}

func queryFunc(cmd *cobra.Command, args []string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	q := knowledge.NewQuerier(Database)

	r, err := q.Query(ctx, args[0])
	if err != nil {
		log.Fatal(err)
	}

	resultsCount := 0
	for r.Cursor.HasMore() {
		var m interface{}
		err := r.Cursor.Read(context.Background(), &m)
		if err != nil {
			log.Fatal(err)
		}

		doc := m.([]interface{})
		ldoc := make([]string, len(doc))
		for i, d := range doc {
			ldoc[i] = fmt.Sprintf("%v", d)
		}
		fmt.Println(ldoc)
		resultsCount++
	}

	totalTime := r.Statistics.Parsing + r.Statistics.Execution

	fmt.Printf("%d results found in %fms\n", resultsCount, float64(totalTime.Microseconds())/1000.0)
}
