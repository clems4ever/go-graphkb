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

func (aoe AndOrExpression) stringInternal(outer bool) string {
	if aoe.Expression != "" {
		return aoe.Expression
	} else if len(aoe.Children) == 1 {
		return aoe.Children[0].stringInternal(false)
	} else if len(aoe.Children) > 1 && aoe.And {
		children := []string{}
		for _, c := range aoe.Children {
			children = append(children, c.stringInternal(false))
		}
		if outer {
			return strings.Join(children, " AND ")
		}
		return fmt.Sprintf("(%s)", strings.Join(children, " AND "))
	} else if len(aoe.Children) > 1 && !aoe.And {
		children := []string{}
		for _, c := range aoe.Children {
			children = append(children, c.stringInternal(false))
		}
		if outer {
			return strings.Join(children, " OR ")
		}
		return fmt.Sprintf("(%s)", strings.Join(children, " OR "))
	}

	return ""
}

func (aoe AndOrExpression) String() string {
	return aoe.stringInternal(true)
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

func FlattenAndOrExpressions(tree AndOrExpression) (AndOrExpression, error) {
	if tree.Expression != "" {
		return tree, nil
	} else if len(tree.Children) == 1 {
		c, err := FlattenAndOrExpressions(tree.Children[0])
		if err != nil {
			return AndOrExpression{}, err
		}
		return c, nil
	} else if len(tree.Children) > 1 && !tree.And {
		orExpr := AndOrExpression{And: false, Children: []AndOrExpression{}}
		for _, c := range tree.Children {
			fc, err := FlattenAndOrExpressions(c)
			if err != nil {
				return AndOrExpression{}, err
			}
			if fc.Expression != "" || len(fc.Children) > 0 && fc.And {
				orExpr.Children = append(orExpr.Children, fc)
			} else if len(fc.Children) > 0 && !fc.And {
				orExpr.Children = append(orExpr.Children, fc.Children...)
			}
		}
		return orExpr, nil
	} else if len(tree.Children) > 1 && tree.And {
		andExpr := AndOrExpression{And: true, Children: []AndOrExpression{}}
		for _, c := range tree.Children {
			fc, err := FlattenAndOrExpressions(c)
			if err != nil {
				return AndOrExpression{}, err
			}
			if fc.Expression != "" || len(fc.Children) > 0 && !fc.And {
				andExpr.Children = append(andExpr.Children, fc)
			} else if len(fc.Children) > 0 && fc.And {
				andExpr.Children = append(andExpr.Children, fc.Children...)
			}
		}
		return andExpr, nil
	}

	return AndOrExpression{}, fmt.Errorf("An AndOrExpression should have an expression of at least one child")
}
