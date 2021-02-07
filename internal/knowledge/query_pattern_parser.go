package knowledge

import "github.com/clems4ever/go-graphkb/internal/query"

// PatternParser is a parser of patterns
type PatternParser struct {
	queryGraph *QueryGraph
}

// MatchScope default match scope. All patterns in the MATCH clause should have this scope.
var MatchScope = Scope{Context: MatchContext, ID: 0}

// NewPatternParser create an instance of pattern parser
func NewPatternParser(queryGraph *QueryGraph) *PatternParser {
	return &PatternParser{
		queryGraph: queryGraph,
	}
}

// ParseRelationshipsPattern parse a relationships pattern
func (ep *PatternParser) ParseRelationshipsPattern(q *query.QueryRelationshipsPattern, scope Scope) error {
	_, i1, err := ep.queryGraph.PushNode(q.QueryNodePattern, scope)
	if err != nil {
		return err
	}

	for _, z := range q.QueryPatternElementChains {
		_, i2, err := ep.queryGraph.PushNode(z.NodePattern, scope)
		if err != nil {
			return err
		}

		_, _, err = ep.queryGraph.PushRelation(z.RelationshipPattern, i1, i2, scope)
		if err != nil {
			return err
		}
		i1 = i2
	}
	return nil
}

// ParsePatternElement parse a pattern element
func (ep *PatternParser) ParsePatternElement(q *query.QueryPatternElement, scope Scope) error {
	_, i1, err := ep.queryGraph.PushNode(q.QueryNodePattern, scope)
	if err != nil {
		return err
	}

	for _, z := range q.QueryPatternElementChains {
		_, i2, err := ep.queryGraph.PushNode(z.NodePattern, scope)
		if err != nil {
			return err
		}

		_, _, err = ep.queryGraph.PushRelation(z.RelationshipPattern, i1, i2, scope)
		if err != nil {
			return err
		}
		i1 = i2
	}
	return nil
}
