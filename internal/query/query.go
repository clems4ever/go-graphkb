package query

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/antlr/antlr4/runtime/Go/antlr/v4"
	"github.com/clems4ever/go-graphkb/internal/parser"
)

// ParsingError represent a parsing error
type ParsingError struct {
	Line    int
	Column  int
	Message string
}

// ParsingErrorListener a listener for raising errors during parsing
type ParsingErrorListener struct {
	*antlr.DefaultErrorListener
	Errors []ParsingError
}

func NewParsingErrorListener() *ParsingErrorListener {
	return &ParsingErrorListener{
		Errors: make([]ParsingError, 0),
	}
}

func (pel *ParsingErrorListener) SyntaxError(recognizer antlr.Recognizer, offendingSymbol interface{}, line, column int, msg string, e antlr.RecognitionException) {
	pel.Errors = append(pel.Errors, ParsingError{
		Line:    line,
		Column:  column,
		Message: msg,
	})
}

// TransformCypher transform an openCypher query into a QueryCypher structure.
func TransformCypher(query string) (*QueryCypher, error) {
	is := antlr.NewInputStream(query)
	lexer := parser.NewCypherLexer(is)
	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	p := parser.NewCypherParser(stream)

	pel := NewParsingErrorListener()
	p.AddErrorListener(pel)

	l := NewCypherVisitor()
	queryCypher := l.Visit(p.OC_Cypher())

	if len(pel.Errors) > 0 {
		errStr := []string{}
		for _, e := range pel.Errors {
			errStr = append(errStr, fmt.Sprintf("line %d:%d - %s", e.Line, e.Column, e.Message))
		}
		return nil, fmt.Errorf("Parsing errors detected: %s", strings.Join(errStr, ", "))
	}

	switch v := queryCypher.(type) {
	case QueryCypher:
		return &v, nil
	case error:
		return nil, v
	}
	return nil, fmt.Errorf("Unable to detect type of IL")
}

// BaseCypherVisitor visitor for cypher
type BaseCypherVisitor struct {
	parser.BaseCypherVisitor

	errors []error
}

// NewCypherVisitor create a visitor for cypher
func NewCypherVisitor() *BaseCypherVisitor {
	return new(BaseCypherVisitor)
}

// VariableType represent a variable type
type VariableType int

const (
	// Node represent the type of a node
	Node VariableType = iota
	// Edge represent the type of edge
	Edge VariableType = iota
	// Unknown represent an unknown type for a variable
	Unknown VariableType = iota
)

// AppendError append one error to visitor
func (cl *BaseCypherVisitor) AppendError(err error) {
	cl.errors = append(cl.errors, err)
}

// AppendErrors append mulitple errors to visitor
func (cl *BaseCypherVisitor) AppendErrors(err ...error) {
	cl.errors = append(cl.errors, err...)
}

// Visit the AST
func (cl *BaseCypherVisitor) Visit(tree antlr.ParseTree) interface{} {
	return tree.Accept(cl)
}

// QueryCypher the representation of the query in IL
type QueryCypher struct {
	QuerySinglePartQuery
}

// VisitOC_Cypher visit cypher
func (cl *BaseCypherVisitor) VisitOC_Cypher(c *parser.OC_CypherContext) interface{} {
	q := QueryCypher{}
	if c.OC_Statement() != nil {
		switch v := c.OC_Statement().Accept(cl).(type) {
		case QuerySinglePartQuery:
			q.QuerySinglePartQuery = v
		case error:
			return v
		}
	}
	return q
}

// VisitOC_Statement visit statement
func (cl *BaseCypherVisitor) VisitOC_Statement(c *parser.OC_StatementContext) interface{} {
	if c.OC_Query() != nil {
		return c.OC_Query().Accept(cl)
	}
	return fmt.Errorf("Unable to parse statement")
}

func (cl *BaseCypherVisitor) VisitOC_Query(c *parser.OC_QueryContext) interface{} {
	if c.OC_RegularQuery() != nil {
		return c.OC_RegularQuery().Accept(cl)
	}
	return fmt.Errorf("Unable to parse Cypher query")
}

