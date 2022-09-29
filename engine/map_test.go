package datalarkengine

import (
	"testing"
)

func TestMapAndLookup(t *testing.T) {
	mustParseSchemaRunScriptAssertOutput(t,
		`
		type Lookup {String:String}
	`,
		"mytypes",
		`
		fruits = mytypes.Lookup(apple='red')
		print(fruits)
		print(fruits['apple'])
	`, `
		map<Lookup>{
			string<String>{"apple"}: string<String>{"red"}
		}
		string<String>{"red"}
`)

	mustParseSchemaRunScriptAssertOutput(t,
		`
		type Lookup {String:String}
	`,
		"mytypes",
		`
		fruits = mytypes.Lookup(_={'banana': 'yellow'})
		print(fruits)
		print(fruits['banana'])
	`, `
		map<Lookup>{
			string<String>{"banana"}: string<String>{"yellow"}
		}
		string<String>{"yellow"}
`)
}

func TestMethodItems(t *testing.T) {
	mustParseSchemaRunScriptAssertOutput(t,
		`
	`,
		`mytypes`,
		`
m = datalark.Map(_={'a': 'apple', 'b': 'banana'})
print(m.items())
`, `
list{
	0: list{
		0: string{"a"}
		1: string{"apple"}
	}
	1: list{
		0: string{"b"}
		1: string{"banana"}
	}
}
`)
}

func TestMapAssign(t *testing.T) {
	mustParseSchemaRunScriptAssertOutput(t,
		`
	`,
		`mytypes`,
		`
m = datalark.Map(_={'a': 'apple', 'b': 'banana'})
m['c'] = 'cherry'
print(len(m))
print(m)
print(len(m))
m['b'] = 'berry'
print(len(m))
print(m)
print(len(m))
`, `
3
map{
	string{"a"}: string{"apple"}
	string{"b"}: string{"banana"}
	string{"c"}: string{"cherry"}
}
3
3
map{
	string{"a"}: string{"apple"}
	string{"b"}: string{"berry"}
	string{"c"}: string{"cherry"}
}
3
`)
}
