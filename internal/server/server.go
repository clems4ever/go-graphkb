package server

import (
	"context"
	"encoding/json"
	"fmt"
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
	"github.com/sirupsen/logrus"
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
				handlers.ReplyWithBadRequest(w, fmt.Errorf("Source %s does not exist", sname))
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

		assetsCount, err := database.CountAssets(r.Context())
		if err != nil {
			handlers.ReplyWithInternalError(w, err)
			return
		}

		relationsCount, err := database.CountRelations(r.Context())
		if err != nil {
			handlers.ReplyWithInternalError(w, err)
			return
		}

		response := DatabaseDetailsResponse{}
		response.AssetsCount = assetsCount
		response.RelationsCount = relationsCount
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			handlers.ReplyWithInternalError(w, err)
			return
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
		if err := graphDB.ReadGraph(r.Context(), source, g); err != nil {
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
		if err := graphDB.FlushAll(r.Context()); err != nil {
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
	postQueryHandler := handlers.PostQuery(database, queryHistorizer)
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

	r.PathPrefix("/debug/pprof/").Handler(http.DefaultServeMux)
	r.HandleFunc("/api/sources", listSourcesHandler).Methods("GET")
	r.HandleFunc("/api/schema", getSourceGraphHandler).Methods("GET")
	r.HandleFunc("/api/database", getDatabaseDetailsHandler).Methods("GET")

	r.HandleFunc("/api/admin/flush", flushDatabaseHandler).Methods("POST")

	r.Handle("/metrics", promhttp.Handler())

	r.HandleFunc("/api/graph/read", getGraphRead(sourcesRegistry, database)).Methods("GET")

	sem := semaphore.NewWeighted(concurrency)

	r.HandleFunc("/api/graph/schema", handlers.PutSchema(sourcesRegistry, graphUpdater, sem)).Methods("PUT")
	r.HandleFunc("/api/graph/assets", handlers.PutAssets(sourcesRegistry, graphUpdater, sem)).Methods("PUT")
	r.HandleFunc("/api/graph/assets", handlers.DeleteAssets(sourcesRegistry, graphUpdater, sem)).Methods("DELETE")
	r.HandleFunc("/api/graph/relations", handlers.PutRelations(sourcesRegistry, graphUpdater, sem)).Methods("PUT")
	r.HandleFunc("/api/graph/relations", handlers.DeleteRelations(sourcesRegistry, graphUpdater, sem)).Methods("DELETE")

	r.HandleFunc("/api/query", postQueryHandler).Methods("POST")
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./web/build/")))

	metrics.StartTimeGauge.Set(float64(time.Now().Unix()))

	var err error
	if viper.GetString("server_tls_cert") != "" {
		logrus.Infof("Listening on %s with TLS enabled, the connection is secure [concurrency=%d]\n", listenInterface, concurrency)
		err = http.ListenAndServeTLS(listenInterface, viper.GetString("server_tls_cert"),
			viper.GetString("server_tls_key"), r)
	} else {
		logrus.Warnf("Listening on %s with TLS disabled. Use `server_tls_cert` option to setup a certificate [concurrency=%d]\n",
			listenInterface, concurrency)
		err = http.ListenAndServe(listenInterface, r)
	}
	if err != nil {
		logrus.Fatal(err)
	}
}
