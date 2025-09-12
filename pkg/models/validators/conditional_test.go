package validators

import (
	"context"
	"fmt"
	"testing"
)

type ConditionalEnum string

const (
	ASAPSchedule     ConditionalEnum = "soon_as_possible_schedule"
	SpecificSchedule ConditionalEnum = "specific_schedule"
	SingleSchedule   ConditionalEnum = "single_datetime_schedule"
	MultipleSchedule ConditionalEnum = "multiple_datetimes_schedule"
) //@Field ScheduleType

var ScheduleTypeList = []ConditionalEnum{
	ASAPSchedule,
	SpecificSchedule,
	SingleSchedule,
	MultipleSchedule,
}

type ConditionalStruct struct {
	Enum                  string
	FieldToValidateIfEnum int `validate:"conditional=Enum:active->required"`
	CheckEmail            bool
	Email                 string `validate:"conditional=CheckEmail:true->email"`
	CondEnum              *ConditionalEnum
	CondValue             *string `validate:"conditional=CondEnum:soon_as_possible_schedule->required"`
}

func TestConditional(t *testing.T) {
	checkNilMap := map[string]bool{
		"conditional": true,
	}
	v, err := NewWithValidateNil(checkNilMap, Conditional())
	if err != nil {
		t.Fatal(err)
	}
	test := "test"
	sgSchedule := SingleSchedule
	aSchedule := ASAPSchedule
	err = v.StructCtx(context.TODO(), ConditionalStruct{
		Enum:                  "active",
		FieldToValidateIfEnum: 5,
		CheckEmail:            false,
		Email:                 "Whatever",
		CondEnum:              &sgSchedule,
		CondValue:             &test,
	})
	if err != nil {
		fmt.Println(err)
		t.Fatal(err)
	}

	err = v.StructCtx(context.TODO(), ConditionalStruct{
		Enum:       "inactive",
		CheckEmail: true,
		Email:      "test@whatever.com",
		CondEnum:   &sgSchedule,
		CondValue:  nil,
	})
	if err != nil {
		t.Fatal(err)
	}

	err = v.StructCtx(context.TODO(), ConditionalStruct{
		CondEnum: &sgSchedule,
	})
	if err != nil {
		t.Fatal(err)
	}

	err = v.StructCtx(context.TODO(), ConditionalStruct{
		CondEnum:  &aSchedule,
		CondValue: &test,
	})
	if err != nil {
		t.Fatal(err)
	}

	structTest := ConditionalStruct{
		CondEnum: &sgSchedule,
	}

	err = v.StructCtx(context.TODO(), &structTest)
	if err != nil {
		t.Fatal(err)
	}
}