func (cl *BaseCypherVisitor) VisitOC_RegularQuery(c *parser.OC_RegularQueryContext) interface{} {
	if c.OC_SingleQuery() != nil {
		return c.OC_SingleQuery().Accept(cl)
	}
	return fmt.Errorf("Unabel to parse regular query")
}

func (cl *BaseCypherVisitor) VisitOC_SingleQuery(c *parser.OC_SingleQueryContext) interface{} {
	if c.OC_SinglePartQuery() != nil {
		v := c.OC_SinglePartQuery().Accept(cl)
		return v
	}
	if c.OC_MultiPartQuery() != nil {
		v := c.OC_MultiPartQuery().Accept(cl)
		return v
	}
	return fmt.Errorf("Unable to parse single query")
}

type QuerySinglePartQuery struct {
	QueryMatches    []QueryMatch
	ProjectionBody  QueryProjectionBody
	WithProjections []QueryWith
}

func (cl *BaseCypherVisitor) VisitOC_SinglePartQuery(c *parser.OC_SinglePartQueryContext) interface{} {
	q := QuerySinglePartQuery{}
	q.QueryMatches = make([]QueryMatch, 0)

	for i := range c.AllOC_ReadingClause() {
		match := c.OC_ReadingClause(i).Accept(cl).(QueryMatch)
		q.QueryMatches = append(q.QueryMatches, match)
	}
	switch v := c.OC_Return().Accept(cl).(type) {
	case QueryProjectionBody:
		q.ProjectionBody = v
	case error:
		return v
	}
	return q
}

func (cl *BaseCypherVisitor) VisitOC_MultiPartQuery(c *parser.OC_MultiPartQueryContext) interface{} {
	q := QuerySinglePartQuery{}
	q.QueryMatches = make([]QueryMatch, 0)

	singlePartQueryContext := c.OC_SinglePartQuery().(*parser.OC_SinglePartQueryContext)
	singlePartQuery := cl.VisitOC_SinglePartQuery(singlePartQueryContext).(QuerySinglePartQuery)

	for i := range c.AllOC_ReadingClause() {
		match := c.OC_ReadingClause(i).Accept(cl).(QueryMatch)
		q.QueryMatches = append(q.QueryMatches, match)
	}

	q.QueryMatches = append(q.QueryMatches, singlePartQuery.QueryMatches...)
	q.ProjectionBody = singlePartQuery.ProjectionBody

	for i := range c.AllOC_With() {
		v := c.OC_With(i).Accept(cl)
		q.WithProjections = append(q.WithProjections,
			v.(QueryWith))
	}

	return q
}

type QueryWith struct {
	Where          *QueryExpression
	ProjectionBody QueryProjectionBody
}

func (cl *BaseCypherVisitor) VisitOC_With(c *parser.OC_WithContext) interface{} {
	q := QueryWith{}
	q.ProjectionBody = c.OC_ProjectionBody().Accept(cl).(QueryProjectionBody)
	if c.OC_Where() != nil {
		switch v := c.OC_Where().Accept(cl).(type) {
		case QueryExpression:
			q.Where = &v
		case error:
			return v
		}
	}
	return q
}

func (cl *BaseCypherVisitor) VisitOC_Return(c *parser.OC_ReturnContext) interface{} {
	return c.OC_ProjectionBody().Accept(cl)
}

type QueryProjectionItem struct {
	Expression QueryExpression
	Alias      string
}

type QueryProjectionBody struct {
	Distinct        bool
	ProjectionItems []QueryProjectionItem
	Limit           *QueryExpression
	Skip            *QueryExpression
}

