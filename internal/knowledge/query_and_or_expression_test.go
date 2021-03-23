package knowledge

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAndOrExpression(t *testing.T) {
	exprs := AndOrExpression{
		And: true,
		Children: []AndOrExpression{
			{
				And: false,
				Children: []AndOrExpression{
					{
						And:        false,
						Expression: "a0.type = 'ip'",
					},
				},
			},
			{
				And: false,
				Children: []AndOrExpression{
					{
						And:        false,
						Expression: "a0.type = 'device'",
					},
				},
			},
			{
				And: false,
				Children: []AndOrExpression{
					{
						And:        false,
						Expression: "a0.type = 'metascan_task_id'",
					},
				},
			},
			{
				And: false,
				Children: []AndOrExpression{
					{
						And:        false,
						Expression: "r0.type = 'observed'",
					},
				},
			},
			{
				And: true,
				Children: []AndOrExpression{
					{
						And:        false,
						Expression: "r0.from_id = a1.id",
					},
					{
						And:        false,
						Expression: "r0.to_id = a0.id",
					},
				},
			},
			{
				And: false,
				Children: []AndOrExpression{
					{
						And:        false,
						Expression: "r1.type = 'scanned'",
					},
				},
			},
			{
				And: true,
				Children: []AndOrExpression{
					{
						And:        false,
						Expression: "r1.from_id = a2.id",
					},
					{
						And:        false,
						Expression: "r1.to_id = a0.id",
					},
				},
			},
		},
	}

	x, err := UnwindOrExpressions(exprs)
	assert.NoError(t, err)
	assert.Len(t, x, 1)
}

func TestFlattenAndOrExpression_Init(t *testing.T) {
	exprs := AndOrExpression{
		And:      true,
		Children: []AndOrExpression{{Expression: "a"}, {Expression: "b"}},
	}
	actualExpr, err := FlattenAndOrExpressions(exprs)
	assert.NoError(t, err)
	assert.Equal(t, AndOrExpression{
		And:      true,
		Children: []AndOrExpression{{Expression: "a"}, {Expression: "b"}}},
		actualExpr)
	assert.Equal(t, "a AND b", actualExpr.String())
}

func TestFlattenAndOrExpression_FlattenAnd(t *testing.T) {
	exprs := AndOrExpression{
		And: true,
		Children: []AndOrExpression{
			{Expression: "a"},
			{And: true, Children: []AndOrExpression{{Expression: "b"}, {Expression: "c"}}},
		},
	}
	actualExpr, err := FlattenAndOrExpressions(exprs)
	assert.NoError(t, err)
	assert.Equal(t, AndOrExpression{
		And: true,
		Children: []AndOrExpression{
			{Expression: "a"}, {Expression: "b"}, {Expression: "c"},
		},
	}, actualExpr)
	assert.Equal(t, "a AND b AND c", actualExpr.String())
}

func TestFlattenAndOrExpression_FlattenMultipleAnds(t *testing.T) {
	exprs := AndOrExpression{
		And: true,
		Children: []AndOrExpression{
			{And: true, Children: []AndOrExpression{{Expression: "a"}}},
			{And: true, Children: []AndOrExpression{{Expression: "b"}, {Expression: "c"}}},
			{And: true, Children: []AndOrExpression{{And: true, Children: []AndOrExpression{{Expression: "d"}}}}},
		},
	}
	actualExpr, err := FlattenAndOrExpressions(exprs)
	assert.NoError(t, err)
	assert.Equal(t, AndOrExpression{
		And: true,
		Children: []AndOrExpression{
			{Expression: "a"}, {Expression: "b"}, {Expression: "c"}, {Expression: "d"},
		},
	}, actualExpr)
	assert.Equal(t, "a AND b AND c AND d", actualExpr.String())
}

func TestFlattenAndOrExpression_FlattenAndsAndOrsIntoAnds(t *testing.T) {
	expr := AndOrExpression{
		And: true,
		Children: []AndOrExpression{
			{And: true, Children: []AndOrExpression{{Expression: "a"}}},
			{And: true, Children: []AndOrExpression{{Expression: "b"}, {Expression: "c"}}},
			{And: false, Children: []AndOrExpression{
				{And: true, Children: []AndOrExpression{
					{And: true, Children: []AndOrExpression{{Expression: "d"}}},
					{Expression: "e"}},
				}},
			},
		},
	}

	expExpr := AndOrExpression{
		And: true,
		Children: []AndOrExpression{
			{Expression: "a"},
			{Expression: "b"},
			{Expression: "c"},
			{Expression: "d"},
			{Expression: "e"},
		},
	}

	actualExpr, err := FlattenAndOrExpressions(expr)
	assert.NoError(t, err)

	assert.Equal(t, expExpr, actualExpr)
	assert.Equal(t, "a AND b AND c AND d AND e", actualExpr.String())
}

func TestFlattenAndOrExpression_FlattenOrs(t *testing.T) {
	expr := AndOrExpression{
		And: false,
		Children: []AndOrExpression{
			{And: false, Children: []AndOrExpression{{Expression: "a"}}},
			{And: false, Children: []AndOrExpression{{Expression: "b"}, {Expression: "c"}}},
			{And: false, Children: []AndOrExpression{
				{And: false, Children: []AndOrExpression{
					{And: false, Children: []AndOrExpression{{Expression: "d"}}},
					{Expression: "e"}},
				}},
			},
		},
	}

	expExpr := AndOrExpression{
		And: false,
		Children: []AndOrExpression{
			{Expression: "a"},
			{Expression: "b"},
			{Expression: "c"},
			{Expression: "d"},
			{Expression: "e"},
		},
	}

	actualExpr, err := FlattenAndOrExpressions(expr)
	assert.NoError(t, err)

	assert.Equal(t, expExpr, actualExpr)
	assert.Equal(t, "a OR b OR c OR d OR e", actualExpr.String())
}

func TestFlattenAndOrExpression_FlattenAndsAndOrsIntoAndsAndOrs(t *testing.T) {
	expr := AndOrExpression{
		And: true,
		Children: []AndOrExpression{
			{And: true, Children: []AndOrExpression{{Expression: "a"}}},
			{And: true, Children: []AndOrExpression{{Expression: "b"}, {Expression: "c"}}},
			{And: false, Children: []AndOrExpression{
				{And: true, Children: []AndOrExpression{{Expression: "d"}}},
				{Expression: "e"},
			},
			},
		},
	}

	expExpr := AndOrExpression{
		And: true,
		Children: []AndOrExpression{
			{Expression: "a"},
			{Expression: "b"},
			{Expression: "c"},
			{And: false, Children: []AndOrExpression{{Expression: "d"}, {Expression: "e"}}},
		},
	}

	actualExpr, err := FlattenAndOrExpressions(expr)
	assert.NoError(t, err)

	assert.Equal(t, expExpr, actualExpr)
	assert.Equal(t, "a AND b AND c AND (d OR e)", actualExpr.String())
}
