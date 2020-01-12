package knowledge

import (
	"testing"

	"github.com/clems4ever/go-graphkb/internal/query"
	"github.com/stretchr/testify/suite"
)

type QueryGraphSuite struct {
	suite.Suite
}

func (s *QueryGraphSuite) TestShouldPushUnamedUntypedNode() {
	g := NewQueryGraph()
	n0, idx0, err := g.PushNode(query.QueryNodePattern{
		Variable: "",
		Labels:   nil,
	})
	s.Require().NoError(err)

	n1, idx1, err := g.PushNode(query.QueryNodePattern{
		Variable: "",
		Labels:   nil,
	})
	s.Require().NoError(err)

	qn := QueryNode{}

	s.Assert().Equal(&qn, n0)
	s.Assert().Equal(0, idx0)

	s.Assert().Equal(&qn, n1)
	s.Assert().Equal(1, idx1)
}

func (s *QueryGraphSuite) TestShouldPushNamedUntypedNode() {
	g := NewQueryGraph()
	n0, idx0, err := g.PushNode(query.QueryNodePattern{
		Variable: "var1",
		Labels:   nil,
	})
	s.Require().NoError(err)

	n1, idx1, err := g.PushNode(query.QueryNodePattern{
		Variable: "var2",
		Labels:   nil,
	})
	s.Require().NoError(err)

	n2, idx2, err := g.PushNode(query.QueryNodePattern{
		Variable: "var1",
		Labels:   nil,
	})
	s.Require().NoError(err)

	qn := QueryNode{}

	s.Assert().Equal(&qn, n0)
	s.Assert().Equal(0, idx0)

	s.Assert().Equal(&qn, n1)
	s.Assert().Equal(1, idx1)

	s.Assert().Equal(n0, n2)
	s.Assert().Equal(0, idx2)
}

func (s *QueryGraphSuite) TestShouldPushUnamedTypedNode() {
	g := NewQueryGraph()
	n0, idx0, err := g.PushNode(query.QueryNodePattern{
		Variable: "",
		Labels:   []string{"t1", "t2"},
	})
	s.Require().NoError(err)

	n1, idx1, err := g.PushNode(query.QueryNodePattern{
		Variable: "",
		Labels:   []string{"t1", "t3"},
	})
	s.Require().NoError(err)

	q0 := QueryNode{Labels: []string{"t1", "t2"}}
	q1 := QueryNode{Labels: []string{"t1", "t3"}}

	s.Assert().Equal(n0, &q0)
	s.Assert().Equal(0, idx0)

	s.Assert().Equal(n1, &q1)
	s.Assert().Equal(1, idx1)
}

func (s *QueryGraphSuite) TestCannotPushTwiceSameNodeWithDifferentTypes() {
	g := NewQueryGraph()
	_, _, err := g.PushNode(query.QueryNodePattern{
		Variable: "var1",
		Labels:   []string{"t1", "t2"},
	})
	s.Require().NoError(err)

	_, _, err = g.PushNode(query.QueryNodePattern{
		Variable: "var1",
		Labels:   []string{"t1", "t3"},
	})
	s.Require().EqualError(err, "Variable 'var1' already defined with a different type")
}

func (s *QueryGraphSuite) TestShouldPushNamedTypedNode() {
	g := NewQueryGraph()
	n0, idx0, err := g.PushNode(query.QueryNodePattern{
		Variable: "var1",
		Labels:   []string{"t1", "t2"},
	})
	s.Require().NoError(err)

	n1, idx1, err := g.PushNode(query.QueryNodePattern{
		Variable: "var2",
		Labels:   []string{"t1", "t3"},
	})
	s.Require().NoError(err)

	q0 := QueryNode{Labels: []string{"t1", "t2"}}
	q1 := QueryNode{Labels: []string{"t1", "t3"}}

	s.Assert().Equal(n0, &q0)
	s.Assert().Equal(0, idx0)

	s.Assert().Equal(n1, &q1)
	s.Assert().Equal(1, idx1)
}

func CreateRelationship(varName string, labels []string) query.QueryRelationshipPattern {
	return query.QueryRelationshipPattern{
		RelationshipDetail: &query.QueryRelationshipDetail{
			Variable: varName,
			Labels:   labels,
		},
	}
}

func (s *QueryGraphSuite) TestShouldPushUnamedUntypedRelation() {
	g := NewQueryGraph()
	_, _, err := g.PushNode(query.QueryNodePattern{
		Variable: "",
		Labels:   []string{"t1", "t2"},
	})
	s.Require().NoError(err)

	pattern := CreateRelationship("", nil)

	n0, idx0, err := g.PushRelation(pattern, 0, 0)
	s.Require().NoError(err)

	n1, idx1, err := g.PushRelation(pattern, 0, 0)
	s.Require().NoError(err)

	qn := QueryRelation{Direction: Either}

	s.Assert().Equal(&qn, n0)
	s.Assert().Equal(0, idx0)

	s.Assert().Equal(&qn, n1)
	s.Assert().Equal(1, idx1)
}

