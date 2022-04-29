package datalarkengine

import (
	"testing"
)

func assertEqual(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Fatalf("%v != %v", a, b)
	}
}

func TestBasic(t *testing.T) {
	var val Value

	val = &NullValue{}
	assertEqual(t, val.Type(), "Null")

	val = &BooleanValue{}
	assertEqual(t, val.Type(), "Boolean")

	val = &IntegerValue{i: 32}
	assertEqual(t, val.Type(), "Integer")
}
