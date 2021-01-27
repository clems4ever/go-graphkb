package client

import (
	"time"

	"github.com/clems4ever/go-graphkb/internal/knowledge"
)

// GraphAPI represent the graph API from a data source point of view
type GraphAPI struct {
	client *GraphClient

	options GraphAPIOptions
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
func (gapi *GraphAPI) CreateTransaction(currentGraph *knowledge.Graph) *Transaction {
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
	transaction.currentGraph = currentGraph
	transaction.parallelization = parallelization
	transaction.chunkSize = chunkSize

	transaction.retryCount = maxRetries
	transaction.retryDelay = retryDelay
	transaction.retryBackoffFactor = retryBackoff
	return transaction
}

// ReadCurrentGraph read the current graph stored in graph kb
func (gapi *GraphAPI) ReadCurrentGraph() (*knowledge.Graph, error) {
	return gapi.client.ReadCurrentGraph()
}
