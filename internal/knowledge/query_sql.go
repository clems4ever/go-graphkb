package knowledge

import (
	"fmt"
	"strings"

	"github.com/clems4ever/go-graphkb/internal/query"
)

type SQLQueryTranslator struct {
	NodeRepository     *QueryNodeRepository
	RelationRepository *QueryRelationRepository
}

func NewSQLQueryTranslator() *SQLQueryTranslator {
	return &SQLQueryTranslator{
		NodeRepository:     NewQueryNodeRepository(),
		RelationRepository: NewQueryRelationRepository(),
	}
}

type AndOrLinks struct {
	And  []AndOrLinks
	Or   []AndOrLinks
	Link *QueryLink
}

type QueryLink struct {
	From  int
	To    int
	Index int
}

type ExpressionType int

const (
	NodeExprType     ExpressionType = iota
	EdgeExprType     ExpressionType = iota
	PropertyExprType ExpressionType = iota
)

type ExpressionBuilder struct {
	NodeRepository     *QueryNodeRepository
	RelationRepository *QueryRelationRepository

	Aggregation    bool
	ExpressionType ExpressionType
}

func NewExpressionBuilder(nodeRepo *QueryNodeRepository, relationRepo *QueryRelationRepository) *ExpressionBuilder {
	return &ExpressionBuilder{
		NodeRepository:     nodeRepo,
		RelationRepository: relationRepo,
	}
}

func (sqt *ExpressionBuilder) buildPropertyOrLabelsExpression(q *query.QueryPropertyOrLabelsExpression) (string, error) {
	sqt.ExpressionType = PropertyExprType
	if q.Atom.Variable != nil {
		hasProperties := false
		propertiesPath := ".*"
		if len(q.PropertyLookups) > 0 {
			hasProperties = true
			propertiesPath = strings.Join(q.PropertyLookups, ".")
		}

		name := *q.Atom.Variable
		_, i := sqt.NodeRepository.ByName(name)
		if i > -1 {
			if !hasProperties {
				sqt.ExpressionType = NodeExprType
			}
			return fmt.Sprintf("a%d%s", i, propertiesPath), nil
		}

		_, i = sqt.RelationRepository.ByName(name)
		if i > -1 {
			if !hasProperties {
				sqt.ExpressionType = EdgeExprType
			}
			return fmt.Sprintf("r%d%s", i, propertiesPath), nil
		}
	} else if q.Atom.Literal != nil {
		if q.Atom.Literal.String != nil {
			return fmt.Sprintf("'%s'", *q.Atom.Literal.String), nil
		} else if q.Atom.Literal.Integer != nil {
			return fmt.Sprintf("%d", *q.Atom.Literal.Integer), nil
		} else if q.Atom.Literal.Double != nil {
			return fmt.Sprintf("%f", *q.Atom.Literal.Double), nil
		} else if q.Atom.Literal.Boolean != nil {
			if *q.Atom.Literal.Boolean {
				return "true", nil
			} else {
				return "false", nil
			}
		}
	} else if q.Atom.FunctionInvocation != nil {
		fnName := strings.ToUpper(q.Atom.FunctionInvocation.FunctionName)
		args := q.Atom.FunctionInvocation.Expressions
		if fnName != "COUNT" {
			return "", fmt.Errorf("Function %s is not supported", fnName)
		}
		argExpr := make([]string, 0)
		for _, a := range args {
			exprBuilder := NewExpressionBuilder(sqt.NodeRepository, sqt.RelationRepository)
			expr, err := exprBuilder.buildExpression(&a)
			if err != nil {
				return "", err
			}
			argExpr = append(argExpr, expr)
		}
		sqt.Aggregation = true
		return fmt.Sprintf("%s(%s)", fnName, strings.Join(argExpr, ", ")), nil
	} else if q.Atom.ParenthesizedExpression != nil {
		exprBuilder := NewExpressionBuilder(sqt.NodeRepository, sqt.RelationRepository)
		expr, err := exprBuilder.buildExpression(q.Atom.ParenthesizedExpression)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("(%s)", expr), nil
	}
	return "", fmt.Errorf("Unable to transform property or labels expression")
}

