package sources

import "github.com/clems4ever/go-graphkb/internal/knowledge"

// Source represent a source of data
type Source interface {
	Start(emitter *knowledge.GraphEmitter) error
	Stop() error
}
