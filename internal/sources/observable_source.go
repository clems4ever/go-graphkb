package sources

import "github.com/clems4ever/go-graphkb/internal/knowledge"

// Observable source is a wrapper of all sources augmenting the schema graph with the 'observed' relation.
// All assets produced by the wrapped source are therefore 'observed' by the source itself
type ObservableSource struct {
	source Source
}

func NewObservableSource(s Source) ObservableSource {
	return ObservableSource{
		source: s,
	}
}

func (os *ObservableSource) Start(e *knowledge.GraphImporter) error {
	return os.source.Start(e)
}

func (os *ObservableSource) Stop() error {
	return os.source.Stop()
}
