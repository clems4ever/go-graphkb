package knowledge

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildBasicSingleSQLSelect_Distinct(t *testing.T) {
	test := func(distinct bool, expected string) func(t *testing.T) {
		return func(t *testing.T) {
			sql, err := buildBasicSingleSQLSelect(
				distinct,
				[]SQLProjection{{Variable: "a0.id"}, {Variable: "a0.value"}, {Variable: "r0.id"}},
				[]SQLFrom{{Value: "asset", Alias: "a0"}, {Value: "relation", Alias: "r0"}},
				[]SQLJoin{},
				[]SQLInnerStructure{},
				AndOrExpression{},
				[]int{}, AndOrExpression{}, map[string]struct{}{}, 0, 0)

			assert.NoError(t, err)
			assert.Equal(t, expected, sql)
		}
	}

	t.Run("DISTINCT false", test(false, "SELECT a0.id, a0.value, r0.id\nFROM (asset a0, relation r0)"))
	t.Run("DISTINCT true", test(true, "SELECT DISTINCT a0.id, a0.value, r0.id\nFROM (asset a0, relation r0)"))
}

func TestBuildBasicSingleSQLSelect_FromAlias(t *testing.T) {
	sql, err := buildBasicSingleSQLSelect(
		false,
		[]SQLProjection{{Variable: "*"}},
		[]SQLFrom{{Value: "asset"}, {Value: "relation", Alias: "r0"}},
		[]SQLJoin{},
		[]SQLInnerStructure{},
		AndOrExpression{},
		[]int{}, AndOrExpression{}, map[string]struct{}{}, 0, 0)

	assert.NoError(t, err)
	assert.Equal(t, "SELECT *\nFROM (asset, relation r0)", sql)
}

func TestBuildBasicSingleSQLSelect_ProjectionAlias(t *testing.T) {
	sql, err := buildBasicSingleSQLSelect(
		false,
		[]SQLProjection{{Variable: "a0.id", Alias: "a0_id"}, {Variable: "a0.value"}},
		[]SQLFrom{{Value: "asset", Alias: "a0"}},
		[]SQLJoin{},
		[]SQLInnerStructure{},
		AndOrExpression{},
		[]int{}, AndOrExpression{}, map[string]struct{}{}, 0, 0)

	assert.NoError(t, err)
	assert.Equal(t, "SELECT a0.id AS a0_id, a0.value\nFROM (asset a0)", sql)
}

func TestBuildBasicSingleSQLSelect_OrExpression(t *testing.T) {
	sql, err := buildBasicSingleSQLSelect(
		false,
		[]SQLProjection{{Variable: "*"}},
		[]SQLFrom{{Value: "asset", Alias: "a0"}},
		[]SQLJoin{},
		[]SQLInnerStructure{},
		AndOrExpression{
			And: false,
			Children: []AndOrExpression{
				{Expression: "a0_id = 'abc'"}, {Expression: "a0.type = 'mytype'"},
			},
		},
		[]int{}, AndOrExpression{}, map[string]struct{}{}, 0, 0)

	assert.NoError(t, err)
	assert.Equal(t, "SELECT *\nFROM (asset a0)\nWHERE a0_id = 'abc' OR a0.type = 'mytype'", sql)
}

func TestBuildBasicSingleSQLSelect_AndExpression(t *testing.T) {
	sql, err := buildBasicSingleSQLSelect(
		false,
		[]SQLProjection{{Variable: "*"}},
		[]SQLFrom{{Value: "asset", Alias: "a0"}},
		[]SQLJoin{},
		[]SQLInnerStructure{},
		AndOrExpression{
			And: true,
			Children: []AndOrExpression{
				{Expression: "a0_id = 'abc'"}, {Expression: "a0.type = 'mytype'"},
			},
		},
		[]int{}, AndOrExpression{}, map[string]struct{}{}, 0, 0)

	assert.NoError(t, err)
	assert.Equal(t, "SELECT *\nFROM (asset a0)\nWHERE a0_id = 'abc' AND a0.type = 'mytype'", sql)
}

