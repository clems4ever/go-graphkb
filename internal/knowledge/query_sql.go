package knowledge

import (
	"fmt"
	"strings"

	"github.com/clems4ever/go-graphkb/internal/query"
)

// SQLQueryTranslator represent an SQL translator object converting cypher queries into SQL
type SQLQueryTranslator struct {
	QueryGraph QueryGraph
}

// NewSQLQueryTranslator create an instance of SQL query translator
func NewSQLQueryTranslator() *SQLQueryTranslator {
	return &SQLQueryTranslator{QueryGraph: NewQueryGraph()}
}

// Projection represent the type and alias of one item in the RETURN statement (called a projection).
type Projection struct {
	Alias          string
	ExpressionType ExpressionType
}

// SQLTranslation the resulting object of the translation
type SQLTranslation struct {
	// Query is the SQL query built from the Cypher query
	Query string
	// ProjectionTypes helps the clients know how to serialize the results
	ProjectionTypes []Projection
}

func buildSQLConstraintsFromPatterns(queryGraph *QueryGraph, constrainedNodes map[int]bool, scope Scope) (*AndOrExpression, []SQLFrom, error) {
	andExpressions := AndOrExpression{And: true}
	from := []SQLFrom{}

	for i, n := range queryGraph.Nodes {
		inScope := false
		inMatchScope := false
		for s := range n.Scopes {
			if s == scope {
				inScope = true
			}
			if s == MatchScope {
				inMatchScope = true
			}
		}
		if !inScope {
			continue
		}

		constraints := AndOrExpression{And: false}

		var alias string
		if scope.Context == WhereContext {
			alias = fmt.Sprintf("aw%d", i)
		} else {
			alias = fmt.Sprintf("a%d", i)
		}
		from = append(from, SQLFrom{Value: "assets", Alias: alias})

		// If the scope is WHERE and the asset is also in the MATCH clause, we make the link between them
		if scope.Context == WhereContext && inMatchScope {
			constraints.Children = append(constraints.Children, AndOrExpression{
				Expression: fmt.Sprintf("%s.id = %s.id", alias, fmt.Sprintf("a%d", i)),
			})
		} else { // otherwise we don't make the link but simply check the types
			for _, label := range n.Labels {
				constraints.Children = append(constraints.Children, AndOrExpression{
					Expression: fmt.Sprintf("%s.type = '%s'", alias, label),
				})
			}
		}

		if len(constraints.Children) > 0 {
			// Append assets constraints
			andExpressions.Children = append(andExpressions.Children, constraints)
		}
	}

	for i, r := range queryGraph.Relations {
		inScope := false
		inMatchScope := false
		for s := range r.Scopes {
			if s == scope {
				inScope = true
				break
			}
			if s == MatchScope {
				inMatchScope = true
			}
		}
		if !inScope {
			continue
		}

		var alias, aliasPrefix, assetAliasPrefix string
		if scope.Context == WhereContext {
			aliasPrefix = "rw"
			assetAliasPrefix = "aw"
		} else {
			aliasPrefix = "r"
			assetAliasPrefix = "a"
		}
		alias = fmt.Sprintf("%s%d", aliasPrefix, i)
		from = append(from, SQLFrom{Value: "relations", Alias: alias})

		constraints := AndOrExpression{And: false}

		// If the scope is WHERE and the asset is also in the MATCH clause, we make the link between them
		if scope.Context == WhereContext && inMatchScope {
			constraints.Children = append(constraints.Children, AndOrExpression{
				Expression: fmt.Sprintf("%s.id = %s.id", alias, fmt.Sprintf("r%d", i)),
			})
		} else { // otherwise we don't make the link but simply check the types
			for _, label := range r.Labels {
				constraints.Children = append(constraints.Children, AndOrExpression{
					Expression: fmt.Sprintf("%s.type = '%s'", alias, label),
				})
			}
		}

		if len(constraints.Children) > 0 {
			andExpressions.Children = append(andExpressions.Children, constraints)
		}

		out := AndOrExpression{
			And: true,
			Children: []AndOrExpression{
				{
					Expression: fmt.Sprintf("%s.from_id = %s%d.id", alias, assetAliasPrefix, r.LeftIdx),
				},
				{
					Expression: fmt.Sprintf("%s.to_id = %s%d.id", alias, assetAliasPrefix, r.RightIdx),
				},
			},
		}

		in := AndOrExpression{
			And: true,
			Children: []AndOrExpression{
				{
					Expression: fmt.Sprintf("%s.from_id = %s%d.id", alias, assetAliasPrefix, r.RightIdx),
				},
				{
					Expression: fmt.Sprintf("%s.to_id = %s%d.id", alias, assetAliasPrefix, r.LeftIdx),
				},
			},
		}

		if r.Direction == Right {
			andExpressions.Children = append(andExpressions.Children, out)
		} else if r.Direction == Left {
			andExpressions.Children = append(andExpressions.Children, in)
		} else if r.Direction == Either {
			oneDirectionOptimization := false
			// Optimization: in this case, finding in any direction is sufficient.
			if len(queryGraph.Relations) == 1 {
				nodesConstrained := false
				for idx := range constrainedNodes {
					if idx == r.LeftIdx {
						nodesConstrained = true
					}
					if idx == r.RightIdx {
						nodesConstrained = true
					}
				}
				if !nodesConstrained {
					n, err := queryGraph.GetNodeByID(r.LeftIdx)
					if err != nil {
						return nil, nil, err
					}

					// if the node has a variable name bound to it, it is considered constrained
					if len(n.Labels) > 0 {
						nodesConstrained = true
					}

					n, err = queryGraph.GetNodeByID(r.RightIdx)
					if err != nil {
						return nil, nil, err
					}

					// if the node has a variable name bound to it, it is considered constrained
					if len(n.Labels) > 0 {
						nodesConstrained = true
					}
				}
				oneDirectionOptimization = !nodesConstrained
			}

			// This optimization is possible because (x)--(y) <=> (y)--(x)
			if oneDirectionOptimization {
				andExpressions.Children = append(andExpressions.Children, out)
			} else {
				// otherwise we need to have an OR expression which is later translated into a union of queries.
				orExpression := AndOrExpression{
					And:      false,
					Children: []AndOrExpression{out, in},
				}
				andExpressions.Children = append(andExpressions.Children, orExpression)
			}
		}
	}
	return &andExpressions, from, nil
}

