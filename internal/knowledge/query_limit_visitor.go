package knowledge

import "github.com/clems4ever/go-graphkb/internal/query"

type QueryLimitVisitor struct {
	ExpressionVisitorBase

	Limit int64
}

// ParseExpression return whether the expression require aggregation
func (qlv *QueryLimitVisitor) ParseExpression(q *query.QueryExpression) error {
	err := NewExpressionParser(qlv).ParseExpression(q)
	if err != nil {
		return err
	}
	return nil
}

func (qlv *QueryLimitVisitor) OnIntegerLiteral(value int64) error {
	qlv.Limit = value
	return nil
}
