package schema

import "context"

type Persistor interface {
	ListSources(ctx context.Context) ([]string, error)

	SaveSchema(ctx context.Context, sourceName string, sg SchemaGraph) error
	LoadSchema(ctx context.Context, sourceName string) (SchemaGraph, error)
}
