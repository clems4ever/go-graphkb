package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/clems4ever/go-graphkb/internal/client"
	"github.com/clems4ever/go-graphkb/internal/knowledge"
	"github.com/clems4ever/go-graphkb/internal/metrics"
	"github.com/clems4ever/go-graphkb/internal/sources"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sync/semaphore"
)

func handleUpdate(registry sources.Registry, fn func(source string, body io.Reader) error, sem *semaphore.Weighted, operationDescriptor string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ok, source, err := IsTokenValid(registry, r)
		if err != nil {
			ReplyWithInternalError(w, err)
			return
		}

		promLabels := prometheus.Labels{
			"source":    source,
			"operation": operationDescriptor,
		}

		if !ok {
			metrics.GraphUpdateRequestsUnauthorizedCounter.
				With(promLabels).
				Inc()
			ReplyWithUnauthorized(w)
			return
		}

		{
			ok = sem.TryAcquire(1)
			if !ok {
				metrics.GraphUpdateRequestsRateLimitedCounter.
					With(promLabels).
					Inc()
				ReplyWithTooManyRequests(w)
				return
			}
			defer sem.Release(1)

			metrics.GraphUpdateRequestsReceivedCounter.
				With(promLabels).
				Inc()

			if err = fn(source, r.Body); err != nil {
				metrics.GraphUpdateRequestsFailedCounter.
					With(promLabels).
					Inc()
				ReplyWithInternalError(w, err)
				return
			}

			metrics.GraphUpdateRequestsSucceededCounter.
				With(promLabels).
				Inc()

			metrics.LastSuccessfulDatasourceUpdateTimestampGauge.
				With(prometheus.Labels{"source": source}).
				Set(float64(time.Now().Unix()))
		}

		_, err = fmt.Fprint(w, "Graph update has been processed")
		if err != nil {
			ReplyWithInternalError(w, err)
			return
		}
	}
}

// PutSchema upsert an asset into the graph of the data source
func PutSchema(registry sources.Registry, graphUpdater *knowledge.GraphUpdater, sem *semaphore.Weighted) http.HandlerFunc {
	return handleUpdate(registry, func(source string, body io.Reader) error {
		requestBody := client.PutGraphSchemaRequestBody{}
		if err := json.NewDecoder(body).Decode(&requestBody); err != nil {
			return err
		}

		// TODO(c.michaud): verify compatibility of the schema with graph updates
		err := graphUpdater.UpdateSchema(source, requestBody.Schema)
		if err != nil {
			return fmt.Errorf("Unable to update the schema: %v", err)
		}

		labels := prometheus.Labels{"source": source}
		metrics.GraphUpdateSchemaUpdatedCounter.
			With(labels).
			Inc()
		return nil
	}, sem, "update_schema")
}

// PutAssets upsert several assets into the graph of the data source
func PutAssets(registry sources.Registry, graphUpdater *knowledge.GraphUpdater, sem *semaphore.Weighted) http.HandlerFunc {
	return handleUpdate(registry, func(source string, body io.Reader) error {
		requestBody := client.PutGraphAssetRequestBody{}
		if err := json.NewDecoder(body).Decode(&requestBody); err != nil {
			return err
		}

		// TODO(c.michaud): verify compatibility of the schema with graph updates
		err := graphUpdater.InsertAssets(source, requestBody.Assets)
		if err != nil {
			return fmt.Errorf("Unable to insert assets: %v", err)
		}
		labels := prometheus.Labels{"source": source}
		metrics.GraphUpdateAssetsInsertedCounter.
			With(labels).
			Add(float64(len(requestBody.Assets)))

		return nil
	}, sem, "insert_assets")
}

// PutRelations upsert multiple relations into the graph of the data source
func PutRelations(registry sources.Registry, graphUpdater *knowledge.GraphUpdater, sem *semaphore.Weighted) http.HandlerFunc {
	return handleUpdate(registry, func(source string, body io.Reader) error {
		requestBody := client.PutGraphRelationRequestBody{}
		if err := json.NewDecoder(body).Decode(&requestBody); err != nil {
			return err
		}

		// TODO(c.michaud): verify compatibility of the schema with graph updates
		err := graphUpdater.InsertRelations(source, requestBody.Relations)
		if err != nil {
			return fmt.Errorf("Unable to insert relation: %v", err)
		}

		labels := prometheus.Labels{"source": source}
		metrics.GraphUpdateRelationsInsertedCounter.
			With(labels).
			Add(float64(len(requestBody.Relations)))
		return nil
	}, sem, "insert_relations")
}

// DeleteAssets delete multiple assets from the graph of the data source
func DeleteAssets(registry sources.Registry, graphUpdater *knowledge.GraphUpdater, sem *semaphore.Weighted) http.HandlerFunc {
	return handleUpdate(registry, func(source string, body io.Reader) error {
		requestBody := client.DeleteGraphAssetRequestBody{}
		if err := json.NewDecoder(body).Decode(&requestBody); err != nil {
			return err
		}

		// TODO(c.michaud): verify compatibility of the schema with graph updates
		err := graphUpdater.RemoveAssets(source, requestBody.Assets)
		if err != nil {
			return fmt.Errorf("Unable to remove assets: %v", err)
		}

		labels := prometheus.Labels{"source": source}
		metrics.GraphUpdateAssetsDeletedCounter.
			With(labels).
			Add(float64(len(requestBody.Assets)))
		return nil
	}, sem, "delete_assets")
}

// DeleteRelations remove multiple relations from the graph of the data source
func DeleteRelations(registry sources.Registry, graphUpdater *knowledge.GraphUpdater, sem *semaphore.Weighted) http.HandlerFunc {
	return handleUpdate(registry, func(source string, body io.Reader) error {
		requestBody := client.DeleteGraphRelationRequestBody{}
		if err := json.NewDecoder(body).Decode(&requestBody); err != nil {
			return err
		}

		// TODO(c.michaud): verify compatibility of the schema with graph updates
		err := graphUpdater.RemoveRelations(source, requestBody.Relations)
		if err != nil {
			return fmt.Errorf("Unable to remove relation: %v", err)
		}

		labels := prometheus.Labels{"source": source}
		metrics.GraphUpdateRelationsDeletedCounter.
			With(labels).
			Add(float64(len(requestBody.Relations)))
		return nil
	}, sem, "delete_relations")
}
