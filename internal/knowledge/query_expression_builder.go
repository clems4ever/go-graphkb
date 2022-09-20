package knowledge

import (
	"fmt"
	"strings"

	"github.com/clems4ever/go-graphkb/internal/query"
)

// ExpressionBuilder build a SQL part based on a Cypher expression
type ExpressionBuilder struct {
	QueryGraph *QueryGraph

	parser  *ExpressionParser
	visitor *SQLExpressionVisitor
}

// NewExpressionBuilder create a new instance of expression builder
func NewExpressionBuilder(queryGraph *QueryGraph) *ExpressionBuilder {
	visitor := SQLExpressionVisitor{queryGraph: queryGraph}
	return &ExpressionBuilder{
		QueryGraph: queryGraph,
		parser:     NewExpressionParser(&visitor, queryGraph),
		visitor:    &visitor,
	}
}

// Build the SQL expression from the Cypher expression
func (eb *ExpressionBuilder) Build(q *query.QueryExpression) (string, error) {
	err := eb.parser.ParseExpression(q)
	if err != nil {
		return "", err
	}
	return eb.visitor.expression, nil
}

// SQLExpressionVisitor visitor used to build the SQL part from the Cypher expression.
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

	// This expression should contain the EXIST(SELECT ...) expression
	// a Cypher where clause containing a pattern is translated as SQL EXIST clause.
	relationshipsPatternExpression string

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
			alias += fmt.Sprintf("%d", typeAndIndex.Index)
		case RelationType:
			alias = "r"
			properties = []string{"id", "from_id", "to_id", "type"}
			alias += fmt.Sprintf("%d", typeAndIndex.Index)
		case PropertyType:
			alias = *sev.variableName
			properties = []string{""}
		}

		if len(sev.propertiesPath) > 0 {
			properties = []string{strings.Join(sev.propertiesPath, ".")}
		}

		projection := []string{}
		for _, p := range properties {
			if typeAndIndex.Type == PropertyType {
				projection = append(projection, alias)
				break
			}
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
	} else if sev.relationshipsPatternExpression != "" {
		sev.propertyLabelsExpression = sev.relationshipsPatternExpression
		sev.relationshipsPatternExpression = ""
	}
	return nil
}

// OnExitFunctionInvocation build the SQL snippet calling the function
func (sev *SQLExpressionVisitor) OnExitFunctionInvocation(name string, distinct bool) error {
	distinctStr := ""
	if distinct {
		distinctStr = "DISTINCT "
	}

	if name == "COUNT" && !distinct {
		sev.expression = "*"
	}

	sev.functionInvocation = fmt.Sprintf("%s(%s%s)", name, distinctStr, sev.expression)
	sev.expression = ""
	return nil
}

// OnExitRelationshipsPattern build the SQL query that will be put in the EXIST clause
func (sev *SQLExpressionVisitor) OnExitRelationshipsPattern(q query.QueryRelationshipsPattern, id int) error {
	scope := Scope{Context: WhereContext, ID: id}

	// Build the constraints for the patterns in the WHERE clause
	joins, from, err := buildSQLConstraintsFromPatterns(sev.queryGraph, nil, scope)
	if err != nil {
		return fmt.Errorf("Unable to deduce SQL constraints for EXISTS query")
	}

	// Build a SELECT query such as SELECT 1 FROM assets a0 WHERE a0.type = 'mytype'.
	// This is then wrapped into an EXISTS SQL clause
	query, err := buildBasicSingleSQLSelect(false, []SQLProjection{{Variable: "1"}}, from, joins[0],
		[]SQLInnerStructure{}, AndOrExpression{}, []int{}, AndOrExpression{}, map[string]struct{}{}, 0, 0)
	if err != nil {
		return fmt.Errorf("Unable to build SQL query for EXISTS query: %v", err)
	}
	sev.relationshipsPatternExpression = fmt.Sprintf("EXISTS (%s)", query)
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
