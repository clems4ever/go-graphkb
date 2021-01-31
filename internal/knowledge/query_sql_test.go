package knowledge

import (
	"strings"
	"testing"

	"github.com/clems4ever/go-graphkb/internal/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type QueryCase struct {
	Cypher   string
	SQL      string
	Error    string
	Selected bool
}

func TestQueryTranslation(t *testing.T) {
	cases := []QueryCase{
		QueryCase{
			Cypher: "MATCH (n:ip) RETURN n",
			SQL: `SELECT a0.id, a0.value, a0.type FROM assets a0
WHERE a0.type = 'ip'`,
		},
		QueryCase{
			Cypher: "MATCH (n:ip), (n:name) RETURN n",
			Error:  "Variable 'n' already defined with a different type",
		},
		QueryCase{
			Cypher: "MATCH (n:ip) RETURN n, n",
			SQL: `SELECT a0.id, a0.value, a0.type, a0.id, a0.value, a0.type FROM assets a0
WHERE a0.type = 'ip'`,
		},
		QueryCase{
			Cypher: "MATCH (n) WHERE n.value = 'prod' RETURN n",
			SQL:    "SELECT a0.id, a0.value, a0.type FROM assets a0\nWHERE a0.value = 'prod'",
		},
		QueryCase{
			Cypher: "MATCH (n) WHERE NOT n.value = 'prod' RETURN n",
			SQL:    "SELECT a0.id, a0.value, a0.type FROM assets a0\nWHERE NOT a0.value = 'prod'",
		},
		QueryCase{
			Cypher: "MATCH (n) WHERE NOT n.value = 'prod' AND n.value = 'preprod' RETURN n",
			SQL:    "SELECT a0.id, a0.value, a0.type FROM assets a0\nWHERE NOT a0.value = 'prod' AND a0.value = 'preprod'",
		},
		QueryCase{
			Cypher: "MATCH (n) WHERE n.value STARTS WITH 'prod' RETURN n",
			SQL:    "SELECT a0.id, a0.value, a0.type FROM assets a0\nWHERE a0.value LIKE 'prod%'",
		},
		QueryCase{
			Cypher: "MATCH (n) WHERE n.value ENDS WITH 'prod' RETURN n",
			SQL:    "SELECT a0.id, a0.value, a0.type FROM assets a0\nWHERE a0.value LIKE '%prod'",
		},
		QueryCase{
			Cypher: "MATCH (n) WHERE n.value CONTAINS 'prod' RETURN n",
			SQL:    "SELECT a0.id, a0.value, a0.type FROM assets a0\nWHERE a0.value LIKE '%prod%'",
		},
		QueryCase{
			Cypher: "MATCH (:variable)-[:has]->(n:name) RETURN n",
			SQL: `
SELECT a1.id, a1.value, a1.type FROM assets a0, assets a1, relations r0
WHERE (((a0.type = 'variable' AND a1.type = 'name') AND r0.type = 'has') AND (r0.from_id = a0.id AND r0.to_id = a1.id))`,
		},
		QueryCase{
			Cypher: "MATCH (:variable)<-[:has]-(n:name) RETURN n",
			SQL: `
SELECT a1.id, a1.value, a1.type FROM assets a0, assets a1, relations r0
WHERE (((a0.type = 'variable' AND a1.type = 'name') AND r0.type = 'has') AND (r0.from_id = a1.id AND r0.to_id = a0.id))`,
		},
		QueryCase{
			Cypher: "MATCH (v:variable)--(n:name) RETURN n",
			SQL: `
(SELECT a1.id, a1.value, a1.type FROM assets a0, assets a1, relations r0
WHERE ((a0.type = 'variable' AND a1.type = 'name') AND (r0.from_id = a0.id AND r0.to_id = a1.id)))
UNION ALL
(SELECT a1.id, a1.value, a1.type FROM assets a0, assets a1, relations r0
WHERE ((a0.type = 'variable' AND a1.type = 'name') AND (r0.from_id = a1.id AND r0.to_id = a0.id)))`,
		},
		QueryCase{
			Cypher: "MATCH (v:variable)-[r]-(n:name) RETURN n",
			SQL: `
(SELECT a1.id, a1.value, a1.type FROM assets a0, assets a1, relations r0
WHERE ((a0.type = 'variable' AND a1.type = 'name') AND (r0.from_id = a0.id AND r0.to_id = a1.id)))
UNION ALL
(SELECT a1.id, a1.value, a1.type FROM assets a0, assets a1, relations r0
WHERE ((a0.type = 'variable' AND a1.type = 'name') AND (r0.from_id = a1.id AND r0.to_id = a0.id)))`,
		},
		QueryCase{
			Cypher: "MATCH (v:variable)-[r]-(n:name) RETURN n LIMIT 10",
			SQL: `
(SELECT a1.id, a1.value, a1.type FROM assets a0, assets a1, relations r0
WHERE ((a0.type = 'variable' AND a1.type = 'name') AND (r0.from_id = a0.id AND r0.to_id = a1.id)))
UNION ALL
(SELECT a1.id, a1.value, a1.type FROM assets a0, assets a1, relations r0
WHERE ((a0.type = 'variable' AND a1.type = 'name') AND (r0.from_id = a1.id AND r0.to_id = a0.id)))
LIMIT 10`,
		},
		QueryCase{
			Cypher: "MATCH (v:variable)-[r]-(n:name) RETURN v.name, COUNT(n.name)",
			SQL: `
SELECT a0.name, COUNT(*) FROM
((SELECT a0.name, COUNT(*) FROM assets a0, assets a1, relations r0
WHERE ((a0.type = 'variable' AND a1.type = 'name') AND (r0.from_id = a0.id AND r0.to_id = a1.id)))
UNION ALL
(SELECT a0.name, COUNT(*) FROM assets a0, assets a1, relations r0
WHERE ((a0.type = 'variable' AND a1.type = 'name') AND (r0.from_id = a1.id AND r0.to_id = a0.id))))
GROUP BY a0.name`,
		},
		QueryCase{
			Cypher: "MATCH (v:variable)-[r]-(n:name) RETURN DISTINCT n.value LIMIT 10",
			SQL: `
(SELECT a1.value FROM assets a0, assets a1, relations r0
WHERE ((a0.type = 'variable' AND a1.type = 'name') AND (r0.from_id = a0.id AND r0.to_id = a1.id)))
UNION
(SELECT a1.value FROM assets a0, assets a1, relations r0
WHERE ((a0.type = 'variable' AND a1.type = 'name') AND (r0.from_id = a1.id AND r0.to_id = a0.id)))
LIMIT 10`,
		},
		QueryCase{
			Cypher: "MATCH (v:variable)-[r]-(n:name) RETURN v.name, COUNT(DISTINCT n.name)",
			SQL: `
SELECT a0.name, COUNT(DISTINCT a1.name) FROM
((SELECT a0.name, COUNT(DISTINCT a1.name) FROM assets a0, assets a1, relations r0
WHERE ((a0.type = 'variable' AND a1.type = 'name') AND (r0.from_id = a0.id AND r0.to_id = a1.id)))
UNION ALL
(SELECT a0.name, COUNT(DISTINCT a1.name) FROM assets a0, assets a1, relations r0
WHERE ((a0.type = 'variable' AND a1.type = 'name') AND (r0.from_id = a1.id AND r0.to_id = a0.id))))
GROUP BY a0.name`,
		},
		QueryCase{
			Cypher: "MATCH (v)-[r]-(n) RETURN n LIMIT 10",
			SQL: `
SELECT a1.id, a1.value, a1.type FROM assets a0, assets a1, relations r0
WHERE (r0.from_id = a0.id AND r0.to_id = a1.id)
LIMIT 10`,
		},
		QueryCase{
			Cypher: "MATCH (v:variable)<-[r]-(n:name), (v)-[r1]->(n) RETURN n",
			SQL: `
SELECT a1.id, a1.value, a1.type FROM assets a0, assets a1, relations r0, relations r1
WHERE (((a0.type = 'variable' AND a1.type = 'name') AND (r0.from_id = a1.id AND r0.to_id = a0.id)) AND (r1.from_id = a0.id AND r1.to_id = a1.id))`,
		},
		QueryCase{
			Cypher: "MATCH (:variable)<-[:has]-(n:name) RETURN n.value",
			SQL: `
SELECT a1.value FROM assets a0, assets a1, relations r0
WHERE (((a0.type = 'variable' AND a1.type = 'name') AND r0.type = 'has') AND (r0.from_id = a1.id AND r0.to_id = a0.id))`,
		},
		QueryCase{
			Cypher: "MATCH (v:variable)<-[r:has]-(n:name) RETURN v, r, n",
			SQL: `
SELECT a0.id, a0.value, a0.type, r0.from_id, r0.to_id, r0.type, a1.id, a1.value, a1.type FROM assets a0, assets a1, relations r0
WHERE (((a0.type = 'variable' AND a1.type = 'name') AND r0.type = 'has') AND (r0.from_id = a1.id AND r0.to_id = a0.id))`,
		},
		QueryCase{
			Cypher: "MATCH (v:variable)<-[r]-(n) RETURN v, r, n",
			SQL: `
SELECT a0.id, a0.value, a0.type, r0.from_id, r0.to_id, r0.type, a1.id, a1.value, a1.type FROM assets a0, assets a1, relations r0
WHERE (a0.type = 'variable' AND (r0.from_id = a1.id AND r0.to_id = a0.id))`,
		},
		QueryCase{
			Cypher: "MATCH (v:variable)<-[:has]-(:name)-[:is_in]->(:program) RETURN v",
			SQL: `
SELECT a0.id, a0.value, a0.type FROM assets a0, assets a1, assets a2, relations r0, relations r1
WHERE ((((((a0.type = 'variable' AND a1.type = 'name') AND a2.type = 'program') AND r0.type = 'has') AND (r0.from_id = a1.id AND r0.to_id = a0.id)) AND r1.type = 'is_in') AND (r1.from_id = a1.id AND r1.to_id = a2.id))`,
		},
		QueryCase{
			Cypher: `MATCH (p:port)<-[:bind]-(c:consul_service)-[:is_in]->(d:datacenter) WHERE d.value = 'pa4'
MATCH (c)-[:is_in]->(e:environment) WHERE e.value = 'preprod'
RETURN c`,
			SQL: `
SELECT a1.id, a1.value, a1.type FROM assets a0, assets a1, assets a2, assets a3, relations r0, relations r1, relations r2
WHERE ((((((((((a0.type = 'port' AND a1.type = 'consul_service') AND a2.type = 'datacenter') AND a3.type = 'environment') AND r0.type = 'bind') AND (r0.from_id = a1.id AND r0.to_id = a0.id)) AND r1.type = 'is_in') AND (r1.from_id = a1.id AND r1.to_id = a2.id)) AND r2.type = 'is_in') AND (r2.from_id = a1.id AND r2.to_id = a3.id)) AND (a2.value = 'pa4' AND a3.value = 'preprod'))`,
		},
		QueryCase{
			Cypher: `MATCH (p:port)<-[:bind]-(c:consul_service)-[:is_in]->(d:datacenter) WHERE d.value = 'pa4'
MATCH (c)-[:is_in]->(e:environment) WHERE e.value <> 'preprod'
RETURN c`,
			SQL: `
SELECT a1.id, a1.value, a1.type FROM assets a0, assets a1, assets a2, assets a3, relations r0, relations r1, relations r2
WHERE ((((((((((a0.type = 'port' AND a1.type = 'consul_service') AND a2.type = 'datacenter') AND a3.type = 'environment') AND r0.type = 'bind') AND (r0.from_id = a1.id AND r0.to_id = a0.id)) AND r1.type = 'is_in') AND (r1.from_id = a1.id AND r1.to_id = a2.id)) AND r2.type = 'is_in') AND (r2.from_id = a1.id AND r2.to_id = a3.id)) AND (a2.value = 'pa4' AND a3.value <> 'preprod'))`,
		},
		QueryCase{
			Cypher: "MATCH (:variable)<-[:has]-(n:name) RETURN n LIMIT 10",
			SQL: `
SELECT a1.id, a1.value, a1.type FROM assets a0, assets a1, relations r0
WHERE (((a0.type = 'variable' AND a1.type = 'name') AND r0.type = 'has') AND (r0.from_id = a1.id AND r0.to_id = a0.id))
LIMIT 10`,
		},
		QueryCase{
			Cypher: "MATCH (:variable)<-[:has]-(n:name) RETURN n SKIP 20 LIMIT 10",
			SQL: `
SELECT a1.id, a1.value, a1.type FROM assets a0, assets a1, relations r0
WHERE (((a0.type = 'variable' AND a1.type = 'name') AND r0.type = 'has') AND (r0.from_id = a1.id AND r0.to_id = a0.id))
LIMIT 10
OFFSET 20`,
		},
		QueryCase{
			Cypher: "MATCH (:variable)<-[:has]-(n:name) RETURN DISTINCT n",
			SQL: `
SELECT DISTINCT a1.id, a1.value, a1.type FROM assets a0, assets a1, relations r0
WHERE (((a0.type = 'variable' AND a1.type = 'name') AND r0.type = 'has') AND (r0.from_id = a1.id AND r0.to_id = a0.id))`,
		},
		QueryCase{
			Cypher: `MATCH (r:rack)<-[:is_in]-(cn:chef_name)-[:is_in]->(e:environment)
WHERE r.value = '01.04'
RETURN e.value, COUNT(cn.value)`,
			SQL: `
SELECT a2.value, COUNT(*) FROM assets a0, assets a1, assets a2, relations r0, relations r1
WHERE (((((((a0.type = 'rack' AND a1.type = 'chef_name') AND a2.type = 'environment') AND r0.type = 'is_in') AND (r0.from_id = a1.id AND r0.to_id = a0.id)) AND r1.type = 'is_in') AND (r1.from_id = a1.id AND r1.to_id = a2.id)) AND a0.value = '01.04')
GROUP BY a2.value`,
		},
		QueryCase{
			Cypher: `MATCH (r:rack)<-[:is_in]-(cn:chef_name) RETURN COUNT(cn.value)`,
			SQL: `
SELECT COUNT(*) FROM assets a0, assets a1, relations r0
WHERE (((a0.type = 'rack' AND a1.type = 'chef_name') AND r0.type = 'is_in') AND (r0.from_id = a1.id AND r0.to_id = a0.id))`,
		},
		QueryCase{
			Cypher: "MATCH (v:variable)-[:has]->(n:name) WHERE v.value = '0x16' AND (n.value = 'myvar' OR n.value = 'myvar2') RETURN n",
			SQL: `
SELECT a1.id, a1.value, a1.type FROM assets a0, assets a1, relations r0
WHERE ((((a0.type = 'variable' AND a1.type = 'name') AND r0.type = 'has') AND (r0.from_id = a0.id AND r0.to_id = a1.id)) AND (a0.value = '0x16' AND a1.value = 'myvar' OR a1.value = 'myvar2'))`,
		},
		QueryCase{
			Cypher: `
MATCH (ip:ip)<-[:observed]-(:device)
WHERE (ip)<-[:scanned]-(:task)
RETURN ip`,
			SQL: `
SELECT a0.id, a0.value, a0.type FROM assets a0, assets a1, assets a2, relations r0, relations r1
WHERE ((((((a0.type = 'ip' AND a1.type = 'device') AND a2.type = 'task') AND r0.type = 'observed') AND (r0.from_id = a1.id AND r0.to_id = a0.id)) AND r1.type = 'scanned') AND (r1.from_id = a2.id AND r1.to_id = a0.id))
			`,
		},
	}

	selectionEnabled := false
	for _, c := range cases {
		selectionEnabled = selectionEnabled || c.Selected
	}

	for _, c := range cases {
		if selectionEnabled && !c.Selected {
			continue
		}
		t.Run(c.Cypher, func(t *testing.T) {
			translator := NewSQLQueryTranslator()
			q, err := query.TransformCypher(c.Cypher)
			require.NoError(t, err)

			sql, err := translator.Translate(q)
			if c.Error != "" {
				require.Error(t, err, "Error on test case %s", c.Cypher)
				if err != nil {
					assert.Equal(t, c.Error, err.Error(), "Error on test case %s", c.Cypher)
				}
			} else {
				require.NoError(t, err, "Error on test case %s", c.Cypher)
				assert.Equal(t, strings.TrimSpace(c.SQL), sql.Query, "Error on test case %s", c.Cypher)
			}
		})
	}
}

