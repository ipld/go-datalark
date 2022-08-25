package datalarkengine

import (
	"testing"
)

func Example_string() {
	mustExecExample(nil, nil,
		"mytypes",
		`
		print(datalark.String('yo'))
	`)

	// Output:
	// string{"yo"}
}

func TestStringUpperMethod(t *testing.T) {
	mustParseSchemaRunScriptAssertOutput(t,
		`
	`,
		`mytypes`,
		`
str = datalark.String('Hello')
print(str.upper())
`,
		`string{"HELLO"}
`)
}

func TestStringCountMethod(t *testing.T) {
	mustParseSchemaRunScriptAssertOutput(t,
		`
	`,
		`mytypes`,
		`
str = datalark.String('Hello')
print(str.count('l'))
`,
		`int{2}
`)
}

func TestStringLenBuiltin(t *testing.T) {
	mustParseSchemaRunScriptAssertOutput(t,
		`
	`,
		`mytypes`,
		`
str = datalark.String('Hello')
print(len(str))
`,
		`5
`)
}
