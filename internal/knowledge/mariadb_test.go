// +build integration

package knowledge

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type MariaDBSuite struct {
	suite.Suite

	database *MariaDB
}

func (s *MariaDBSuite) SetupSuite() {
	SchemaRegistrySingleton = *NewSchemaRegistry()
	SchemaRegistrySingleton.AddAssetType("source")
	SchemaRegistrySingleton.AddRelationType("observed")
	SchemaRegistrySingleton.AddAssetType("ip")
	SchemaRegistrySingleton.AddAssetType("hostname")
	SchemaRegistrySingleton.AddRelationType("linked")

	s.database = NewMariaDB("root", "example", "", "test_db")
}

func (s *MariaDBSuite) SetupTest() {
	err := s.database.FlushAll()
	s.Require().NoError(err)

	err = s.database.InitializeSchema()
	s.Require().NoError(err)
}

func (s *MariaDBSuite) TestShouldRemoveAssetIfThereIsNoMoreEdge() {
	sourceGraph1 := NewSourceGraph("mysource")
	ip1 := sourceGraph1.AddAsset("ip", "127.0.0.1")
	ip2 := sourceGraph1.AddAsset("ip", "192.168.0.1")
	sourceGraph1.AddRelation(ip1, "linked", ip2)

	bulkCreation := GenerateGraphUpdatesBulk(
		nil,
		sourceGraph1.Graph)

	err := s.database.UpdateGraph("mysource", bulkCreation)
	s.Require().NoError(err)

	assetCount, err := s.database.CountAssets()
	s.Require().NoError(err)
	s.Require().Equal(int64(3), assetCount)

	relationCount, err := s.database.CountRelations()
	s.Require().NoError(err)
	s.Require().Equal(int64(3), relationCount)

	// Create new graph which is a subset of graph 1
	sourceGraph2 := NewSourceGraph("mysource")
	sourceGraph2.AddAsset("ip", "127.0.0.1")
	bulkRemoval := GenerateGraphUpdatesBulk(sourceGraph1.Graph, sourceGraph2.Graph)

	err = s.database.UpdateGraph("mysource", bulkRemoval)
	s.Require().NoError(err)

	assetCount, err = s.database.CountAssets()
	s.Require().NoError(err)
	s.Require().Equal(int64(2), assetCount)

	relationCount, err = s.database.CountRelations()
	s.Require().NoError(err)
	s.Require().Equal(int64(1), relationCount)
}

func (s *MariaDBSuite) TestShouldWriteAndReadBackGraph() {
	sourceGraph := NewSourceGraph("mysource")
	ip1 := sourceGraph.AddAsset("ip", "127.0.0.1")
	ip2 := sourceGraph.AddAsset("ip", "192.168.0.1")
	sourceGraph.AddRelation(ip1, "linked", ip2)
	host1 := sourceGraph.AddAsset("hostname", "myhost1")
	host2 := sourceGraph.AddAsset("hostname", "myhost2")
	sourceGraph.AddRelation(ip1, "linked", host1)
	sourceGraph.AddRelation(ip2, "linked", host2)

	s.Assert().Len(sourceGraph.Assets(), 5)
	s.Assert().Len(sourceGraph.Relations(), 7)

	bulk := GenerateGraphUpdatesBulk(nil, sourceGraph.Graph)

	err := s.database.UpdateGraph("mysource", bulk)
	s.Require().NoError(err)

	newGraph := NewGraph()
	err = s.database.ReadGraph("mysource", newGraph)
	s.Require().NoError(err)

	s.Assert().Len(newGraph.Assets(), 5)
	s.Assert().Len(newGraph.Relations(), 7)

	s.Assert().True(sourceGraph.Equal(newGraph))
}

func TestMariaDBSuite(t *testing.T) {
	suite.Run(t, new(MariaDBSuite))
}