func (cl *BaseCypherVisitor) VisitOC_ProjectionBody(c *parser.OC_ProjectionBodyContext) interface{} {
	q := QueryProjectionBody{}
	q.Distinct = c.DISTINCT() != nil

	if c.OC_ProjectionItems() != nil {
		switch v := c.OC_ProjectionItems().Accept(cl).(type) {
		case []QueryProjectionItem:
			q.ProjectionItems = v
		case error:
			return v
		}
	} else {
		return fmt.Errorf("Unable to parse projection items")
	}

	if c.OC_Limit() != nil {
		q.Limit = new(QueryExpression)
		*q.Limit = c.OC_Limit().Accept(cl).(QueryExpression)
	}
	if c.OC_Skip() != nil {
		q.Skip = new(QueryExpression)
		*q.Skip = c.OC_Skip().Accept(cl).(QueryExpression)
	}
	return q
}

func (cl *BaseCypherVisitor) VisitOC_Limit(c *parser.OC_LimitContext) interface{} {
	return c.OC_Expression().Accept(cl)
}

func (cl *BaseCypherVisitor) VisitOC_Skip(c *parser.OC_SkipContext) interface{} {
	return c.OC_Expression().Accept(cl)
}

func (cl *BaseCypherVisitor) VisitOC_ProjectionItems(c *parser.OC_ProjectionItemsContext) interface{} {
	items := make([]QueryProjectionItem, 0)
	for i := range c.AllOC_ProjectionItem() {
		item := c.OC_ProjectionItem(i).Accept(cl).(QueryProjectionItem)
		items = append(items, item)
	}
	return items
}

func (cl *BaseCypherVisitor) VisitOC_ProjectionItem(c *parser.OC_ProjectionItemContext) interface{} {
	item := QueryProjectionItem{}

	item.Expression = c.OC_Expression().Accept(cl).(QueryExpression)

	if c.OC_Variable() != nil {
		item.Alias = c.OC_Variable().Accept(cl).(string)
	} else {
		item.Alias = c.GetText()
	}
	return item
}

func (cl *BaseCypherVisitor) VisitOC_ReadingClause(c *parser.OC_ReadingClauseContext) interface{} {
	return c.OC_Match().Accept(cl)
}

type QueryMatch struct {
	PatternElements []QueryPatternElement
	Where           *QueryExpression
}

func (cl *BaseCypherVisitor) VisitOC_Match(c *parser.OC_MatchContext) interface{} {
	q := QueryMatch{}
	q.PatternElements = c.OC_Pattern().Accept(cl).([]QueryPatternElement)
	if c.OC_Where() != nil {
		switch v := c.OC_Where().Accept(cl).(type) {
		case QueryExpression:
			q.Where = &v
		case error:
			return v
		}
	}
	return q
}

func (cl *BaseCypherVisitor) VisitOC_Where(c *parser.OC_WhereContext) interface{} {
	return c.OC_Expression().Accept(cl).(QueryExpression)
}

type QueryExpression struct {
	OrExpression QueryOrExpression
}

func (cl *BaseCypherVisitor) VisitOC_Expression(c *parser.OC_ExpressionContext) interface{} {
	q := QueryExpression{}
	q.OrExpression = c.OC_OrExpression().Accept(cl).(QueryOrExpression)
	return q
}

type QueryOrExpression struct {
	XorExpressions []QueryXorExpression
}

func (cl *BaseCypherVisitor) VisitOC_OrExpression(c *parser.OC_OrExpressionContext) interface{} {
	q := QueryOrExpression{}
	items := make([]QueryXorExpression, 0)
	for i := range c.AllOC_XorExpression() {
		items = append(items, c.OC_XorExpression(i).Accept(cl).(QueryXorExpression))
	}
	q.XorExpressions = items
	return q
}

type QueryXorExpression struct {
	AndExpressions []QueryAndExpression
}

func (cl *BaseCypherVisitor) VisitOC_XorExpression(c *parser.OC_XorExpressionContext) interface{} {
	q := QueryXorExpression{}
	items := make([]QueryAndExpression, 0)
	for i := range c.AllOC_AndExpression() {
		items = append(items, c.OC_AndExpression(i).Accept(cl).(QueryAndExpression))
	}
	q.AndExpressions = items
	return q
}

type QueryAndExpression struct {
	NotExpressions []QueryNotExpression
}