func TestUnwindOrExpressions(t *testing.T) {
	And := func(e ...AndOrExpression) AndOrExpression {
		return AndOrExpression{
			And:      true,
			Children: e,
		}
	}
	Or := func(e ...AndOrExpression) AndOrExpression {
		return AndOrExpression{
			And:      false,
			Children: e,
		}
	}
	Expr := func(e string) AndOrExpression {
		return AndOrExpression{
			Expression: e,
		}
	}

	t.Run("single", func(t *testing.T) {
		exprA := Expr("a")

		unwoundExpr, err := UnwindOrExpressions(exprA)
		require.NoError(t, err)
		require.Len(t, unwoundExpr, 1)

		expected := And(exprA)
		assert.Equal(t, expected, unwoundExpr[0])
	})

	t.Run("and", func(t *testing.T) {
		exprA := Expr("a")
		exprB := Expr("b")
		expr := And(exprA, exprB)

		unwoundExpr, err := UnwindOrExpressions(expr)
		require.NoError(t, err)
		require.Len(t, unwoundExpr, 1)

		expected := And(And(exprA), And(exprB))
		assert.Equal(t, expected, unwoundExpr[0])
	})

	t.Run("and_and_and", func(t *testing.T) {
		exprA := Expr("a")
		exprB := Expr("b")
		exprC := Expr("c")
		exprD := Expr("d")
		expr := And(exprA, And(exprB, And(exprC, exprD)))

		unwoundExpr, err := UnwindOrExpressions(expr)
		require.NoError(t, err)
		require.Len(t, unwoundExpr, 1)

		expected := And(And(exprA), And(And(exprB), And(And(exprC), And(exprD))))
		assert.Equal(t, expected, unwoundExpr[0])
	})

	t.Run("or", func(t *testing.T) {
		exprA := Expr("a")
		exprB := Expr("b")
		expr := Or(exprA, exprB)

		unwoundExpr, err := UnwindOrExpressions(expr)
		require.NoError(t, err)
		require.Len(t, unwoundExpr, 2)

		assert.Equal(t, And(exprA), unwoundExpr[0])
		assert.Equal(t, And(exprB), unwoundExpr[1])
	})

	t.Run("and_or", func(t *testing.T) {
		exprA := Expr("a")
		exprB := Expr("b")
		exprC := Expr("c")
		exprD := Expr("d")
		expr := And(Or(exprA, exprB), Or(exprC, exprD))

		unwoundExpr, err := UnwindOrExpressions(expr)
		require.NoError(t, err)
		require.Len(t, unwoundExpr, 4)

		assert.Equal(t, And(And(exprA), And(exprC)), unwoundExpr[0])
		assert.Equal(t, And(And(exprA), And(exprD)), unwoundExpr[1])
		assert.Equal(t, And(And(exprB), And(exprC)), unwoundExpr[2])
		assert.Equal(t, And(And(exprB), And(exprD)), unwoundExpr[3])
	})

	t.Run("or_and", func(t *testing.T) {
		exprA := Expr("a")
		exprB := Expr("b")
		exprC := Expr("c")
		exprD := Expr("d")
		expr := Or(And(exprA, exprB), And(exprC, exprD))

		unwoundExpr, err := UnwindOrExpressions(expr)
		require.NoError(t, err)
		require.Len(t, unwoundExpr, 2)

		assert.Equal(t, And(And(exprA), And(exprB)), unwoundExpr[0])
		assert.Equal(t, And(And(exprC), And(exprD)), unwoundExpr[1])
	})

	t.Run("complex", func(t *testing.T) {
		exprA := Expr("a")
		exprB := Expr("b")
		exprC := Expr("c")
		exprD := Expr("d")
		exprE := Expr("e")
		expr := Or(Or(exprA, exprB), And(exprC, Or(exprD, exprE)))

		unwoundExpr, err := UnwindOrExpressions(expr)
		require.NoError(t, err)
		require.Len(t, unwoundExpr, 4)

		assert.Equal(t, And(exprA), unwoundExpr[0])
		assert.Equal(t, And(exprB), unwoundExpr[1])
		assert.Equal(t, And(And(exprC), And(exprD)), unwoundExpr[2])
		assert.Equal(t, And(And(exprC), And(exprE)), unwoundExpr[3])
	})
}
