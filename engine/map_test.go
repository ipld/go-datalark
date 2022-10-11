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

func TestMapLookupWithMutation(t *testing.T) {
	mustParseSchemaRunScriptAssertOutput(t,
		`
	`,
		`mytypes`,
		`
m = datalark.Map(_={'a': 'apple', 'b': 'banana'})
m['a'] = 'apricot'
m['c'] = 'cherry'
print(m.get('a'))
print(m.get('b'))
print(m.get('c'))
`, `
string{"apricot"}
string{"banana"}
string{"cherry"}
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

	mustParseSchemaRunScriptAssertOutput(t,
		`
	`,
		`mytypes`,
		`
m = datalark.Map(_={'a': 'apple', 'b': 'banana'})
m['c'] = 'cherry'
m['b'] = 'berry'
print(m.items())
`, `
list{
	0: list{
		0: string{"a"}
		1: string{"apple"}
	}
	1: list{
		0: string{"b"}
		1: string{"berry"}
	}
	2: list{
		0: string{"c"}
		1: string{"cherry"}
	}
}
`)
}

func TestMethodKeys(t *testing.T) {
	mustParseSchemaRunScriptAssertOutput(t,
		`
	`,
		`mytypes`,
		`
m = datalark.Map(_={'a': 'apple', 'b': 'banana'})
print(m.keys())
`, `
list{
	0: string{"a"}
	1: string{"b"}
}
`)
}

func TestMethodValues(t *testing.T) {
	mustParseSchemaRunScriptAssertOutput(t,
		`
	`,
		`mytypes`,
		`
m = datalark.Map(_={'a': 'apple', 'b': 'banana'})
print(m.values())
`, `
list{
	0: string{"apple"}
	1: string{"banana"}
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

func TestMethodClear(t *testing.T) {
	mustParseSchemaRunScriptAssertOutput(t,
		`
	`,
		`mytypes`,
		`
m = datalark.Map(_={'a': 'apple', 'b': 'banana'})
m.clear()
print(m.values())
`, `
list{}
`)
}

func TestMethodGet(t *testing.T) {
	mustParseSchemaRunScriptAssertOutput(t,
		`
	`,
		`mytypes`,
		`
m = datalark.Map(_={'a': 'apple', 'b': 'banana'})
print(m.get('a'))
print(m.get('a', 'apricot'))
print(m.get('c'))
print(m.get('c', 'cherry'))
`, `
string{"apple"}
string{"apple"}
None
string{"cherry"}
`)

	mustParseSchemaRunScriptAssertOutput(t,
		`
	`,
		`mytypes`,
		`
m = datalark.Map(_={'a': 'apple', 'b': 'banana', 'c': 'cherry'})
m['a'] = 'apricot'
m.pop('c')
m['d'] = 'date'
print(m.get('a'))
print(m.get('b'))
print(m.get('c'))
print(m.get('d'))
`, `
string{"apricot"}
string{"banana"}
None
string{"date"}
`)
}