func (cl *BaseCypherVisitor) VisitOC_AndExpression(c *parser.OC_AndExpressionContext) interface{} {
	q := QueryAndExpression{}
	items := make([]QueryNotExpression, 0)
	for i := range c.AllOC_NotExpression() {
		items = append(items, c.OC_NotExpression(i).Accept(cl).(QueryNotExpression))
	}
	q.NotExpressions = items
	return q
}

type QueryNotExpression struct {
	Not                  bool
	ComparisonExpression QueryComparisonExpression
}

func (cl *BaseCypherVisitor) VisitOC_NotExpression(c *parser.OC_NotExpressionContext) interface{} {
	q := QueryNotExpression{}
	q.Not = len(c.AllNOT())%2 == 1
	q.ComparisonExpression = c.OC_ComparisonExpression().Accept(cl).(QueryComparisonExpression)
	return q
}

type QueryComparisonExpression struct {
	AddOrSubtractExpression      QueryAddOrSubtractExpression
	PartialComparisonExpressions []QueryPartialComparisonExpression
}

type AddOrSubstractExpression string
type ComparisonOperator int

const (
	UnknownComparison = iota
	Equal             = iota
	NotEqual          = iota
	Less              = iota
	Greater           = iota
	LessOrEqual       = iota
	GreaterOrEqual    = iota
)

func (cl *BaseCypherVisitor) VisitOC_ComparisonExpression(c *parser.OC_ComparisonExpressionContext) interface{} {
	q := QueryComparisonExpression{}
	q.AddOrSubtractExpression = c.OC_AddOrSubtractExpression().Accept(cl).(QueryAddOrSubtractExpression)
	q.PartialComparisonExpressions = make([]QueryPartialComparisonExpression, 0)
	for i := range c.AllOC_PartialComparisonExpression() {
		q.PartialComparisonExpressions = append(q.PartialComparisonExpressions,
			c.OC_PartialComparisonExpression(i).Accept(cl).(QueryPartialComparisonExpression))
	}
	return q
}

type QueryPartialComparisonExpression struct {
	ComparisonOperator      ComparisonOperator
	AddOrSubtractExpression QueryAddOrSubtractExpression
}

func (cl *BaseCypherVisitor) VisitOC_PartialComparisonExpression(c *parser.OC_PartialComparisonExpressionContext) interface{} {
	q := QueryPartialComparisonExpression{}
	q.ComparisonOperator = UnknownComparison
	opStr := c.GetChild(0).GetPayload().(*antlr.CommonToken).GetText()
	switch opStr {
	case "=":
		q.ComparisonOperator = Equal
	case "<>":
		q.ComparisonOperator = NotEqual
	case "<":
		q.ComparisonOperator = Less
	case ">":
		q.ComparisonOperator = Greater
	case "<=":
		q.ComparisonOperator = LessOrEqual
	case ">=":
		q.ComparisonOperator = GreaterOrEqual
	}
	q.AddOrSubtractExpression = c.OC_AddOrSubtractExpression().Accept(cl).(QueryAddOrSubtractExpression)
	return q
}

type QueryPartialAddOrSubtractExpression struct {
	AddOrSubtractOperator          AddOrSubtractOperator
	MultipleDivideModuloExpression QueryMultipleDivideModuloExpression
}

type QueryAddOrSubtractExpression struct {
	MultipleDivideModuloExpression QueryMultipleDivideModuloExpression
	PartialAddOrSubtractExpression []QueryPartialAddOrSubtractExpression
}

func (cl *BaseCypherVisitor) VisitOC_AddOrSubtractExpression(c *parser.OC_AddOrSubtractExpressionContext) interface{} {
	q := QueryAddOrSubtractExpression{}
	q.MultipleDivideModuloExpression = c.OC_MultiplyDivideModuloExpression(0).Accept(cl).(QueryMultipleDivideModuloExpression)
	items := make([]QueryPartialAddOrSubtractExpression, 0)
	for i := 1; i < len(c.AllOC_MultiplyDivideModuloExpression()); i++ {
		qi := QueryPartialAddOrSubtractExpression{}
		qi.AddOrSubtractOperator = Add
		qi.MultipleDivideModuloExpression = c.OC_MultiplyDivideModuloExpression(i).Accept(cl).(QueryMultipleDivideModuloExpression)
		items = append(items, qi)
	}
	q.PartialAddOrSubtractExpression = items
	return q
}