func (sqt *ExpressionBuilder) buildStringListNullOperatorExpression(q *query.QueryStringListNullOperatorExpression) (string, error) {
	expr, err := sqt.buildPropertyOrLabelsExpression(&q.PropertyOrLabelsExpression)
	if err != nil {
		return "", err
	}

	for i := range q.StringOperatorExpression {
		stringExpression := q.StringOperatorExpression[i]

		if stringExpression.PropertyOrLabelsExpression.Atom.Literal == nil {
			return "", fmt.Errorf("Expression must be a literal to be used with string operator")
		}

		if stringExpression.PropertyOrLabelsExpression.Atom.Literal.String == nil {
			return "", fmt.Errorf("Expression must be a string literal to be used with string operator")
		}

		rightExpr := *stringExpression.PropertyOrLabelsExpression.Atom.Literal.String

		switch stringExpression.Operator {
		case query.StartsWithOperator:
			expr = fmt.Sprintf("%s LIKE '%s%%'", expr, rightExpr)
		case query.EndsWithOperator:
			expr = fmt.Sprintf("%s LIKE '%%%s'", expr, rightExpr)
		case query.ContainsOperator:
			expr = fmt.Sprintf("%s LIKE '%%%s%%'", expr, rightExpr)
		}
	}
	return expr, nil
}

func (sqt *ExpressionBuilder) buildUnaryAddOrSubtractExpression(q *query.QueryUnaryAddOrSubtractExpression) (string, error) {
	operatorStr := ""
	if q.Negation {
		operatorStr = "-"
	}

	expr, err := sqt.buildStringListNullOperatorExpression(&q.StringListNullOperatorExpression)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s%s", operatorStr, expr), nil
}

func (sqt *ExpressionBuilder) buildPowerOfExpression(q *query.QueryPowerOfExpression) (string, error) {
	leftExpr, err := sqt.buildUnaryAddOrSubtractExpression(&q.QueryUnaryAddOrSubtractExpressions[0])
	if err != nil {
		return "", err
	}

	for i := 1; i < len(q.QueryUnaryAddOrSubtractExpressions); i++ {
		rightExpr, err := sqt.buildUnaryAddOrSubtractExpression(&q.QueryUnaryAddOrSubtractExpressions[i])
		if err != nil {
			return "", err
		}
		leftExpr = fmt.Sprintf("%s ^ %s", leftExpr, rightExpr)
	}
	return leftExpr, nil
}

func (sqt *ExpressionBuilder) buildMultipleDivideModuloExpressions(q *query.QueryMultipleDivideModuloExpression) (string, error) {
	leftExpr, err := sqt.buildPowerOfExpression(&q.PowerOfExpression)
	if err != nil {
		return "", err
	}

	for _, pmdme := range q.PartialMultipleDivideModuloExpressions {
		operatorStr := ""
		rightExpr, err := sqt.buildPowerOfExpression(&pmdme.QueryPowerOfExpression)
		if err != nil {
			return "", err
		}
		switch pmdme.MultiplyDivideOperator {
		case query.Multiply:
			operatorStr = "*"
		case query.Divide:
			operatorStr = "/"
		case query.Modulo:
			operatorStr = "%"
		}
		leftExpr = fmt.Sprintf("%s %s %s", leftExpr, operatorStr, rightExpr)
	}
	return leftExpr, nil
}

