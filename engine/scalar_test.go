package datalarkengine

import (
	"testing"
)

func assertEqual(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Fatalf("%v != %v", a, b)
	}
}

func TestBasicTypes(t *testing.T) {
/*
	var val Value

	val = &NullValue{}
	assertEqual(t, val.Type(), "Null")

	val = &BoolValue{}
	assertEqual(t, val.Type(), "Bool")

	val = &IntegerValue{i: 32}
	assertEqual(t, val.Type(), "Integer")

	val = &FloatValue{f: 4.56}
	assertEqual(t, val.Type(), "Float")

	val = &StringValue{s: "test"}
	assertEqual(t, val.Type(), "String")
*/
}
