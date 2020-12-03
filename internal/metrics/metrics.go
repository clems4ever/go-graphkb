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

// ********************* GRAPH UPDATE ENQUEUING ******************

// GraphUpdateEnqueuingRequestsReceivedCounter reports the number of authorized updates request received by the webserver
var GraphUpdateEnqueuingRequestsReceivedCounter = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "go_graphkb_graph_update_enqueuing_requests_received_counter",
	Help: "The number of graph update enqueuing requests received by the web server but not yet processed",
}, []string{"source"})

// GraphUpdateEnqueuingRequestsFailedCounter reports the number of failed update requests
var GraphUpdateEnqueuingRequestsFailedCounter = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "go_graphkb_graph_update_enqueuing_requests_failed_counter",
	Help: "The number of failed graph update enqueuing requests",
}, []string{"source", "status_code"})

// GraphUpdateEnqueuingRequestsSucceededCounter reports the number of successful enqueuing requests
var GraphUpdateEnqueuingRequestsSucceededCounter = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "go_graphkb_graph_update_enqueuing_requests_succeeded_counter",
	Help: "The number of successful graph update enqueuing requests",
}, []string{"source"})

// ********************* GRAPH PROCESSING *************************

// GraphUpdatesProcessingRequestedCounter reports the number of updates to be processed and imported into DB.
var GraphUpdatesProcessingRequestedCounter = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "go_graphkb_graph_updates_processing_requested_counter",
	Help: "The number of update requests to be processed and imported into DB",
}, []string{"source"})

// GraphUpdatesProcessingSucceededCounter reports the number of updates which have successfully been processed
var GraphUpdatesProcessingSucceededCounter = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "go_graphkb_graph_updates_processing_succeeded_counter",
	Help: "The number of update requests processed successfully",
}, []string{"source"})

// GraphUpdatesProcessingFailedCounter reports the number of updates which failed to be processed
var GraphUpdatesProcessingFailedCounter = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "go_graphkb_graph_updates_failed_counter",
	Help: "The number of update requests which failed",
}, []string{"source"})
