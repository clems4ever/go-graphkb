package knowledge

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAndOrExpression(t *testing.T) {
	exprs := AndOrExpression{
		And: true,
		Children: []AndOrExpression{
			AndOrExpression{
				And: false,
				Children: []AndOrExpression{
					AndOrExpression{
						And:        false,
						Expression: "a0.type = 'ip'",
					},
				},
			},
			AndOrExpression{
				And: false,
				Children: []AndOrExpression{
					AndOrExpression{
						And:        false,
						Expression: "a0.type = 'device'",
					},
				},
			},
			AndOrExpression{
				And: false,
				Children: []AndOrExpression{
					AndOrExpression{
						And:        false,
						Expression: "a0.type = 'metascan_task_id'",
					},
				},
			},
			AndOrExpression{
				And: false,
				Children: []AndOrExpression{
					AndOrExpression{
						And:        false,
						Expression: "r0.type = 'observed'",
					},
				},
			},
			AndOrExpression{
				And: true,
				Children: []AndOrExpression{
					AndOrExpression{
						And:        false,
						Expression: "r0.from_id = a1.id",
					},
					AndOrExpression{
						And:        false,
						Expression: "r0.to_id = a0.id",
					},
				},
			},
			AndOrExpression{
				And: false,
				Children: []AndOrExpression{
					AndOrExpression{
						And:        false,
						Expression: "r1.type = 'scanned'",
					},
				},
			},
			AndOrExpression{
				And: true,
				Children: []AndOrExpression{
					AndOrExpression{
						And:        false,
						Expression: "r1.from_id = a2.id",
					},
					AndOrExpression{
						And:        false,
						Expression: "r1.to_id = a0.id",
					},
				},
			},
			AndOrExpression{
				And: true,
				Children: []AndOrExpression{
					AndOrExpression{
						And:        false,
						Expression: "",
					},
				},
			},
		},
	}

	x, err := UnwindOrExpressions(exprs)
	assert.NoError(t, err)
	assert.Len(t, x, 4)
}
