package sources

import "github.com/clems4ever/go-graphkb/internal/knowledge"

// DataSource represent a source of data
type DataSource interface {
	Start(emitter *knowledge.DataSource) error
	Stop() error
}
