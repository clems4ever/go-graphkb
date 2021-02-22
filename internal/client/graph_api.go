package client

import (
	"fmt"
	"time"

	"github.com/clems4ever/go-graphkb/internal/knowledge"
	"github.com/sirupsen/logrus"
)

// GraphAPI represent the graph API from a data source point of view
type GraphAPI struct {
	client *GraphClient

	options GraphAPIOptions

	currentGraph *knowledge.Graph
}

// GraphAPIOptions options to pass to build graph API
type GraphAPIOptions struct {
	// URL to GraphKB
	URL string
	// Auth token for this data source.
	AuthToken string
	// Skip verifying the certificate when using https
	SkipVerify bool

	// The level of parallelization for streaming updates, i.e., number of HTTP requests sent in parallel.
	Parallelization int

	// The size of a chunk of updates, i.e., number of assets or relations sent in one HTTP request to the streaming API.
	ChunkSize int

	// Max number of retries before giving up (default is 10)
	MaxRetries int
	// The base delay between retries (default is 5 seconds). This delay is multiplied by the backoff factor.
	RetryDelay time.Duration

	// The backoff factor (default is 1.01)
	RetryBackoffFactor float64
}

// NewGraphAPI create an emitter of graph
func NewGraphAPI(options GraphAPIOptions) *GraphAPI {
	return &GraphAPI{
		client:  NewGraphClient(options.URL, options.AuthToken, options.SkipVerify),
		options: options,
	}
}

// CreateTransaction create a full graph transaction. This kind of transaction will diff the new graph
// with previous version of it.
func (gapi *GraphAPI) CreateTransaction() (*Transaction, error) {
	if gapi.currentGraph == nil {
		logrus.Debug("transaction: fetching remote graph")
		g, err := gapi.ReadCurrentGraph()
		if err != nil {
			return nil, fmt.Errorf("create transaction: %w", err)
		}
		gapi.currentGraph = g
	}

	var parallelization = gapi.options.Parallelization
	if parallelization == 0 {
		parallelization = 30
	}

	var chunkSize = gapi.options.ChunkSize
	if chunkSize == 0 {
		chunkSize = 1000
	}

	var maxRetries = gapi.options.MaxRetries
	if maxRetries == 0 {
		maxRetries = 10
	}

	var retryDelay = gapi.options.RetryDelay
	if retryDelay == 0 {
		retryDelay = 5 * time.Second
	}

	var retryBackoff = gapi.options.RetryBackoffFactor
	if retryBackoff == 0.0 {
		retryBackoff = 1.01
	}

	transaction := new(Transaction)
	transaction.newGraph = knowledge.NewGraph()
	transaction.binder = knowledge.NewGraphBinder(transaction.newGraph)
	transaction.client = gapi.client
	transaction.currentGraph = gapi.currentGraph
	transaction.parallelization = parallelization
	transaction.chunkSize = chunkSize

	transaction.retryCount = maxRetries
	transaction.retryDelay = retryDelay
	transaction.retryBackoffFactor = retryBackoff

	transaction.onError = func(err error) {
		// there was an error, we don't know the remote graph state.
		// we clear the cached copy to refresh in on the next run.
		logrus.Debug("transaction: clearing graph cache because of error:", err)
		gapi.currentGraph = nil
	}

	transaction.onSuccess = func(g *knowledge.Graph) {
		// tx was successful, we updated to local graph cache to
		// speed up the next tx.
		gapi.currentGraph = g
	}

	return transaction, nil
}

// ReadCurrentGraph read the current graph stored in graph kb
func (gapi *GraphAPI) ReadCurrentGraph() (*knowledge.Graph, error) {
	return gapi.client.ReadCurrentGraph()
}