// Translate a Cypher query into a SQL model
func (sqt *SQLQueryTranslator) Translate(query *query.QueryCypher, includeDataSourceInResults bool) (*SQLTranslation, error) {
	constrainedNodes := make(map[int]bool)

	filterExpressions := AndOrExpression{And: true}
	whereExpressions := AndOrExpression{And: true}
	for _, x := range query.QuerySinglePartQuery.QueryMatches {
		parser := NewPatternParser(&sqt.QueryGraph)
		for _, y := range x.PatternElements {
			err := parser.ParsePatternElement(&y, MatchScope)
			if err != nil {
				return nil, err
			}
		}

		if x.Where != nil {
			whereVisitor := NewQueryWhereVisitor(&sqt.QueryGraph, includeDataSourceInResults)
			whereExpression, err := whereVisitor.ParseExpression(x.Where)
			if err != nil {
				return nil, err
			}
			for _, v := range whereVisitor.Variables {
				typeAndIndex, err := sqt.QueryGraph.FindVariable(v)
				if err != nil {
					return nil, err
				}
				constrainedNodes[typeAndIndex.Index] = true
			}

			// We only append the expression if it's not empty
			if whereExpression != "" {
				whereExpressions.Children = append(whereExpressions.Children,
					AndOrExpression{And: true, Expression: whereExpression})
			}
		}
	}

	// Build the constraints for the patterns in MATCH clause
	expr, f, err := buildSQLConstraintsFromPatterns(&sqt.QueryGraph, constrainedNodes, MatchScope)
	if err != nil {
		return nil, fmt.Errorf("Unable to build SQL constraints from patterns in the MATCH clause: %v", err)
	}
	from := f

	if expr.Expression != "" || len(expr.Children) > 0 {
		filterExpressions.Children = append(filterExpressions.Children, *expr)
	}

	if whereExpressions.Expression != "" || len(whereExpressions.Children) > 0 {
		filterExpressions.Children = append(filterExpressions.Children, whereExpressions)
	}

	projections := make([]SQLProjection, 0)
	projectionTypes := make([]Projection, 0)
	groupByIndices := []int{}
	groupByRequired := false
	variableIndices := []int{}

	for i, p := range query.QuerySinglePartQuery.ProjectionBody.ProjectionItems {
		projectionVisitor := NewProjectionVisitor(&sqt.QueryGraph)
		err := projectionVisitor.ParseExpression(&p.Expression)
		if err != nil {
			return nil, err
		}

		for _, proj := range projectionVisitor.Projections {
			if proj.Function != "" {
				projections = append(projections, SQLProjection{
					Function: &SQLFunction{Name: proj.Function, Distinct: proj.Distinct},
					Variable: proj.Variable})
				groupByRequired = true
			} else if proj.Variable != "" {
				projections = append(projections, SQLProjection{Variable: proj.Variable})
				variableIndices = append(variableIndices, i)
			} else {
				return nil, fmt.Errorf("Unable to detect type of projection")
			}
		}

		projectionTypes = append(projectionTypes, Projection{
			Alias:          p.Alias,
			ExpressionType: projectionVisitor.ExpressionType,
		})
	}
	// If group by is required, we group by all variables except the aggregation functions
	if groupByRequired {
		groupByIndices = variableIndices
	}

	limit := 0
	if query.QuerySinglePartQuery.ProjectionBody.Limit != nil {
		limitVisitor := NewQueryLimitVisitor(&sqt.QueryGraph)
		err := limitVisitor.ParseExpression(
			query.QuerySinglePartQuery.ProjectionBody.Limit)
		if err != nil {
			return nil, err
		}
		limit = int(limitVisitor.Limit)
	}

	offset := 0
	if query.QuerySinglePartQuery.ProjectionBody.Skip != nil {
		if limit == 0 {
			return nil, fmt.Errorf("SKIP must be used in combination with limit")
		}
		skipVisitor := NewQuerySkipVisitor(&sqt.QueryGraph)
		err := skipVisitor.ParseExpression(
			query.QuerySinglePartQuery.ProjectionBody.Skip)
		if err != nil {
			return nil, err
		}
		offset = int(skipVisitor.Skip)
	}

	var sqlQuery string

	innerSQL := SQLStructure{
		Distinct:        query.QuerySinglePartQuery.ProjectionBody.Distinct,
		Projections:     projections,
		FromEntries:     from,
		WhereExpression: filterExpressions,
		GroupByIndices:  groupByIndices,
		Limit:           limit,
		Offset:          offset,
	}

	if includeDataSourceInResults {
		from = []SQLFrom{}
		whereExpr := AndOrExpression{
			And:      true,
			Children: []AndOrExpression{},
		}

		for i := range innerSQL.Projections {
			if innerSQL.Projections[i].Function != nil {
				innerSQL.Projections[i].Alias = fmt.Sprintf("%s_%s",
					strings.ReplaceAll(innerSQL.Projections[i].Variable, ".", "_"), innerSQL.Projections[i].Function.Name)
			} else {
				innerSQL.Projections[i].Alias = strings.ReplaceAll(innerSQL.Projections[i].Variable, ".", "_")
			}
		}

		var projItems []SQLProjection

		projItems = append(projItems, SQLProjection{Variable: "s0.*"})

		for i := range sqt.QueryGraph.Nodes {
			alias := fmt.Sprintf("a%d", i)
			sourceAlias := fmt.Sprintf("%s_s", alias)
			from = append(from, SQLFrom{Value: "assets_by_source", Alias: fmt.Sprintf("%s_bs", alias)})
			from = append(from, SQLFrom{Value: "sources", Alias: sourceAlias})

			projItems = append(projItems, SQLProjection{Variable: fmt.Sprintf("%s.name", sourceAlias)})

			whereExpr.Children = append(whereExpr.Children, AndOrExpression{
				And: true,
				Children: []AndOrExpression{
					{Expression: fmt.Sprintf("%s_bs.asset_id = %s_id", alias, alias)},
					{Expression: fmt.Sprintf("%s_bs.source_id = %s_s.id", alias, alias)},
				},
			})
		}

		for i := range sqt.QueryGraph.Relations {
			alias := fmt.Sprintf("r%d", i)
			sourceAlias := fmt.Sprintf("%s_s", alias)
			from = append(from, SQLFrom{Value: "relations_by_source", Alias: fmt.Sprintf("%s_bs", alias)})
			from = append(from, SQLFrom{Value: "sources", Alias: sourceAlias})

			projItems = append(projItems, SQLProjection{Variable: fmt.Sprintf("%s.name", sourceAlias)})

			whereExpr.Children = append(whereExpr.Children, AndOrExpression{
				And: true,
				Children: []AndOrExpression{
					{Expression: fmt.Sprintf("%s_bs.relation_id = %s_id", alias, alias)},
					{Expression: fmt.Sprintf("%s_bs.source_id = %s_s.id", alias, alias)},
				},
			})
		}

		sqlQuery, err = buildSQLSelect(SQLStructure{
			Distinct:    false,
			Projections: projItems,
			FromEntries: from,
			FromStructures: []SQLInnerStructure{
				{
					Alias:     "s0",
					Structure: innerSQL,
				},
			},
			WhereExpression: whereExpr,
		})
	} else {
		sqlQuery, err = buildSQLSelect(innerSQL)
	}

	if err != nil {
		return nil, err
	}

	return &SQLTranslation{
		Query:           sqlQuery,
		ProjectionTypes: projectionTypes,
	}, nil
}