func (sqt *ExpressionBuilder) buildAddOrSubtractExpression(q *query.QueryAddOrSubtractExpression) (string, error) {
	leftExpr, err := sqt.buildMultipleDivideModuloExpressions(&q.MultipleDivideModuloExpression)
	if err != nil {
		return "", err
	}
	for _, pase := range q.PartialAddOrSubtractExpression {
		operatorStr := ""
		rightExpr, err := sqt.buildMultipleDivideModuloExpressions(&pase.MultipleDivideModuloExpression)
		if err != nil {
			return "", err
		}
		switch pase.AddOrSubtractOperator {
		case query.Add:
			operatorStr = "+"
		case query.Subtract:
			operatorStr = "-"
		}
		leftExpr = fmt.Sprintf("%s %s %s", leftExpr, operatorStr, rightExpr)
	}
	return leftExpr, nil
}

func (sqt *ExpressionBuilder) buildComparisonExpression(q *query.QueryComparisonExpression) (string, error) {
	leftExpr, err := sqt.buildAddOrSubtractExpression(&q.AddOrSubtractExpression)
	if err != nil {
		return "", err
	}

	for _, pce := range q.PartialComparisonExpressions {
		operatorStr := ""
		rightExpr, err := sqt.buildAddOrSubtractExpression(&pce.AddOrSubtractExpression)
		if err != nil {
			return "", err
		}

		switch pce.ComparisonOperator {
		case query.Equal:
			operatorStr = "="
		case query.NotEqual:
			operatorStr = "<>"
		case query.Less:
			operatorStr = "<"
		case query.LessOrEqual:
			operatorStr = "<="
		case query.Greater:
			operatorStr = ">"
		case query.GreaterOrEqual:
			operatorStr = ">="
		}
		leftExpr = fmt.Sprintf("%s %s %s", leftExpr, operatorStr, rightExpr)
	}
	return leftExpr, nil
}

func (sqt *ExpressionBuilder) buildExpression(q *query.QueryExpression) (string, error) {
	orFilters := make([]string, 0)
	for _, xorExpr := range q.OrExpression.XorExpressions {
		xorFilters := make([]string, 0)
		for _, andExpr := range xorExpr.AndExpressions {
			andFilters := make([]string, 0)
			for _, notExpr := range andExpr.NotExpressions {
				expr, err := sqt.buildComparisonExpression(&notExpr.ComparisonExpression)
				if err != nil {
					return "", err
				}
				if notExpr.Not {
					expr = fmt.Sprintf("NOT %s", expr)
				}
				andFilters = append(andFilters, expr)
			}
			xorFilters = append(xorFilters, strings.Join(andFilters, " AND "))
		}
		orFilters = append(orFilters, strings.Join(xorFilters, " XOR "))
	}
	return strings.Join(orFilters, " OR "), nil
}

type Projection struct {
	Alias          string
	ExpressionType ExpressionType
}

type SQLTranslation struct {
	Query           string
	ProjectionTypes []Projection
}

func BuildAndOrExpressionFromLinks(links AndOrLinks) (string, error) {
	if links.Link != nil {
		rName := fmt.Sprintf("r%d", links.Link.Index)
		fromName := fmt.Sprintf("a%d", links.Link.From)
		toName := fmt.Sprintf("a%d", links.Link.To)
		return fmt.Sprintf("%s.from_id = %s.id AND %s.to_id = %s.id", rName, fromName, rName, toName), nil
	} else if len(links.And) > 0 {
		exprs := make([]string, 0)
		for i := range links.And {
			expr, err := BuildAndOrExpressionFromLinks(links.And[i])
			if err != nil {
				return "", err
			}
			if expr != "" {
				e := expr
				if len(links.And) > 1 {
					e = fmt.Sprintf("(%s)", expr)
				}
				exprs = append(exprs, e)
			}
		}
		return strings.Join(exprs, " AND "), nil
	} else if len(links.Or) > 0 {
		exprs := make([]string, 0)
		for i := range links.Or {
			expr, err := BuildAndOrExpressionFromLinks(links.Or[i])
			if err != nil {
				return "", err
			}
			if expr != "" {
				e := expr
				if len(links.Or) > 1 {
					e = fmt.Sprintf("(%s)", expr)
				}
				exprs = append(exprs, e)
			}
		}
		return strings.Join(exprs, " OR "), nil
	}
	return "", nil
}

