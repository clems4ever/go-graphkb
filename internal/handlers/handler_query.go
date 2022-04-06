package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/clems4ever/go-graphkb/internal/history"
	"github.com/clems4ever/go-graphkb/internal/knowledge"
	"github.com/patrickmn/go-cache"
	"github.com/spf13/viper"
)

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

// PostQuery post endpoint to query the graph
func PostQuery(database knowledge.GraphDB, queryHistorizer history.Historizer, cacheTTL time.Duration) http.HandlerFunc {
	cache := cache.New(cacheTTL, cacheTTL*2)

	return func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			ReplyWithInternalError(w, err)
			return
		}

		cacheKey := string(body)

		var response []byte

		if res, ok := cache.Get(cacheKey); ok {
			response = res.([]byte)
		} else {
			res, err := executeQuery(r.Context(), database, queryHistorizer, body)
			if err != nil {
				ReplyWithInternalError(w, err)
				return
			}
			cache.Set(cacheKey, res, cacheTTL)
			response = res
		}

		_, err = w.Write(response)
		if err != nil {
			ReplyWithInternalError(w, err)
		}
	}
}

func executeQuery(ctx context.Context, database knowledge.GraphDB, queryHistorizer history.Historizer, body []byte) ([]byte, error) {

	requestBody := QueryRequestBody{}
	err := json.Unmarshal(body, &requestBody)
	if err != nil {
		return nil, err
	}

	if requestBody.Query == "" {
		return nil, fmt.Errorf("empty request")
	}

	QueryMaxTime := viper.GetDuration("query_max_time")
	if QueryMaxTime == 0 {
		QueryMaxTime = 30 * time.Second
	}
	querier := knowledge.NewQuerier(database, queryHistorizer)
	ctx, cancel := context.WithTimeout(context.Background(), QueryMaxTime)
	defer cancel()

	res, err := querier.Query(ctx, requestBody.Query)
	if err != nil {
		return nil, err
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
			return nil, err
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
		sourcesByID, err := database.GetAssetSources(ctx, ids)
		if err != nil {
			return nil, err
		}

		for i, row := range items {
			for j, col := range row {
				switch v := col.(type) {
				case knowledge.AssetWithID:
					sources, ok := sourcesByID[v.ID]
					if !ok {
						return nil, fmt.Errorf("Unable to find sources of asset with ID %s", v.ID)
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

		sourcesByID, err = database.GetRelationSources(ctx, ids)
		if err != nil {
			return nil, err
		}

		for i, row := range items {
			for j, col := range row {
				switch v := col.(type) {
				case knowledge.RelationWithID:
					sources, ok := sourcesByID[v.ID]
					if !ok {
						return nil, fmt.Errorf("Unable to find sources of relation with ID %s", v.ID)
					}
					items[i][j] = RelationWithIDAndSources{
						RelationWithID: v,
						Sources:        sources,
					}
				}
			}
		}

	}

	return json.Marshal(QueryResponseBody{
		Items:           items,
		Columns:         columns,
		ExecutionTimeMs: res.Statistics.Execution / time.Millisecond,
	})
}
