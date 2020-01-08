package server

import (
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
	responseJSON, err := sg.ToJSON()
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
		fmt.Println(err)
	}
}

func getSourceGraph(db schema.Persistor) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		availableSourceNames, err := db.ListSources(context.Background())
		if err != nil {
			replyWithInternalError(w, err)
			return
		}
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

func listSources(db schema.Persistor) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sourceNames, err := db.ListSources(r.Context())
		if err != nil {
			replyWithInternalError(w, err)
			return
		}

		err = json.NewEncoder(w).Encode(sourceNames)
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
			_, err = w.Write([]byte("Empty query parameter"))
			if err != nil {
				replyWithInternalError(w, err)
			}
			return
		}

		querier := knowledge.NewQuerier(database)
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		res, err := querier.Query(ctx, requestBody.Query)

		if err != nil {
			replyWithInternalError(w, err)
			return
		}

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

// Secret is the secret provider function for basic auth
func Secret(user, realm string) string {
	if user == "admin" {
		return viper.GetString("password")
	}
	return ""
}

// StartServer start the web server
func StartServer(database knowledge.GraphDB, schemaPersistor schema.Persistor) {
	r := mux.NewRouter()

	listSourcesHandler := listSources(schemaPersistor)
	getSourceGraphHandler := getSourceGraph(schemaPersistor)
	getDatabaseDetailsHandler := getDatabaseDetails(database)
	postQueryHandler := postQuery(database)

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
	}

	r.HandleFunc("/api/sources", listSourcesHandler).Methods("GET")
	r.HandleFunc("/api/schema", getSourceGraphHandler).Methods("GET")
	r.HandleFunc("/api/database", getDatabaseDetailsHandler).Methods("GET")
	r.HandleFunc("/api/query", postQueryHandler).Methods("POST")
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./web/build/")))

	bindInterface := fmt.Sprintf(":%d", viper.GetInt32("port"))
	fmt.Printf("Listening on %s\n", bindInterface)

	err := http.ListenAndServe(bindInterface, r)
	if err != nil {
		log.Fatal(err)
	}
}
