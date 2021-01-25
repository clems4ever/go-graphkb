package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/clems4ever/go-graphkb/internal/handlers"
	"github.com/clems4ever/go-graphkb/internal/history"
	"github.com/clems4ever/go-graphkb/internal/knowledge"
	"github.com/clems4ever/go-graphkb/internal/metrics"
	"github.com/clems4ever/go-graphkb/internal/schema"
	"github.com/clems4ever/go-graphkb/internal/sources"
	"github.com/clems4ever/go-graphkb/internal/utils"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/sync/semaphore"

	auth "github.com/abbot/go-http-auth"

	"github.com/gorilla/mux"
	"github.com/spf13/viper"
)

func getSourceGraph(registry sources.Registry, db schema.Persistor) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sources := []string{}
		availableSources := []string{}

		sourceToToken, err := registry.ListSources(r.Context())
		if err != nil {
			handlers.ReplyWithInternalError(w, err)
			return
		}
		for k := range sourceToToken {
			availableSources = append(availableSources, k)
		}

		sourcesParams, ok := r.URL.Query()["sources"]
		if ok {
			if sourcesParams[0] != "" {
				for _, s := range sourcesParams {
					sources = append(sources, strings.Split(s, ",")...)
				}
			}
		} else {
			sources = availableSources
		}

		sg := schema.NewSchemaGraph()
		for _, sname := range sources {
			if !utils.IsStringInSlice(sname, availableSources) {
				w.WriteHeader(http.StatusBadRequest)
				fmt.Printf("Source %s is not available", sname)
				return
			}
			g, err := db.LoadSchema(context.Background(), sname)
			if err != nil {
				handlers.ReplyWithInternalError(w, err)
				return
			}
			sg.Merge(g)
		}
		handlers.ReplyWithSourceGraph(w, &sg)
	}
}

func listSources(registry sources.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sourcesToTokens, err := registry.ListSources(r.Context())
		if err != nil {
			handlers.ReplyWithInternalError(w, err)
			return
		}

		sources := []string{}
		for k := range sourcesToTokens {
			sources = append(sources, k)
		}

		err = json.NewEncoder(w).Encode(sources)
		if err != nil {
			handlers.ReplyWithInternalError(w, err)
		}
	}
}

func getDatabaseDetails(database knowledge.GraphDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type DatabaseDetailsResponse struct {
			AssetsCount    int64 `json:"assets_count"`
			RelationsCount int64 `json:"relations_count"`
		}

		assetsCount, err := database.CountAssets()
		if err != nil {
			handlers.ReplyWithInternalError(w, err)
			return
		}

		relationsCount, err := database.CountRelations()
		if err != nil {
			handlers.ReplyWithInternalError(w, err)
			return
		}

		response := DatabaseDetailsResponse{}
		response.AssetsCount = assetsCount
		response.RelationsCount = relationsCount
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

func postQuery(database knowledge.GraphDB, queryHistorizer history.Historizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type QueryRequestBody struct {
			Query string `json:"q"`
		}

		type ColumnType struct {
			Name string `json:"name"`
			Type string `json:"type"`
		}

		type QueryResponseBody struct {
			Items           [][]interface{} `json:"items"`
			Columns         []ColumnType    `json:"columns"`
			ExecutionTimeMs time.Duration   `json:"execution_time_ms"`
		}

		requestBody := QueryRequestBody{}
		err := json.NewDecoder(r.Body).Decode(&requestBody)
		if err != nil {
			handlers.ReplyWithInternalError(w, err)
			return
		}

		if requestBody.Query == "" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Println("Empty query parameter")
			_, err = w.Write([]byte("Empty query parameter"))
			if err != nil {
				handlers.ReplyWithInternalError(w, err)
			}
			return
		}

		querier := knowledge.NewQuerier(database, queryHistorizer)
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		res, err := querier.Query(ctx, requestBody.Query)
		if err != nil {
			handlers.ReplyWithInternalError(w, err)
			return
		}
		defer res.Cursor.Close()

		columns := make([]ColumnType, 0)
		for _, p := range res.Projections {
			var colType string
			switch p.ExpressionType {
			case knowledge.NodeExprType:
				colType = "asset"
			case knowledge.EdgeExprType:
				colType = "relation"
			default:
				colType = "property"
			}
			columns = append(columns, ColumnType{
				Name: p.Alias,
				Type: colType,
			})
		}

		items := make([][]interface{}, 0)
		for res.Cursor.HasMore() {
			var d interface{}
			err := res.Cursor.Read(context.Background(), &d)
			if err != nil {
				handlers.ReplyWithInternalError(w, err)
				return
			}

			dCols := d.([]interface{})

			rowDocs := make([]interface{}, 0)

			for _, x := range dCols {
				switch v := x.(type) {
				case knowledge.AssetWithID:
					rowDocs = append(rowDocs, v)
				case knowledge.RelationWithID:
					rowDocs = append(rowDocs, v)
				default:
					rowDocs = append(rowDocs, v)
				}
			}
			items = append(items, rowDocs)
		}

		response := QueryResponseBody{
			Items:           items,
			Columns:         columns,
			ExecutionTimeMs: res.Statistics.Execution / time.Millisecond,
		}

		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			handlers.ReplyWithInternalError(w, err)
		}
	}
}

