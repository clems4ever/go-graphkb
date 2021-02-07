package knowledge

import (
	"fmt"
	"strings"

	"github.com/clems4ever/go-graphkb/internal/query"
)

// ExpressionType expression type
type ExpressionType int

const (
	// NodeExprType node expression type
	NodeExprType ExpressionType = iota
	// EdgeExprType edge expression type
	EdgeExprType ExpressionType = iota
	// PropertyExprType property expression type
	PropertyExprType ExpressionType = iota
)

// ExpressionParser is a parser of expression
type ExpressionParser struct {
	visitor    ExpressionVisitor
	queryGraph *QueryGraph

	// This ID is incremented for every pattern found in the expression
	patternIDGenerator int
}

// NewExpressionParser create a new instance of expression parser.
func NewExpressionParser(visitor ExpressionVisitor, queryGraph *QueryGraph) *ExpressionParser {
	return &ExpressionParser{
		visitor:            visitor,
		queryGraph:         queryGraph,
		patternIDGenerator: 0,
	}
}

// ParsePropertyOrLabelsExpression parse property or labels expression
func (ep *ExpressionParser) ParsePropertyOrLabelsExpression(q *query.QueryPropertyOrLabelsExpression) error {
	err := ep.visitor.OnEnterPropertyOrLabelsExpression(*q)
	if err != nil {
		return err
	}

	if len(q.PropertyKeys) > 0 {
		if err := ep.visitor.OnVariablePropertiesPath(q.PropertyKeys); err != nil {
			return err
		}
	}

	if q.Atom.Variable != nil {
		err := ep.visitor.OnVariable(*q.Atom.Variable)
		if err != nil {
			return err
		}
	} else if q.Atom.Literal != nil {
		if q.Atom.Literal.String != nil {
			err := ep.visitor.OnStringLiteral(*q.Atom.Literal.String)
			if err != nil {
				return err
			}
		} else if q.Atom.Literal.Integer != nil {
			err := ep.visitor.OnIntegerLiteral(*q.Atom.Literal.Integer)
			if err != nil {
				return err
			}
		} else if q.Atom.Literal.Double != nil {
			err := ep.visitor.OnDoubleLiteral(*q.Atom.Literal.Double)
			if err != nil {
				return err
			}
		} else if q.Atom.Literal.Boolean != nil {
			err := ep.visitor.OnBooleanLiteral(*q.Atom.Literal.Boolean)
			if err != nil {
				return err
			}
		}
	} else if q.Atom.FunctionInvocation != nil {
		fnName := strings.ToUpper(q.Atom.FunctionInvocation.FunctionName)
		distinct := q.Atom.FunctionInvocation.Distinct
		err := ep.visitor.OnEnterFunctionInvocation(fnName, distinct)
		if err != nil {
			return err
		}

		args := q.Atom.FunctionInvocation.Expressions
		for _, a := range args {
			if err := ep.ParseExpression(&a); err != nil {
				return err
			}
		}
		err = ep.visitor.OnExitFunctionInvocation(fnName, distinct)
		if err != nil {
			return err
		}
	} else if q.Atom.ParenthesizedExpression != nil {
		err := ep.visitor.OnEnterParenthesizedExpression()
		if err != nil {
			return err
		}
		err = ep.ParseExpression(q.Atom.ParenthesizedExpression)
		if err != nil {
			return err
		}
		err = ep.visitor.OnExitParenthesizedExpression()
		if err != nil {
			return err
		}
	} else if q.Atom.RelationshipsPattern != nil {
		parser := NewPatternParser(ep.queryGraph)
		// Parse the pattern to push the nodes and relations into the query graph
		err := parser.ParseRelationshipsPattern(
			q.Atom.RelationshipsPattern,
			Scope{Context: WhereContext, ID: ep.patternIDGenerator})
		if err != nil {
			return err
		}
		defer func() { ep.patternIDGenerator++ }()

		err = ep.visitor.OnEnterRelationshipsPattern(*q.Atom.RelationshipsPattern, ep.patternIDGenerator)
		if err != nil {
			return err
		}

		err = ep.ParseRelationshipsPattern(q.Atom.RelationshipsPattern)
		if err != nil {
			return err
		}

		err = ep.visitor.OnExitRelationshipsPattern(*q.Atom.RelationshipsPattern, ep.patternIDGenerator)
		if err != nil {
			return err
		}

	} else {
		return fmt.Errorf("Unable to parse property or labels expression")
	}

	err = ep.visitor.OnExitPropertyOrLabelsExpression(*q)
	if err != nil {
		return err
	}

	return nil
}

