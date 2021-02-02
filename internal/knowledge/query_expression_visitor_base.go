package knowledge

import (
	"github.com/clems4ever/go-graphkb/internal/query"
)

// ExpressionVisitorBase expression visitor base interface
type ExpressionVisitorBase struct{}

func (evb *ExpressionVisitorBase) OnEnterRelationshipsPattern() error {
	return nil
}
func (evb *ExpressionVisitorBase) OnExitRelationshipsPattern() error {
	return nil
}

func (evb *ExpressionVisitorBase) OnEnterNodePattern() error {
	return nil
}
func (evb *ExpressionVisitorBase) OnExitNodePattern() error {
	return nil
}
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
