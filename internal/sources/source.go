package sources

import "github.com/clems4ever/go-graphkb/internal/knowledge"

// Source represent a source of data
type Source interface {
	Name() string
	Graph() (*knowledge.SchemaGraph, error)

	Start(emitter *knowledge.GraphEmitter) error
	Stop() error
}
