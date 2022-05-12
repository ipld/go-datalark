package datalarkengine

import (
	"testing"
)

func assertEqual(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Errorf("%v != %v", a, b)
	}
}

func TestBasicTypes(t *testing.T) {
	var val Value

	val = NewNull()
	assertEqual(t, val.String(), "null")
	assertEqual(t, val.Type(), "datalark.null")

	val = NewBool(true)
	assertEqual(t, val.String(), "bool{true}")
	assertEqual(t, val.Type(), "datalark.bool")

	val = NewInt(34)
	assertEqual(t, val.String(), "int{34}")
	assertEqual(t, val.Type(), "datalark.int")

	val = NewFloat(7.2)
	assertEqual(t, val.String(), "float{7.2}")
	assertEqual(t, val.Type(), "datalark.float")

	val = NewString("hi")
	assertEqual(t, val.String(), "string{\"hi\"}")
	assertEqual(t, val.Type(), "datalark.string")

	val = NewBytes([]byte{0x12, 0x56, 0x90})
	assertEqual(t, val.String(), "bytes{125690}")
	assertEqual(t, val.Type(), "datalark.bytes")

	val = NewLink(newTestLink())
	assertEqual(t, val.String(), "link{bafkqabiaaebagba}")
	assertEqual(t, val.Type(), "datalark.link")
}

func TestBasicScript(t *testing.T) {
	mustParseSchemaRunScriptAssertOutput(t, "", "", `
b = datalark.Bool(True)
print(b)

n = datalark.Int(34)
print(n)

f = datalark.Float(7.2)
print(f)

s = datalark.String('hi')
print(s)

d = datalark.Bytes(bytes([0x12, 0x56, 0x90]))
print(d)
`,
		`bool{true}
int{34}
float{7.2}
string{"hi"}
bytes{125690}
`,
	)
}