func (sqt *SQLQueryTranslator) Translate(queryIL *query.QueryIL) (*SQLTranslation, error) {
	matchWhereFilters := make([]string, 0)

	matchLinks := AndOrLinks{}
	matchLinks.And = make([]AndOrLinks, 0)
	for _, x := range queryIL.QueryCypher.QuerySinglePartQuery.QueryMatches {
		orLinks := make([]AndOrLinks, 0)
		for _, y := range x.PatternElements {
			i1, err := sqt.NodeRepository.PushNode(&y.QueryNodePattern)
			if err != nil {
				return nil, err
			}

			andLinks := make([]AndOrLinks, 0)

			for _, z := range y.QueryPatternElementChains {
				i2, err := sqt.NodeRepository.PushNode(&z.QueryNodePattern)
				if err != nil {
					return nil, err
				}

				var qrd query.QueryRelationshipDetail

				if z.RelationshipPattern.RelationshipDetail != nil {
					qrd = *z.RelationshipPattern.RelationshipDetail
				} else {
					qrd.Labels = make([]string, 0)
				}
				idx, err := sqt.RelationRepository.PushRelation(&qrd)
				if err != nil {
					return nil, err
				}

				qinbound := QueryLink{From: i2, To: i1, Index: idx}
				qoutbound := QueryLink{From: i1, To: i2, Index: idx}

				if !z.RelationshipPattern.LeftArrow && !z.RelationshipPattern.RightArrow {
					andLinks2 := AndOrLinks{Link: &qinbound}
					andLinks3 := AndOrLinks{Link: &qoutbound}
					orLinks := AndOrLinks{Or: []AndOrLinks{andLinks2, andLinks3}}
					andLinks = append(andLinks, orLinks)
				} else if z.RelationshipPattern.LeftArrow && z.RelationshipPattern.RightArrow {
					andLinks = append(andLinks,
						AndOrLinks{And: []AndOrLinks{AndOrLinks{Link: &qoutbound}, AndOrLinks{Link: &qinbound}}})
				} else if z.RelationshipPattern.LeftArrow {
					andLinks = append(andLinks, AndOrLinks{Link: &qinbound})
				} else if z.RelationshipPattern.RightArrow {
					andLinks = append(andLinks, AndOrLinks{Link: &qoutbound})
				}
				i1 = i2
			}
			orLinks = append(orLinks, AndOrLinks{And: andLinks})
		}
		matchLinks.And = append(matchLinks.And, AndOrLinks{Or: orLinks})

		if x.Where != nil {
			whereFilter, err := NewExpressionBuilder(sqt.NodeRepository, sqt.RelationRepository).
				buildExpression(x.Where)
			if err != nil {
				return nil, err
			}
			matchWhereFilters = append(matchWhereFilters, whereFilter)
		}
	}

	projections := make([]string, 0)
	projectionTypes := make([]Projection, 0)
	from := make([]string, 0)
	groupingKeys := make([]string, 0)
	aggregationRequired := false

	for _, p := range queryIL.QuerySinglePartQuery.ProjectionBody.ProjectionItems {
		builder := NewExpressionBuilder(sqt.NodeRepository, sqt.RelationRepository)
		projection, err := builder.buildExpression(&p.Expression)
		if err != nil {
			return nil, err
		}

		if !builder.Aggregation {
			groupingKeys = append(groupingKeys, projection)
		} else {
			aggregationRequired = true
		}

		projections = append(projections, projection)
		projectionTypes = append(projectionTypes, Projection{
			Alias:          p.Alias,
			ExpressionType: builder.ExpressionType,
		})
	}

	typeFiltersAnd := make([]string, 0)
	for i, n := range sqt.NodeRepository.Nodes() {
		from = append(from, fmt.Sprintf("assets a%d", i))

		typeFiltersOr := make([]string, 0)
		for _, l := range n.Labels {
			typeFiltersOr = append(typeFiltersOr, fmt.Sprintf("a%d.type = '%s'", i, l))
		}
		if len(typeFiltersOr) > 0 {
			if len(typeFiltersOr) == 1 {
				typeFiltersAnd = append(typeFiltersAnd, strings.Join(typeFiltersOr, " OR "))
			} else {
				typeFiltersAnd = append(typeFiltersAnd, fmt.Sprintf("(%s)", strings.Join(typeFiltersOr, " OR ")))
			}
		}
	}
	for i, r := range sqt.RelationRepository.Relations() {
		from = append(from, fmt.Sprintf("relations r%d", i))
		typeFiltersOr := make([]string, 0)
		for _, l := range r.Labels {
			typeFiltersOr = append(typeFiltersOr, fmt.Sprintf("r%d.type = '%s'", i, l))
		}
		if len(typeFiltersOr) > 0 {
			if len(typeFiltersOr) == 1 {
				typeFiltersAnd = append(typeFiltersAnd, strings.Join(typeFiltersOr, " OR "))
			} else {
				typeFiltersAnd = append(typeFiltersAnd, fmt.Sprintf("(%s)", strings.Join(typeFiltersOr, " OR ")))
			}

		}
	}

	// Compute the relation constraints
	relationFilter, err := BuildAndOrExpressionFromLinks(matchLinks)
	if err != nil {
		return nil, err
	}

	distinctStr := ""
	if queryIL.QueryCypher.QuerySinglePartQuery.ProjectionBody.Distinct {
		distinctStr = "DISTINCT "
	}

	sqlQuery := fmt.Sprintf("SELECT %s%s FROM %s",
		distinctStr,
		strings.Join(projections, ", "),
		strings.Join(from, ", "))

	where := make([]string, 0)
	if len(typeFiltersAnd) > 0 {
		where = append(where, strings.Join(typeFiltersAnd, " AND "))
	}
	if relationFilter != "" {
		where = append(where, fmt.Sprintf("(%s)", relationFilter))
	}
	if len(matchWhereFilters) > 0 {
		where = append(where, strings.Join(matchWhereFilters, " AND "))
	}

	if len(where) > 0 {
		sqlQuery += fmt.Sprintf("\nWHERE %s", strings.Join(where, "\nAND "))
	}

	if aggregationRequired && len(groupingKeys) > 0 {
		sqlQuery = fmt.Sprintf("%s\nGROUP BY %s", sqlQuery, strings.Join(groupingKeys, ", "))
	}

	limit := queryIL.QueryCypher.QuerySinglePartQuery.ProjectionBody.Limit
	skip := queryIL.QueryCypher.QuerySinglePartQuery.ProjectionBody.Skip
	if limit != nil {
		expr, err := NewExpressionBuilder(sqt.NodeRepository, sqt.RelationRepository).
			buildExpression(limit)
		if err != nil {
			return nil, err
		}
		sqlQuery = fmt.Sprintf("%s\nLIMIT %s", sqlQuery, expr)
	}

	if skip != nil {
		if limit == nil {
			return nil, fmt.Errorf("SKIP must be used in combination with limit")
		}
		expr, err := NewExpressionBuilder(sqt.NodeRepository, sqt.RelationRepository).
			buildExpression(skip)
		if err != nil {
			return nil, err
		}
		sqlQuery = fmt.Sprintf("%s\nOFFSET %s", sqlQuery, expr)
	}

	return &SQLTranslation{
		Query:           sqlQuery,
		ProjectionTypes: projectionTypes,
	}, nil
}

