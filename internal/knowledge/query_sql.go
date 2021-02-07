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

func (sqt *SQLQueryTranslator) buildSQLSelect(
	distinct bool, projections []string, projectionTypes []Projection, fromTables []string,
	whereExpressions AndOrExpression, groupBy []int, limit int, offset int) (string, error) {
	var sqlQuery string

	andExpressions, err := UnwindOrExpressions(whereExpressions)
	if err != nil {
		return "", err
	}

	if len(andExpressions) > 1 {
		singleQueries := []string{}
		for _, where := range andExpressions {
			singleQuery, err := sqt.buildSingleSQLSelect(false, projections, fromTables, where, nil, 0, 0)
			if err != nil {
				return "", err
			}
			singleQueries = append(singleQueries, fmt.Sprintf("(%s)", singleQuery))
		}
		if distinct {
			sqlQuery = strings.Join(singleQueries, "\nUNION\n")
		} else {
			sqlQuery = strings.Join(singleQueries, "\nUNION ALL\n")
		}

		if len(groupBy) > 0 {
			groupByProjections := []string{}
			for i := range groupBy {
				groupByProjections = append(groupByProjections, projections[groupBy[i]])
			}

			sqlQuery = fmt.Sprintf("SELECT %s FROM\n(%s)\nGROUP BY %s",
				strings.Join(projections, ", "), sqlQuery, strings.Join(groupByProjections, ","))
		}

		if limit > 0 {
			sqlQuery += fmt.Sprintf("\nLIMIT %d", limit)
		}

		if offset > 0 {
			sqlQuery += fmt.Sprintf("\nOFFSET %d", offset)
		}

	} else {
		and := AndOrExpression{And: true, Children: andExpressions}
		singleQuery, err := sqt.buildSingleSQLSelect(distinct, projections, fromTables, and, groupBy, limit, offset)
		if err != nil {
			return "", err
		}
		sqlQuery = singleQuery
	}

	return sqlQuery, nil
}

func buildBasicSingleSQLSelect(
	distinct bool, projections []string, fromTables []string,
	whereExpressions AndOrExpression) (string, error) {

	projectionsStr := ""
	if distinct {
		projectionsStr += "DISTINCT "
	}

	projectionsStr += strings.Join(projections, ", ")
	fromTablesStr := strings.Join(fromTables, ", ")

	sqlQuery := fmt.Sprintf("SELECT %s FROM %s", projectionsStr, fromTablesStr)
	whereExprStr, err := BuildAndOrExpression(whereExpressions)
	if err != nil {
		return "", err
	}

	if whereExprStr != "" {
		sqlQuery += fmt.Sprintf("\nWHERE %s", whereExprStr)
	}
	return sqlQuery, nil
}

func (sqt *SQLQueryTranslator) buildSingleSQLSelect(
	distinct bool, projections []string, fromTables []string,
	whereExpressions AndOrExpression, groupBy []int, limit int, offset int) (string, error) {

	sqlQuery, err := buildBasicSingleSQLSelect(distinct, projections, fromTables, whereExpressions)
	if err != nil {
		return "", err
	}

	if len(groupBy) > 0 {
		groupByProjection := make([]string, len(groupBy))
		for i := range groupBy {
			groupByProjection[i] = projections[groupBy[i]]
		}
		sqlQuery += fmt.Sprintf("\nGROUP BY %s", strings.Join(groupByProjection, ", "))
	}

	if limit > 0 {
		sqlQuery += fmt.Sprintf("\nLIMIT %d", limit)
	}

	if offset > 0 {
		sqlQuery += fmt.Sprintf("\nOFFSET %d", offset)
	}

	return sqlQuery, nil
}

func buildSQLConstraintsFromPatterns(queryGraph *QueryGraph, constrainedNodes map[int]bool, scope Scope) (*AndOrExpression, []string, error) {
	andExpressions := AndOrExpression{And: true}
	from := make([]string, 0)

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
		from = append(from, fmt.Sprintf("assets %s", alias))

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
		from = append(from, fmt.Sprintf("relations %s", alias))

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
func (sqt *SQLQueryTranslator) Translate(query *query.QueryCypher) (*SQLTranslation, error) {
	constrainedNodes := make(map[int]bool)

	// The list of strings that will be concatenated with , in the top-level FROM clause
	var from []string

	filterExpressions := AndOrExpression{And: true}
	for _, x := range query.QuerySinglePartQuery.QueryMatches {
		parser := NewPatternParser(&sqt.QueryGraph)
		for _, y := range x.PatternElements {
			err := parser.ParsePatternElement(&y, MatchScope)
			if err != nil {
				return nil, err
			}
		}

		// Build the constraints for the patterns in MATCH clause
		expr, f, err := buildSQLConstraintsFromPatterns(&sqt.QueryGraph, constrainedNodes, MatchScope)
		if err != nil {
			return nil, fmt.Errorf("Unable to build SQL constraints from patterns in the MATCH clause: %v", err)
		}
		from = f

		if expr.Expression != "" || len(expr.Children) > 0 {
			filterExpressions.Children = append(filterExpressions.Children, *expr)
		}

		if x.Where != nil {
			whereVisitor := NewQueryWhereVisitor(&sqt.QueryGraph)
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
				filterExpressions.Children = append(filterExpressions.Children,
					AndOrExpression{Expression: whereExpression})
			}
		}
	}

	projections := make([]string, 0)
	projectionTypes := make([]Projection, 0)

	unaggregatedProjectionItems := []int{}
	aggregationRequired := false

	for i, p := range query.QuerySinglePartQuery.ProjectionBody.ProjectionItems {
		projectionVisitor := NewProjectionVisitor(&sqt.QueryGraph)
		err := projectionVisitor.ParseExpression(&p.Expression)
		if err != nil {
			return nil, err
		}

		projection, err := NewExpressionBuilder(&sqt.QueryGraph).Build(&p.Expression)
		if err != nil {
			return nil, err
		}

		if !projectionVisitor.Aggregation {
			unaggregatedProjectionItems = append(unaggregatedProjectionItems, i)
		} else {
			aggregationRequired = true
		}

		projections = append(projections, projection)
		projectionTypes = append(projectionTypes, Projection{
			Alias:          p.Alias,
			ExpressionType: projectionVisitor.ExpressionType,
		})
	}

	if !aggregationRequired {
		unaggregatedProjectionItems = nil
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

	sqlQuery, err := sqt.buildSQLSelect(
		query.QuerySinglePartQuery.ProjectionBody.Distinct,
		projections,
		projectionTypes,
		from,
		filterExpressions,
		unaggregatedProjectionItems,
		limit, offset)

	if err != nil {
		return nil, err
	}

	return &SQLTranslation{
		Query:           sqlQuery,
		ProjectionTypes: projectionTypes,
	}, nil
}
