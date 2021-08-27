package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/clems4ever/go-graphkb/internal/history"
	"github.com/clems4ever/go-graphkb/internal/knowledge"
	"github.com/spf13/viper"
)

// PostQuery post endpoint to query the graph
func PostQuery(database knowledge.GraphDB, queryHistorizer history.Historizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type QueryRequestBody struct {
			Query          string `json:"q"`
			IncludeSources bool   `json:"include_sources"`
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

		type AssetWithIDAndSources struct {
			Sources []string `json:"sources,omitempty"`
			knowledge.AssetWithID
		}

		type RelationWithIDAndSources struct {
			Sources []string `json:"sources,omitempty"`
			knowledge.RelationWithID
		}

		requestBody := QueryRequestBody{}
		err := json.NewDecoder(r.Body).Decode(&requestBody)
		if err != nil {
			ReplyWithInternalError(w, err)
			return
		}

		if requestBody.Query == "" {
			w.WriteHeader(http.StatusBadRequest)
			_, err = w.Write([]byte("Empty 'query' parameter"))
			if err != nil {
				ReplyWithInternalError(w, err)
			}
			return
		}

		query_timewait := viper.GetDuration("query_timeout")
		if query_timewait != 0 {
			query_timewait = 30
		}
		querier := knowledge.NewQuerier(database, queryHistorizer)
		ctx, cancel := context.WithTimeout(context.Background(), query_timewait*time.Second)
		defer cancel()

		res, err := querier.Query(ctx, requestBody.Query)
		if err != nil {
			ReplyWithInternalError(w, err)
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

		assetIDs := make(map[string]struct{})
		relationIDs := make(map[string]struct{})

		items := make([][]interface{}, 0)
		for res.Cursor.HasMore() {
			var d interface{}
			err := res.Cursor.Read(context.Background(), &d)
			if err != nil {
				ReplyWithInternalError(w, err)
				return
			}

			dCols := d.([]interface{})

			rowDocs := make([]interface{}, 0)

			for _, x := range dCols {
				switch v := x.(type) {
				case knowledge.AssetWithID:
					rowDocs = append(rowDocs, v)
					if requestBody.IncludeSources {
						assetIDs[v.ID] = struct{}{}
					}
				case knowledge.RelationWithID:
					rowDocs = append(rowDocs, v)
					if requestBody.IncludeSources {
						relationIDs[v.ID] = struct{}{}
					}
				default:
					rowDocs = append(rowDocs, v)
				}
			}
			items = append(items, rowDocs)
		}

		if requestBody.IncludeSources {
			ids := []string{}
			for k := range assetIDs {
				ids = append(ids, k)
			}
			sourcesByID, err := database.GetAssetSources(r.Context(), ids)
			if err != nil {
				ReplyWithInternalError(w, err)
				return
			}

			for i, row := range items {
				for j, col := range row {
					switch v := col.(type) {
					case knowledge.AssetWithID:
						sources, ok := sourcesByID[v.ID]
						if !ok {
							ReplyWithInternalError(w, fmt.Errorf("Unable to find sources of asset with ID %s", v.ID))
							return
						}
						items[i][j] = AssetWithIDAndSources{
							AssetWithID: v,
							Sources:     sources,
						}
					}
				}
			}

			ids = []string{}
			for k := range relationIDs {
				ids = append(ids, k)
			}

			sourcesByID, err = database.GetRelationSources(r.Context(), ids)
			if err != nil {
				ReplyWithInternalError(w, err)
				return
			}

			for i, row := range items {
				for j, col := range row {
					switch v := col.(type) {
					case knowledge.RelationWithID:
						sources, ok := sourcesByID[v.ID]
						if !ok {
							ReplyWithInternalError(w, fmt.Errorf("Unable to find sources of relation with ID %s", v.ID))
							return
						}
						items[i][j] = RelationWithIDAndSources{
							RelationWithID: v,
							Sources:        sources,
						}
					}
				}
			}

		}

		response := QueryResponseBody{
			Items:           items,
			Columns:         columns,
			ExecutionTimeMs: res.Statistics.Execution / time.Millisecond,
		}

		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			ReplyWithInternalError(w, err)
		}
	}
}
