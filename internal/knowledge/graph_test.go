package knowledge

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type GraphSuite struct {
	suite.Suite
}

func (s *GraphSuite) SetupSuite() {
	SchemaRegistrySingleton = *NewSchemaRegistry()
	SchemaRegistrySingleton.AddAssetType("ip")
	SchemaRegistrySingleton.AddAssetType("hostname")
	SchemaRegistrySingleton.AddRelationType("linked")
}

func (s *GraphSuite) TestShouldTestGraphsAreEqual() {
	g := NewGraph()

	ip1 := g.AddAsset("ip", "127.0.0.1")
	ip2 := g.AddAsset("ip", "127.0.0.2")
	g.AddRelation(ip1, "linked", ip2)

	s.Assert().True(g.Equal(g))
}

func (s *GraphSuite) TestShouldTestCopiedGraphsAreEqual() {
	g := NewGraph()

	ip1 := g.AddAsset("ip", "127.0.0.1")
	ip2 := g.AddAsset("ip", "127.0.0.2")
	g.AddRelation(ip1, "linked", ip2)

	g2 := g.Copy()

	s.Assert().True(g.Equal(g2))
	s.Assert().True(g2.Equal(g))
}

func (s *GraphSuite) TestShouldTestGraphsAreDifferent() {
	g := NewGraph()

	ip1 := g.AddAsset("ip", "127.0.0.1")
	ip2 := g.AddAsset("ip", "127.0.0.2")
	g.AddRelation(ip1, "linked", ip2)

	g2 := g.Copy()
	g2.AddAsset("ip", "127.0.0.3")

	s.Assert().False(g.Equal(g2))
	s.Assert().False(g2.Equal(g))
}

func TestGraphSuite(t *testing.T) {
	suite.Run(t, new(GraphSuite))
}
