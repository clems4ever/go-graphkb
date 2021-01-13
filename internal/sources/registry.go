package sources

import "context"

// Registry is a regostry of data sources with their auth tokens
type Registry interface {
	// List data sources with their authentication tokens
	ListSources(ctx context.Context) (map[string]string, error)
}
