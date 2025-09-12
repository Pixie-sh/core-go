package pubsub

// Subscriber it's an entity that wants to process and be published with certain types of content
type Subscriber[T any] interface {
	ID() string
	Publish(msg T)
}

// Publisher represent an entity that from other external source
// can publish to it subscribers certain types of content
type Publisher[T any] interface {
	Subscribe(sub Subscriber[T])
}
