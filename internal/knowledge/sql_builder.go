package knowledge

import (
	"fmt"
	"strings"
)

type SQLFunction struct {
	Name     string
	Distinct bool
}

// SQLProjection represent a projection item with an optional alias name
type SQLProjection struct {
	// If alias is empty there won't be any aliasing with AS keyword.
	Alias    string
	Variable string
	// If function is not empty, variable should be provided and then the SQL expression should be <Function>(<Variable>) AS <Alias>
	Function *SQLFunction
}

// SQLFrom represent a from item with an optional alias name
type SQLFrom struct {
	// Alias can be empty.
	Alias string
	Value string
}

type SQLJoin struct {
	Table string
	Alias string
	On    string
	Index string
}

// SQLInnerStructure represent a SQL inner structure with an optional alias name
type SQLInnerStructure struct {
	// If alias is empty there won't be any aliasing with AS keyword.
	Alias     string
	Structure SQLStructure
}

// SQLStructure represent a SQL structure for building the query string
type SQLStructure struct {
	Distinct        bool
	Projections     []SQLProjection
	FromEntries     []SQLFrom
	FromStructures  []SQLInnerStructure
	WhereExpression AndOrExpression
	JoinEntries     [][]SQLJoin
	GroupByIndices  []int
	Limit           int
	Offset          int
}

func buildSQLSelect(structure SQLStructure) (string, error) {
	var sqlQuery string

	unwoundAndExpressions, err := UnwindOrExpressions(structure.WhereExpression)
	if err != nil {
		return "", fmt.Errorf("Unable to unwind where expression: %w", err)
	}

	andExpressions := []AndOrExpression{}
	// Expression strings are filled up here
	for _, e := range unwoundAndExpressions {
		exp, err := FlattenAndOrExpressions(e)
		if err != nil {
			return "", fmt.Errorf("Unable to flatten expression: %w", err)
		}
		andExpressions = append(andExpressions, exp)
	}

	// If we end up with multiple and expressions after unwind or we have forked JOINs, we must derive an union query.
	if len(andExpressions) > 1 || len(structure.JoinEntries) > 1 {
		singleQueries := []string{}
		for _, where := range andExpressions {
			// If groupBy is required, we need aliases for all columns
			if len(structure.GroupByIndices) > 0 {
				for i, p := range structure.Projections {
					if p.Function != nil {
						structure.Projections[i].Alias = fmt.Sprintf("%s_%s", strings.ReplaceAll(p.Variable, ".", "_"), p.Function.Name)
					} else {
						structure.Projections[i].Alias = strings.ReplaceAll(p.Variable, ".", "_")
					}
				}
			}

			// In that case, groupBy, limit and offset should be applied to the union instead of to all queries in the global query.
			joinEntries := []SQLJoin{}
			if len(structure.JoinEntries) > 0 {
				joinEntries = structure.JoinEntries[0]
			}
			singleQuery, err := buildBasicSingleSQLSelect(false, structure.Projections, structure.FromEntries, joinEntries,
				structure.FromStructures, where, structure.GroupByIndices, 0, 0)
			if err != nil {
				return "", err
			}
			singleQueries = append(singleQueries, fmt.Sprintf("(%s)", singleQuery))
		}

		for _, join := range structure.JoinEntries {

			if len(structure.GroupByIndices) > 0 {
				for i, p := range structure.Projections {
					if p.Function != nil {
						structure.Projections[i].Alias = fmt.Sprintf("%s_%s", strings.ReplaceAll(p.Variable, ".", "_"), p.Function.Name)
					} else {
						structure.Projections[i].Alias = strings.ReplaceAll(p.Variable, ".", "_")
					}
				}
			}

			where := AndOrExpression{}
			if len(andExpressions) > 0 {
				where = andExpressions[0]
			}
			// In that case, groupBy, limit and offset should be applied to the union instead of to all queries in the global query.
			singleQuery, err := buildBasicSingleSQLSelect(false, structure.Projections, structure.FromEntries, join,
				structure.FromStructures, where, structure.GroupByIndices, 0, 0)
			if err != nil {
				return "", err
			}
			singleQueries = append(singleQueries, fmt.Sprintf("(%s)", singleQuery))
		}

		if structure.Distinct {
			sqlQuery = strings.Join(singleQueries, "\nUNION\n")
		} else {
			sqlQuery = strings.Join(singleQueries, "\nUNION ALL\n")
		}

		if len(structure.GroupByIndices) > 0 {
			groupByProjections := []string{}
			for i := range structure.GroupByIndices {
				if structure.Projections[structure.GroupByIndices[i]].Alias != "" {
					groupByProjections = append(groupByProjections, fmt.Sprintf("x.%s", structure.Projections[structure.GroupByIndices[i]].Alias))
				} else {
					return "", fmt.Errorf("The projections must be aliased for group by to work")
				}
			}

			projectionsSQL := []string{}
			for _, p := range structure.Projections {
				if p.Alias == "" {
					return "", fmt.Errorf("The projections must be aliased for group by to work")
				}

				if p.Function != nil && p.Function.Name == "COUNT" {
					projectionsSQL = append(projectionsSQL, fmt.Sprintf("SUM(%s)", p.Alias))
				} else {
					projectionsSQL = append(projectionsSQL, p.Alias)
				}
			}

			sqlQuery = fmt.Sprintf("SELECT %s\nFROM\n(%s) AS x\nGROUP BY %s",
				strings.Join(projectionsSQL, ", "), sqlQuery, strings.Join(groupByProjections, ","))
		}

		if structure.Limit > 0 {
			sqlQuery += fmt.Sprintf("\nLIMIT %d", structure.Limit)
		}

		if structure.Offset > 0 {
			sqlQuery += fmt.Sprintf("\nOFFSET %d", structure.Offset)
		}

	} else { // We don't need union since there were only and expressions in the constraints
		where := AndOrExpression{}
		if len(andExpressions) > 0 {
			where = andExpressions[0]
		}
		joinEntries := []SQLJoin{}
		if len(structure.JoinEntries) > 0 {
			joinEntries = structure.JoinEntries[0]
		}
		singleQuery, err := buildBasicSingleSQLSelect(structure.Distinct, structure.Projections, structure.FromEntries,
			joinEntries, structure.FromStructures, where, structure.GroupByIndices, structure.Limit, structure.Offset)
		if err != nil {
			return "", err
		}
		sqlQuery = singleQuery
	}

	return sqlQuery, nil
}

