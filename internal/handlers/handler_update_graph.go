package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/clems4ever/go-graphkb/internal/client"
	"github.com/clems4ever/go-graphkb/internal/knowledge"
	"github.com/clems4ever/go-graphkb/internal/sources"
	"golang.org/x/sync/semaphore"
)

var updatesSemaphore = semaphore.NewWeighted(int64(32))

func handleUpdate(registry sources.Registry, fn func(source string, body io.Reader) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ok, source, err := IsTokenValid(registry, r)
		if err != nil {
			ReplyWithInternalError(w, err)
			return
		}

		if !ok {
			ReplyWithUnauthorized(w)
			return
		}

		{
			ok = updatesSemaphore.TryAcquire(1)
			if !ok {
				ReplyWithTooManyRequests(w)
				return
			}
			defer updatesSemaphore.Release(1)

			if err = fn(source, r.Body); err != nil {
				ReplyWithInternalError(w, err)
				return
			}
		}

		_, err = bytes.NewBufferString("Graph has been received and will be processed soon").WriteTo(w)
		if err != nil {
			ReplyWithInternalError(w, err)
			return
		}
	}
}

// PutSchema upsert an asset into the graph of the data source
func PutSchema(registry sources.Registry, graphUpdater *knowledge.GraphUpdater) http.HandlerFunc {
	return handleUpdate(registry, func(source string, body io.Reader) error {
		requestBody := client.PutGraphSchemaRequestBody{}
		if err := json.NewDecoder(body).Decode(&requestBody); err != nil {
			return err
		}

		// TODO(c.michaud): verify compatibility of the schema with graph updates
		graphUpdater.UpdateSchema(source, requestBody.Schema)
		return nil
	})
}

// PutAsset upsert an asset into the graph of the data source
func PutAsset(registry sources.Registry, graphUpdater *knowledge.GraphUpdater) http.HandlerFunc {
	return handleUpdate(registry, func(source string, body io.Reader) error {
		requestBody := client.PutGraphAssetRequestBody{}
		if err := json.NewDecoder(body).Decode(&requestBody); err != nil {
			return err
		}

		// TODO(c.michaud): verify compatibility of the schema with graph updates
		graphUpdater.UpsertAsset(source, requestBody.Asset)
		return nil
	})
}

// PutRelation upsert a relation into the graph of the data source
func PutRelation(registry sources.Registry, graphUpdater *knowledge.GraphUpdater) http.HandlerFunc {
	return handleUpdate(registry, func(source string, body io.Reader) error {
		requestBody := client.PutGraphRelationRequestBody{}
		if err := json.NewDecoder(body).Decode(&requestBody); err != nil {
			return err
		}

		// TODO(c.michaud): verify compatibility of the schema with graph updates
		graphUpdater.UpsertRelation(source, requestBody.Relation)
		return nil
	})
}

// DeleteAsset delete an asset from the graph of the data source
func DeleteAsset(registry sources.Registry, graphUpdater *knowledge.GraphUpdater) http.HandlerFunc {
	return handleUpdate(registry, func(source string, body io.Reader) error {
		requestBody := client.DeleteGraphAssetRequestBody{}
		if err := json.NewDecoder(body).Decode(&requestBody); err != nil {
			return err
		}

		// TODO(c.michaud): verify compatibility of the schema with graph updates
		graphUpdater.RemoveAsset(source, requestBody.Asset)
		return nil
	})
}

// DeleteRelation upsert a relation into the graph of the data source
func DeleteRelation(registry sources.Registry, graphUpdater *knowledge.GraphUpdater) http.HandlerFunc {
	return handleUpdate(registry, func(source string, body io.Reader) error {
		requestBody := client.DeleteGraphRelationRequestBody{}
		if err := json.NewDecoder(body).Decode(&requestBody); err != nil {
			return err
		}

		// TODO(c.michaud): verify compatibility of the schema with graph updates
		graphUpdater.RemoveRelation(source, requestBody.Relation)
		return nil
	})
}