// ParseRelationshipsPattern parse a query relationships pattern
func (ep *ExpressionParser) ParseRelationshipsPattern(q *query.QueryRelationshipsPattern) error {
	err := ep.visitor.OnNodePattern(q.QueryNodePattern)
	if err != nil {
		return err
	}

	for _, pc := range q.QueryPatternElementChains {
		err = ep.visitor.OnRelationshipPattern(pc.RelationshipPattern)
		if err != nil {
			return err
		}

		err = ep.visitor.OnNodePattern(pc.NodePattern)
		if err != nil {
			return err
		}
	}
	return nil
}

// ParseStringListNullOperatorExpression parse string list null operator expression
func (ep *ExpressionParser) ParseStringListNullOperatorExpression(q *query.QueryStringListNullOperatorExpression) error {
	err := ep.visitor.OnEnterStringListNullOperatorExpression(*q)
	if err != nil {
		return err
	}

	err = ep.ParsePropertyOrLabelsExpression(&q.PropertyOrLabelsExpression)
	if err != nil {
		return err
	}

	for i := range q.StringOperatorExpression {
		stringExpression := q.StringOperatorExpression[i]

		if stringExpression.PropertyOrLabelsExpression.Atom.Literal == nil {
			return fmt.Errorf("Expression must be a literal to be used with string operator")
		}

		if stringExpression.PropertyOrLabelsExpression.Atom.Literal.String == nil {
			return fmt.Errorf("Expression must be a string literal to be used with string operator")
		}

		err := ep.visitor.OnStringOperator(stringExpression.Operator)
		if err != nil {
			return err
		}

		err = ep.ParsePropertyOrLabelsExpression(&stringExpression.PropertyOrLabelsExpression)
		if err != nil {
			return err
		}
	}

	err = ep.visitor.OnExitStringListNullOperatorExpression(*q)
	if err != nil {
		return err
	}

	return nil
}

// ParseUnaryAddOrSubtractExpression parse unary add or subtract expression
func (ep *ExpressionParser) ParseUnaryAddOrSubtractExpression(q *query.QueryUnaryAddOrSubtractExpression) error {
	err := ep.ParseStringListNullOperatorExpression(&q.StringListNullOperatorExpression)
	if err != nil {
		return err
	}
	return nil
}

// ParsePowerOfExpression parse power of expression
func (ep *ExpressionParser) ParsePowerOfExpression(q *query.QueryPowerOfExpression) error {
	err := ep.visitor.OnEnterPowerOfExpression()
	if err != nil {
		return err
	}

	for i := 0; i < len(q.QueryUnaryAddOrSubtractExpressions); i++ {
		err = ep.ParseUnaryAddOrSubtractExpression(&q.QueryUnaryAddOrSubtractExpressions[i])
		if err != nil {
			return err
		}
	}

	err = ep.visitor.OnExitPowerOfExpression()
	if err != nil {
		return err
	}
	return nil
}

// ParseMultipleDivideModuloExpression parse multiple divide modulo expression
func (ep *ExpressionParser) ParseMultipleDivideModuloExpression(q *query.QueryMultipleDivideModuloExpression) error {
	err := ep.visitor.OnEnterMultipleDivideModuloExpression()
	if err != nil {
		return err
	}

	err = ep.ParsePowerOfExpression(&q.PowerOfExpression)
	if err != nil {
		return err
	}

	for _, pmdme := range q.PartialMultipleDivideModuloExpressions {
		err = ep.ParsePowerOfExpression(&pmdme.QueryPowerOfExpression)
		if err != nil {
			return err
		}

		err = ep.visitor.OnMultiplyDivideModuloOperator(pmdme.MultiplyDivideOperator)
		if err != nil {
			return err
		}
	}
	err = ep.visitor.OnExitMultipleDivideModuloExpression()
	if err != nil {
		return err
	}
	return nil
}

