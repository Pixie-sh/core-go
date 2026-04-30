package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Implementation tests live in github.com/pixie-sh/clone-go. The cases here
// are wrapper-level smoke tests: they verify that the re-export shim and the
// generic type alias actually delegate correctly.

type smokePerson struct {
	Name string
	Tags []string
}

type smokeNode struct {
	Data int
	Self *smokeNode
}

type smokeCloner struct {
	Marker int
}

func (s *smokeCloner) Clone() *smokeCloner {
	out := *s
	out.Marker++
	return &out
}

func TestClone_WrapperDelegates(t *testing.T) {
	p := &smokePerson{Name: "Alice", Tags: []string{"a"}}
	c := Clone(p)
	assert.NotSame(t, p, c)
	c.Name = "Bob"
	c.Tags[0] = "MUT"
	assert.Equal(t, "Alice", p.Name)
	assert.Equal(t, "a", p.Tags[0])
}

func TestCloneable_AliasIdentity(t *testing.T) {
	// *smokeCloner satisfies types.Cloneable[smokeCloner]; the alias means it
	// also satisfies clone.Cloneable[smokeCloner], so Clone hits the escape
	// hatch and increments Marker rather than going through reflect.
	src := &smokeCloner{Marker: 1}
	out := Clone(src)
	assert.Equal(t, 2, out.Marker, "Cloneable.Clone() must run via the alias-bridged dispatch")
}

func TestCloneSlowly_WrapperHandlesCycle(t *testing.T) {
	n := &smokeNode{Data: 1}
	n.Self = n
	c := CloneSlowly(n)
	assert.NotSame(t, n, c)
	assert.Same(t, c, c.Self, "self-cycle must close on the cloned node")
}

func TestClone_NilThroughWrapper(t *testing.T) {
	assert.Nil(t, Clone[smokePerson](nil))
	assert.Nil(t, CloneSlowly[smokePerson](nil))
}
