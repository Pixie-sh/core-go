package validators

import (
	"context"
	"testing"
)

type listOf struct {
	List []testNotRepeatedValidator `validate:"not_repeated=Index"`
}

type testNotRepeatedValidator struct {
	Index int
	Value string
}

func TestNotRepeated(t *testing.T) {

	v, err := New(NotRepeated)
	if err != nil {
		t.Error(err)
	}

	err = v.StructCtx(context.Background(), listOf{List: []testNotRepeatedValidator{
		{
			Index: 1,
			Value: "a",
		},
		{
			Index: 2,
			Value: "a",
		},
	}})
	if err != nil {
		t.Error(err)
	}

	err = v.StructCtx(context.Background(), listOf{List: []testNotRepeatedValidator{
		{
			Index: 1,
			Value: "a",
		},
		{
			Index: 1,
			Value: "a",
		},
	}})

	if err == nil {
		t.Fatal("err should not be nil")
	}
}
