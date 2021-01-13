package knowledge

// DataSource an interface for data source allowing to read current graph and push graph updates
type DataSource struct {
	api *GraphAPI
}

// NewDataSource create an emitter of graph
func NewDataSource(api *GraphAPI) *DataSource {
	return &DataSource{api: api}
}

// CreateTransaction create a full graph transaction. This kind of transaction will diff the new graph
// with previous version of it.
func (ds *DataSource) CreateTransaction(currentGraph *Graph) *Transaction {
	transaction := new(Transaction)
	transaction.newGraph = NewGraph()
	transaction.binder = NewGraphBinder(transaction.newGraph)
	transaction.api = ds.api
	transaction.currentGraph = currentGraph
	return transaction
}

// ReadCurrentGraph read the graph related to that data source
func (ds *DataSource) ReadCurrentGraph() (*Graph, error) {
	return ds.api.ReadCurrentGraph()
}
