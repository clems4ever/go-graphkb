package schema

import "context"

// Persistor is a persistor of schema
type Persistor interface {
	SaveSchema(ctx context.Context, sourceName string, sg SchemaGraph) error
	LoadSchema(ctx context.Context, sourceName string) (SchemaGraph, error)
}
