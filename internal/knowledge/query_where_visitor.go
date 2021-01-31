package knowledge

import "github.com/clems4ever/go-graphkb/internal/query"

type QueryWhereVisitor struct {
	ExpressionVisitorBase

	Variables  []string
	queryGraph *QueryGraph
}

// NewQueryWhereVisitor create an instance of query where visitor.
func NewQueryWhereVisitor(queryGraph *QueryGraph) *QueryWhereVisitor {
	return &QueryWhereVisitor{
		queryGraph: queryGraph,
	}
}

// ParseExpression return whether the expression require aggregation
func (qwv *QueryWhereVisitor) ParseExpression(q *query.QueryExpression) (string, error) {
	expression, err := NewExpressionBuilder(qwv.queryGraph).Build(q)
	if err != nil {
		return "", err
	}
	return expression, nil
}

func (qwv *QueryWhereVisitor) OnVariable(name string) error {
	qwv.Variables = append(qwv.Variables, name)
	return nil
}
