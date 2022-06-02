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

// ProcessedRelationTuple tuple for existing relationships
type ProcessedRelationTuple struct {
	processedRelation      *QueryRelation
	processedRelationAlias string
}

func isRelationOptimizable(queryGraph *QueryGraph, r QueryRelation) (bool, error) {

	if len(queryGraph.Relations) > 1 {
		return false, nil
	}

	nodesConstrained := false

	n, err := queryGraph.GetNodeByID(r.LeftIdx)
	if err != nil {
		return false, err
	}

	// if the node has a variable name bound to it, it is considered constrained
	if len(n.Labels) > 0 {
		nodesConstrained = true
	}

	n, err = queryGraph.GetNodeByID(r.RightIdx)
	if err != nil {
		return false, err
	}

	// if the node has a variable name bound to it, it is considered constrained
	if len(n.Labels) > 0 {
		nodesConstrained = true
	}

	return !nodesConstrained, nil
}

func buildSQLConstraintsFromPatterns(queryGraph *QueryGraph, constrainedNodes map[int]bool, scope Scope) ([][]SQLJoin, []SQLFrom, error) {
	from := []SQLFrom{}
	relationSet := make(map[*QueryRelation]string)
	assetSet := make(map[*QueryNode]struct{ alias string })
	joinCollections := [][]SQLJoin{}
	joins := []SQLJoin{}
	relationCount := 0

	// Process every node as it is encountered in the query
	for i, n := range queryGraph.Nodes {
		inScope := false
		for s := range n.Scopes {
			if s == scope {
				inScope = true
			}
		}
		if !inScope {
			continue
		}

		relations := queryGraph.GetRelationsByNodeId(i)

		var processedRelations []ProcessedRelationTuple

		relationExists := false
		var existingRelationAlias string
		relationToAssetExists := false

		// Look for already visited relationships to link them in each JOIN
		for _, relation := range relations {
			existingRelationAlias, relationExists = relationSet[relation]
			if !relationExists {
				continue
			}
			// Get the node at the other side of the relationship to link them together
			if relation.LeftIdx == i {
				neighbor, err := queryGraph.GetNodeByID(relation.RightIdx)
				if err != nil {
					return nil, nil, err
				}
				_, relationToAssetExists = assetSet[neighbor]
				if relationToAssetExists {
					processedRelations = append(processedRelations, struct {
						processedRelation      *QueryRelation
						processedRelationAlias string
					}{processedRelation: relation, processedRelationAlias: existingRelationAlias})
				}
			} else {
				neighbor, err := queryGraph.GetNodeByID(relation.LeftIdx)
				if err != nil {
					return nil, nil, err
				}
				_, relationToAssetExists = assetSet[neighbor]
				if relationToAssetExists {
					processedRelations = append(processedRelations, struct {
						processedRelation      *QueryRelation
						processedRelationAlias string
					}{processedRelation: relation, processedRelationAlias: existingRelationAlias})
				}
			}

		}

		var alias string
		if scope.Context == WhereContext {
			alias = fmt.Sprintf("aw%d", i)
		} else {
			alias = fmt.Sprintf("a%d", i)
		}

		// If no relationship of this node has been visited yet, then we have no information on this node.
		// Scan the assets table again for this particular node.
		if !relationToAssetExists {
			if len(n.Labels) == 0 {
				from = append(from, SQLFrom{Value: "assets", Alias: alias})
			}

			for _, label := range n.Labels {
				from = append(from, SQLFrom{Value: "assets", Alias: label})

				if scope.Context == WhereContext {
					joins = append(joins, SQLJoin{
						Table: "assets",
						Alias: alias,
						On:    fmt.Sprintf("%s.type = '%s' AND %s.id = %s.id", alias, label, alias, strings.ReplaceAll(alias, "w", "")),
					})
				} else {
					joins = append(joins, SQLJoin{
						Table: "assets",
						Alias: alias,
						On:    fmt.Sprintf("%s.type = '%s' AND %s.id = %s.id", alias, label, alias, label),
					})
				}

			}
			// If a relationship exists, we link this asset to it in the JOIN.
		} else {
			var exp []string

			if len(n.Labels) > 0 {
				for _, label := range n.Labels {
					exp = append(exp, fmt.Sprintf("%s.type = '%s'", alias, label))
				}
			}

			for _, processedRelationStruct := range processedRelations {

				processedRelation := processedRelationStruct.processedRelation
				processedRelationAlias := processedRelationStruct.processedRelationAlias
				if (processedRelation.Direction == Right && processedRelation.LeftIdx == i) || (processedRelation.Direction == Left && processedRelation.RightIdx == i) {
					exp = append(exp, fmt.Sprintf("%s.from_id = %s.id", processedRelationAlias, alias))
				} else if (processedRelation.Direction == Right && processedRelation.RightIdx == i) || (processedRelation.Direction == Left && processedRelation.LeftIdx == i) {
					exp = append(exp, fmt.Sprintf("%s.to_id = %s.id", processedRelationAlias, alias))
				} else {
					// If the relationship has no direction, we assume it is a left directed relationship
					// Previously we assumed the relationship was to the left and forked the same relation to the right.
					if processedRelation.RightIdx == i {
						exp = append(exp, fmt.Sprintf("%s.from_id = %s.id", processedRelationAlias, alias))
					} else {
						exp = append(exp, fmt.Sprintf("%s.to_id = %s.id", processedRelationAlias, alias))
					}

				}
			}
			joins = append(joins, SQLJoin{
				Table: "assets",
				Alias: alias,
				On:    strings.Join(exp, " AND "),
				Index: "PRIMARY",
			})

		}

		// For each node, visit each of its relationship
		for _, relation := range relations {
			_, relationExists = relationSet[relation]
			if relationExists {
				continue
			}
			inScope := false
			for s := range relation.Scopes {
				if s == scope {
					inScope = true
					break
				}
			}
			if !inScope {
				continue
			}

			var ralias, aliasPrefix, assetAliasPrefix string
			if scope.Context == WhereContext {
				aliasPrefix = "rw"
				assetAliasPrefix = "aw"
			} else {
				aliasPrefix = "r"
				assetAliasPrefix = "a"
			}
			ralias = fmt.Sprintf("%s%d", aliasPrefix, relationCount)
			relationCount++
			var exps []string

			if len(relation.Labels) > 0 {
				for _, label := range relation.Labels {
					exps = append(exps, fmt.Sprintf("%s.type = '%s'", ralias, label))
				}
			}

			index := ""

			if relation.Direction == Right && relation.LeftIdx == i {
				exps = append(exps, fmt.Sprintf("%s.from_id = %s%d.id", ralias, assetAliasPrefix, relation.LeftIdx))
				index = "full_relation_type_from_to_idx"
			} else if relation.Direction == Right && relation.RightIdx == i {
				exps = append(exps, fmt.Sprintf("%s.to_id = %s%d.id", ralias, assetAliasPrefix, relation.RightIdx))
				index = "full_relation_type_to_from_idx"
			} else if relation.Direction == Left && relation.LeftIdx == i {
				exps = append(exps, fmt.Sprintf("%s.to_id = %s%d.id", ralias, assetAliasPrefix, relation.LeftIdx))
				index = "full_relation_type_to_from_idx"
			} else if relation.Direction == Left && relation.RightIdx == i {
				exps = append(exps, fmt.Sprintf("%s.from_id = %s%d.id", ralias, assetAliasPrefix, relation.RightIdx))
				index = "full_relation_type_from_to_idx"
			} else {
				// If a relationship is undirected
				optimize, err := isRelationOptimizable(queryGraph, *relation)

				if err != nil {
					return nil, nil, err
				}

				if !optimize { // if this relation is optimizable, then just on direction is needed as (v) -- (q) <==> (q) -- (v)
					// if not, we need to fork a join to translate it into an UNION when building the SQL query

					queryGraphClone := queryGraph.Clone()

					// Calculate the join if the graph was directed to the right
					queryGraphClone.Relations[relation.id].Direction = Right
					forkedJoinCollection, _, err := buildSQLConstraintsFromPatterns(queryGraphClone, constrainedNodes, scope)
					if err != nil {
						return nil, nil, err
					}
					joinCollections = append(joinCollections, forkedJoinCollection...)

				}

				// Assume a left directed relationship
				if relation.LeftIdx == i {
					exps = append(exps, fmt.Sprintf("%s.to_id = %s%d.id", ralias, assetAliasPrefix, relation.LeftIdx))
				} else if relation.RightIdx == i {
					exps = append(exps, fmt.Sprintf("%s.from_id = %s%d.id", ralias, assetAliasPrefix, relation.RightIdx))
				}

			}

			// Hash the relationship so we know if we've seen it before
			relationSet[relation] = ralias

			joins = append(joins, SQLJoin{
				Table: "relations",
				Alias: ralias,
				On:    strings.Join(exps, " AND "),
				Index: index,
			})

		}
		// Hash the node so we can make sure we've processed its relationship
		assetSet[&queryGraph.Nodes[i]] = struct{ alias string }{alias: alias}

	}

	joinCollections = append(joinCollections, joins)

	return joinCollections, from, nil
}

