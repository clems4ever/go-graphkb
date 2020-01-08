package sources

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/clems4ever/go-graphkb/internal/knowledge"
	"github.com/clems4ever/go-graphkb/internal/schema"
	"github.com/spf13/viper"
)

type CSVSource struct {
	dataPath string
}

func NewCSVSource() *CSVSource {
	csvSource := new(CSVSource)
	csvSource.dataPath = viper.GetString("path")

	if csvSource.dataPath == "" {
		log.Fatal(fmt.Errorf("Unable to detect CSV file path in configuration. Check patch configuration is provided"))
	}

	return csvSource
}

func (cs *CSVSource) Start(importer *knowledge.GraphImporter) error {
	file, err := os.Open(cs.dataPath)
	if err != nil {
		return err
	}
	defer file.Close()

	r := csv.NewReader(file)

	previousGraph, err := importer.ReadCurrentGraph()
	if err != nil {
		return fmt.Errorf("Unable to read previous graph: %v", err)
	}

	tx := importer.CreateTransaction(previousGraph)

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

		relationType := schema.RelationType{
			FromType: schema.AssetType(record[0]),
			ToType:   schema.AssetType(record[3]),
			Type:     schema.RelationKeyType(record[2]),
		}

		tx.Relate(record[1], relationType, record[4])
	}
	_, err = tx.Commit()
	fmt.Println("CSV data has been sent successfully")
	return err
}

func (cs *CSVSource) Stop() error {
	return nil
}
