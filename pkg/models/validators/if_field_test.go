package validators

import (
	"context"
	"testing"
)

type FakeStruct struct {
	FieldPredicate            string
	FieldToEventuallyValidate *int `validate:"if_field=FieldPredicate:atum->required,numeric"`
}

func TesIfField(t *testing.T) {
	v, err := New(IfField())
	if err != nil {
		t.Fatal(err)
	}

	five := 5
	err = v.StructCtx(context.TODO(), FakeStruct{
		FieldPredicate:            "atum",
		FieldToEventuallyValidate: &five,
	})
	if err != nil {
		t.Fatal(err)
	}

	err = v.StructCtx(context.TODO(), FakeStruct{
		FieldPredicate: "atum",
	})
	if err != nil {
		t.Fatal(err)
	}
}
