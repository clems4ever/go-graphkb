package importers

import "github.com/clems4ever/go-graphkb/internal/knowledge"

// Importer represent an importer of data
type Importer interface {
	Start(emitter *knowledge.GraphImporter) error
	Stop() error
}