func TestBuildBasicSingleSQLSelect_NestedExpressions(t *testing.T) {
	sql, err := buildBasicSingleSQLSelect(
		false,
		[]SQLProjection{{Variable: "*"}},
		[]SQLFrom{{Value: "asset", Alias: "a0"}},
		[]SQLJoin{},
		[]SQLInnerStructure{},
		AndOrExpression{
			And: false,
			Children: []AndOrExpression{
				{Expression: "a0_id = 'abc'"},
				{And: true, Children: []AndOrExpression{
					{Expression: "a0.type = 'mytype'"}, {Expression: "a0.value = 'myvalue'"}},
				},
			},
		},
		[]int{}, AndOrExpression{}, map[string]struct{}{}, 0, 0)

	assert.NoError(t, err)
	assert.Equal(t, "SELECT *\nFROM (asset a0)\nWHERE a0_id = 'abc' OR (a0.type = 'mytype' AND a0.value = 'myvalue')", sql)
}

func TestBuildBasicSingleSQLSelect_LIMIT(t *testing.T) {
	sql, err := buildBasicSingleSQLSelect(
		false,
		[]SQLProjection{{Variable: "*"}},
		[]SQLFrom{{Value: "asset"}},
		[]SQLJoin{},
		[]SQLInnerStructure{},
		AndOrExpression{},
		[]int{}, AndOrExpression{}, map[string]struct{}{}, 10, 0)

	assert.NoError(t, err)
	assert.Equal(t, "SELECT *\nFROM (asset)\nLIMIT 10", sql)
}

func TestBuildBasicSingleSQLSelect_OFFSET(t *testing.T) {
	sql, err := buildBasicSingleSQLSelect(
		false,
		[]SQLProjection{{Variable: "*"}},
		[]SQLFrom{{Value: "asset"}},
		[]SQLJoin{},
		[]SQLInnerStructure{},
		AndOrExpression{},
		[]int{}, AndOrExpression{}, map[string]struct{}{}, 10, 20)

	assert.NoError(t, err)
	assert.Equal(t, "SELECT *\nFROM (asset)\nLIMIT 10\nOFFSET 20", sql)
}

func TestBuildBasicSingleSQLSelect_GroupBy(t *testing.T) {
	sql, err := buildBasicSingleSQLSelect(
		false,
		[]SQLProjection{{Variable: "id"}, {Variable: "name"}, {Variable: "key"}},
		[]SQLFrom{{Value: "asset"}},
		[]SQLJoin{},
		[]SQLInnerStructure{},
		AndOrExpression{},
		[]int{0, 2}, AndOrExpression{}, map[string]struct{}{}, 0, 0)

	assert.NoError(t, err)
	assert.Equal(t, "SELECT id, name, key\nFROM (asset)\nGROUP BY id, name", sql)
}

func TestBuildSQLSelect_UnwindOrExprIntoUnion(t *testing.T) {
	sql, err := buildSQLSelect(
		SQLStructure{
			Distinct:    false,
			Projections: []SQLProjection{{Variable: "id"}, {Variable: "name"}, {Variable: "key"}},
			FromEntries: []SQLFrom{{Value: "asset"}},
			WhereExpression: AndOrExpression{
				And: false,
				Children: []AndOrExpression{
					{Expression: "id == 56"},
					{Expression: "name == 'myname'"},
				},
			},
		})

	assert.NoError(t, err)
	assert.Equal(t, "(SELECT id, name, key\nFROM (asset)\nWHERE id == 56)\nUNION ALL\n(SELECT id, name, key\nFROM (asset)\nWHERE name == 'myname')", sql)
}

