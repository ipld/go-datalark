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
	// Null
	dv, err := starlarkToDatalarkValue(starlark.None)
	if err != nil {
		t.Fatal(err)
	}
	expectNull := NewNull()
	assertDatalark(t, expectNull, dv)

	// Bool
	dv, err = starlarkToDatalarkValue(starlark.Bool(true))
	if err != nil {
		t.Fatal(err)
	}
	expectBool := NewBool(true)
	assertDatalark(t, expectBool, dv)

	// Int
	dv, err = starlarkToDatalarkValue(starlark.MakeInt(3))
	if err != nil {
		t.Fatal(err)
	}
	expectInt := NewInt(3)
	assertDatalark(t, expectInt, dv)

	// Float
	dv, err = starlarkToDatalarkValue(starlark.Float(5.5))
	if err != nil {
		t.Fatal(err)
	}
	expectFloat := NewFloat(5.5)
	assertDatalark(t, expectFloat, dv)

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

	// Bytes
	dv, err = starlarkToDatalarkValue(starlark.Bytes([]byte{0x07, 0x08, 0x09}))
	if err != nil {
		t.Fatal(err)
	}
	expectBytes := NewBytes([]byte{0x07, 0x08, 0x09})
	assertDatalark(t, expectBytes, dv)
}
