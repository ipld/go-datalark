package datalarkengine

import (
	"testing"
)

// Test map construction using restructuring
func TestMapRestructuring(t *testing.T) {
	stdout, err := runScript(nil, "", `
m = datalark.Map(_={"a": "apple"})
print(m)
`)
	if err != nil {
		t.Fatal(err)
	}
	assertEqual(t, stdout, `map{
	string{"a"}: string{"apple"}
}
`)

	// map<string,int>
	stdout, err = runScript(nil, "", `
m = datalark.Map(_={"n": 123})
print(m)
`)
	if err != nil {
		t.Fatal(err)
	}
	assertEqual(t, stdout, `map{
	string{"n"}: int{123}
}
`)

}

// Test map construction using kwargs, resulting stringified value
// has deterministic key order
func TestMapKwargs(t *testing.T) {
	stdout, err := runScript(nil, "", `
m = datalark.Map(a="apple", b="banana")
print(m)
`)
	if err != nil {
		t.Fatal(err)
	}

	expect := `map{
	string{"a"}: string{"apple"}
	string{"b"}: string{"banana"}
}
`
	if stdout != expect {
		t.Errorf("unexpected output: %v", stdout)
	}
}

// Test list construction using positional args
func TestListPositional(t *testing.T) {
	stdout, err := runScript(nil, "", `
n = datalark.List(3, 4, 5)
print(n)
`)
	if err != nil {
		t.Fatal(err)
	}
	assertEqual(t, stdout, `list{
	0: int{3}
	1: int{4}
	2: int{5}
}
`)
}

// Test list construction using a nested list, does not do the same
// thing as restructuring
func TestListNestedList(t *testing.T) {
	stdout, err := runScript(nil, "", `
n = datalark.List([3, 4, 5])
print(n)
`)
	if err != nil {
		t.Fatal(err)
	}
	assertEqual(t, stdout, `list{
	0: list{
		0: int{3}
		1: int{4}
		2: int{5}
	}
}
`)
}

// Test list construction using restructuring
func TestListRestructuring(t *testing.T) {
	stdout, err := runScript(nil, "", `
n = datalark.List(_=[3, 4, 5])
print(n)
`)
	if err != nil {
		t.Fatal(err)
	}
	assertEqual(t, stdout, `list{
	0: int{3}
	1: int{4}
	2: int{5}
}
`)
}

// Test union construction using kwargs
func TestUnionKwargs(t *testing.T) {
	mustParseSchemaRunScriptAssertOutput(t,
		`
		type NameOrNum union {
			| String "name"
			| Int    "num"
		} representation keyed
	`,
		"mytypes",
		`
		print(mytypes.NameOrNum(String="Alice"))
	`, `union<NameOrNum>{string<String>{"Alice"}}
`)

	mustParseSchemaRunScriptAssertOutput(t,
		`
		type NameOrNum union {
			| String "name"
			| Int    "num"
		} representation keyed
	`,
		"mytypes",
		`
		print(mytypes.NameOrNum(Int=42))
	`, `union<NameOrNum>{int<Int>{42}}
`)
}

// Test union construction using positional arg
func TestUnionPositional(t *testing.T) {
	mustParseSchemaRunScriptAssertOutput(t,
		`
		type NameOrNum union {
			| String "name"
			| Int    "num"
		} representation keyed
	`,
		"mytypes",
		`
		print(mytypes.NameOrNum("Alice"))
	`, `union<NameOrNum>{string<String>{"Alice"}}
`)

	mustParseSchemaRunScriptAssertOutput(t,
		`
		type NameOrNum union {
			| String "name"
			| Int    "num"
		} representation keyed
	`,
		"mytypes",
		`
		print(mytypes.NameOrNum(42))
	`, `union<NameOrNum>{int<Int>{42}}
`)
}

// Test union construction using restructuring
func TestUnionRestructuring(t *testing.T) {
	mustParseSchemaRunScriptAssertOutput(t,
		`
		type NameOrNum union {
			| String "name"
			| Int    "num"
		} representation keyed
	`,
		"mytypes",
		`
		print(mytypes.NameOrNum(_={"String": "Bob"}))
	`, `union<NameOrNum>{string<String>{"Bob"}}
`)

	mustParseSchemaRunScriptAssertOutput(t,
		`
		type NameOrNum union {
			| String "name"
			| Int    "num"
		} representation keyed
	`,
		"mytypes",
		`
		print(mytypes.NameOrNum(_={"Int": 42}))
	`, `union<NameOrNum>{int<Int>{42}}
`)
}

// Test structs and unions can be built from string representations
func TestStructAndUnionReprs(t *testing.T) {
	mustParseSchemaRunScriptAssertOutput(t,
		`
		type Beta union {
			| Gamma "gamma:"
			| Delta "delta:"
		} representation stringprefix

		type Gamma string

		type Delta struct {
			x String
			y String
		} representation stringjoin {
			join ","
		}
	`,
		"mytypes",
		`
		print(mytypes.Delta("1,2"))
	`, `struct<Delta>{
	x: string<String>{"1"}
	y: string<String>{"2"}
}
`)

	mustParseSchemaRunScriptAssertOutput(t,
		`
		type Beta union {
			| Gamma "gamma:"
			| Delta "delta:"
		} representation stringprefix

		type Gamma string

		type Delta struct {
			x String
			y String
		} representation stringjoin {
			join ","
		}
	`,
		"mytypes",
		`
		print(mytypes.Beta("gamma:1,2"))
	`, `union<Beta>{string<Gamma>{"1,2"}}
`)

	mustParseSchemaRunScriptAssertOutput(t,
		`
		type Beta union {
			| Gamma "gamma:"
			| Delta "delta:"
		} representation stringprefix

		type Gamma string

		type Delta struct {
			x String
			y String
		} representation stringjoin {
			join ","
		}
	`,
		"mytypes",
		`
		print(mytypes.Beta("delta:1,2"))
	`, `
union<Beta>{struct<Delta>{
	x: string<String>{"1"}
	y: string<String>{"2"}
}}
`)

}

func TestStructTypedAndRepr(t *testing.T) {
	mustParseSchemaRunScriptAssertOutput(t,
		`
type Alpha struct {
    beta Beta (rename "b")
} representation map

type Beta string
`,
		"mytypes",
		// TODO: 4th case is not working, is `rename` implemented?
		`
		print(mytypes.Alpha("cat"))
		print(mytypes.Alpha(_={"beta": "meow"}))
		print(mytypes.Alpha.Typed(_={"Beta": "cat"}))
		#print(mytypes.Alpha.Repr(_={"b": "meow"}))
	`, `
struct<Alpha>{
	beta: string<Beta>{"cat"}
}
struct<Alpha>{
	beta: string<Beta>{"meow"}
}
struct<Alpha>{
	beta: string<Beta>{"cat"}
}
`)

}
