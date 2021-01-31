package knowledge

import (
	"fmt"
	"strings"

	"github.com/clems4ever/go-graphkb/internal/query"
)

type ExpressionType int

const (
	NodeExprType     ExpressionType = iota
	EdgeExprType     ExpressionType = iota
	PropertyExprType ExpressionType = iota
)

type ExpressionVisitor interface {
	OnEnterPropertyOrLabelsExpression(e query.QueryPropertyOrLabelsExpression) error
	OnExitPropertyOrLabelsExpression(e query.QueryPropertyOrLabelsExpression) error

	OnEnterStringListNullOperatorExpression(e query.QueryStringListNullOperatorExpression) error
	OnExitStringListNullOperatorExpression(e query.QueryStringListNullOperatorExpression) error

	OnVariable(name string) error
	OnVariablePropertiesPath(propertiesPath []string) error

	OnStringLiteral(value string) error
	OnDoubleLiteral(value float64) error
	OnIntegerLiteral(value int64) error
	OnBooleanLiteral(value bool) error

	OnEnterFunctionInvocation(name string, distinct bool) error
	OnExitFunctionInvocation(name string, distinct bool) error

	OnEnterParenthesizedExpression() error
	OnExitParenthesizedExpression() error

	OnStringOperator(operator query.StringOperator) error

	OnEnterUnaryExpression() error
	OnExitUnaryExpression() error

	OnEnterPowerOfExpression() error
	OnExitPowerOfExpression() error

	OnEnterMultipleDivideModuloExpression() error
	OnExitMultipleDivideModuloExpression() error
	OnMultiplyDivideModuloOperator(operator query.MultiplyDivideModuloOperator) error

	OnEnterAddOrSubtractExpression() error
	OnExitAddOrSubtractExpression() error
	OnAddOrSubtractOperator(operator query.AddOrSubtractOperator) error

	OnEnterComparisonExpression() error
	OnExitComparisonExpression() error
	OnComparisonOperator(operator query.ComparisonOperator) error

	OnEnterNotExpression(not bool) error
	OnExitNotExpression(not bool) error

	OnEnterAndExpression() error
	OnExitAndExpression() error

	OnEnterXorExpression() error
	OnExitXorExpression() error

	OnEnterOrExpression() error
	OnExitOrExpression() error

	OnEnterExpression() error
	OnExitExpression() error
}

type ExpressionParser struct {
	visitor ExpressionVisitor
}

func NewExpressionParser(visitor ExpressionVisitor) *ExpressionParser {
	return &ExpressionParser{
		visitor: visitor,
	}
}

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
	} else {
		return fmt.Errorf("Unable to parse property or labels expression")
	}

	err = ep.visitor.OnExitPropertyOrLabelsExpression(*q)
	if err != nil {
		return err
	}

	return nil
}

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

func (ep *ExpressionParser) ParseUnaryAddOrSubtractExpression(q *query.QueryUnaryAddOrSubtractExpression) error {
	err := ep.ParseStringListNullOperatorExpression(&q.StringListNullOperatorExpression)
	if err != nil {
		return err
	}
	return nil
}

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

type ExpressionVisitorBase struct{}

func (evb *ExpressionVisitorBase) OnEnterPropertyOrLabelsExpression(e query.QueryPropertyOrLabelsExpression) error {
	return nil
}
func (evb *ExpressionVisitorBase) OnExitPropertyOrLabelsExpression(e query.QueryPropertyOrLabelsExpression) error {
	return nil
}
func (evb *ExpressionVisitorBase) OnEnterStringListNullOperatorExpression(e query.QueryStringListNullOperatorExpression) error {
	return nil
}
func (evb *ExpressionVisitorBase) OnExitStringListNullOperatorExpression(e query.QueryStringListNullOperatorExpression) error {
	return nil
}
func (evb *ExpressionVisitorBase) OnVariable(name string) error                           { return nil }
func (evb *ExpressionVisitorBase) OnVariablePropertiesPath(propertiesPath []string) error { return nil }
func (evb *ExpressionVisitorBase) OnStringLiteral(value string) error                     { return nil }
func (evb *ExpressionVisitorBase) OnDoubleLiteral(value float64) error                    { return nil }
func (evb *ExpressionVisitorBase) OnIntegerLiteral(value int64) error                     { return nil }
func (evb *ExpressionVisitorBase) OnBooleanLiteral(value bool) error                      { return nil }
func (evb *ExpressionVisitorBase) OnEnterFunctionInvocation(name string, distinct bool) error {
	return nil
}
func (evb *ExpressionVisitorBase) OnExitFunctionInvocation(name string, distinct bool) error {
	return nil
}
func (evb *ExpressionVisitorBase) OnEnterParenthesizedExpression() error                { return nil }
func (evb *ExpressionVisitorBase) OnExitParenthesizedExpression() error                 { return nil }
func (evb *ExpressionVisitorBase) OnStringOperator(operator query.StringOperator) error { return nil }
func (evb *ExpressionVisitorBase) OnEnterUnaryExpression() error                        { return nil }
func (evb *ExpressionVisitorBase) OnExitUnaryExpression() error                         { return nil }
func (evb *ExpressionVisitorBase) OnEnterPowerOfExpression() error                      { return nil }
func (evb *ExpressionVisitorBase) OnExitPowerOfExpression() error                       { return nil }
func (evb *ExpressionVisitorBase) OnEnterMultipleDivideModuloExpression() error         { return nil }
func (evb *ExpressionVisitorBase) OnExitMultipleDivideModuloExpression() error          { return nil }
func (evb *ExpressionVisitorBase) OnMultiplyDivideModuloOperator(operator query.MultiplyDivideModuloOperator) error {
	return nil
}
func (evb *ExpressionVisitorBase) OnEnterAddOrSubtractExpression() error { return nil }
func (evb *ExpressionVisitorBase) OnExitAddOrSubtractExpression() error  { return nil }
func (evb *ExpressionVisitorBase) OnAddOrSubtractOperator(operator query.AddOrSubtractOperator) error {
	return nil
}
func (evb *ExpressionVisitorBase) OnEnterComparisonExpression() error { return nil }
func (evb *ExpressionVisitorBase) OnExitComparisonExpression() error  { return nil }
func (evb *ExpressionVisitorBase) OnComparisonOperator(operator query.ComparisonOperator) error {
	return nil
}
func (evb *ExpressionVisitorBase) OnEnterNotExpression(not bool) error { return nil }
func (evb *ExpressionVisitorBase) OnExitNotExpression(not bool) error  { return nil }
func (evb *ExpressionVisitorBase) OnEnterAndExpression() error         { return nil }
func (evb *ExpressionVisitorBase) OnExitAndExpression() error          { return nil }
func (evb *ExpressionVisitorBase) OnEnterXorExpression() error         { return nil }
func (evb *ExpressionVisitorBase) OnExitXorExpression() error          { return nil }
func (evb *ExpressionVisitorBase) OnEnterOrExpression() error          { return nil }
func (evb *ExpressionVisitorBase) OnExitOrExpression() error           { return nil }
func (evb *ExpressionVisitorBase) OnEnterExpression() error            { return nil }
func (evb *ExpressionVisitorBase) OnExitExpression() error             { return nil }