func (s *QueryGraphSuite) TestShouldPushNamedUntypedRelation() {
	g := NewQueryGraph()
	_, _, err := g.PushNode(query.QueryNodePattern{
		Variable: "",
		Labels:   []string{"t1", "t2"},
	})
	s.Require().NoError(err)

	pattern0 := CreateRelationship("var1", nil)
	pattern1 := CreateRelationship("var2", nil)

	n0, idx0, err := g.PushRelation(pattern0, 0, 0)
	s.Require().NoError(err)

	n1, idx1, err := g.PushRelation(pattern1, 0, 0)
	s.Require().NoError(err)

	n2, idx2, err := g.PushRelation(pattern0, 0, 0)
	s.Require().NoError(err)

	qn := QueryRelation{Direction: Either}

	s.Assert().Equal(&qn, n0)
	s.Assert().Equal(0, idx0)

	s.Assert().Equal(&qn, n1)
	s.Assert().Equal(1, idx1)

	s.Assert().Equal(n0, n2)
	s.Assert().Equal(0, idx2)
}

func (s *QueryGraphSuite) TestShouldPushUnamedTypedRelation() {
	g := NewQueryGraph()
	_, _, err := g.PushNode(query.QueryNodePattern{
		Variable: "",
		Labels:   []string{"t1", "t2"},
	})
	s.Require().NoError(err)

	pattern0 := CreateRelationship("", []string{"t1", "t2"})
	pattern1 := CreateRelationship("", []string{"t1", "t3"})

	n0, idx0, err := g.PushRelation(pattern0, 0, 0)
	s.Require().NoError(err)

	n1, idx1, err := g.PushRelation(pattern1, 0, 0)
	s.Require().NoError(err)

	q0 := QueryRelation{Labels: []string{"t1", "t2"}, Direction: Either}
	q1 := QueryRelation{Labels: []string{"t1", "t3"}, Direction: Either}

	s.Assert().Equal(n0, &q0)
	s.Assert().Equal(0, idx0)

	s.Assert().Equal(n1, &q1)
	s.Assert().Equal(1, idx1)
}

func (s *QueryGraphSuite) TestCannotPushTwiceSameRelationWithDifferentTypes() {
	g := NewQueryGraph()
	_, _, err := g.PushNode(query.QueryNodePattern{
		Variable: "",
		Labels:   []string{"t1", "t2"},
	})
	s.Require().NoError(err)

	pattern0 := CreateRelationship("var1", []string{"t1", "t2"})
	pattern1 := CreateRelationship("var1", []string{"t1", "t3"})

	_, _, err = g.PushRelation(pattern0, 0, 0)
	s.Require().NoError(err)

	_, _, err = g.PushRelation(pattern1, 0, 0)
	s.Require().EqualError(err, "Variable 'var1' already defined with a different type")
}

func (s *QueryGraphSuite) TestShouldPushNamedTypedRelation() {
	g := NewQueryGraph()
	_, _, err := g.PushNode(query.QueryNodePattern{
		Variable: "",
		Labels:   []string{"t1", "t2"},
	})
	s.Require().NoError(err)

	pattern0 := CreateRelationship("var1", []string{"t1", "t2"})
	pattern1 := CreateRelationship("var2", []string{"t1", "t3"})

	n0, idx0, err := g.PushRelation(pattern0, 0, 0)
	s.Require().NoError(err)

	n1, idx1, err := g.PushRelation(pattern1, 0, 0)
	s.Require().NoError(err)

	q0 := QueryRelation{Labels: []string{"t1", "t2"}, Direction: Either}
	q1 := QueryRelation{Labels: []string{"t1", "t3"}, Direction: Either}

	s.Assert().Equal(&q0, n0)
	s.Assert().Equal(0, idx0)

	s.Assert().Equal(&q1, n1)
	s.Assert().Equal(1, idx1)
}

func (s *QueryGraphSuite) TestCannotPushNodeThenRelationWithSameName() {
	g := NewQueryGraph()
	_, _, err := g.PushNode(query.QueryNodePattern{
		Variable: "var1",
		Labels:   []string{"t1", "t2"},
	})
	s.Require().NoError(err)

	pattern0 := CreateRelationship("var1", []string{"t1", "t2"})
	_, _, err = g.PushRelation(pattern0, 0, 0)
	s.Require().EqualError(err, "Variable 'var1' is assigned to a different type")
}

func (s *QueryGraphSuite) TestCannotPushRelationThenNodeWithSameName() {
	g := NewQueryGraph()
	_, _, err := g.PushNode(query.QueryNodePattern{
		Variable: "var1",
		Labels:   []string{"t1", "t2"},
	})
	s.Require().NoError(err)

	pattern0 := CreateRelationship("var2", []string{"t1", "t2"})
	_, _, err = g.PushRelation(pattern0, 0, 0)
	s.Require().NoError(err)

	_, _, err = g.PushNode(query.QueryNodePattern{
		Variable: "var2",
		Labels:   []string{"t1", "t2"},
	})
	s.Require().EqualError(err, "Variable 'var2' is assigned to a different type")
}

func (s *QueryGraphSuite) TestCannotPushARelationBoundToUnexistingNode() {
	g := NewQueryGraph()

	pattern0 := CreateRelationship("var2", []string{"t1", "t2"})
	_, _, err := g.PushRelation(pattern0, 0, 0)
	s.Require().EqualError(err, "Cannot push relation bound to an unexisting node")
}

func TestShouldRunQueryGraphSuite(t *testing.T) {
	suite.Run(t, new(QueryGraphSuite))
}
