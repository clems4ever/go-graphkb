package knowledge

import (
	"fmt"
	"strings"

	"github.com/clems4ever/go-graphkb/internal/query"
)

type FunctionInvocationContext struct {
	Distinct       bool
	FunctionName   string
	Expression     string
	VariableName   string
	VariableType   ExpressionType
	PropertiesPath []string
}

type ProjectionItem struct {
	Variable string
	Function string
	Distinct bool
}

type ProjectionVisitor struct {
	ExpressionVisitorBase

	queryGraph *QueryGraph

	Aggregation    bool
	TypeAndIndex   TypeAndIndex
	ExpressionType ExpressionType
	Projections    []ProjectionItem

	// This context is instantiated as soon as a function call is detected.
	// It's used because an expression can contain a function call which is applied to an expression so the visitor
	// is used both in the context of the projection but also of the function call. When this context is nil, we refer
	// to the projection context, otherwise to the function call context.
	functionInvocationContext *FunctionInvocationContext

	etype          ExpressionType
	propertiesPath []string
	variableName   string
}

func NewProjectionVisitor(queryGraph *QueryGraph) *ProjectionVisitor {
	return &ProjectionVisitor{
		queryGraph: queryGraph,
	}
}

// ParseExpression return whether the expression require aggregation
func (pv *ProjectionVisitor) ParseExpression(q *query.QueryExpression) error {
	err := NewExpressionParser(pv, pv.queryGraph).ParseExpression(q)
	if err != nil {
		return err
	}
	return nil
}

// OnExitFunctionInvocation called when the ExitFunctionInvocation is parsed. Name is the name of the function.
func (pv *ProjectionVisitor) OnEnterFunctionInvocation(name string, distinct bool) error {
	if name == "COUNT" {
		pv.functionInvocationContext = new(FunctionInvocationContext)
		pv.functionInvocationContext.Distinct = distinct
		pv.functionInvocationContext.FunctionName = name
		pv.Aggregation = true
	} else {
		return fmt.Errorf("Function %s is not supported", name)
	}
	return nil
}

// OnExitFunctionInvocation called when the ExitFunctionInvocation is parsed. Name is the name of the function.
func (pv *ProjectionVisitor) OnExitFunctionInvocation(name string, distinct bool) error {
	return nil
}

func (pv *ProjectionVisitor) OnVariable(name string) error {
	typeAndIndex, err := pv.queryGraph.FindVariable(name)
	if err != nil {
		return err
	}

	var etype ExpressionType

	switch typeAndIndex.Type {
	case NodeType:
		etype = NodeExprType
	case RelationType:
		etype = EdgeExprType
	default:
		etype = PropertyExprType
	}

	if pv.functionInvocationContext != nil {
		pv.functionInvocationContext.VariableType = etype
		pv.functionInvocationContext.VariableName = name
	} else {
		pv.etype = etype
		pv.variableName = name
	}
	return nil
}

func (pv *ProjectionVisitor) OnVariablePropertiesPath(properties []string) error {
	if pv.functionInvocationContext != nil {
		pv.functionInvocationContext.PropertiesPath = properties
	} else {
		pv.propertiesPath = properties
	}
	return nil
}

func (pv *ProjectionVisitor) OnExitPropertyOrLabelsExpression(e query.QueryPropertyOrLabelsExpression) error {
	projections := []ProjectionItem{}

	if pv.variableName != "" {
		typeAndIndex, err := pv.queryGraph.FindVariable(pv.variableName)
		if err != nil {
			return err
		}

		var properties []string
		var alias string

		switch typeAndIndex.Type {
		case NodeType:
			alias = "a"
			properties = []string{"id", "value", "type"}
		case RelationType:
			alias = "r"
			properties = []string{"id", "from_id", "to_id", "type"}
		}
		alias += fmt.Sprintf("%d", typeAndIndex.Index)

		if len(pv.propertiesPath) > 0 {
			properties = []string{strings.Join(pv.propertiesPath, ".")}
		}

		pv.ExpressionType = pv.etype
		for _, p := range properties {
			projections = append(projections, ProjectionItem{Variable: fmt.Sprintf("%s.%s", alias, p)})
		}
	} else if pv.functionInvocationContext != nil {
		typeAndIndex, err := pv.queryGraph.FindVariable(pv.functionInvocationContext.VariableName)
		if err != nil {
			return err
		}

		var alias string

		switch typeAndIndex.Type {
		case NodeType:
			alias = "a"
		case RelationType:
			alias = "r"
		}
		alias += fmt.Sprintf("%d", typeAndIndex.Index)

		pv.ExpressionType = PropertyExprType
		if len(pv.functionInvocationContext.PropertiesPath) == 0 {
			variable := fmt.Sprintf("%s.id", alias)
			projections = append(projections, ProjectionItem{
				Function: pv.functionInvocationContext.FunctionName,
				Variable: variable,
				Distinct: pv.functionInvocationContext.Distinct,
			})
		} else {
			variable := fmt.Sprintf("%s.%s",
				alias, strings.Join(pv.functionInvocationContext.PropertiesPath, "."))
			projections = append(projections, ProjectionItem{
				Function: pv.functionInvocationContext.FunctionName,
				Variable: variable,
				Distinct: pv.functionInvocationContext.Distinct,
			})
		}
	}

	pv.Projections = projections
	pv.propertiesPath = nil
	return nil
}
