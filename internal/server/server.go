package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/clems4ever/go-graphkb/internal/history"
	"github.com/clems4ever/go-graphkb/internal/importers"

	auth "github.com/abbot/go-http-auth"
	"github.com/clems4ever/go-graphkb/internal/knowledge"
	"github.com/clems4ever/go-graphkb/internal/schema"
	"github.com/clems4ever/go-graphkb/internal/utils"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
)

func replyWithSourceGraph(w http.ResponseWriter, sg *schema.SchemaGraph) {
	responseJSON, err := json.Marshal(sg)
	if err != nil {
		replyWithInternalError(w, err)
		return
	}

	if _, err := w.Write(responseJSON); err != nil {
		replyWithInternalError(w, err)
	}
}

func replyWithInternalError(w http.ResponseWriter, err error) {
	fmt.Println(err)
	w.WriteHeader(http.StatusInternalServerError)
	_, werr := w.Write([]byte(err.Error()))
	if werr != nil {
		fmt.Println(werr)
	}
}

func replyWithUnauthorized(w http.ResponseWriter) {
	w.WriteHeader(http.StatusUnauthorized)
	_, werr := w.Write([]byte("Unauthorized"))
	if werr != nil {
		fmt.Println(werr)
	}
}

func getSourceGraph(registry importers.Registry, db schema.Persistor) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		importers := []string{}
		availableImporters := []string{}

		importerToToken, err := registry.ListImporters(r.Context())
		if err != nil {
			replyWithInternalError(w, err)
			return
		}
		for k := range importerToToken {
			availableImporters = append(availableImporters, k)
		}

		sourcesParams, ok := r.URL.Query()["sources"]
		if ok {
			if sourcesParams[0] != "" {
				for _, s := range sourcesParams {
					importers = append(importers, strings.Split(s, ",")...)
				}
			}
		} else {
			importers = availableImporters
		}

		sg := schema.NewSchemaGraph()
		for _, sname := range importers {
			if !utils.IsStringInSlice(sname, availableImporters) {
				w.WriteHeader(http.StatusBadRequest)
				fmt.Printf("Source %s is not available", sname)
				return
			}
			g, err := db.LoadSchema(context.Background(), sname)
			if err != nil {
				replyWithInternalError(w, err)
				return
			}
			sg.Merge(g)
		}
		replyWithSourceGraph(w, &sg)
	}
}

func listImporters(registry importers.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		importersToToken, err := registry.ListImporters(r.Context())
		if err != nil {
			replyWithInternalError(w, err)
			return
		}

		importers := []string{}
		for k := range importersToToken {
			importers = append(importers, k)
		}

		err = json.NewEncoder(w).Encode(importers)
		if err != nil {
			replyWithInternalError(w, err)
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
			replyWithInternalError(w, err)
			return
		}

		relationsCount, err := database.CountRelations()
		if err != nil {
			replyWithInternalError(w, err)
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
			replyWithInternalError(w, err)
			return
		}

		if requestBody.Query == "" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Println("Empty query parameter")
			_, err = w.Write([]byte("Empty query parameter"))
			if err != nil {
				replyWithInternalError(w, err)
			}
			return
		}

		querier := knowledge.NewQuerier(database, queryHistorizer)
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		res, err := querier.Query(ctx, requestBody.Query)
		if err != nil {
			replyWithInternalError(w, err)
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
				replyWithInternalError(w, err)
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
			replyWithInternalError(w, err)
		}
	}
}

func isTokenValid(registry importers.Registry, r *http.Request) (bool, string, error) {
	token, ok := r.URL.Query()["token"]

	if !ok || len(token) != 1 {
		return false, "", fmt.Errorf("Unable to detect token query parameter")
	}

	importerToToken, err := registry.ListImporters(r.Context())

	if err != nil {
		return false, "", err
	}

	for sn, t := range importerToToken {
		if t == token[0] {
			return true, sn, nil
		}
	}

	return false, "", nil
}

