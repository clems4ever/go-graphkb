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

func getSourceGraph(availableSourceNames []string, db schema.Persistor) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sourcesParams, ok := r.URL.Query()["sources"]
		var sourceNames []string

		if ok && len(sourcesParams) > 0 {
			sourceNames = make([]string, 0)
			for _, s := range sourcesParams {
				sourceNames = append(sourceNames, strings.Split(s, ",")...)
			}
		} else {
			sourceNames = availableSourceNames
		}

		sg := schema.NewSchemaGraph()
		for _, sname := range sourceNames {
			if !utils.IsStringInSlice(sname, availableSourceNames) {
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

func listSources(db schema.Persistor, sourceNames []string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := json.NewEncoder(w).Encode(sourceNames)
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

func postQuery(database knowledge.GraphDB) http.HandlerFunc {
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

		querier := knowledge.NewQuerier(database)
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

type GraphSnapshotRequestBody struct {
	Token string `json:"token"`
}

func isTokenValid(r *http.Request) (bool, string) {
	token, ok := r.URL.Query()["token"]

	if !ok || len(token) != 1 {
		return false, ""
	}

	sourceToToken := viper.GetStringMap("sources")

	found := false
	sourceName := ""

	for sn, t := range sourceToToken {
		if v, ok := t.(string); ok && v == token[0] {
			found = true
			sourceName = sn
			break
		}
	}

	if !found {
		return false, ""
	}

	return true, sourceName
}

func getGraphRead(graphDB knowledge.GraphDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ok, source := isTokenValid(r)
		if !ok {
			replyWithUnauthorized(w)
			return
		}

		g := knowledge.NewGraph()
		if err := graphDB.ReadGraph(source, g); err != nil {
			replyWithInternalError(w, err)
			return
		}

		gJson, err := json.Marshal(g)
		if err != nil {
			replyWithInternalError(w, err)
			return
		}

		if _, err := w.Write(gJson); err != nil {
			replyWithInternalError(w, err)
		}
	}
}

func postGraphUpdates(graphUpdatesC chan knowledge.SourceSubGraphUpdates) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ok, source := isTokenValid(r)
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

		_, err := bytes.NewBufferString("Graph has been received and will be processed soon").WriteTo(w)
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
func StartServer(listenInterface string, database knowledge.GraphDB,
	schemaPersistor schema.Persistor,
	graphUpdatesC chan knowledge.SourceSubGraphUpdates) {

	sourcesToToken := viper.GetStringMap("sources")
	sources := []string{}
	for s := range sourcesToToken {
		sources = append(sources, s)
	}

	r := mux.NewRouter()

	listSourcesHandler := listSources(schemaPersistor, sources)
	getSourceGraphHandler := getSourceGraph(sources, schemaPersistor)
	getDatabaseDetailsHandler := getDatabaseDetails(database)
	postQueryHandler := postQuery(database)
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

	r.HandleFunc("/api/graph/read", getGraphRead(database)).Methods("GET")
	r.HandleFunc("/api/graph/update", postGraphUpdates(graphUpdatesC)).Methods("POST")

	r.HandleFunc("/api/query", postQueryHandler).Methods("POST")
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./web/build/")))

	fmt.Printf("Listening on %s\n", listenInterface)

	var err error
	if viper.GetString("tls_cert") != "" {
		fmt.Println("Server is using TLS, the connection is secure")
		err = http.ListenAndServeTLS(listenInterface, viper.GetString("tls_cert"),
			viper.GetString("tls_key"), r)
	} else {
		fmt.Println("[WARNING] Server is NOT using TLS, the connection is not secure")
		err = http.ListenAndServe(listenInterface, r)
	}
	if err != nil {
		log.Fatal(err)
	}
}
