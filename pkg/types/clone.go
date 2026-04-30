package types

import (
	"github.com/pixie-sh/clone-go"
)

// Cloneable is a type alias for clone.Cloneable[T] preserved for backward
// compatibility with downstream code that imports it as types.Cloneable[T].
// Because this is a Go 1.24 generic type alias (not a fresh interface
// definition), implementations of types.Cloneable[T] satisfy
// clone.Cloneable[T] for the dispatch check inside clone.Clone.
type Cloneable[T any] = clone.Cloneable[T]

// Clone delegates to clone.Clone. See github.com/pixie-sh/clone-go.
func Clone[T any](o *T) *T { return clone.Clone(o) }

// CloneSlowly delegates to clone.CloneSlowly. See github.com/pixie-sh/clone-go.
func CloneSlowly[T any](o *T) *T { return clone.CloneSlowly(o) }
