package knowledge

import (
	"fmt"
	"strings"
)

// AndOrExpression represent a AND or OR expression
type AndOrExpression struct {
	And        bool // true for And and false for Or
	Children   []AndOrExpression
	Expression string
}

func (aoe AndOrExpression) String() string {
	if aoe.Expression != "" {
		return aoe.Expression
	}

	if len(aoe.Children) > 0 {
		op := " AND "
		if !aoe.And {
			op = " OR "
		}
		exprs := make([]string, 0)
		for _, e := range aoe.Children {
			exprs = append(exprs, e.String())
		}
		return strings.Join(exprs, op)
	}
	return "(empty)"
}

// BuildAndOrExpression build a string representation of AndOrExpression.
func BuildAndOrExpression(tree AndOrExpression) (string, error) {
	if tree.Expression != "" {
		return tree.Expression, nil
	} else if tree.And {
		exprs := make([]string, 0)
		for i := range tree.Children {
			expr, err := BuildAndOrExpression(tree.Children[i])
			if err != nil {
				return "", err
			}
			if expr != "" {
				exprs = append(exprs, expr)
			}
		}
		if len(exprs) > 1 {
			return fmt.Sprintf("(%s)", strings.Join(exprs, " AND ")), nil
		}
		return strings.Join(exprs, " AND "), nil
	} else if !tree.And {
		exprs := make([]string, 0)
		for i := range tree.Children {
			expr, err := BuildAndOrExpression(tree.Children[i])
			if err != nil {
				return "", err
			}

			if expr != "" {
				exprs = append(exprs, expr)
			}
		}
		if len(exprs) > 1 {
			return fmt.Sprintf("(%s)", strings.Join(exprs, " OR ")), nil
		}
		return strings.Join(exprs, " OR "), nil
	}
	return "", nil
}

// CrossProductExpressions computes the cross product of 2 sets of expressions. This is used to transform OR expressions into a union of AND expressions.
func CrossProductExpressions(and1 []AndOrExpression, and2 []AndOrExpression) []AndOrExpression {
	outExpr := []AndOrExpression{}
	for i := range and1 {
		for j := range and2 {
			children := AndOrExpression{
				And:      true,
				Children: []AndOrExpression{and1[i], and2[j]},
			}
			outExpr = append(outExpr, children)
		}
	}
	return outExpr
}

// UnwindOrExpressions in order to transform query with or relations into a union
// query which is more performant, an AndOrExpression is transformed into a list of AndExpressions
func UnwindOrExpressions(tree AndOrExpression) ([]AndOrExpression, error) {
	if tree.Expression != "" {
		child := AndOrExpression{Children: []AndOrExpression{tree}, And: true}
		return []AndOrExpression{child}, nil
	} else if !tree.And {
		exprs := []AndOrExpression{}
		for i := range tree.Children {
			nestedExpressions, err := UnwindOrExpressions(tree.Children[i])
			if err != nil {
				return nil, err
			}
			exprs = append(exprs, nestedExpressions...)
		}
		return exprs, nil
	} else if tree.And {
		exprs := []AndOrExpression{}
		for i := range tree.Children {
			expr, err := UnwindOrExpressions(tree.Children[i])
			if err != nil {
				return nil, err
			}

			if len(exprs) == 0 {
				exprs = append(exprs, expr...)
			} else {
				exprs = CrossProductExpressions(exprs, expr)
			}
		}
		return exprs, nil
	}
	return nil, fmt.Errorf("Unable to detect kind of node")
}