func TestBuildSQLSelect_GroupBy_Limit_Offset(t *testing.T) {
	sql, err := buildSQLSelect(
		SQLStructure{
			Distinct:    false,
			Projections: []SQLProjection{{Variable: "id"}, {Variable: "name"}, {Variable: "key"}},
			FromEntries: []SQLFrom{{Value: "asset"}},
			WhereExpression: AndOrExpression{
				And: true,
				Children: []AndOrExpression{
					{Expression: "id == 56"},
					{Expression: "name == 'myname'"},
				},
			},
			GroupByIndices: []int{0, 2},
			Limit:          10,
			Offset:         20,
		})

	assert.NoError(t, err)
	assert.Equal(t, "SELECT id, name, key\nFROM (asset)\nWHERE id == 56 AND name == 'myname'\nGROUP BY id, name\nLIMIT 10\nOFFSET 20", sql)
}

func TestBuildBasicSingleSQLSelect_InnerSELECT(t *testing.T) {
	sql, err := buildBasicSingleSQLSelect(
		false,
		[]SQLProjection{{Variable: "*"}},
		[]SQLFrom{{Value: "asset", Alias: "a0"}},
		[]SQLJoin{},
		[]SQLInnerStructure{
			{
				Alias: "s0",
				Structure: SQLStructure{
					Distinct:        false,
					Projections:     []SQLProjection{{Variable: "id", Alias: "r_id"}, {Variable: "from_id", Alias: "r_from_id"}},
					FromEntries:     []SQLFrom{{Value: "relations"}},
					WhereExpression: AndOrExpression{Expression: "type = 'mytype'"},
				},
			},
		},
		AndOrExpression{And: true, Children: []AndOrExpression{{Expression: "s0.r_id > 8"}, {Expression: "a0.id = s0.r_from_id"}}},
		[]int{}, AndOrExpression{}, map[string]struct{}{}, 0, 0)

	assert.NoError(t, err)
	assert.Equal(t, "SELECT *\nFROM (asset a0, (SELECT id AS r_id, from_id AS r_from_id\nFROM (relations)\nWHERE type = 'mytype') AS s0)\nWHERE s0.r_id > 8 AND a0.id = s0.r_from_id", sql)
}

func TestBuildBasicSingleSQLSelect_JOIN(t *testing.T) {
	sql, err := buildBasicSingleSQLSelect(
		false,
		[]SQLProjection{{Variable: "*"}},
		[]SQLFrom{{Value: "assets", Alias: "variable"}},
		[]SQLJoin{{
			Table: "assets",
			Alias: "a0",
			On:    "a0.type = 'variable' AND a0.id = variable.id",
		}},
		[]SQLInnerStructure{},
		AndOrExpression{},
		[]int{}, AndOrExpression{}, map[string]struct{}{}, 0, 0)

	assert.NoError(t, err)
	assert.Equal(t, "SELECT *\nFROM (assets variable)\nJOIN assets a0 ON a0.type = 'variable' AND a0.id = variable.id", sql)
}

func TestBuildBasicSingleSQLSelect_JOINs(t *testing.T) {
	sql, err := buildBasicSingleSQLSelect(
		false,
		[]SQLProjection{{Variable: "*"}},
		[]SQLFrom{{Value: "assets", Alias: "variable"}},
		[]SQLJoin{
			{
				Table: "assets",
				Alias: "a0",
				On:    "a0.type = 'variable' AND a0.id = variable.id",
			},
			{
				Table: "relations",
				Alias: "r0",
				On:    "r0.type = 'is' AND r0.from_id = a0.id",
			},
			{
				Table: "assets",
				Alias: "a1",
				On:    "a1.type = 'scope' AND r0.to_id = a1.id",
			},
		},
		[]SQLInnerStructure{},
		AndOrExpression{},
		[]int{}, AndOrExpression{}, map[string]struct{}{}, 0, 0)

	assert.NoError(t, err)
	assert.Equal(t, "SELECT *\nFROM (assets variable)\nJOIN assets a0 ON a0.type = 'variable' AND a0.id = variable.id\nJOIN relations r0 ON r0.type = 'is' AND r0.from_id = a0.id\nJOIN assets a1 ON a1.type = 'scope' AND r0.to_id = a1.id", sql)
}
