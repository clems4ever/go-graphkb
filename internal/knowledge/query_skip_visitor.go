package knowledge

import "github.com/clems4ever/go-graphkb/internal/query"

type QuerySkipVisitor struct {
	ExpressionVisitorBase

	Skip int64
}

// ParseExpression return whether the expression require aggregation
func (qsv *QuerySkipVisitor) ParseExpression(q *query.QueryExpression) error {
	err := NewExpressionParser(qsv).ParseExpression(q)
	if err != nil {
		return err
	}
	return nil
}

func (qsv *QuerySkipVisitor) OnIntegerLiteral(value int64) error {
	qsv.Skip = value
	return nil
}