// ParseAddOrSubtractExpression parse add or subtract expression
func (ep *ExpressionParser) ParseAddOrSubtractExpression(q *query.QueryAddOrSubtractExpression) error {
	err := ep.visitor.OnEnterAddOrSubtractExpression()
	if err != nil {
		return err
	}

	err = ep.ParseMultipleDivideModuloExpression(&q.MultipleDivideModuloExpression)
	if err != nil {
		return err
	}
	for _, pase := range q.PartialAddOrSubtractExpression {
		err = ep.ParseMultipleDivideModuloExpression(&pase.MultipleDivideModuloExpression)
		if err != nil {
			return err
		}
		err = ep.visitor.OnAddOrSubtractOperator(pase.AddOrSubtractOperator)
		if err != nil {
			return err
		}
	}

	err = ep.visitor.OnExitAddOrSubtractExpression()
	if err != nil {
		return err
	}
	return nil
}

// ParseComparisonExpression parse comparison expression
func (ep *ExpressionParser) ParseComparisonExpression(q *query.QueryComparisonExpression) error {
	err := ep.visitor.OnEnterComparisonExpression()
	if err != nil {
		return err
	}

	err = ep.ParseAddOrSubtractExpression(&q.AddOrSubtractExpression)
	if err != nil {
		return err
	}

	for _, pce := range q.PartialComparisonExpressions {
		err = ep.visitor.OnComparisonOperator(pce.ComparisonOperator)
		if err != nil {
			return err
		}

		err := ep.ParseAddOrSubtractExpression(&pce.AddOrSubtractExpression)
		if err != nil {
			return err
		}
	}

	err = ep.visitor.OnExitComparisonExpression()
	if err != nil {
		return err
	}
	return nil
}

// ParseNotExpression parse not expression
func (ep *ExpressionParser) ParseNotExpression(q *query.QueryNotExpression) error {
	err := ep.visitor.OnEnterNotExpression(q.Not)
	if err != nil {
		return err
	}

	err = ep.ParseComparisonExpression(&q.ComparisonExpression)
	if err != nil {
		return err
	}

	err = ep.visitor.OnExitNotExpression(q.Not)
	if err != nil {
		return err
	}
	return err
}

// ParseXorExpression parse xor expression
func (ep *ExpressionParser) ParseXorExpression(q *query.QueryXorExpression) error {
	err := ep.visitor.OnEnterXorExpression()
	if err != nil {
		return err
	}

	for _, andExpr := range q.AndExpressions {
		err = ep.visitor.OnEnterAndExpression()
		if err != nil {
			return err
		}

		for _, notExpr := range andExpr.NotExpressions {
			err = ep.ParseNotExpression(&notExpr)
			if err != nil {
				return err
			}
		}

		err = ep.visitor.OnExitAndExpression()
		if err != nil {
			return err
		}
	}
	err = ep.visitor.OnExitXorExpression()
	if err != nil {
		return err
	}
	return nil
}

// ParseOrExpression parse or expression
func (ep *ExpressionParser) ParseOrExpression(q *query.QueryOrExpression) error {
	var err error
	err = ep.visitor.OnEnterOrExpression()
	if err != nil {
		return err
	}

	for _, xorExpr := range q.XorExpressions {
		err = ep.ParseXorExpression(&xorExpr)
		if err != nil {
			return err
		}
	}

	err = ep.visitor.OnExitOrExpression()
	if err != nil {
		return err
	}
	return nil
}

// ParseExpression parse expression
func (ep *ExpressionParser) ParseExpression(q *query.QueryExpression) error {
	err := ep.visitor.OnEnterExpression()
	if err != nil {
		return err
	}
	err = ep.ParseOrExpression(&q.OrExpression)
	if err != nil {
		return err
	}
	err = ep.visitor.OnExitExpression()
	if err != nil {
		return err
	}
	return nil
}
