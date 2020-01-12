package knowledge

import (
	"testing"

	"github.com/antlr/antlr4/runtime/Go/antlr"
	"github.com/clems4ever/go-graphkb/internal/parser"
	"github.com/clems4ever/go-graphkb/internal/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func CypherToExpr(cypher string) query.QueryExpression {
	is := antlr.NewInputStream(cypher)
	lexer := parser.NewCypherLexer(is)
	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	p := parser.NewCypherParser(stream)

	l := query.NewCypherVisitor()
	queryExpression := l.Visit(p.OC_Expression())
	return queryExpression.(query.QueryExpression)
}

type ExpressionTestCase struct {
	Cypher   string
	SQL      string
	Selected bool
}

func TestShouldBuildExpression(t *testing.T) {
	anySelected := false
	for _, tc := range testCases {
		if tc.Selected {
			anySelected = true
		}
	}

	for _, tc := range testCases {
		if anySelected && !tc.Selected {
			continue
		}
		t.Run(tc.Cypher, func(t *testing.T) {
			qg := NewQueryGraph()
			ep := NewExpressionBuilder(&qg)

			_, _, err := qg.PushNode(query.QueryNodePattern{
				Variable: "a",
			})
			require.NoError(t, err)
			_, _, err = qg.PushNode(query.QueryNodePattern{
				Variable: "b",
			})
			require.NoError(t, err)

			expr := CypherToExpr(tc.Cypher)

			sql, err := ep.Build(&expr)
			require.NoError(t, err)
			assert.Equal(t, tc.SQL, sql)
		})
	}
}

var testCases = []ExpressionTestCase{
	ExpressionTestCase{
		Cypher: "(a.value)",
		SQL:    "(a0.value)",
	},
	ExpressionTestCase{
		Cypher: "a.value OR b.value",
		SQL:    "a0.value OR a1.value",
	},
	ExpressionTestCase{
		Cypher: "a.value AND b.value",
		SQL:    "a0.value AND a1.value",
	},
	ExpressionTestCase{
		Cypher: "NOT a.value",
		SQL:    "NOT a0.value",
	},
	ExpressionTestCase{
		Cypher: "NOT NOT a.value",
		SQL:    "a0.value",
	},
	ExpressionTestCase{
		Cypher: "a.value CONTAINS 'abc'",
		SQL:    "a0.value LIKE '%abc%'",
	},
	ExpressionTestCase{
		Cypher: "a.value STARTS WITH 'abc'",
		SQL:    "a0.value LIKE 'abc%'",
	},
	ExpressionTestCase{
		Cypher: "a.value ENDS WITH 'abc'",
		SQL:    "a0.value LIKE '%abc'",
	},
	ExpressionTestCase{
		Cypher: "COUNT(a.value)",
		SQL:    "COUNT(a0.value)",
	},
	ExpressionTestCase{
		Cypher: "a.value < b.value",
		SQL:    "a0.value < a1.value",
	},
	ExpressionTestCase{
		Cypher: "a.value = b.value",
		SQL:    "a0.value = a1.value",
	},
	ExpressionTestCase{
		Cypher: "a.value <> b.value",
		SQL:    "a0.value <> a1.value",
	},
	ExpressionTestCase{
		Cypher: "a.value = 'abc'",
		SQL:    "a0.value = 'abc'",
	},
	ExpressionTestCase{
		Cypher: "a",
		SQL:    "a0.*",
	},
	ExpressionTestCase{
		Cypher: "a.value",
		SQL:    "a0.value",
	},
	ExpressionTestCase{
		Cypher: "'abc'",
		SQL:    "'abc'",
	},
	ExpressionTestCase{
		Cypher: "2",
		SQL:    "2",
	},
	ExpressionTestCase{
		Cypher: "2.5",
		SQL:    "2.500000",
	},
	ExpressionTestCase{
		Cypher: "true",
		SQL:    "true",
	},
	ExpressionTestCase{
		Cypher: "false",
		SQL:    "false",
	},
	ExpressionTestCase{
		Cypher: "TRUE",
		SQL:    "true",
	},
	ExpressionTestCase{
		Cypher: "FALSE",
		SQL:    "false",
	},
}
