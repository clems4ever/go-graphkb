package knowledge

import (
	"testing"

	"github.com/clems4ever/go-graphkb/internal/schema"
	"github.com/clems4ever/go-graphkb/internal/utils"
	"github.com/stretchr/testify/suite"
)

type SourceUpdatesSuite struct {
	suite.Suite
}

func (s *SourceUpdatesSuite) TestShouldUpsertForCreatingGraph() {
	g := NewGraph()
	ip1, err := g.AddAsset("ip", "127.0.0.1")
	s.Require().NoError(err)
	ip2, err := g.AddAsset("ip", "192.168.0.1")
	s.Require().NoError(err)

	rel := g.AddRelation(ip1, "linked", ip2)

	bulk := GenerateGraphUpdatesBulk(g)

	s.Require().Len(bulk.GetAssetUpserts(), 2)
	s.Require().Len(bulk.GetRelationUpserts(), 1)
	s.Require().Len(bulk.GetAssetRemovals(), 0)
	s.Require().Len(bulk.GetRelationRemovals(), 0)

	s.Assert().ElementsMatch(bulk.GetAssetUpserts(), []Asset{Asset(ip2), Asset(ip1)})
	s.Assert().ElementsMatch(bulk.GetRelationUpserts(), []Relation{rel})
}

func (s *SourceUpdatesSuite) TestShouldUpsertAssets() {
	g1 := NewGraph()
	ip1, err := g1.AddAsset("ip", "127.0.0.1")
	s.Require().NoError(err)
	ip2, err := g1.AddAsset("ip", "192.168.0.1")
	s.Require().NoError(err)

	g1.AddRelation(ip1, "linked", ip2)
	g1.Clean()

	ip3, err := g1.AddAsset("ip", "10.0.0.1")
	s.Require().NoError(err)
	ip4, err := g1.AddAsset("ip", "10.0.0.2")
	s.Require().NoError(err)

	bulk := GenerateGraphUpdatesBulk(g1)

	s.Require().Len(bulk.GetAssetUpserts(), 2)
	s.Require().Len(bulk.GetRelationUpserts(), 0)
	s.Require().Len(bulk.GetAssetRemovals(), 2)
	s.Require().Len(bulk.GetRelationRemovals(), 1)

	s.Assert().ElementsMatch(bulk.GetAssetUpserts(), []Asset{Asset(ip3), Asset(ip4)})
}

func (s *SourceUpdatesSuite) TestShouldUpsertRelations() {
	g1 := NewGraph()
	ip1, err := g1.AddAsset("ip", "127.0.0.1")
	s.Require().NoError(err)
	ip2, err := g1.AddAsset("ip", "192.168.0.1")
	s.Require().NoError(err)

	g1.AddRelation(ip1, "linked", ip2)
	g1.Clean()

	ip3, err := g1.AddAsset("ip", "10.0.0.1")
	s.Require().NoError(err)
	r1 := g1.AddRelation(ip3, "linked", ip1)
	r2 := g1.AddRelation(ip3, "linked", ip2)

	bulk := GenerateGraphUpdatesBulk(g1)

	s.Require().Len(bulk.GetAssetUpserts(), 1)
	s.Require().Len(bulk.GetRelationUpserts(), 2)
	s.Require().Len(bulk.GetAssetRemovals(), 2)
	s.Require().Len(bulk.GetRelationRemovals(), 1)

	s.Assert().ElementsMatch(bulk.GetAssetUpserts(), []Asset{Asset(ip3)})
	s.Assert().ElementsMatch(bulk.GetRelationUpserts(), []Relation{r1, r2})
}

