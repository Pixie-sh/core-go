package validators

import (
	"context"
	"testing"
	"time"
)

type todayOrLater struct {
	TodayOrLater    time.Time  `validate:"today_or_later"`
	TodayOrLaterPtr *time.Time `validate:"today_or_later"`
}

func TestTodayOrLater(t *testing.T) {

	v, err := New(TodayOrLater)
	if err != nil {
		t.Error(err)
	}

	olderThanNow := time.Date(2020, time.March, 29, 1, 1, 0, 0, time.UTC)
	notOlderThanNow := time.Date(2999, time.March, 29, 1, 1, 0, 0, time.UTC)
	err = v.StructCtx(context.Background(), todayOrLater{
		TodayOrLater:    olderThanNow,
		TodayOrLaterPtr: &olderThanNow,
	})
	if err == nil {
		t.Error("err should not be nil")
	}

	err = v.StructCtx(context.Background(), todayOrLater{
		TodayOrLater:    notOlderThanNow,
		TodayOrLaterPtr: &notOlderThanNow,
	})

	if err != nil {
		t.Error(err)
	}

}
