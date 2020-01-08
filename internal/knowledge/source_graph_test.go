package knowledge

import (
	"testing"

	"github.com/clems4ever/go-graphkb/internal/schema"
	"github.com/stretchr/testify/suite"
)

type SourceGraphSuite struct {
	suite.Suite
}

func (s *SourceGraphSuite) TestShouldHaveSourceInEmptyGraph() {
	sourceGraph := NewSourceGraph("mysource")

	s.Assert().Len(sourceGraph.Assets(), 1)
	s.Assert().Len(sourceGraph.Relations(), 0)

	sourceAsset := Asset{Type: "source", Key: "mysource"}
	s.Assert().Equal(sourceAsset, sourceGraph.Assets()[0])
}

func (s *SourceGraphSuite) TestShouldHaveObservedLinksTowardAllAssets() {
	sourceGraph := NewSourceGraph("mysource")

	source := Asset{Type: "source", Key: "mysource"}
	ip1 := sourceGraph.AddAsset("ip", "127.0.0.1")
	ip2 := sourceGraph.AddAsset("ip", "192.168.0.1")
	linkRelation := sourceGraph.AddRelation(ip1, "linked", ip2)

	sourceObservedIP1Relation := Relation{
		Type: schema.RelationKeyType("observed"),
		From: AssetKey(source),
		To:   ip1,
	}

	sourceObservedIP2Relation := Relation{
		Type: schema.RelationKeyType("observed"),
		From: AssetKey(source),
		To:   ip2,
	}

	s.Assert().Len(sourceGraph.Assets(), 3)
	s.Assert().Len(sourceGraph.Relations(), 3)

	s.Assert().Contains(sourceGraph.Assets(), Asset(ip1))
	s.Assert().Contains(sourceGraph.Assets(), Asset(ip2))
	s.Assert().Contains(sourceGraph.Assets(), source)

	s.Assert().Contains(sourceGraph.Relations(), linkRelation)
	s.Assert().Contains(sourceGraph.Relations(), sourceObservedIP1Relation)
	s.Assert().Contains(sourceGraph.Relations(), sourceObservedIP2Relation)
}

func TestSourceGraphSuite(t *testing.T) {
	suite.Run(t, new(SourceGraphSuite))
}
