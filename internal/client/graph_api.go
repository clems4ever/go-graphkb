package client

import (
	"github.com/clems4ever/go-graphkb/internal/knowledge"
)

// GraphAPI represent the graph API from a data source point of view
type GraphAPI struct {
	client *GraphClient

	parallelization int
	chunkSize       int
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
}

// NewGraphAPI create an emitter of graph
func NewGraphAPI(options GraphAPIOptions) *GraphAPI {
	return &GraphAPI{
		client:          NewGraphClient(options.URL, options.AuthToken, options.SkipVerify),
		parallelization: options.Parallelization,
		chunkSize:       options.ChunkSize,
	}
}

// CreateTransaction create a full graph transaction. This kind of transaction will diff the new graph
// with previous version of it.
func (gapi *GraphAPI) CreateTransaction(currentGraph *knowledge.Graph) *Transaction {
	var parallelization = gapi.parallelization
	if parallelization == 0 {
		parallelization = 30
	}

	var chunkSize = gapi.chunkSize
	if chunkSize == 0 {
		chunkSize = 1000
	}

	transaction := new(Transaction)
	transaction.newGraph = knowledge.NewGraph()
	transaction.binder = knowledge.NewGraphBinder(transaction.newGraph)
	transaction.client = gapi.client
	transaction.currentGraph = currentGraph
	transaction.parallelization = parallelization
	transaction.chunkSize = chunkSize
	return transaction
}

// ReadCurrentGraph read the current graph stored in graph kb
func (gapi *GraphAPI) ReadCurrentGraph() (*knowledge.Graph, error) {
	return gapi.client.ReadCurrentGraph()
}
