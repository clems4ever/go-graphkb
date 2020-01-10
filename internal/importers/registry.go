package importers

import "context"

// Registry is a regostry of importers with their auth tokens
type Registry interface {
	// List importers with their authentication tokens
	ListImporters(ctx context.Context) (map[string]string, error)
}
