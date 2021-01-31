package knowledge

import "github.com/clems4ever/go-graphkb/internal/query"

type QuerySkipVisitor struct {
	ExpressionVisitorBase

	queryGraph *QueryGraph
	Skip       int64
}

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

func (qsv *QuerySkipVisitor) OnIntegerLiteral(value int64) error {
	qsv.Skip = value
	return nil
}
