package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/clems4ever/go-graphkb/internal/knowledge"
)

const MaxIds = 20000

func postAssetSources(database knowledge.GraphDB, fetcherFn func(context.Context, []string) (map[string][]string, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type RequestBody struct {
			IDs []string `json:"ids"`
		}

		type ResponseBody struct {
			Results map[string][]string `json:"results"`
		}

		requestBody := RequestBody{}
		err := json.NewDecoder(r.Body).Decode(&requestBody)
		if err != nil {
			ReplyWithInternalError(w, err)
			return
		}

		if len(requestBody.IDs) > MaxIds {
			ReplyWithBadRequest(w, fmt.Errorf("A maximum of %d IDs can be requested in one query", MaxIds))
			return
		}

		idsSet := make(map[string]struct{})
		for _, id := range requestBody.IDs {
			idsSet[id] = struct{}{}
		}

		ids := []string{}
		for k := range idsSet {
			ids = append(ids, k)
		}

		sources, err := fetcherFn(r.Context(), ids)
		if err != nil {
			ReplyWithInternalError(w, err)
			return
		}

		response := ResponseBody{
			Results: sources,
		}

		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			ReplyWithInternalError(w, err)
			return
		}
	}
}

// PostQueryAssetsSources post endpoint to retrieve the sources of a given set of assets
func PostQueryAssetsSources(database knowledge.GraphDB) http.HandlerFunc {
	return postAssetSources(database, database.GetAssetSources)
}

// PostQueryAssetsSources post endpoint to retrieve the sources of a given set of assets
func PostQueryRelationsSources(database knowledge.GraphDB) http.HandlerFunc {
	return postAssetSources(database, database.GetRelationSources)
}