func getGraphRead(registry sources.Registry, graphDB knowledge.GraphDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ok, source, err := handlers.IsTokenValid(registry, r)
		if err != nil {
			handlers.ReplyWithInternalError(w, err)
			return
		}

		if !ok {
			handlers.ReplyWithUnauthorized(w)
			return
		}

		g := knowledge.NewGraph()
		if err := graphDB.ReadGraph(source, g); err != nil {
			handlers.ReplyWithInternalError(w, err)
			return
		}

		gJSON, err := json.Marshal(g)
		if err != nil {
			handlers.ReplyWithInternalError(w, err)
			return
		}

		if _, err := w.Write(gJSON); err != nil {
			handlers.ReplyWithInternalError(w, err)
		}
	}
}

func flushDatabase(graphDB knowledge.GraphDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := graphDB.FlushAll(); err != nil {
			handlers.ReplyWithInternalError(w, err)
			return
		}

		if err := graphDB.InitializeSchema(); err != nil {
			handlers.ReplyWithInternalError(w, err)
			return
		}
	}
}

// Secret is the secret provider function for basic auth
func Secret(user, realm string) string {
	if user == "admin" {
		return viper.GetString("password")
	}
	return ""
}

// StartServer start the web server
func StartServer(listenInterface string,
	database knowledge.GraphDB,
	schemaPersistor schema.Persistor,
	sourcesRegistry sources.Registry,
	queryHistorizer history.Historizer,
	concurrency int64) {

	r := mux.NewRouter()

	graphUpdater := knowledge.NewGraphUpdater(database, schemaPersistor)

	listSourcesHandler := listSources(sourcesRegistry)
	getSourceGraphHandler := getSourceGraph(sourcesRegistry, schemaPersistor)
	getDatabaseDetailsHandler := getDatabaseDetails(database)
	postQueryHandler := postQuery(database, queryHistorizer)
	flushDatabaseHandler := flushDatabase(database)

	if viper.GetString("password") != "" {
		authenticator := auth.NewBasicAuthenticator("example.com", Secret)

		AuthMiddleware := func(h http.HandlerFunc) http.HandlerFunc {
			return authenticator.Wrap(func(w http.ResponseWriter, ar *auth.AuthenticatedRequest) {
				h.ServeHTTP(w, &ar.Request)
			})
		}

		listSourcesHandler = AuthMiddleware(listSourcesHandler)
		getSourceGraphHandler = AuthMiddleware(getSourceGraphHandler)
		getDatabaseDetailsHandler = AuthMiddleware(getDatabaseDetailsHandler)
		postQueryHandler = AuthMiddleware(postQueryHandler)
		flushDatabaseHandler = AuthMiddleware(flushDatabaseHandler)
	}

	r.HandleFunc("/api/sources", listSourcesHandler).Methods("GET")
	r.HandleFunc("/api/schema", getSourceGraphHandler).Methods("GET")
	r.HandleFunc("/api/database", getDatabaseDetailsHandler).Methods("GET")

	r.HandleFunc("/api/admin/flush", flushDatabaseHandler).Methods("POST")

	r.Handle("/metrics", promhttp.Handler())

	r.HandleFunc("/api/graph/read", getGraphRead(sourcesRegistry, database)).Methods("GET")

	sem := semaphore.NewWeighted(concurrency)

	r.HandleFunc("/api/graph/schema", handlers.PutSchema(sourcesRegistry, graphUpdater, sem)).Methods("PUT")
	r.HandleFunc("/api/graph/asset", handlers.PutAsset(sourcesRegistry, graphUpdater, sem)).Methods("PUT")
	r.HandleFunc("/api/graph/relation", handlers.PutRelation(sourcesRegistry, graphUpdater, sem)).Methods("PUT")
	r.HandleFunc("/api/graph/asset", handlers.DeleteAsset(sourcesRegistry, graphUpdater, sem)).Methods("DELETE")
	r.HandleFunc("/api/graph/relation", handlers.DeleteRelation(sourcesRegistry, graphUpdater, sem)).Methods("DELETE")

	r.HandleFunc("/api/query", postQueryHandler).Methods("POST")
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./web/build/")))

	metrics.StartTimeGauge.Set(float64(time.Now().Unix()))

	var err error
	if viper.GetString("server_tls_cert") != "" {
		fmt.Printf("Listening on %s with TLS enabled, the connection is secure [concurrency=%d]\n", listenInterface, concurrency)
		err = http.ListenAndServeTLS(listenInterface, viper.GetString("server_tls_cert"),
			viper.GetString("server_tls_key"), r)
	} else {
		fmt.Printf("[WARNING] Listening on %s with TLS disabled. Use `server_tls_cert` option to setup a certificate [concurrency=%d]\n",
			listenInterface, concurrency)
		err = http.ListenAndServe(listenInterface, r)
	}
	if err != nil {
		log.Fatal(err)
	}
}
