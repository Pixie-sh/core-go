package validators

import (
	"context"
	"testing"
)

type listOfWeekdaysUint64 struct {
	List []uint64 `validate:"is_weekday"`
}

type listOfWeekdaysUint32 struct {
	List []uint32 `validate:"is_weekday"`
}

func TestIsWeekdayUint64(t *testing.T) {

	v, err := New(IsWeekday)
	if err != nil {
		t.Error(err)
	}

	err = v.StructCtx(context.Background(), listOfWeekdaysUint64{
		List: []uint64{0, 5, 7},
	})
	if err == nil {
		t.Error("err should be nil")
	}

	err = v.StructCtx(context.Background(), listOfWeekdaysUint64{
		List: []uint64{0, 5, 6},
	})
	if err != nil {
		t.Error(err)
	}
}

func TestIsWeekdayUint32(t *testing.T) {

	v, err := New(IsWeekday)
	if err != nil {
		t.Error(err)
	}

	err = v.StructCtx(context.Background(), listOfWeekdaysUint32{
		List: []uint32{0, 5, 7},
	})
	if err == nil {
		t.Error("err should be nil")
	}

	err = v.StructCtx(context.Background(), listOfWeekdaysUint32{
		List: []uint32{0, 5, 6},
	})
	if err != nil {
		t.Error(err)
	}
}
