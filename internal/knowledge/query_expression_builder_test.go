package knowledge

import (
	"testing"

	"github.com/antlr/antlr4/runtime/Go/antlr/v4"
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
			}, MatchScope)
			require.NoError(t, err)
			_, _, err = qg.PushNode(query.QueryNodePattern{
				Variable: "b",
			}, MatchScope)
			require.NoError(t, err)

			require.NoError(t, err)
			_, _, err = qg.PushRelation(query.QueryRelationshipPattern{
				RelationshipDetail: &query.QueryRelationshipDetail{
					Variable: "r",
				},
			}, 0, 0, MatchScope)
			require.NoError(t, err)

			expr := CypherToExpr(tc.Cypher)

			sql, err := ep.Build(&expr)
			require.NoError(t, err)
			assert.Equal(t, tc.SQL, sql)
		})
	}
}

var testCases = []ExpressionTestCase{
	{
		Cypher: "(a.value)",
		SQL:    "(a0.value)",
	},
	{
		Cypher: "a.value OR b.value",
		SQL:    "a0.value OR a1.value",
	},
	{
		Cypher: "a.value AND b.value",
		SQL:    "a0.value AND a1.value",
	},
	{
		Cypher: "NOT a.value",
		SQL:    "NOT a0.value",
	},
	{
		Cypher: "NOT NOT a.value",
		SQL:    "a0.value",
	},
	{
		Cypher: "a.value CONTAINS 'abc'",
		SQL:    "a0.value LIKE '%abc%'",
	},
	{
		Cypher: "a.value STARTS WITH 'abc'",
		SQL:    "a0.value LIKE 'abc%'",
	},
	{
		Cypher: "a.value ENDS WITH 'abc'",
		SQL:    "a0.value LIKE '%abc'",
	},
	{
		Cypher: "COUNT(a.value)",
		SQL:    "COUNT(*)",
	},
	{
		Cypher: "a.value < b.value",
		SQL:    "a0.value < a1.value",
	},
	{
		Cypher: "a.value = b.value",
		SQL:    "a0.value = a1.value",
	},
	{
		Cypher: "a.value <> b.value",
		SQL:    "a0.value <> a1.value",
	},
	{
		Cypher: "a.value = 'abc'",
		SQL:    "a0.value = 'abc'",
	},
	{
		Cypher: "a",
		SQL:    "a0.id, a0.value, a0.type",
	},
	{
		Cypher: "r",
		SQL:    "r0.id, r0.from_id, r0.to_id, r0.type",
	},
	{
		Cypher: "a.value",
		SQL:    "a0.value",
	},
	{
		Cypher: "'abc'",
		SQL:    "'abc'",
	},
	{
		Cypher: "2",
		SQL:    "2",
	},
	{
		Cypher: "2.5",
		SQL:    "2.500000",
	},
	{
		Cypher: "true",
		SQL:    "true",
	},
	{
		Cypher: "false",
		SQL:    "false",
	},
	{
		Cypher: "TRUE",
		SQL:    "true",
	},
	{
		Cypher: "FALSE",
		SQL:    "false",
	},
}