type MultiplyDivideModuloOperator int

const (
	Multiply MultiplyDivideModuloOperator = iota
	Divide   MultiplyDivideModuloOperator = iota
	Modulo   MultiplyDivideModuloOperator = iota
)

type QueryPartialMultipleDivideModuloExpression struct {
	MultiplyDivideOperator MultiplyDivideModuloOperator
	QueryPowerOfExpression QueryPowerOfExpression
}

type QueryMultipleDivideModuloExpression struct {
	PowerOfExpression                      QueryPowerOfExpression
	PartialMultipleDivideModuloExpressions []QueryPartialMultipleDivideModuloExpression
}

func (cl *BaseCypherVisitor) VisitOC_MultiplyDivideModuloExpression(c *parser.OC_MultiplyDivideModuloExpressionContext) interface{} {
	q := QueryMultipleDivideModuloExpression{}
	q.PowerOfExpression = c.OC_PowerOfExpression(0).Accept(cl).(QueryPowerOfExpression)

	items := make([]QueryPartialMultipleDivideModuloExpression, 0)
	for i := 1; i < len(c.AllOC_PowerOfExpression()); i++ {
		qi := QueryPartialMultipleDivideModuloExpression{}
		qi.MultiplyDivideOperator = Multiply
		qi.QueryPowerOfExpression = c.OC_PowerOfExpression(i).Accept(cl).(QueryPowerOfExpression)
		items = append(items, qi)
	}
	q.PartialMultipleDivideModuloExpressions = items
	return q
}

type QueryPowerOfExpression struct {
	QueryUnaryAddOrSubtractExpressions []QueryUnaryAddOrSubtractExpression
}

func (cl *BaseCypherVisitor) VisitOC_PowerOfExpression(c *parser.OC_PowerOfExpressionContext) interface{} {
	q := QueryPowerOfExpression{}
	q.QueryUnaryAddOrSubtractExpressions = make([]QueryUnaryAddOrSubtractExpression, 0)
	for i := range c.AllOC_UnaryAddOrSubtractExpression() {
		q.QueryUnaryAddOrSubtractExpressions = append(q.QueryUnaryAddOrSubtractExpressions,
			c.OC_UnaryAddOrSubtractExpression(i).Accept(cl).(QueryUnaryAddOrSubtractExpression))
	}
	return q
}

type AddOrSubtractOperator int

const (
	Add      AddOrSubtractOperator = iota
	Subtract AddOrSubtractOperator = iota
)

type QueryUnaryAddOrSubtractExpression struct {
	StringListNullOperatorExpression QueryStringListNullOperatorExpression
	Negation                         bool
}

func (cl *BaseCypherVisitor) VisitOC_UnaryAddOrSubtractExpression(c *parser.OC_UnaryAddOrSubtractExpressionContext) interface{} {
	q := QueryUnaryAddOrSubtractExpression{}
	q.StringListNullOperatorExpression = c.OC_StringListNullOperatorExpression().Accept(cl).(QueryStringListNullOperatorExpression)
	return q
}

type QueryStringListNullOperatorExpression struct {
	PropertyOrLabelsExpression QueryPropertyOrLabelsExpression
	StringOperatorExpression   []QueryStringOperatorExpression
}

func (cl *BaseCypherVisitor) VisitOC_StringListNullOperatorExpression(c *parser.OC_StringListNullOperatorExpressionContext) interface{} {
	q := QueryStringListNullOperatorExpression{}
	q.PropertyOrLabelsExpression = c.OC_PropertyOrLabelsExpression().Accept(cl).(QueryPropertyOrLabelsExpression)

	items := make([]QueryStringOperatorExpression, 0)
	for i := range c.AllOC_StringOperatorExpression() {
		items = append(items, c.OC_StringOperatorExpression(i).Accept(cl).(QueryStringOperatorExpression))
	}
	q.StringOperatorExpression = items
	return q
}

