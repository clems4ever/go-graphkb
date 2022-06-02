package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// StartTimeGauge reports the start time of the instance
var StartTimeGauge = promauto.NewGauge(prometheus.GaugeOpts{
	Name: "go_graphkb_start_timestamp_gauge",
	Help: "The timestamp of the time when the app started",
})

// GraphQueryTimeExecution reports the time execution in ms for queries.
var GraphQueryTimeExecution = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "go_graphkb_graph_query_execution_time_ms",
	Help:    "The time execution in ms of queries.",
	Buckets: prometheus.ExponentialBucketsRange(0.1, 1000*60*30, 15),
}, []string{})

// ********************* GRAPH METRICS ******************

// GraphAssetsTotalGauge reports the number of nodes in the graph of a given source
var GraphAssetsTotalGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
	Name: "go_graphkb_graph_assets_total_gauge",
	Help: "The number of nodes in the graph for a given source",
}, []string{"source"})

// GraphRelationsTotalGauge reports the number of nodes in the graph of a given source
var GraphRelationsTotalGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
	Name: "go_graphkb_graph_relations_total_gauge",
	Help: "The number of edges in the graph for a given source",
}, []string{"source"})

// GraphAssetsAggregatedGauge reports the number of nodes in the graph with the various sources merged
var GraphAssetsAggregatedGauge = promauto.NewGauge(prometheus.GaugeOpts{
	Name: "go_graphkb_graph_assets_aggregated_gauge",
	Help: "The number of nodes in the graph with the various datasource graphs merged",
})

// GraphRelationsAggregatedGauge reports the number of nodes in the graph with the various sources merged
var GraphRelationsAggregatedGauge = promauto.NewGauge(prometheus.GaugeOpts{
	Name: "go_graphkb_graph_relations_aggregated_gauge",
	Help: "The number of edges in the graph with the various datasource graphs merged",
})

// ********************* GRAPH UPDATE REQUESTS ******************

// GraphUpdateRequestsReceivedCounter reports the number of authorized and not rate limited updates requests received by the webserver
var GraphUpdateRequestsReceivedCounter = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "go_graphkb_graph_update_requests_received_counter",
	Help: "The number of graph updates (insertion or removal of assets or relations) received. A request is considered received if authorization is valid and the request has not been rate limited",
}, []string{"source", "operation"})

// GraphUpdateRequestsRateLimitedCounter reports the number of unauthorized updates requests received by the webserver
var GraphUpdateRequestsRateLimitedCounter = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "go_graphkb_graph_update_requests_rate_limited_counter",
	Help: "The number of graph updates which were rate limited",
}, []string{"source", "operation"})

// GraphUpdateRequestsUnauthorizedCounter reports the number of unauthorized updates requests received by the webserver
var GraphUpdateRequestsUnauthorizedCounter = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "go_graphkb_graph_update_requests_unauthorized_counter",
	Help: "The number of graph updates which were unauthorized",
}, []string{"source", "operation"})

// GraphUpdateRequestsFailedCounter reports the number of failed update requests
var GraphUpdateRequestsFailedCounter = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "go_graphkb_graph_update_requests_failed_counter",
	Help: "The number of failed graph update",
}, []string{"source", "operation"})

// GraphUpdateRequestsSucceededCounter reports the number of successful update requests
var GraphUpdateRequestsSucceededCounter = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "go_graphkb_graph_update_requests_succeeded_counter",
	Help: "The number of succeeded graph update",
}, []string{"source", "operation"})

// ********************* GRAPH UPDATE COUNTERS ******************

// GraphUpdateSchemaUpdatedCounter reports the number of schema updated since the start of the process
var GraphUpdateSchemaUpdatedCounter = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "go_graphkb_graph_update_schema_updated_counter",
	Help: "The number of assets inserted since the start of the process",
}, []string{"source"})

// GraphUpdateAssetsInsertedCounter reports the number of assets inserted since the start of the process
var GraphUpdateAssetsInsertedCounter = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "go_graphkb_graph_update_assets_inserted_counter",
	Help: "The number of assets inserted since the start of the process",
}, []string{"source"})

// GraphUpdateRelationsInsertedCounter reports the number of relations inserted since the start of the process
var GraphUpdateRelationsInsertedCounter = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "go_graphkb_graph_update_relations_inserted_counter",
	Help: "The number of relations inserted since the start of the process",
}, []string{"source"})

// GraphUpdateAssetsDeletedCounter reports the number of assets inserted since the start of the process
var GraphUpdateAssetsDeletedCounter = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "go_graphkb_graph_update_assets_deleted_counter",
	Help: "The number of assets deleted since the start of the process",
}, []string{"source"})

// GraphUpdateRelationsDeletedCounter reports the number of relations deleted since the start of the process
var GraphUpdateRelationsDeletedCounter = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "go_graphkb_graph_update_relations_deleted_counter",
	Help: "The number of relations deleted since the start of the process",
}, []string{"source"})

// ********************* SOURCES ******************

// LastSuccessfulDatasourceUpdateTimestampGauge reports the timestamp of the last successful update operation for a given source
var LastSuccessfulDatasourceUpdateTimestampGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
	Name: "go_graphkb_last_successful_datasource_update_timestamp_gauge",
	Help: "The timestamp of the last successful operation of the data source",
}, []string{"source"})

// DatabaseMetricsGauge reports some technical metrics about the database (database implementation specific)
var DatabaseMetricsGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
	Name: "go_graphkb_database_metrics_gauge",
	Help: "Metrics gathered from the database",
}, []string{"name"})