// QueryNode represent a node type
type QueryNode struct {
	Labels []string
}

// QueryNodeRepository represent a node repository
type QueryNodeRepository struct {
	All           []QueryNode
	Anonymous     []QueryNode
	Named         map[string]QueryNode
	NamedPosition map[string]int
}

// NewQueryNodeRepository create a node repository
func NewQueryNodeRepository() *QueryNodeRepository {
	qnr := new(QueryNodeRepository)
	qnr.Anonymous = make([]QueryNode, 0)
	qnr.Named = make(map[string]QueryNode)
	qnr.All = make([]QueryNode, 0)
	qnr.NamedPosition = make(map[string]int)
	return qnr
}

// PushNode push a node in the repository
func (qnr *QueryNodeRepository) PushNode(q *query.QueryNodePattern) (int, error) {
	var idx int
	node := QueryNode{Labels: q.Labels}

	toAdd := false

	if q.Variable == "" {
		qnr.Anonymous = append(qnr.Anonymous, node)
		toAdd = true
	} else {
		v, found := qnr.Named[q.Variable]
		if !found {
			qnr.Named[q.Variable] = node
			qnr.NamedPosition[q.Variable] = len(qnr.All)
			toAdd = true
		} else if found {
			if len(node.Labels) > 0 && !StringSliceElementsEqual(node.Labels, v.Labels) {
				return -1, fmt.Errorf("Redefinition of variable %s with different type", q.Variable)
			}
			idx = qnr.NamedPosition[q.Variable]
		}
	}

	if toAdd {
		idx = len(qnr.All)
		qnr.All = append(qnr.All, node)
	}
	return idx, nil
}

