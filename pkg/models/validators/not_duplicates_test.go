package validators

import (
	"context"
	"testing"
)

type listOfStringNotDuplicated struct {
	List []string `validate:"no_duplicates"`
}

type listOfUint64NotDuplicated struct {
	List []uint64 `validate:"no_duplicates"`
}

func TestNotDuplicatedString(t *testing.T) {

	v, err := New(NoDuplicates)
	if err != nil {
		t.Error(err)
	}

	err = v.StructCtx(context.Background(), listOfStringNotDuplicated{
		List: []string{"#test", "#testAgain"},
	})
	if err != nil {
		t.Error(err)
	}

	err = v.StructCtx(context.Background(), listOfStringNotDuplicated{
		List: []string{"#test", "#test", "#anotherOne"},
	})
	if err == nil {
		t.Error("err shouldn't be nil")
	}

	testStruct := listOfStringNotDuplicated{
		List: []string{"#test", "#testAgain"},
	}

	err = v.StructCtx(context.Background(), &testStruct)
	if err != nil {
		t.Error(err)
	}

}

func TestNotDuplicatedUint64(t *testing.T) {

	v, err := New(NoDuplicates)
	if err != nil {
		t.Error(err)
	}

	err = v.StructCtx(context.Background(), listOfUint64NotDuplicated{
		List: []uint64{1, 2, 3},
	})
	if err != nil {
		t.Error(err)
	}

	err = v.StructCtx(context.Background(), listOfUint64NotDuplicated{
		List: []uint64{1, 1, 3},
	})
	if err == nil {
		t.Error("err shouldn't be nil")
	}
}
