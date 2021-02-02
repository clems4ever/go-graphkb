package knowledge

import "github.com/clems4ever/go-graphkb/internal/query"

// QueryLimitVisitor a visitor for the limit clause
type QueryLimitVisitor struct {
	ExpressionVisitorBase

	Limit int64

	queryGraph *QueryGraph
}

// NewQueryLimitVisitor create an instance of query limit visitor
func NewQueryLimitVisitor(queryGraph *QueryGraph) *QueryLimitVisitor {
	return &QueryLimitVisitor{
		queryGraph: queryGraph,
	}
}

// ParseExpression return whether the expression require aggregation
func (qlv *QueryLimitVisitor) ParseExpression(q *query.QueryExpression) error {
	err := NewExpressionParser(qlv, qlv.queryGraph).ParseExpression(q)
	if err != nil {
		return err
	}
	return nil
}

// OnIntegerLiteral handler called when an integer literal is visited
func (qlv *QueryLimitVisitor) OnIntegerLiteral(value int64) error {
	qlv.Limit = value
	return nil
}