// Nodes return the list of nodes contained in repository
func (qnr *QueryNodeRepository) Nodes() []QueryNode {
	return qnr.All
}

// ByName return the query node attached to given variable name
func (qnr *QueryNodeRepository) ByName(name string) (*QueryNode, int) {
	v, ok := qnr.Named[name]
	if !ok {
		return nil, -1
	}
	return &v, qnr.NamedPosition[name]
}

// QueryRelation represent a relation
type QueryRelation struct {
	Labels []string
}

// QueryRelationRepository is a repository of relations
type QueryRelationRepository struct {
	All           []QueryRelation
	Anonymous     []QueryRelation
	Named         map[string]QueryRelation
	NamedPosition map[string]int
}

// NewQueryRelationRepository create a relation repository
func NewQueryRelationRepository() *QueryRelationRepository {
	qnr := new(QueryRelationRepository)
	qnr.Anonymous = make([]QueryRelation, 0)
	qnr.Named = make(map[string]QueryRelation)
	qnr.All = make([]QueryRelation, 0)
	qnr.NamedPosition = make(map[string]int)
	return qnr
}

func StringSliceElementsEqual(s1 []string, s2 []string) bool {
	if len(s1) != len(s2) {
		return false
	}

	for i := range s1 {
		el1 := s1[i]
		found := false
		for j := range s2 {
			if el1 == s2[j] {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// PushRelation push a relation into the repository
func (qnr *QueryRelationRepository) PushRelation(q *query.QueryRelationshipDetail) (int, error) {
	var idx int
	node := QueryRelation{Labels: q.Labels}

	toAdd := false

	if q.Variable == "" {
		qnr.Anonymous = append(qnr.Anonymous, node)
		toAdd = true
	} else {
		v, found := qnr.Named[q.Variable]
		if !found {
			qnr.Named[q.Variable] = node
			qnr.NamedPosition[q.Variable] = len(qnr.All)
			toAdd = true
		} else if found {
			if len(node.Labels) > 0 && !StringSliceElementsEqual(q.Labels, v.Labels) {
				return -1, fmt.Errorf("Redefinition of variable %s with different type", q.Variable)
			}
			idx = qnr.NamedPosition[q.Variable]
		}
	}

	if toAdd {
		idx = len(qnr.All)
		qnr.All = append(qnr.All, node)
	}
	return idx, nil
}

// Relations return the list of relations contained in the repository
func (qnr *QueryRelationRepository) Relations() []QueryRelation {
	return qnr.All
}

// ByName return the relation attached to given variable name
func (qnr *QueryRelationRepository) ByName(name string) (*QueryRelation, int) {
	v, ok := qnr.Named[name]
	if !ok {
		return nil, -1
	}
	return &v, qnr.NamedPosition[name]
}