// Translate a Cypher query into a SQL model
func (sqt *SQLQueryTranslator) Translate(query *query.QueryCypher) (*SQLTranslation, error) {
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
			whereVisitor := NewQueryWhereVisitor(&sqt.QueryGraph) // where conditions of cql
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
	//TODO returns a set of where expressions that matches a list of from expressions -> (WHERE a0.type = 'subnet'...... from a0 assets)
	joins, f, err := buildSQLConstraintsFromPatterns(&sqt.QueryGraph, constrainedNodes, MatchScope)
	if err != nil {
		return nil, fmt.Errorf("Unable to build SQL constraints from patterns in the MATCH clause: %v", err)
	}
	from := f

	if whereExpressions.Expression != "" || len(whereExpressions.Children) > 0 {
		filterExpressions.Children = append(filterExpressions.Children, whereExpressions) // CQL where conditions
	}

	projections := make([]SQLProjection, 0)
	projectionTypes := make([]Projection, 0)
	groupByIndices := []int{}
	groupByRequired := false
	variableIndices := []int{}

	for i, p := range query.QuerySinglePartQuery.ProjectionBody.ProjectionItems { // Here's the select statement
		projectionVisitor := NewProjectionVisitor(&sqt.QueryGraph)
		err := projectionVisitor.ParseExpression(&p.Expression) // Lots of interfaces, gets to the return statement (projection) and returns them to be parsed as the SELECT
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
		WhereExpression: whereExpressions,
		JoinEntries:     joins,
		GroupByIndices:  groupByIndices,
		Limit:           limit,
		Offset:          offset,
	}

	sqlQuery, err = buildSQLSelect(innerSQL)

	if err != nil {
		return nil, err
	}

	return &SQLTranslation{
		Query:           sqlQuery,
		ProjectionTypes: projectionTypes,
	}, nil
}
