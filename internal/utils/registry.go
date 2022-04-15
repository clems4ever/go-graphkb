package utils

type Registry[T comparable, U any] struct {
	values map[T]U
}

func NewRegistry[T comparable, U any]() *Registry[T, U] {
	return &Registry[T, U]{values: make(map[T]U)}
}

func (r *Registry[T, U]) Get(k T) (U, bool) {
	v, ok := r.values[k]
	return v, ok
}

func (r *Registry[T, U]) Set(k T, v U) {
	r.values[k] = v
}

func (r *Registry[T, U]) Del(k T) {
	delete(r.values, k)
}
