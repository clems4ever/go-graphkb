package knowledge

import "github.com/clems4ever/go-graphkb/internal/query"

// QuerySkipVisitor a visitor for the skip clause
type QuerySkipVisitor struct {
	ExpressionVisitorBase

	queryGraph *QueryGraph
	Skip       int64
}

// NewQuerySkipVisitor create an instance of the skip visitor
func NewQuerySkipVisitor(queryGraph *QueryGraph) *QuerySkipVisitor {
	return &QuerySkipVisitor{
		queryGraph: queryGraph,
	}
}

// ParseExpression return whether the expression require aggregation
func (qsv *QuerySkipVisitor) ParseExpression(q *query.QueryExpression) error {
	err := NewExpressionParser(qsv, qsv.queryGraph).ParseExpression(q)
	if err != nil {
		return err
	}
	return nil
}

// OnIntegerLiteral handler called when an integer literal is visited
func (qsv *QuerySkipVisitor) OnIntegerLiteral(value int64) error {
	qsv.Skip = value
	return nil
}