type StringOperator int

const (
	StartsWithOperator StringOperator = 0
	EndsWithOperator   StringOperator = 1
	ContainsOperator   StringOperator = 2
)

type QueryStringOperatorExpression struct {
	PropertyOrLabelsExpression QueryPropertyOrLabelsExpression
	Operator                   StringOperator
}

func (cl *BaseCypherVisitor) VisitOC_StringOperatorExpression(c *parser.OC_StringOperatorExpressionContext) interface{} {
	q := QueryStringOperatorExpression{}
	q.PropertyOrLabelsExpression = c.OC_PropertyOrLabelsExpression().Accept(cl).(QueryPropertyOrLabelsExpression)
	if c.CONTAINS() != nil {
		q.Operator = ContainsOperator
	} else if c.STARTS() != nil {
		q.Operator = StartsWithOperator
	} else if c.ENDS() != nil {
		q.Operator = EndsWithOperator
	} else {
		cl.AppendError(fmt.Errorf("Unable to detect string operator"))
		return nil
	}
	return q
}

type QueryPropertyOrLabelsExpression struct {
	Atom         QueryAtom
	PropertyKeys []string
}

func (cl *BaseCypherVisitor) VisitOC_PropertyOrLabelsExpression(c *parser.OC_PropertyOrLabelsExpressionContext) interface{} {
	q := QueryPropertyOrLabelsExpression{}
	q.Atom = c.OC_Atom().Accept(cl).(QueryAtom)

	propLookups := make([]string, 0)
	for i := range c.AllOC_PropertyLookup() {
		propLookups = append(propLookups, c.OC_PropertyLookup(i).Accept(cl).(string))
	}
	q.PropertyKeys = propLookups
	return q
}

func (cl *BaseCypherVisitor) VisitOC_PropertyLookup(c *parser.OC_PropertyLookupContext) interface{} {
	return c.OC_PropertyKeyName().GetText()
}

type QueryAtom struct {
	Variable                *string
	Literal                 *QueryLiteral
	FunctionInvocation      *QueryFunctionInvocation
	ParenthesizedExpression *QueryExpression
	RelationshipsPattern    *QueryRelationshipsPattern
}

func (cl *BaseCypherVisitor) VisitOC_Atom(c *parser.OC_AtomContext) interface{} {
	q := QueryAtom{}
	if c.OC_Variable() != nil {
		q.Variable = new(string)
		*q.Variable = c.OC_Variable().GetText()
	} else if c.OC_Literal() != nil {
		q.Literal = new(QueryLiteral)
		*q.Literal = c.OC_Literal().Accept(cl).(QueryLiteral)
	} else if c.OC_FunctionInvocation() != nil {
		q.FunctionInvocation = new(QueryFunctionInvocation)
		*q.FunctionInvocation = c.OC_FunctionInvocation().Accept(cl).(QueryFunctionInvocation)
	} else if c.OC_ParenthesizedExpression() != nil {
		q.ParenthesizedExpression = new(QueryExpression)
		*q.ParenthesizedExpression = c.OC_ParenthesizedExpression().Accept(cl).(QueryExpression)
	} else if c.OC_RelationshipsPattern() != nil {
		q.RelationshipsPattern = new(QueryRelationshipsPattern)
		*q.RelationshipsPattern = c.OC_RelationshipsPattern().Accept(cl).(QueryRelationshipsPattern)
	}
	return q
}

type QueryRelationshipsPattern struct {
	QueryNodePattern
	QueryPatternElementChains []QueryPatternElementChain
}

func (cl *BaseCypherVisitor) VisitOC_RelationshipsPattern(c *parser.OC_RelationshipsPatternContext) interface{} {
	relPattern := QueryRelationshipsPattern{}
	relPattern.QueryNodePattern = c.OC_NodePattern().Accept(cl).(QueryNodePattern)

	elemChains := make([]QueryPatternElementChain, 0)
	for i := range c.AllOC_PatternElementChain() {
		elemChains = append(elemChains, c.OC_PatternElementChain(i).Accept(cl).(QueryPatternElementChain))
	}

	relPattern.QueryPatternElementChains = elemChains
	return relPattern
}

