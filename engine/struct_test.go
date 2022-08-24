package datalarkengine

import (
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestStructs(t *testing.T) {
	mustParseSchemaRunScriptAssertOutput(t,
		`
		type FooBar struct {
			foo String
			bar String
		}
	`,
		"mytypes",
		`
		print(mytypes.FooBar(foo="one", bar="two"))
	`, `
		struct<FooBar>{
			foo: string<String>{"one"}
			bar: string<String>{"two"}
		}
	`)
}

func TestStructUnordered(t *testing.T) {
	mustParseSchemaRunScriptAssertOutput(t,
		`
		type FooBar struct {
			foo String
			bar String
		}
	`,
		"mytypes",
		`
		print(mytypes.FooBar(bar="two", foo="one"))
	`, `
		struct<FooBar>{
			foo: string<String>{"one"}
			bar: string<String>{"two"}
		}
	`)
}

func TestStructFieldAccess(t *testing.T) {
	mustParseSchemaRunScriptAssertOutput(t,
		`
		type FooBar struct {
			foo String
			bar String
		}
	`,
		"mytypes",
		`
		f = mytypes.FooBar(foo="one", bar="two")
		print(f.foo)
	`, `
		string<String>{"one"}
	`)
}

func TestStructWrongNumberOfFields1(t *testing.T) {
	defines := mustParseSchemaDefines(t,
		`
		type FooBar struct {
			foo String
			bar String
			baz String
		}
	`)
	_, err := runScript(defines,
		"mytypes",
		`
		f = mytypes.FooBar("one", "two")
		print(f)
	`)
	if err == nil {
		t.Fatalf("expected error, did not get one")
	}
	expectErr := `expected 3 values (foo,bar,baz), only got 2`
	qt.Assert(t, err.Error(), qt.Equals, expectErr)
}

func TestStructWrongNumberOfFields2(t *testing.T) {
	defines := mustParseSchemaDefines(t,
		`
		type Animals struct {
			cat String
			dog String
			eel String
		}
	`)
	_, err := runScript(defines,
		"mytypes",
		`
		f = mytypes.Animals(dog="bark", cat="meow")
		print(f)
	`)
	if err == nil {
		t.Fatalf("expected error, did not get one")
	}
	expectErr := `expected 3 values (cat,dog,eel), only got 2`
	qt.Assert(t, err.Error(), qt.Equals, expectErr)
}

func TestStructCorrectNumberOfFields2(t *testing.T) {
	defines := mustParseSchemaDefines(t,
		`
		type Animals struct {
			cat String
			dog String
			eel String
		}
	`)
	res, err := runScript(defines,
		"mytypes",
		`
		f = mytypes.Animals(dog="bark", eel="zap", cat="meow")
		print(f)
	`)
	if err != nil {
		t.Fatal(err)
	}
	expectOut := `struct<Animals>{
	cat: string<String>{"meow"}
	dog: string<String>{"bark"}
	eel: string<String>{"zap"}
}
`
	qt.Assert(t, res, qt.Equals, expectOut)
}

func TestEmptyConstructor(t *testing.T) {
	mustParseSchemaRunScriptAssertOutput(t,
		`
		type FooBar struct {
			foo optional String
		}
	`,
		"mytypes",
		`
		print(mytypes.FooBar())
	`, `
		struct<FooBar>{
			foo: absent
		}
	`)
}
