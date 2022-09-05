package datalarkengine

import (
	"go.starlark.net/starlark"
	"testing"
)

func assertDatalark(t *testing.T, expect, actual Value) {
	t.Helper()
	if expect.Type() != actual.Type() {
		t.Errorf("type mismatch, expect %s, actual %s", expect.Type(), actual.Type())
	}
	if expect.String() != actual.String() {
		t.Errorf("value mismatch, expect %s, actual %s", expect.String(), actual.String())
	}
}

func TestStarlarkToDatalarkValue(t *testing.T) {
	// Int
	dv, err := starlarkToDatalarkValue(starlark.MakeInt(3))
	if err != nil {
		t.Fatal(err)
	}
	expectInt := NewInt(3)
	assertDatalark(t, expectInt, dv)

	// String
	dv, err = starlarkToDatalarkValue(starlark.String("apple"))
	if err != nil {
		t.Fatal(err)
	}
	expectString := NewString("apple")
	assertDatalark(t, expectString, dv)

	// List
	values := []starlark.Value{starlark.MakeInt(1), starlark.MakeInt(2)}
	dv, err = starlarkToDatalarkValue(starlark.NewList(values))
	if err != nil {
		t.Fatal(err)
	}
	expectList, err := NewList(starlark.NewList(values))
	if err != nil {
		t.Fatal(err)
	}
	assertDatalark(t, expectList, dv)
}