func buildBasicSingleSQLSelect(
	distinct bool, projections []SQLProjection, fromEntries []SQLFrom, joinEntries []SQLJoin, fromStructures []SQLInnerStructure,
	whereExpressions AndOrExpression, groupBy []int, limit int, offset int) (string, error) {

	projectionsStr := ""
	if distinct {
		projectionsStr += "DISTINCT "
	}

	projectionsSQL := []string{}
	for _, p := range projections {
		leftSide := ""
		if p.Function != nil {
			distinctStr := ""
			if p.Function.Distinct {
				distinctStr = "DISTINCT "
			}
			leftSide = fmt.Sprintf("%s(%s%s)", p.Function.Name, distinctStr, p.Variable)
		} else {
			leftSide = p.Variable
		}

		if p.Alias != "" {
			projectionsSQL = append(projectionsSQL, fmt.Sprintf("%s AS %s", leftSide, p.Alias))
		} else {
			projectionsSQL = append(projectionsSQL, leftSide)
		}
	}

	projectionsStr += strings.Join(projectionsSQL, ", ")

	fromTablesSQL := []string{}
	for _, f := range fromEntries {
		if f.Alias != "" {
			fromTablesSQL = append(fromTablesSQL, fmt.Sprintf("%s %s", f.Value, f.Alias))
		} else {
			fromTablesSQL = append(fromTablesSQL, f.Value)
		}
	}

	for _, f := range fromStructures {
		sql, err := buildSQLSelect(f.Structure)
		if err != nil {
			return "", fmt.Errorf("Unable to build inner SQL structure: %v", err)
		}

		if f.Alias != "" {
			fromTablesSQL = append(fromTablesSQL, fmt.Sprintf("(%s) AS %s", sql, f.Alias))
		} else {
			fromTablesSQL = append(fromTablesSQL, fmt.Sprintf("(%s)", sql))
		}
	}

	fromTablesStr := strings.Join(fromTablesSQL, ", ")

	var sb strings.Builder
	for _, j := range joinEntries {
		sb.WriteString(fmt.Sprintf("\nJOIN %s %s ", j.Table, j.Alias))

		if j.Index != "" {
			sb.WriteString(fmt.Sprintf("FORCE INDEX FOR JOIN (%s) ", j.Index))
		}

		sb.WriteString(fmt.Sprintf("ON %s", j.On))
	}

	joins := sb.String()

	sqlQuery := fmt.Sprintf("SELECT %s\nFROM (%s)%s", projectionsStr, fromTablesStr, joins)
	whereExprStr := whereExpressions.String()

	if whereExprStr != "" {
		sqlQuery += fmt.Sprintf("\nWHERE %s", whereExprStr)
	}

	if len(groupBy) > 0 {
		groupByProjection := make([]string, len(groupBy))
		for i := range groupBy {
			if projections[groupBy[i]].Alias != "" {
				groupByProjection[i] = projections[groupBy[i]].Alias
			} else {
				if projections[groupBy[i]].Function != nil {
					return "", fmt.Errorf("Unable to group by function, there should be an alias")
				}
				groupByProjection[i] = projections[groupBy[i]].Variable
			}
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
