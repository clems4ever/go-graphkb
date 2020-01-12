package knowledge

import "fmt"

import "strings"

import "github.com/clems4ever/go-graphkb/internal/query"

type ExpressionBuilder struct {
	QueryGraph *QueryGraph

	parser  *ExpressionParser
	visitor *SQLExpressionVisitor
}

func NewExpressionBuilder(queryGraph *QueryGraph) *ExpressionBuilder {
	visitor := SQLExpressionVisitor{}
	visitor.queryGraph = queryGraph
	return &ExpressionBuilder{
		QueryGraph: queryGraph,
		parser:     NewExpressionParser(&visitor),
		visitor:    &visitor,
	}
}

func (eb *ExpressionBuilder) Build(q *query.QueryExpression) (string, error) {
	err := eb.parser.ParseExpression(q)
	if err != nil {
		return "", err
	}
	return eb.visitor.expression, nil
}

type SQLExpressionVisitor struct {
	ExpressionVisitorBase

	queryGraph *QueryGraph

	propertiesPath []string

	variableName   *string
	stringLiteral  *string
	integerLiteral *int64
	doubleLiteral  *float64
	boolLiteral    *bool

	parenthesizedExpression string

	functionInvocation string

	propertyLabelsExpression string

	comparisonExpression string
	comparisonOperator   query.ComparisonOperator

	stringExpression string
	stringOperator   query.StringOperator

	notExpressions []string
	andExpressions []string
	orExpression   string
	expression     string
}

func (sev *SQLExpressionVisitor) OnVariable(name string) error {
	sev.variableName = new(string)
	*sev.variableName = name
	return nil
}

func (sev *SQLExpressionVisitor) OnStringLiteral(value string) error {
	sev.stringLiteral = new(string)
	*sev.stringLiteral = value
	return nil
}

func (sev *SQLExpressionVisitor) OnIntegerLiteral(value int64) error {
	sev.integerLiteral = new(int64)
	*sev.integerLiteral = value
	return nil
}

func (sev *SQLExpressionVisitor) OnDoubleLiteral(value float64) error {
	sev.doubleLiteral = new(float64)
	*sev.doubleLiteral = value
	return nil
}

func (sev *SQLExpressionVisitor) OnBooleanLiteral(value bool) error {
	sev.boolLiteral = new(bool)
	*sev.boolLiteral = value
	return nil
}

func (sev *SQLExpressionVisitor) OnVariablePropertiesPath(propertiesPath []string) error {
	sev.propertiesPath = propertiesPath
	return nil
}

func (sev *SQLExpressionVisitor) OnExitParenthesizedExpression() error {
	sev.parenthesizedExpression = sev.expression
	sev.expression = ""
	return nil
}

func (sev *SQLExpressionVisitor) OnExitPropertyOrLabelsExpression(e query.QueryPropertyOrLabelsExpression) error {
	if sev.variableName != nil {
		var properties []string
		typeAndIndex, err := sev.queryGraph.FindVariable(*sev.variableName)
		if err != nil {
			return err
		}

		alias := ""
		switch typeAndIndex.Type {
		case NodeType:
			alias = "a"
			properties = []string{"id", "value", "type"}
		case RelationType:
			alias = "r"
			properties = []string{"from_id", "to_id", "type"}
		}
		alias += fmt.Sprintf("%d", typeAndIndex.Index)
		if len(sev.propertiesPath) > 0 {
			properties = []string{strings.Join(sev.propertiesPath, ".")}
		}

		projection := []string{}
		for _, p := range properties {
			projection = append(projection, fmt.Sprintf("%s.%s", alias, p))
		}

		sev.propertyLabelsExpression = strings.Join(projection, ", ")
		sev.variableName = nil
		sev.propertiesPath = nil
	} else if sev.stringLiteral != nil {
		sev.propertyLabelsExpression = fmt.Sprintf("'%s'", *sev.stringLiteral)
		sev.stringLiteral = nil
	} else if sev.integerLiteral != nil {
		sev.propertyLabelsExpression = fmt.Sprintf("%d", *sev.integerLiteral)
		sev.integerLiteral = nil
	} else if sev.doubleLiteral != nil {
		sev.propertyLabelsExpression = fmt.Sprintf("%f", *sev.doubleLiteral)
		sev.doubleLiteral = nil
	} else if sev.boolLiteral != nil {
		value := "false"
		if *sev.boolLiteral {
			value = "true"
		}
		sev.propertyLabelsExpression = value
		sev.boolLiteral = nil
	} else if sev.functionInvocation != "" {
		sev.propertyLabelsExpression = sev.functionInvocation
		sev.functionInvocation = ""
	} else if sev.parenthesizedExpression != "" {
		sev.propertyLabelsExpression = fmt.Sprintf("(%s)", sev.parenthesizedExpression)
		sev.parenthesizedExpression = ""
	}
	return nil
}

