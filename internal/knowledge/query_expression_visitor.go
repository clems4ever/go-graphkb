package knowledge

import (
	"github.com/clems4ever/go-graphkb/internal/query"
)

// ExpressionVisitor a visitor of expression
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

	OnEnterRelationshipsPattern(q query.QueryRelationshipsPattern, id int) error
	OnExitRelationshipsPattern(q query.QueryRelationshipsPattern, id int) error

	OnNodePattern(q query.QueryNodePattern) error
	OnRelationshipPattern(q query.QueryRelationshipPattern) error

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
