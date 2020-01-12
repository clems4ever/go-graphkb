package knowledge

import "github.com/clems4ever/go-graphkb/internal/query"

import "fmt"

type ProjectionVisitor struct {
	ExpressionVisitorBase

	QueryGraph *QueryGraph

	Aggregation    bool
	ExpressionType ExpressionType
	TypeAndIndex   TypeAndIndex
}

// ParseExpression return whether the expression require aggregation
func (pv *ProjectionVisitor) ParseExpression(q *query.QueryExpression) error {
	err := NewExpressionParser(pv).ParseExpression(q)
	if err != nil {
		return err
	}
	return nil
}

func (pv *ProjectionVisitor) OnEnterFunctionInvocation(name string) error {
	if name == "COUNT" {
		pv.Aggregation = true
	} else {
		return fmt.Errorf("Function %s is not supported", name)
	}
	return nil
}

func (pv *ProjectionVisitor) OnVariable(name string) error {
	typeAndIndex, err := pv.QueryGraph.FindVariable(name)
	if err != nil {
		return err
	}

	switch typeAndIndex.Type {
	case NodeType:
		pv.ExpressionType = NodeExprType
	case RelationType:
		pv.ExpressionType = EdgeExprType
	default:
		pv.ExpressionType = PropertyExprType
	}
	pv.TypeAndIndex = typeAndIndex
	return nil
}
