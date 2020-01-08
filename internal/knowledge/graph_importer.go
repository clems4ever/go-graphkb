package knowledge

// GraphImporter an interface for importers allowing to read current graph and push graph updates
type GraphImporter struct {
	api *GraphAPI
}

// NewGraphEmitter create an emitter of graph
func NewGraphImporter(api *GraphAPI) *GraphImporter {
	return &GraphImporter{api: api}
}

// CreateTransaction create a full graph transaction. This kind of transaction will diff the new graph
// with previous version of it.
func (gi *GraphImporter) CreateTransaction(currentGraph *Graph) *Transaction {
	transaction := new(Transaction)
	transaction.newGraph = NewGraph()
	transaction.binder = NewGraphBinder(transaction.newGraph)
	transaction.api = gi.api
	transaction.currentGraph = currentGraph
	return transaction
}

func (gi *GraphImporter) ReadCurrentGraph() (*Graph, error) {
	return gi.api.ReadCurrentGraph()
}
