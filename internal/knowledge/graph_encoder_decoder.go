package knowledge

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
)

// ******************************* DECODER ***********************************

// GraphEncoder represent a graph encoder
type GraphEncoder struct {
	writer io.Writer

	jsonEncoder *json.Encoder
}

// NewGraphEncoder create an instance of a graph encoder
func NewGraphEncoder(w io.Writer) *GraphEncoder {
	return &GraphEncoder{
		writer:      w,
		jsonEncoder: json.NewEncoder(w),
	}
}

// EncodeRelation encode a relation
func (ge *GraphEncoder) EncodeRelation(relation Relation) error {
	_, err := fmt.Fprint(ge.writer, "R")
	if err != nil {
		return err
	}
	err = ge.jsonEncoder.Encode(relation)
	if err != nil {
		return err
	}
	return nil
}

// EncodeAsset encode an asset
func (ge *GraphEncoder) EncodeAsset(asset Asset) error {
	_, err := fmt.Fprint(ge.writer, "A")
	if err != nil {
		return err
	}
	err = ge.jsonEncoder.Encode(asset)
	if err != nil {
		return err
	}
	return nil
}

// ******************************* DECODER ***********************************

// GraphDecoder represent a graph decoder
type GraphDecoder struct {
	reader io.Reader

	jsonDecoder *json.Decoder
}

// NewGraphDecoder create an instance of a graph encoder
func NewGraphDecoder(r io.Reader) *GraphDecoder {
	return &GraphDecoder{
		reader:      r,
		jsonDecoder: json.NewDecoder(r),
	}
}

// Decode decode a graph
func (ge *GraphDecoder) Decode(graph *Graph) error {
	scanner := bufio.NewScanner(ge.reader)
	for scanner.Scan() {
		line := scanner.Text()
		switch line[0] {
		case 'A':
			var asset Asset
			err := json.Unmarshal([]byte(line[1:]), &asset)
			if err != nil {
				return err
			}
			graph.AddAsset(asset.Type, asset.Key)
		case 'R':
			var relation Relation
			err := json.Unmarshal([]byte(line[1:]), &relation)
			if err != nil {
				return err
			}

			graph.AddRelation(relation.From, relation.Type, relation.To)
		}
	}
	return nil
}
