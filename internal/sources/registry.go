package sources

// Registry a registry singleton
var Registry registry

// Registry a registry of sources
type registry struct {
	sources map[string]ObservableSource
}

func init() {
	Registry.sources = make(map[string]ObservableSource)
}

// Add add a source
func (r *registry) Add(s Source) {
	r.sources[s.Name()] = NewObservableSource(s)
}

func (r *registry) Exist(sourceName string) bool {
	_, ok := r.sources[sourceName]
	return ok
}

// Get get a source by name
func (r *registry) Get(sourceName string) *ObservableSource {
	if s, ok := r.sources[sourceName]; ok {
		return &s
	}
	return nil
}

func (r *registry) GetAll() []ObservableSource {
	sources := make([]ObservableSource, 0)
	for _, s := range r.sources {
		sources = append(sources, s)
	}
	return sources
}

func (r *registry) GetAllNames() []string {
	sources := make([]string, 0)
	for _, s := range r.sources {
		sources = append(sources, s.Name())
	}
	return sources
}
