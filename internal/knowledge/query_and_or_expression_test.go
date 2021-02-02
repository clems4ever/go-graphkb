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