func getGraphRead(registry importers.Registry, graphDB knowledge.GraphDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ok, source, err := isTokenValid(registry, r)
		if err != nil {
			replyWithInternalError(w, err)
			return
		}

		if !ok {
			replyWithUnauthorized(w)
			return
		}

		g := knowledge.NewGraph()
		if err := graphDB.ReadGraph(source, g); err != nil {
			replyWithInternalError(w, err)
			return
		}

		gJSON, err := json.Marshal(g)
		if err != nil {
			replyWithInternalError(w, err)
			return
		}

		if _, err := w.Write(gJSON); err != nil {
			replyWithInternalError(w, err)
		}
	}
}

func postGraphUpdates(registry importers.Registry, graphUpdatesC chan knowledge.SourceSubGraphUpdates) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ok, source, err := isTokenValid(registry, r)
		if err != nil {
			replyWithInternalError(w, err)
			return
		}

		if !ok {
			replyWithUnauthorized(w)
			return
		}

		requestBody := knowledge.GraphUpdateRequestBody{}
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			replyWithInternalError(w, err)
			return
		}

		// TODO(c.michaud): verify compatibility of the schema with graph updates

		graphUpdatesC <- knowledge.SourceSubGraphUpdates{
			Updates: *requestBody.Updates,
			Schema:  requestBody.Schema,
			Source:  source,
		}

		_, err = bytes.NewBufferString("Graph has been received and will be processed soon").WriteTo(w)
		if err != nil {
			replyWithInternalError(w, err)
			return
		}
	}
}

func flushDatabase(graphDB knowledge.GraphDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := graphDB.FlushAll(); err != nil {
			replyWithInternalError(w, err)
			return
		}

		if err := graphDB.InitializeSchema(); err != nil {
			replyWithInternalError(w, err)
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
	importersRegistry importers.Registry,
	queryHistorizer history.Historizer,
	graphUpdatesC chan knowledge.SourceSubGraphUpdates) {

	r := mux.NewRouter()

	listImportersHandler := listImporters(importersRegistry)
	getSourceGraphHandler := getSourceGraph(importersRegistry, schemaPersistor)
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

		listImportersHandler = AuthMiddleware(listImportersHandler)
		getSourceGraphHandler = AuthMiddleware(getSourceGraphHandler)
		getDatabaseDetailsHandler = AuthMiddleware(getDatabaseDetailsHandler)
		postQueryHandler = AuthMiddleware(postQueryHandler)
		flushDatabaseHandler = AuthMiddleware(flushDatabaseHandler)
	}

	r.HandleFunc("/api/sources", listImportersHandler).Methods("GET")
	r.HandleFunc("/api/schema", getSourceGraphHandler).Methods("GET")
	r.HandleFunc("/api/database", getDatabaseDetailsHandler).Methods("GET")

	r.HandleFunc("/api/admin/flush", flushDatabaseHandler).Methods("POST")

	r.HandleFunc("/api/graph/read", getGraphRead(importersRegistry, database)).Methods("GET")
	r.HandleFunc("/api/graph/update", postGraphUpdates(importersRegistry, graphUpdatesC)).Methods("POST")

	r.HandleFunc("/api/query", postQueryHandler).Methods("POST")
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./web/build/")))

	var err error
	if viper.GetString("server_tls_cert") != "" {
		fmt.Printf("Listening on %s with TLS enabled, the connection is secure\n", listenInterface)
		err = http.ListenAndServeTLS(listenInterface, viper.GetString("server_tls_cert"),
			viper.GetString("server_tls_key"), r)
	} else {
		fmt.Printf("[WARNING] Listening on %s with TLS disabled. Use `server_tls_cert` option to setup a certificate\n",
			listenInterface)
		err = http.ListenAndServe(listenInterface, r)
	}
	if err != nil {
		log.Fatal(err)
	}
}
