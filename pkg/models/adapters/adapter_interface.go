package adapters

// Adapter is a generic interface for adapting business layer entities to API layer models.
type Adapter[T any, R any] interface {
	Adapt(entity T) R
	AdaptCollection(entity []T) []R
}
