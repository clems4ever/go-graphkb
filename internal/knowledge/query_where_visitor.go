package knowledge

import "github.com/clems4ever/go-graphkb/internal/query"

type QueryWhereVisitor struct {
	ExpressionVisitorBase

	Variables []string
}

// ParseExpression return whether the expression require aggregation
func (qwv *QueryWhereVisitor) ParseExpression(q *query.QueryExpression, qg *QueryGraph) (string, error) {
	expression, err := NewExpressionBuilder(qg).Build(q)
	if err != nil {
		return "", err
	}
	return expression, nil
}

func (qwv *QueryWhereVisitor) OnVariable(name string) error {
	qwv.Variables = append(qwv.Variables, name)
	return nil
}