func (s *SourceUpdatesSuite) TestShouldRemoveGraph() {
	g1 := NewGraph()
	ip1, err := g1.AddAsset("ip", "127.0.0.1")
	s.Require().NoError(err)
	ip2, err := g1.AddAsset("ip", "192.168.0.1")
	s.Require().NoError(err)
	r := g1.AddRelation(ip1, "linked", ip2)
	g1.Clean()

	bulk := GenerateGraphUpdatesBulk(g1)

	s.Require().Len(bulk.GetAssetUpserts(), 0)
	s.Require().Len(bulk.GetRelationUpserts(), 0)
	s.Require().Len(bulk.GetAssetRemovals(), 2)
	s.Require().Len(bulk.GetRelationRemovals(), 1)

	s.Assert().ElementsMatch(bulk.GetAssetRemovals(), []Asset{Asset(ip1), Asset(ip2)})
	s.Assert().ElementsMatch(bulk.GetRelationRemovals(), []Relation{r})
}

func (s *SourceUpdatesSuite) TestShouldGenerateBulkOfSubgraph() {
	g1 := NewGraph()
	ip1, err := g1.AddAsset("ip", "127.0.0.1")
	s.Require().NoError(err)
	ip2, err := g1.AddAsset("ip", "192.168.0.1")
	s.Require().NoError(err)
	r := g1.AddRelation(ip1, "linked", ip2)

	g1.Clean()
	_, err = g1.AddAsset("ip", "127.0.0.1")
	s.Require().NoError(err)

	bulk := GenerateGraphUpdatesBulk(g1)

	s.Require().Len(bulk.GetAssetUpserts(), 0)
	s.Require().Len(bulk.GetRelationUpserts(), 0)
	s.Require().Len(bulk.GetAssetRemovals(), 1)
	s.Require().Len(bulk.GetRelationRemovals(), 1)

	s.Assert().ElementsMatch(bulk.GetAssetRemovals(), []Asset{Asset(ip2)})
	s.Assert().ElementsMatch(bulk.GetRelationRemovals(), []Relation{r})
}

func (s *SourceUpdatesSuite) TestShouldGenerateBulkForMixedAdditionsAndRemovals() {
	g1 := NewGraph()
	ip1, err := g1.AddAsset("ip", "127.0.0.1")
	s.Require().NoError(err)
	ip2, err := g1.AddAsset("ip", "192.168.0.1")
	s.Require().NoError(err)
	r := g1.AddRelation(ip1, "linked", ip2)

	g1.Clean()
	_, err = g1.AddAsset("ip", "127.0.0.1")
	s.Require().NoError(err)
	ip3, err := g1.AddAsset("ip", "10.0.0.1")
	s.Require().NoError(err)
	r2 := g1.AddRelation(ip3, "linked", ip2)

	bulk := GenerateGraphUpdatesBulk(g1)

	s.Require().Len(bulk.GetAssetUpserts(), 1)
	s.Require().Len(bulk.GetRelationUpserts(), 1)
	s.Require().Len(bulk.GetAssetRemovals(), 1)
	s.Require().Len(bulk.GetRelationRemovals(), 1)

	s.Assert().ElementsMatch(bulk.GetAssetUpserts(), []Asset{Asset(ip3)})
	s.Assert().ElementsMatch(bulk.GetRelationUpserts(), []Relation{r2})
	s.Assert().ElementsMatch(bulk.GetAssetRemovals(), []Asset{Asset(ip2)})
	s.Assert().ElementsMatch(bulk.GetRelationRemovals(), []Relation{r})
}

func (s *SourceUpdatesSuite) TestAssetValidation() {
	asset := schema.AssetType("asset")
	reg := schema.AssetValidationRegistry.(*utils.Registry[schema.AssetType, []schema.AssetValidationFunc])

	defer reg.Del(asset)
	schema.AddAssetValidator(asset, func(s string) bool {
		return s == "foo"
	})

	g := NewGraph()

	_, err := g.AddAsset(asset, "bar")
	s.Require().Error(err)
	s.Require().Empty(g.Assets())

	_, err = g.AddAsset(asset, "foo")
	s.Require().NoError(err)

	assets := []Asset{}
	for a := range g.Assets() {
		assets = append(assets, a)
	}
	s.Require().Equal([]Asset{{Type: asset, Key: "foo"}}, assets)
}

func TestGraphUpdatesSuite(t *testing.T) {
	suite.Run(t, new(SourceUpdatesSuite))
}
