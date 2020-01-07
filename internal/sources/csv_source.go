package sources

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"

	"github.com/clems4ever/go-graphkb/internal/knowledge"
	"github.com/spf13/viper"
)

type CSVSource struct {
	schemaPath string
	dataPath   string
}

func NewCSVSource() *CSVSource {
	csvSource := new(CSVSource)
	csvSource.schemaPath = viper.GetString("sources.csv.schema")
	csvSource.dataPath = viper.GetString("sources.csv.data")
	return csvSource
}

func (cs *CSVSource) Name() string {
	return "csv"
}

func (cs *CSVSource) Graph() (*knowledge.SchemaGraph, error) {
	file, err := os.Open(cs.schemaPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	r := csv.NewReader(file)

	graph := knowledge.NewSchemaGraph()

	header := true

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("Unable to read schema in CSV file: %v", err)
		}

		// Skip header line
		if header {
			header = false
			continue
		}

		fromType := graph.AddAsset(record[0])
		toType := graph.AddAsset(record[2])
		graph.AddRelation(fromType, record[1], toType)
	}
	return &graph, nil
}

func (cs *CSVSource) Start(emitter *knowledge.GraphEmitter) error {
	file, err := os.Open(cs.dataPath)
	if err != nil {
		return err
	}
	defer file.Close()

	r := csv.NewReader(file)

	previousGraph, err := emitter.Read()
	if err != nil {
		return err
	}

	tx := emitter.CreateCompleteGraphTransaction(previousGraph)

	header := true

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("Unable to read schema in CSV file: %v", err)
		}

		// Skip header line
		if header {
			header = false
			continue
		}

		relationType := knowledge.RelationType{
			FromType: knowledge.AssetType(record[0]),
			ToType:   knowledge.AssetType(record[3]),
			Type:     knowledge.RelationKeyType(record[2]),
		}

		tx.Relate(record[1], relationType, record[4])
	}
	tx.Commit()
	return nil
}

func (cs *CSVSource) Stop() error {
	return nil
}
