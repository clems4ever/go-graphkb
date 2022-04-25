package knowledge

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type GraphSuite struct {
	suite.Suite
}

func (s *GraphSuite) TestShouldTestGraphsAreEqual() {
	g := NewGraph()

	ip1, err := g.AddAsset("ip", "127.0.0.1")
	s.Require().NoError(err)
	ip2, err := g.AddAsset("ip", "127.0.0.2")
	s.Require().NoError(err)
	g.AddRelation(ip1, "linked", ip2)

	s.Assert().True(g.Equal(g))
}

func (s *GraphSuite) TestShouldTestCopiedGraphsAreEqual() {
	g := NewGraph()

	ip1, err := g.AddAsset("ip", "127.0.0.1")
	s.Require().NoError(err)
	ip2, err := g.AddAsset("ip", "127.0.0.2")
	s.Require().NoError(err)
	g.AddRelation(ip1, "linked", ip2)

	g2 := g.Copy()

	s.Assert().True(g.Equal(g2))
	s.Assert().True(g2.Equal(g))
}

func (s *GraphSuite) TestShouldTestGraphsAreDifferent() {
	g := NewGraph()

	ip1, err := g.AddAsset("ip", "127.0.0.1")
	s.Require().NoError(err)
	ip2, err := g.AddAsset("ip", "127.0.0.2")
	s.Require().NoError(err)
	g.AddRelation(ip1, "linked", ip2)

	g2 := g.Copy()
	g2.AddAsset("ip", "127.0.0.3")

	s.Assert().False(g.Equal(g2))
	s.Assert().False(g2.Equal(g))
}

func (s *GraphSuite) TestShouldTestGraphsHasAsset() {
	g := NewGraph()

	ip1, err := g.AddAsset("ip", "127.0.0.1")
	s.Require().NoError(err)
	ip2, err := g.AddAsset("ip", "127.0.0.2")
	s.Require().NoError(err)
	g.AddRelation(ip1, "linked", ip2)

	s.Assert().True(g.HasAsset(Asset(ip1)))
	s.Assert().False(g.HasAsset(Asset{Type: "abc", Key: "abc"}))
}

func (s *GraphSuite) TestShouldTestGraphsHasRelation() {
	g := NewGraph()

	ip1, err := g.AddAsset("ip", "127.0.0.1")
	s.Require().NoError(err)
	ip2, err := g.AddAsset("ip", "127.0.0.2")
	s.Require().NoError(err)
	rel := g.AddRelation(ip1, "linked", ip2)

	s.Assert().True(g.HasRelation(Relation(rel)))
	s.Assert().False(g.HasRelation(Relation{
		Type: "test",
		From: AssetKey{Type: "test", Key: "test"},
		To:   AssetKey{Type: "test", Key: "test"},
	}))
}

func TestGraphSuite(t *testing.T) {
	suite.Run(t, new(GraphSuite))
}
