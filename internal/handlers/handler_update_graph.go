package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/clems4ever/go-graphkb/internal/client"
	"github.com/clems4ever/go-graphkb/internal/knowledge"
	"github.com/clems4ever/go-graphkb/internal/metrics"
	"github.com/clems4ever/go-graphkb/internal/sources"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sync/semaphore"
)

var updatesSemaphore = semaphore.NewWeighted(int64(1))

// PostGraphUpdates POST endpoint updating a graph
func PostGraphUpdates(registry sources.Registry, graphUpdater *knowledge.GraphUpdater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ok, source, err := IsTokenValid(registry, r)
		if err != nil {
			metrics.GraphUpdateEnqueuingRequestsFailedCounter.
				With(
					prometheus.Labels{
						"source":      "",
						"status_code": strconv.FormatInt(http.StatusInternalServerError, 10)}).
				Inc()
			ReplyWithInternalError(w, err)
			return
		}

		if !ok {
			metrics.GraphUpdateEnqueuingRequestsFailedCounter.
				With(prometheus.Labels{
					"source":      source,
					"status_code": strconv.FormatInt(http.StatusUnauthorized, 10)}).
				Inc()
			ReplyWithUnauthorized(w)
			return
		}

		metrics.GraphUpdateEnqueuingRequestsReceivedCounter.
			With(prometheus.Labels{"source": source}).
			Inc()

		requestBody := client.GraphUpdateRequestBody{}
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			ReplyWithInternalError(w, err)
			return
		}

		{
			ok = updatesSemaphore.TryAcquire(1)
			if !ok {
				ReplyWithTooManyRequests(w)
				return
			}
			defer updatesSemaphore.Release(1)

			// TODO(c.michaud): verify compatibility of the schema with graph updates
			graphUpdater.Update(knowledge.SourceSubGraphUpdates{
				Updates: *requestBody.Updates,
				Schema:  requestBody.Schema,
				Source:  source,
			})

		}

		_, err = bytes.NewBufferString("Graph has been received and will be processed soon").WriteTo(w)
		if err != nil {
			metrics.GraphUpdateEnqueuingRequestsFailedCounter.
				With(prometheus.Labels{
					"source":      source,
					"status_code": strconv.FormatInt(http.StatusUnauthorized, 10)}).
				Inc()
			ReplyWithInternalError(w, err)
			return
		}

		metrics.GraphUpdateEnqueuingRequestsSucceededCounter.
			With(prometheus.Labels{"source": source}).
			Inc()
	}
}
