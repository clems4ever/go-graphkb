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

func (os *ObservableSource) Name() string {
	return os.source.Name()
}

func (os *ObservableSource) Graph() (*knowledge.SchemaGraph, error) {
	sg, err := os.source.Graph()
	if err != nil {
		return nil, err
	}

	source := sg.AddAsset("source")
	for _, a := range sg.Assets() {
		if a == source {
			continue
		}
		sg.AddRelation(source, "observed", a)
	}
	return sg, nil
}

func (os *ObservableSource) Start(e *knowledge.GraphEmitter) error {
	return os.source.Start(e)
}

func (os *ObservableSource) Stop() error {
	return os.source.Stop()
}