func (sev *SQLExpressionVisitor) OnExitFunctionInvocation(name string) error {
	sev.functionInvocation = fmt.Sprintf("%s(%s)", name, sev.expression)
	sev.expression = ""
	return nil
}

func (sev *SQLExpressionVisitor) OnExitStringListNullOperatorExpression(e query.QueryStringListNullOperatorExpression) error {
	if sev.stringExpression != "" {
		expression := sev.propertyLabelsExpression[1 : len(sev.propertyLabelsExpression)-1]

		switch sev.stringOperator {
		case query.ContainsOperator:
			sev.stringExpression = fmt.Sprintf("%s LIKE '%%%s%%'", sev.stringExpression, expression)
		case query.EndsWithOperator:
			sev.stringExpression = fmt.Sprintf("%s LIKE '%%%s'", sev.stringExpression, expression)
		case query.StartsWithOperator:
			sev.stringExpression = fmt.Sprintf("%s LIKE '%s%%'", sev.stringExpression, expression)
		}
	} else {
		sev.stringExpression = sev.propertyLabelsExpression
	}
	sev.propertyLabelsExpression = ""
	return nil
}

func (sev *SQLExpressionVisitor) OnStringOperator(operator query.StringOperator) error {
	sev.stringOperator = operator
	sev.stringExpression = sev.propertyLabelsExpression
	sev.propertyLabelsExpression = ""
	return nil
}

func (sev *SQLExpressionVisitor) OnExitComparisonExpression() error {
	if sev.comparisonExpression != "" {
		operatorStr := ""
		switch sev.comparisonOperator {
		case query.Equal:
			operatorStr = "="
		case query.NotEqual:
			operatorStr = "<>"
		case query.Less:
			operatorStr = "<"
		case query.LessOrEqual:
			operatorStr = "<="
		case query.Greater:
			operatorStr = ">"
		case query.GreaterOrEqual:
			operatorStr = ">="
		}

		sev.comparisonExpression = fmt.Sprintf("%s %s %s", sev.comparisonExpression,
			operatorStr, sev.stringExpression)
	} else {
		sev.comparisonExpression = sev.stringExpression
	}
	sev.stringExpression = ""
	return nil
}

func (sev *SQLExpressionVisitor) OnComparisonOperator(operator query.ComparisonOperator) error {
	sev.comparisonOperator = operator
	sev.comparisonExpression = sev.stringExpression
	sev.stringExpression = ""
	return nil
}

func (sev *SQLExpressionVisitor) OnExitNotExpression(not bool) error {
	expr := sev.comparisonExpression
	if not {
		expr = "NOT " + expr
	}
	sev.notExpressions = append(sev.notExpressions, expr)
	sev.comparisonExpression = ""
	return nil
}

func (sev *SQLExpressionVisitor) OnExitAndExpression() error {
	if sev.andExpressions == nil {
		sev.andExpressions = []string{}
	}
	sev.andExpressions = append(sev.andExpressions, strings.Join(sev.notExpressions, " AND "))
	sev.notExpressions = nil
	return nil
}

func (sev *SQLExpressionVisitor) OnExitOrExpression() error {
	sev.orExpression = strings.Join(sev.andExpressions, " OR ")
	sev.andExpressions = nil
	return nil
}

func (sev *SQLExpressionVisitor) OnExitExpression() error {
	sev.expression = sev.orExpression
	sev.orExpression = ""
	return nil
}
