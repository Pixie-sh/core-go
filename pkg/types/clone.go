package types

import (
	clone "github.com/huandu/go-clone/generic"
)

// Cloneable defines the interface for types that can be cloned.
type Cloneable[T any] interface {
	// Clone creates a deep copy of the object.
	Clone() *T
}

// Clone creates a deep copy of the object passed as argument.
// Only works for types that are json serializable, and all fields must be exported.
func Clone[T any](o *T) *T {
	// check if o implements the Cloneable interface
	// if it does, call its Clone method
	if c, ok := (interface{})(o).(Cloneable[T]); ok {
		return c.Clone()
	}

	return clone.Clone(o)
}

// CloneSlowly creates a deep copy of the object passed as argument.
// Only works for types that are json serializable, and all fields must be exported.
// Use this function when the object has pointers to itself.
func CloneSlowly[T any](o *T) *T {
	return clone.Slowly(o)
}
