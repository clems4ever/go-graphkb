// Code generated from Cypher.g4 by ANTLR 4.7.2. DO NOT EDIT.

package parser // Cypher

import "github.com/antlr/antlr4/runtime/Go/antlr"

// CypherListener is a complete listener for a parse tree produced by CypherParser.
type CypherListener interface {
	antlr.ParseTreeListener

	// EnterOC_Cypher is called when entering the oC_Cypher production.
	EnterOC_Cypher(c *OC_CypherContext)

	// EnterOC_Statement is called when entering the oC_Statement production.
	EnterOC_Statement(c *OC_StatementContext)

	// EnterOC_Query is called when entering the oC_Query production.
	EnterOC_Query(c *OC_QueryContext)

	// EnterOC_RegularQuery is called when entering the oC_RegularQuery production.
	EnterOC_RegularQuery(c *OC_RegularQueryContext)

	// EnterOC_Union is called when entering the oC_Union production.
	EnterOC_Union(c *OC_UnionContext)

	// EnterOC_SingleQuery is called when entering the oC_SingleQuery production.
	EnterOC_SingleQuery(c *OC_SingleQueryContext)

	// EnterOC_SinglePartQuery is called when entering the oC_SinglePartQuery production.
	EnterOC_SinglePartQuery(c *OC_SinglePartQueryContext)

	// EnterOC_MultiPartQuery is called when entering the oC_MultiPartQuery production.
	EnterOC_MultiPartQuery(c *OC_MultiPartQueryContext)

	// EnterOC_UpdatingClause is called when entering the oC_UpdatingClause production.
	EnterOC_UpdatingClause(c *OC_UpdatingClauseContext)

	// EnterOC_ReadingClause is called when entering the oC_ReadingClause production.
	EnterOC_ReadingClause(c *OC_ReadingClauseContext)

	// EnterOC_Match is called when entering the oC_Match production.
	EnterOC_Match(c *OC_MatchContext)

	// EnterOC_Unwind is called when entering the oC_Unwind production.
	EnterOC_Unwind(c *OC_UnwindContext)

	// EnterOC_Merge is called when entering the oC_Merge production.
	EnterOC_Merge(c *OC_MergeContext)

	// EnterOC_MergeAction is called when entering the oC_MergeAction production.
	EnterOC_MergeAction(c *OC_MergeActionContext)

	// EnterOC_Create is called when entering the oC_Create production.
	EnterOC_Create(c *OC_CreateContext)

	// EnterOC_Set is called when entering the oC_Set production.
	EnterOC_Set(c *OC_SetContext)

	// EnterOC_SetItem is called when entering the oC_SetItem production.
	EnterOC_SetItem(c *OC_SetItemContext)

	// EnterOC_Delete is called when entering the oC_Delete production.
	EnterOC_Delete(c *OC_DeleteContext)

	// EnterOC_Remove is called when entering the oC_Remove production.
	EnterOC_Remove(c *OC_RemoveContext)

	// EnterOC_RemoveItem is called when entering the oC_RemoveItem production.
	EnterOC_RemoveItem(c *OC_RemoveItemContext)

	// EnterOC_InQueryCall is called when entering the oC_InQueryCall production.
	EnterOC_InQueryCall(c *OC_InQueryCallContext)

	// EnterOC_StandaloneCall is called when entering the oC_StandaloneCall production.
	EnterOC_StandaloneCall(c *OC_StandaloneCallContext)

	// EnterOC_YieldItems is called when entering the oC_YieldItems production.
	EnterOC_YieldItems(c *OC_YieldItemsContext)

	// EnterOC_YieldItem is called when entering the oC_YieldItem production.
	EnterOC_YieldItem(c *OC_YieldItemContext)

	// EnterOC_With is called when entering the oC_With production.
	EnterOC_With(c *OC_WithContext)

	// EnterOC_Return is called when entering the oC_Return production.
	EnterOC_Return(c *OC_ReturnContext)

	// EnterOC_ProjectionBody is called when entering the oC_ProjectionBody production.
	EnterOC_ProjectionBody(c *OC_ProjectionBodyContext)

	// EnterOC_ProjectionItems is called when entering the oC_ProjectionItems production.
	EnterOC_ProjectionItems(c *OC_ProjectionItemsContext)

	// EnterOC_ProjectionItem is called when entering the oC_ProjectionItem production.
	EnterOC_ProjectionItem(c *OC_ProjectionItemContext)

	// EnterOC_Order is called when entering the oC_Order production.
	EnterOC_Order(c *OC_OrderContext)

	// EnterOC_Skip is called when entering the oC_Skip production.
	EnterOC_Skip(c *OC_SkipContext)

	// EnterOC_Limit is called when entering the oC_Limit production.
	EnterOC_Limit(c *OC_LimitContext)

	// EnterOC_SortItem is called when entering the oC_SortItem production.
	EnterOC_SortItem(c *OC_SortItemContext)

	// EnterOC_Where is called when entering the oC_Where production.
	EnterOC_Where(c *OC_WhereContext)

	// EnterOC_Pattern is called when entering the oC_Pattern production.
	EnterOC_Pattern(c *OC_PatternContext)

	// EnterOC_PatternPart is called when entering the oC_PatternPart production.
	EnterOC_PatternPart(c *OC_PatternPartContext)

	// EnterOC_AnonymousPatternPart is called when entering the oC_AnonymousPatternPart production.
	EnterOC_AnonymousPatternPart(c *OC_AnonymousPatternPartContext)

	// EnterOC_PatternElement is called when entering the oC_PatternElement production.
	EnterOC_PatternElement(c *OC_PatternElementContext)

	// EnterOC_NodePattern is called when entering the oC_NodePattern production.
	EnterOC_NodePattern(c *OC_NodePatternContext)

	// EnterOC_PatternElementChain is called when entering the oC_PatternElementChain production.
	EnterOC_PatternElementChain(c *OC_PatternElementChainContext)

	// EnterOC_RelationshipPattern is called when entering the oC_RelationshipPattern production.
	EnterOC_RelationshipPattern(c *OC_RelationshipPatternContext)

	// EnterOC_RelationshipDetail is called when entering the oC_RelationshipDetail production.
	EnterOC_RelationshipDetail(c *OC_RelationshipDetailContext)

	// EnterOC_Properties is called when entering the oC_Properties production.
	EnterOC_Properties(c *OC_PropertiesContext)

	// EnterOC_RelationshipTypes is called when entering the oC_RelationshipTypes production.
	EnterOC_RelationshipTypes(c *OC_RelationshipTypesContext)

	// EnterOC_NodeLabels is called when entering the oC_NodeLabels production.
	EnterOC_NodeLabels(c *OC_NodeLabelsContext)

	// EnterOC_NodeLabel is called when entering the oC_NodeLabel production.
	EnterOC_NodeLabel(c *OC_NodeLabelContext)

	// EnterOC_RangeLiteral is called when entering the oC_RangeLiteral production.
	EnterOC_RangeLiteral(c *OC_RangeLiteralContext)

	// EnterOC_LabelName is called when entering the oC_LabelName production.
	EnterOC_LabelName(c *OC_LabelNameContext)

	// EnterOC_RelTypeName is called when entering the oC_RelTypeName production.
	EnterOC_RelTypeName(c *OC_RelTypeNameContext)

	// EnterOC_Expression is called when entering the oC_Expression production.
	EnterOC_Expression(c *OC_ExpressionContext)

	// EnterOC_OrExpression is called when entering the oC_OrExpression production.
	EnterOC_OrExpression(c *OC_OrExpressionContext)

	// EnterOC_XorExpression is called when entering the oC_XorExpression production.
	EnterOC_XorExpression(c *OC_XorExpressionContext)

	// EnterOC_AndExpression is called when entering the oC_AndExpression production.
	EnterOC_AndExpression(c *OC_AndExpressionContext)

	// EnterOC_NotExpression is called when entering the oC_NotExpression production.
	EnterOC_NotExpression(c *OC_NotExpressionContext)

	// EnterOC_ComparisonExpression is called when entering the oC_ComparisonExpression production.
	EnterOC_ComparisonExpression(c *OC_ComparisonExpressionContext)

	// EnterOC_AddOrSubtractExpression is called when entering the oC_AddOrSubtractExpression production.
	EnterOC_AddOrSubtractExpression(c *OC_AddOrSubtractExpressionContext)

	// EnterOC_MultiplyDivideModuloExpression is called when entering the oC_MultiplyDivideModuloExpression production.
	EnterOC_MultiplyDivideModuloExpression(c *OC_MultiplyDivideModuloExpressionContext)

	// EnterOC_PowerOfExpression is called when entering the oC_PowerOfExpression production.
	EnterOC_PowerOfExpression(c *OC_PowerOfExpressionContext)

	// EnterOC_UnaryAddOrSubtractExpression is called when entering the oC_UnaryAddOrSubtractExpression production.
	EnterOC_UnaryAddOrSubtractExpression(c *OC_UnaryAddOrSubtractExpressionContext)

	// EnterOC_StringListNullOperatorExpression is called when entering the oC_StringListNullOperatorExpression production.
	EnterOC_StringListNullOperatorExpression(c *OC_StringListNullOperatorExpressionContext)

	// EnterOC_ListOperatorExpression is called when entering the oC_ListOperatorExpression production.
	EnterOC_ListOperatorExpression(c *OC_ListOperatorExpressionContext)

	// EnterOC_StringOperatorExpression is called when entering the oC_StringOperatorExpression production.
	EnterOC_StringOperatorExpression(c *OC_StringOperatorExpressionContext)

	// EnterOC_NullOperatorExpression is called when entering the oC_NullOperatorExpression production.
	EnterOC_NullOperatorExpression(c *OC_NullOperatorExpressionContext)

	// EnterOC_PropertyOrLabelsExpression is called when entering the oC_PropertyOrLabelsExpression production.
	EnterOC_PropertyOrLabelsExpression(c *OC_PropertyOrLabelsExpressionContext)

	// EnterOC_Atom is called when entering the oC_Atom production.
	EnterOC_Atom(c *OC_AtomContext)

	// EnterOC_Literal is called when entering the oC_Literal production.
	EnterOC_Literal(c *OC_LiteralContext)

	// EnterOC_BooleanLiteral is called when entering the oC_BooleanLiteral production.
	EnterOC_BooleanLiteral(c *OC_BooleanLiteralContext)

	// EnterOC_ListLiteral is called when entering the oC_ListLiteral production.
	EnterOC_ListLiteral(c *OC_ListLiteralContext)

	// EnterOC_PartialComparisonExpression is called when entering the oC_PartialComparisonExpression production.
	EnterOC_PartialComparisonExpression(c *OC_PartialComparisonExpressionContext)

	// EnterOC_ParenthesizedExpression is called when entering the oC_ParenthesizedExpression production.
	EnterOC_ParenthesizedExpression(c *OC_ParenthesizedExpressionContext)

	// EnterOC_RelationshipsPattern is called when entering the oC_RelationshipsPattern production.
	EnterOC_RelationshipsPattern(c *OC_RelationshipsPatternContext)

	// EnterOC_FilterExpression is called when entering the oC_FilterExpression production.
	EnterOC_FilterExpression(c *OC_FilterExpressionContext)

	// EnterOC_IdInColl is called when entering the oC_IdInColl production.
	EnterOC_IdInColl(c *OC_IdInCollContext)

	// EnterOC_FunctionInvocation is called when entering the oC_FunctionInvocation production.
	EnterOC_FunctionInvocation(c *OC_FunctionInvocationContext)

	// EnterOC_FunctionName is called when entering the oC_FunctionName production.
	EnterOC_FunctionName(c *OC_FunctionNameContext)

	// EnterOC_ExplicitProcedureInvocation is called when entering the oC_ExplicitProcedureInvocation production.
	EnterOC_ExplicitProcedureInvocation(c *OC_ExplicitProcedureInvocationContext)

	// EnterOC_ImplicitProcedureInvocation is called when entering the oC_ImplicitProcedureInvocation production.
	EnterOC_ImplicitProcedureInvocation(c *OC_ImplicitProcedureInvocationContext)

	// EnterOC_ProcedureResultField is called when entering the oC_ProcedureResultField production.
	EnterOC_ProcedureResultField(c *OC_ProcedureResultFieldContext)

	// EnterOC_ProcedureName is called when entering the oC_ProcedureName production.
	EnterOC_ProcedureName(c *OC_ProcedureNameContext)

	// EnterOC_Namespace is called when entering the oC_Namespace production.
	EnterOC_Namespace(c *OC_NamespaceContext)

	// EnterOC_ListComprehension is called when entering the oC_ListComprehension production.
	EnterOC_ListComprehension(c *OC_ListComprehensionContext)

	// EnterOC_PatternComprehension is called when entering the oC_PatternComprehension production.
	EnterOC_PatternComprehension(c *OC_PatternComprehensionContext)

	// EnterOC_PropertyLookup is called when entering the oC_PropertyLookup production.
	EnterOC_PropertyLookup(c *OC_PropertyLookupContext)

	// EnterOC_CaseExpression is called when entering the oC_CaseExpression production.
	EnterOC_CaseExpression(c *OC_CaseExpressionContext)

	// EnterOC_CaseAlternatives is called when entering the oC_CaseAlternatives production.
	EnterOC_CaseAlternatives(c *OC_CaseAlternativesContext)

	// EnterOC_Variable is called when entering the oC_Variable production.
	EnterOC_Variable(c *OC_VariableContext)

	// EnterOC_NumberLiteral is called when entering the oC_NumberLiteral production.
	EnterOC_NumberLiteral(c *OC_NumberLiteralContext)

	// EnterOC_MapLiteral is called when entering the oC_MapLiteral production.
	EnterOC_MapLiteral(c *OC_MapLiteralContext)

	// EnterOC_Parameter is called when entering the oC_Parameter production.
	EnterOC_Parameter(c *OC_ParameterContext)

	// EnterOC_PropertyExpression is called when entering the oC_PropertyExpression production.
	EnterOC_PropertyExpression(c *OC_PropertyExpressionContext)

	// EnterOC_PropertyKeyName is called when entering the oC_PropertyKeyName production.
	EnterOC_PropertyKeyName(c *OC_PropertyKeyNameContext)

	// EnterOC_IntegerLiteral is called when entering the oC_IntegerLiteral production.
	EnterOC_IntegerLiteral(c *OC_IntegerLiteralContext)

	// EnterOC_DoubleLiteral is called when entering the oC_DoubleLiteral production.
	EnterOC_DoubleLiteral(c *OC_DoubleLiteralContext)

	// EnterOC_SchemaName is called when entering the oC_SchemaName production.
	EnterOC_SchemaName(c *OC_SchemaNameContext)

	// EnterOC_ReservedWord is called when entering the oC_ReservedWord production.
	EnterOC_ReservedWord(c *OC_ReservedWordContext)

	// EnterOC_SymbolicName is called when entering the oC_SymbolicName production.
	EnterOC_SymbolicName(c *OC_SymbolicNameContext)

	// EnterOC_LeftArrowHead is called when entering the oC_LeftArrowHead production.
	EnterOC_LeftArrowHead(c *OC_LeftArrowHeadContext)

	// EnterOC_RightArrowHead is called when entering the oC_RightArrowHead production.
	EnterOC_RightArrowHead(c *OC_RightArrowHeadContext)

	// EnterOC_Dash is called when entering the oC_Dash production.
	EnterOC_Dash(c *OC_DashContext)

	// ExitOC_Cypher is called when exiting the oC_Cypher production.
	ExitOC_Cypher(c *OC_CypherContext)

	// ExitOC_Statement is called when exiting the oC_Statement production.
	ExitOC_Statement(c *OC_StatementContext)

	// ExitOC_Query is called when exiting the oC_Query production.
	ExitOC_Query(c *OC_QueryContext)

	// ExitOC_RegularQuery is called when exiting the oC_RegularQuery production.
	ExitOC_RegularQuery(c *OC_RegularQueryContext)

	// ExitOC_Union is called when exiting the oC_Union production.
	ExitOC_Union(c *OC_UnionContext)

	// ExitOC_SingleQuery is called when exiting the oC_SingleQuery production.
	ExitOC_SingleQuery(c *OC_SingleQueryContext)

	// ExitOC_SinglePartQuery is called when exiting the oC_SinglePartQuery production.
	ExitOC_SinglePartQuery(c *OC_SinglePartQueryContext)

	// ExitOC_MultiPartQuery is called when exiting the oC_MultiPartQuery production.
	ExitOC_MultiPartQuery(c *OC_MultiPartQueryContext)

	// ExitOC_UpdatingClause is called when exiting the oC_UpdatingClause production.
	ExitOC_UpdatingClause(c *OC_UpdatingClauseContext)

	// ExitOC_ReadingClause is called when exiting the oC_ReadingClause production.
	ExitOC_ReadingClause(c *OC_ReadingClauseContext)

	// ExitOC_Match is called when exiting the oC_Match production.
	ExitOC_Match(c *OC_MatchContext)

	// ExitOC_Unwind is called when exiting the oC_Unwind production.
	ExitOC_Unwind(c *OC_UnwindContext)

	// ExitOC_Merge is called when exiting the oC_Merge production.
	ExitOC_Merge(c *OC_MergeContext)

	// ExitOC_MergeAction is called when exiting the oC_MergeAction production.
	ExitOC_MergeAction(c *OC_MergeActionContext)

	// ExitOC_Create is called when exiting the oC_Create production.
	ExitOC_Create(c *OC_CreateContext)

	// ExitOC_Set is called when exiting the oC_Set production.
	ExitOC_Set(c *OC_SetContext)

	// ExitOC_SetItem is called when exiting the oC_SetItem production.
	ExitOC_SetItem(c *OC_SetItemContext)

	// ExitOC_Delete is called when exiting the oC_Delete production.
	ExitOC_Delete(c *OC_DeleteContext)

	// ExitOC_Remove is called when exiting the oC_Remove production.
	ExitOC_Remove(c *OC_RemoveContext)

	// ExitOC_RemoveItem is called when exiting the oC_RemoveItem production.
	ExitOC_RemoveItem(c *OC_RemoveItemContext)

	// ExitOC_InQueryCall is called when exiting the oC_InQueryCall production.
	ExitOC_InQueryCall(c *OC_InQueryCallContext)

	// ExitOC_StandaloneCall is called when exiting the oC_StandaloneCall production.
	ExitOC_StandaloneCall(c *OC_StandaloneCallContext)

	// ExitOC_YieldItems is called when exiting the oC_YieldItems production.
	ExitOC_YieldItems(c *OC_YieldItemsContext)

	// ExitOC_YieldItem is called when exiting the oC_YieldItem production.
	ExitOC_YieldItem(c *OC_YieldItemContext)

	// ExitOC_With is called when exiting the oC_With production.
	ExitOC_With(c *OC_WithContext)

	// ExitOC_Return is called when exiting the oC_Return production.
	ExitOC_Return(c *OC_ReturnContext)

	// ExitOC_ProjectionBody is called when exiting the oC_ProjectionBody production.
	ExitOC_ProjectionBody(c *OC_ProjectionBodyContext)

	// ExitOC_ProjectionItems is called when exiting the oC_ProjectionItems production.
	ExitOC_ProjectionItems(c *OC_ProjectionItemsContext)

	// ExitOC_ProjectionItem is called when exiting the oC_ProjectionItem production.
	ExitOC_ProjectionItem(c *OC_ProjectionItemContext)

	// ExitOC_Order is called when exiting the oC_Order production.
	ExitOC_Order(c *OC_OrderContext)

	// ExitOC_Skip is called when exiting the oC_Skip production.
	ExitOC_Skip(c *OC_SkipContext)

	// ExitOC_Limit is called when exiting the oC_Limit production.
	ExitOC_Limit(c *OC_LimitContext)

	// ExitOC_SortItem is called when exiting the oC_SortItem production.
	ExitOC_SortItem(c *OC_SortItemContext)

	// ExitOC_Where is called when exiting the oC_Where production.
	ExitOC_Where(c *OC_WhereContext)

	// ExitOC_Pattern is called when exiting the oC_Pattern production.
	ExitOC_Pattern(c *OC_PatternContext)

	// ExitOC_PatternPart is called when exiting the oC_PatternPart production.
	ExitOC_PatternPart(c *OC_PatternPartContext)

	// ExitOC_AnonymousPatternPart is called when exiting the oC_AnonymousPatternPart production.
	ExitOC_AnonymousPatternPart(c *OC_AnonymousPatternPartContext)

	// ExitOC_PatternElement is called when exiting the oC_PatternElement production.
	ExitOC_PatternElement(c *OC_PatternElementContext)

	// ExitOC_NodePattern is called when exiting the oC_NodePattern production.
	ExitOC_NodePattern(c *OC_NodePatternContext)

	// ExitOC_PatternElementChain is called when exiting the oC_PatternElementChain production.
	ExitOC_PatternElementChain(c *OC_PatternElementChainContext)

	// ExitOC_RelationshipPattern is called when exiting the oC_RelationshipPattern production.
	ExitOC_RelationshipPattern(c *OC_RelationshipPatternContext)

	// ExitOC_RelationshipDetail is called when exiting the oC_RelationshipDetail production.
	ExitOC_RelationshipDetail(c *OC_RelationshipDetailContext)

	// ExitOC_Properties is called when exiting the oC_Properties production.
	ExitOC_Properties(c *OC_PropertiesContext)

	// ExitOC_RelationshipTypes is called when exiting the oC_RelationshipTypes production.
	ExitOC_RelationshipTypes(c *OC_RelationshipTypesContext)

	// ExitOC_NodeLabels is called when exiting the oC_NodeLabels production.
	ExitOC_NodeLabels(c *OC_NodeLabelsContext)

	// ExitOC_NodeLabel is called when exiting the oC_NodeLabel production.
	ExitOC_NodeLabel(c *OC_NodeLabelContext)

	// ExitOC_RangeLiteral is called when exiting the oC_RangeLiteral production.
	ExitOC_RangeLiteral(c *OC_RangeLiteralContext)

	// ExitOC_LabelName is called when exiting the oC_LabelName production.
	ExitOC_LabelName(c *OC_LabelNameContext)

	// ExitOC_RelTypeName is called when exiting the oC_RelTypeName production.
	ExitOC_RelTypeName(c *OC_RelTypeNameContext)

	// ExitOC_Expression is called when exiting the oC_Expression production.
	ExitOC_Expression(c *OC_ExpressionContext)

	// ExitOC_OrExpression is called when exiting the oC_OrExpression production.
	ExitOC_OrExpression(c *OC_OrExpressionContext)

	// ExitOC_XorExpression is called when exiting the oC_XorExpression production.
	ExitOC_XorExpression(c *OC_XorExpressionContext)

	// ExitOC_AndExpression is called when exiting the oC_AndExpression production.
	ExitOC_AndExpression(c *OC_AndExpressionContext)

	// ExitOC_NotExpression is called when exiting the oC_NotExpression production.
	ExitOC_NotExpression(c *OC_NotExpressionContext)

	// ExitOC_ComparisonExpression is called when exiting the oC_ComparisonExpression production.
	ExitOC_ComparisonExpression(c *OC_ComparisonExpressionContext)

	// ExitOC_AddOrSubtractExpression is called when exiting the oC_AddOrSubtractExpression production.
	ExitOC_AddOrSubtractExpression(c *OC_AddOrSubtractExpressionContext)

	// ExitOC_MultiplyDivideModuloExpression is called when exiting the oC_MultiplyDivideModuloExpression production.
	ExitOC_MultiplyDivideModuloExpression(c *OC_MultiplyDivideModuloExpressionContext)

	// ExitOC_PowerOfExpression is called when exiting the oC_PowerOfExpression production.
	ExitOC_PowerOfExpression(c *OC_PowerOfExpressionContext)

	// ExitOC_UnaryAddOrSubtractExpression is called when exiting the oC_UnaryAddOrSubtractExpression production.
	ExitOC_UnaryAddOrSubtractExpression(c *OC_UnaryAddOrSubtractExpressionContext)

	// ExitOC_StringListNullOperatorExpression is called when exiting the oC_StringListNullOperatorExpression production.
	ExitOC_StringListNullOperatorExpression(c *OC_StringListNullOperatorExpressionContext)

	// ExitOC_ListOperatorExpression is called when exiting the oC_ListOperatorExpression production.
	ExitOC_ListOperatorExpression(c *OC_ListOperatorExpressionContext)

	// ExitOC_StringOperatorExpression is called when exiting the oC_StringOperatorExpression production.
	ExitOC_StringOperatorExpression(c *OC_StringOperatorExpressionContext)

	// ExitOC_NullOperatorExpression is called when exiting the oC_NullOperatorExpression production.
	ExitOC_NullOperatorExpression(c *OC_NullOperatorExpressionContext)

	// ExitOC_PropertyOrLabelsExpression is called when exiting the oC_PropertyOrLabelsExpression production.
	ExitOC_PropertyOrLabelsExpression(c *OC_PropertyOrLabelsExpressionContext)

	// ExitOC_Atom is called when exiting the oC_Atom production.
	ExitOC_Atom(c *OC_AtomContext)

	// ExitOC_Literal is called when exiting the oC_Literal production.
	ExitOC_Literal(c *OC_LiteralContext)

	// ExitOC_BooleanLiteral is called when exiting the oC_BooleanLiteral production.
	ExitOC_BooleanLiteral(c *OC_BooleanLiteralContext)

	// ExitOC_ListLiteral is called when exiting the oC_ListLiteral production.
	ExitOC_ListLiteral(c *OC_ListLiteralContext)

	// ExitOC_PartialComparisonExpression is called when exiting the oC_PartialComparisonExpression production.
	ExitOC_PartialComparisonExpression(c *OC_PartialComparisonExpressionContext)

	// ExitOC_ParenthesizedExpression is called when exiting the oC_ParenthesizedExpression production.
	ExitOC_ParenthesizedExpression(c *OC_ParenthesizedExpressionContext)

	// ExitOC_RelationshipsPattern is called when exiting the oC_RelationshipsPattern production.
	ExitOC_RelationshipsPattern(c *OC_RelationshipsPatternContext)

	// ExitOC_FilterExpression is called when exiting the oC_FilterExpression production.
	ExitOC_FilterExpression(c *OC_FilterExpressionContext)

	// ExitOC_IdInColl is called when exiting the oC_IdInColl production.
	ExitOC_IdInColl(c *OC_IdInCollContext)

	// ExitOC_FunctionInvocation is called when exiting the oC_FunctionInvocation production.
	ExitOC_FunctionInvocation(c *OC_FunctionInvocationContext)

	// ExitOC_FunctionName is called when exiting the oC_FunctionName production.
	ExitOC_FunctionName(c *OC_FunctionNameContext)

	// ExitOC_ExplicitProcedureInvocation is called when exiting the oC_ExplicitProcedureInvocation production.
	ExitOC_ExplicitProcedureInvocation(c *OC_ExplicitProcedureInvocationContext)

	// ExitOC_ImplicitProcedureInvocation is called when exiting the oC_ImplicitProcedureInvocation production.
	ExitOC_ImplicitProcedureInvocation(c *OC_ImplicitProcedureInvocationContext)

	// ExitOC_ProcedureResultField is called when exiting the oC_ProcedureResultField production.
	ExitOC_ProcedureResultField(c *OC_ProcedureResultFieldContext)

	// ExitOC_ProcedureName is called when exiting the oC_ProcedureName production.
	ExitOC_ProcedureName(c *OC_ProcedureNameContext)

	// ExitOC_Namespace is called when exiting the oC_Namespace production.
	ExitOC_Namespace(c *OC_NamespaceContext)

	// ExitOC_ListComprehension is called when exiting the oC_ListComprehension production.
	ExitOC_ListComprehension(c *OC_ListComprehensionContext)

	// ExitOC_PatternComprehension is called when exiting the oC_PatternComprehension production.
	ExitOC_PatternComprehension(c *OC_PatternComprehensionContext)

	// ExitOC_PropertyLookup is called when exiting the oC_PropertyLookup production.
	ExitOC_PropertyLookup(c *OC_PropertyLookupContext)

	// ExitOC_CaseExpression is called when exiting the oC_CaseExpression production.
	ExitOC_CaseExpression(c *OC_CaseExpressionContext)

	// ExitOC_CaseAlternatives is called when exiting the oC_CaseAlternatives production.
	ExitOC_CaseAlternatives(c *OC_CaseAlternativesContext)

	// ExitOC_Variable is called when exiting the oC_Variable production.
	ExitOC_Variable(c *OC_VariableContext)

	// ExitOC_NumberLiteral is called when exiting the oC_NumberLiteral production.
	ExitOC_NumberLiteral(c *OC_NumberLiteralContext)

	// ExitOC_MapLiteral is called when exiting the oC_MapLiteral production.
	ExitOC_MapLiteral(c *OC_MapLiteralContext)

	// ExitOC_Parameter is called when exiting the oC_Parameter production.
	ExitOC_Parameter(c *OC_ParameterContext)

	// ExitOC_PropertyExpression is called when exiting the oC_PropertyExpression production.
	ExitOC_PropertyExpression(c *OC_PropertyExpressionContext)

	// ExitOC_PropertyKeyName is called when exiting the oC_PropertyKeyName production.
	ExitOC_PropertyKeyName(c *OC_PropertyKeyNameContext)

	// ExitOC_IntegerLiteral is called when exiting the oC_IntegerLiteral production.
	ExitOC_IntegerLiteral(c *OC_IntegerLiteralContext)

	// ExitOC_DoubleLiteral is called when exiting the oC_DoubleLiteral production.
	ExitOC_DoubleLiteral(c *OC_DoubleLiteralContext)

	// ExitOC_SchemaName is called when exiting the oC_SchemaName production.
	ExitOC_SchemaName(c *OC_SchemaNameContext)

	// ExitOC_ReservedWord is called when exiting the oC_ReservedWord production.
	ExitOC_ReservedWord(c *OC_ReservedWordContext)

	// ExitOC_SymbolicName is called when exiting the oC_SymbolicName production.
	ExitOC_SymbolicName(c *OC_SymbolicNameContext)

	// ExitOC_LeftArrowHead is called when exiting the oC_LeftArrowHead production.
	ExitOC_LeftArrowHead(c *OC_LeftArrowHeadContext)

	// ExitOC_RightArrowHead is called when exiting the oC_RightArrowHead production.
	ExitOC_RightArrowHead(c *OC_RightArrowHeadContext)

	// ExitOC_Dash is called when exiting the oC_Dash production.
	ExitOC_Dash(c *OC_DashContext)
}