func (cl *BaseCypherVisitor) VisitOC_ParenthesizedExpression(c *parser.OC_ParenthesizedExpressionContext) interface{} {
	return c.OC_Expression().Accept(cl)
}

type QueryFunctionInvocation struct {
	FunctionName string
	Expressions  []QueryExpression
	Distinct     bool
}

func (cl *BaseCypherVisitor) VisitOC_FunctionInvocation(c *parser.OC_FunctionInvocationContext) interface{} {
	q := QueryFunctionInvocation{}
	expressions := make([]QueryExpression, 0)
	if c.DISTINCT() != nil {
		q.Distinct = true
	}

	for i := range c.AllOC_Expression() {
		expressions = append(expressions, c.OC_Expression(i).Accept(cl).(QueryExpression))
	}
	q.FunctionName = c.OC_FunctionName().GetText()
	q.Expressions = expressions
	return q
}

type QueryLiteral struct {
	String  *string
	Integer *int64
	Double  *float64
	Boolean *bool
}

func (cl *BaseCypherVisitor) VisitOC_Literal(c *parser.OC_LiteralContext) interface{} {
	q := QueryLiteral{}
	if c.StringLiteral() != nil {
		q.String = new(string)
		token := c.StringLiteral().GetText()
		*q.String = token[1 : len(token)-1]
	} else if c.OC_NumberLiteral() != nil {
		switch v := c.OC_NumberLiteral().Accept(cl).(type) {
		case int64:
			q.Integer = new(int64)
			*q.Integer = v
		case float64:
			q.Double = new(float64)
			*q.Double = v
		}
	} else if c.OC_BooleanLiteral() != nil {
		q.Boolean = new(bool)
		*q.Boolean = c.OC_BooleanLiteral().Accept(cl).(bool)
	}
	return q
}

func (cl *BaseCypherVisitor) VisitOC_NumberLiteral(c *parser.OC_NumberLiteralContext) interface{} {
	if c.OC_IntegerLiteral() != nil {
		return c.OC_IntegerLiteral().Accept(cl)
	} else if c.OC_DoubleLiteral() != nil {
		return c.OC_DoubleLiteral().Accept(cl)
	}
	cl.AppendError(fmt.Errorf("Unable to detect number literal"))
	return nil
}

func (cl *BaseCypherVisitor) VisitOC_IntegerLiteral(c *parser.OC_IntegerLiteralContext) interface{} {
	x, err := strconv.ParseInt(c.GetText(), 10, 64)
	if err != nil {
		cl.AppendError(err)
		return nil
	}
	return x
}

func (cl *BaseCypherVisitor) VisitOC_DoubleLiteral(c *parser.OC_DoubleLiteralContext) interface{} {
	x, err := strconv.ParseFloat(c.GetText(), 64)
	if err != nil {
		cl.AppendError(err)
		return nil
	}
	return x
}

func (cl *BaseCypherVisitor) VisitOC_BooleanLiteral(c *parser.OC_BooleanLiteralContext) interface{} {
	return c.TRUE() != nil
}

func (cl *BaseCypherVisitor) VisitOC_Pattern(c *parser.OC_PatternContext) interface{} {
	items := make([]QueryPatternElement, 0)
	for i := range c.AllOC_PatternPart() {
		items = append(items, c.OC_PatternPart(i).Accept(cl).(QueryPatternElement))
	}
	return items
}

func (cl *BaseCypherVisitor) VisitOC_PatternPart(c *parser.OC_PatternPartContext) interface{} {
	return c.OC_AnonymousPatternPart().Accept(cl)
}

func (cl *BaseCypherVisitor) VisitOC_AnonymousPatternPart(c *parser.OC_AnonymousPatternPartContext) interface{} {
	return c.OC_PatternElement().Accept(cl)
}

type QueryPatternElement struct {
	QueryNodePattern
	QueryPatternElementChains []QueryPatternElementChain
}

