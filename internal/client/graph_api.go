package client

import (
	"github.com/clems4ever/go-graphkb/internal/knowledge"
)

// GraphAPI represent the graph API from a data source point of view
type GraphAPI struct {
	client *GraphClient
}

// GraphAPIOptions options to pass to build graph API
type GraphAPIOptions struct {
	URL        string
	AuthToken  string
	SkipVerify bool
}

// NewGraphAPI create an emitter of graph
func NewGraphAPI(options GraphAPIOptions) *GraphAPI {
	return &GraphAPI{
		client: NewGraphClient(options.URL, options.AuthToken, options.SkipVerify),
	}
}

// CreateTransaction create a full graph transaction. This kind of transaction will diff the new graph
// with previous version of it.
func (gapi *GraphAPI) CreateTransaction(currentGraph *knowledge.Graph) *Transaction {
	transaction := new(Transaction)
	transaction.newGraph = knowledge.NewGraph()
	transaction.binder = knowledge.NewGraphBinder(transaction.newGraph)
	transaction.client = gapi.client
	transaction.currentGraph = currentGraph
	transaction.parallelization = 30
	return transaction
}

// ReadCurrentGraph read the current graph stored in graph kb
func (gapi *GraphAPI) ReadCurrentGraph() (*knowledge.Graph, error) {
	return gapi.client.ReadCurrentGraph()
}
