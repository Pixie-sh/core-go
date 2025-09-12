package validators

import (
	"context"
	"testing"
	"time"
)

type testTodayOrBeforeValidator struct {
	DatePtr *time.Time `validate:"today_or_before"`
	Date    time.Time  `validate:"today_or_before"`
}

func TestTodayOrBefore(t *testing.T) {

	v, err := New(TodayOrBefore)
	if err != nil {
		t.Error(err)
	}
	beforeDate := time.Now().AddDate(-50, 0, 0)
	err = v.StructCtx(context.Background(), testTodayOrBeforeValidator{
		DatePtr: &beforeDate,
		Date:    beforeDate,
	})
	if err != nil {
		t.Error(err)
	}

	afterDate := time.Now().AddDate(50, 0, 0)
	err = v.StructCtx(context.Background(), testTodayOrBeforeValidator{
		DatePtr: &afterDate,
		Date:    afterDate,
	})

	if err == nil {
		t.Fatal("err should not be nil")
	}
}