func (cl *BaseCypherVisitor) VisitOC_PatternElement(c *parser.OC_PatternElementContext) interface{} {
	q := QueryPatternElement{}
	q.QueryPatternElementChains = make([]QueryPatternElementChain, 0)
	q.QueryNodePattern = c.OC_NodePattern().Accept(cl).(QueryNodePattern)

	for i := range c.AllOC_PatternElementChain() {
		q.QueryPatternElementChains = append(q.QueryPatternElementChains,
			c.OC_PatternElementChain(i).Accept(cl).(QueryPatternElementChain))
	}
	return q
}

type QueryPatternElementChain struct {
	RelationshipPattern QueryRelationshipPattern
	NodePattern         QueryNodePattern
}

func (cl *BaseCypherVisitor) VisitOC_PatternElementChain(c *parser.OC_PatternElementChainContext) interface{} {
	q := QueryPatternElementChain{}
	q.RelationshipPattern = c.OC_RelationshipPattern().Accept(cl).(QueryRelationshipPattern)
	q.NodePattern = c.OC_NodePattern().Accept(cl).(QueryNodePattern)
	return q
}

type QueryRelationshipPattern struct {
	RelationshipDetail *QueryRelationshipDetail
	LeftArrow          bool
	RightArrow         bool
}

func (cl *BaseCypherVisitor) VisitOC_RelationshipPattern(c *parser.OC_RelationshipPatternContext) interface{} {
	rp := QueryRelationshipPattern{}
	if c.OC_RelationshipDetail() != nil {
		rp.RelationshipDetail = new(QueryRelationshipDetail)
		*rp.RelationshipDetail = c.OC_RelationshipDetail().Accept(cl).(QueryRelationshipDetail)
	}
	rp.LeftArrow = c.OC_LeftArrowHead() != nil
	rp.RightArrow = c.OC_RightArrowHead() != nil
	return rp
}

// QueryRelationshipDetail object representing a relation [var:label]
type QueryRelationshipDetail struct {
	Variable string
	Labels   []string
}

func (cl *BaseCypherVisitor) VisitOC_RelationshipDetail(c *parser.OC_RelationshipDetailContext) interface{} {
	rs := QueryRelationshipDetail{}
	if c.OC_Variable() != nil {
		rs.Variable = c.OC_Variable().GetText()
	}
	if c.OC_RelationshipTypes() != nil {
		rs.Labels = c.OC_RelationshipTypes().Accept(cl).([]string)
	}
	return rs
}

func (cl *BaseCypherVisitor) VisitOC_RelationshipTypes(c *parser.OC_RelationshipTypesContext) interface{} {
	items := make([]string, 0)
	for i := range c.AllOC_RelTypeName() {
		items = append(items, c.OC_RelTypeName(i).GetText())
	}
	return items
}

type QueryNodePattern struct {
	Variable string
	Labels   []string
}

func (cl *BaseCypherVisitor) VisitOC_NodePattern(c *parser.OC_NodePatternContext) interface{} {
	q := QueryNodePattern{}
	if c.OC_NodeLabels() != nil {
		q.Labels = c.OC_NodeLabels().Accept(cl).([]string)
	}

	if c.OC_Variable() != nil {
		q.Variable = c.OC_Variable().Accept(cl).(string)
	}
	return q
}

func (cl *BaseCypherVisitor) VisitOC_NodeLabels(c *parser.OC_NodeLabelsContext) interface{} {
	labels := make([]string, 0)
	for i := range c.AllOC_NodeLabel() {
		labels = append(labels, c.OC_NodeLabel(i).Accept(cl).(string))
	}
	return labels
}

func (cl *BaseCypherVisitor) VisitOC_NodeLabel(c *parser.OC_NodeLabelContext) interface{} {
	return c.OC_LabelName().Accept(cl)
}

func (cl *BaseCypherVisitor) VisitOC_LabelName(c *parser.OC_LabelNameContext) interface{} {
	return c.GetText()
}

func (cl *BaseCypherVisitor) VisitOC_Variable(c *parser.OC_VariableContext) interface{} {
	return c.GetText()
}
