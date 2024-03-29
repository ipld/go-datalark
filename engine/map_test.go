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
	// unmodified map, get keys, some found and some missing, print the map
	mustParseSchemaRunScriptAssertOutput(t,
		`
	`,
		`mytypes`,
		`
m = datalark.Map(_={'a': 'apple', 'b': 'banana'})
print(m.get('a'))
print(m.get('b'))
print(m.get('c'))
print(m)
`, `
string{"apple"}
string{"banana"}
None
map{
	string{"a"}: string{"apple"}
	string{"b"}: string{"banana"}
}
`)

	// replace a key and add a new key
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

	// delete a key
	mustParseSchemaRunScriptAssertOutput(t,
		`
	`,
		`mytypes`,
		`
m = datalark.Map(_={'a': 'apple', 'b': 'banana'})
m.pop('b')
print(m.get('a'))
print(m.get('b'))
print(m)
`, `
string{"apple"}
None
map{
	string{"a"}: string{"apple"}
}
`)

}

func TestMapAddAndDelete(t *testing.T) {
	// delete a key and then reassign it
	mustParseSchemaRunScriptAssertOutput(t,
		`
	`,
		`mytypes`,
		`
m = datalark.Map(_={'a': 'apple', 'b': 'banana', 'c': 'cherry'})
m.pop('b')
print(m.get('b'))
print(m)

n = datalark.Map(_={'a': 'apple', 'b': 'banana', 'c': 'cherry'})
n.pop('b')
n['b'] = 'berry'
print(n.get('b'))
print(n)
`, `
None
map{
	string{"a"}: string{"apple"}
	string{"c"}: string{"cherry"}
}
string{"berry"}
map{
	string{"a"}: string{"apple"}
	string{"b"}: string{"berry"}
	string{"c"}: string{"cherry"}
}
`)

	// replace a key and then delete it
	mustParseSchemaRunScriptAssertOutput(t,
		`
	`,
		`mytypes`,
		`
m = datalark.Map(_={'a': 'apple', 'b': 'banana', 'c': 'cherry'})
m['c'] = 'cantaloupe'
print(m.get('c'))
print(m)

n = datalark.Map(_={'a': 'apple', 'b': 'banana', 'c': 'cherry'})
n['c'] = 'cantaloupe'
n.pop('c')
print(n.get('c'))
print(n)
`, `
string{"cantaloupe"}
map{
	string{"a"}: string{"apple"}
	string{"b"}: string{"banana"}
	string{"c"}: string{"cantaloupe"}
}
None
map{
	string{"a"}: string{"apple"}
	string{"b"}: string{"banana"}
}
`)

	// add a new key and then delete it
	mustParseSchemaRunScriptAssertOutput(t,
		`
	`,
		`mytypes`,
		`
m = datalark.Map(_={'a': 'apple', 'b': 'banana', 'c': 'cherry'})
m['d'] = 'durian'
print(m.get('d'))
print(m)

n = datalark.Map(_={'a': 'apple', 'b': 'banana', 'c': 'cherry'})
n['d'] = 'durian'
n.pop('d')
print(n.get('d'))
print(n)
`, `
string{"durian"}
map{
	string{"a"}: string{"apple"}
	string{"b"}: string{"banana"}
	string{"c"}: string{"cherry"}
	string{"d"}: string{"durian"}
}
None
map{
	string{"a"}: string{"apple"}
	string{"b"}: string{"banana"}
	string{"c"}: string{"cherry"}
}
`)

}

func TestMethodItems(t *testing.T) {
	// method .items()
	mustParseSchemaRunScriptAssertOutput(t,
		`
	`,
		`mytypes`,
		`
m = datalark.Map(_={'a': 'apple', 'b': 'banana', 'c': 'cherry'})
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
	2: list{
		0: string{"c"}
		1: string{"cherry"}
	}
}
`)

	// method .items() with replacement node
	mustParseSchemaRunScriptAssertOutput(t,
		`
	`,
		`mytypes`,
		`
m = datalark.Map(_={'a': 'apple', 'b': 'banana', 'c': 'cherry'})
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

	// method .items() with added node
	mustParseSchemaRunScriptAssertOutput(t,
		`
	`,
		`mytypes`,
		`
m = datalark.Map(_={'a': 'apple', 'b': 'banana', 'c': 'cherry'})
m['d'] = 'durian'
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
	2: list{
		0: string{"c"}
		1: string{"cherry"}
	}
	3: list{
		0: string{"d"}
		1: string{"durian"}
	}
}
`)

	// method .items() with deleted node
	mustParseSchemaRunScriptAssertOutput(t,
		`
	`,
		`mytypes`,
		`
m = datalark.Map(_={'a': 'apple', 'b': 'banana', 'c': 'cherry'})
m.pop('a')
print(m.items())
`, `
list{
	0: list{
		0: string{"b"}
		1: string{"banana"}
	}
	1: list{
		0: string{"c"}
		1: string{"cherry"}
	}
}
`)

}

func TestMethodKeys(t *testing.T) {
	// method .keys()
	mustParseSchemaRunScriptAssertOutput(t,
		`
	`,
		`mytypes`,
		`
m = datalark.Map(_={'a': 'apple', 'b': 'banana', 'c': 'cherry'})
print(m.keys())
`, `
list{
	0: string{"a"}
	1: string{"b"}
	2: string{"c"}
}
`)

	// method .keys() with replacement node
	mustParseSchemaRunScriptAssertOutput(t,
		`
	`,
		`mytypes`,
		`
m = datalark.Map(_={'a': 'apple', 'b': 'banana', 'c': 'cherry'})
m['b'] = 'berry'
print(m.keys())
`, `
list{
	0: string{"a"}
	1: string{"b"}
	2: string{"c"}
}
`)

	// method .keys() with added node
	mustParseSchemaRunScriptAssertOutput(t,
		`
	`,
		`mytypes`,
		`
m = datalark.Map(_={'a': 'apple', 'b': 'banana', 'c': 'cherry'})
m['d'] = 'durian'
print(m.keys())
`, `
list{
	0: string{"a"}
	1: string{"b"}
	2: string{"c"}
	3: string{"d"}
}
`)

	// method .keys() with deleted node
	mustParseSchemaRunScriptAssertOutput(t,
		`
	`,
		`mytypes`,
		`
m = datalark.Map(_={'a': 'apple', 'b': 'banana', 'c': 'cherry'})
m.pop('a')
print(m.keys())
`, `
list{
	0: string{"b"}
	1: string{"c"}
}
`)
}

func TestMethodValues(t *testing.T) {
	// method .values()
	mustParseSchemaRunScriptAssertOutput(t,
		`
	`,
		`mytypes`,
		`
m = datalark.Map(_={'a': 'apple', 'b': 'banana', 'c': 'cherry'})
print(m.values())
`, `
list{
	0: string{"apple"}
	1: string{"banana"}
	2: string{"cherry"}
}
`)

	// method .values() with replacement node
	mustParseSchemaRunScriptAssertOutput(t,
		`
	`,
		`mytypes`,
		`
m = datalark.Map(_={'a': 'apple', 'b': 'banana', 'c': 'cherry'})
m['b'] = 'berry'
print(m.values())
`, `
list{
	0: string{"apple"}
	1: string{"berry"}
	2: string{"cherry"}
}
`)

	// method .values() with added node
	mustParseSchemaRunScriptAssertOutput(t,
		`
	`,
		`mytypes`,
		`
m = datalark.Map(_={'a': 'apple', 'b': 'banana', 'c': 'cherry'})
m['d'] = 'durian'
print(m.values())
`, `
list{
	0: string{"apple"}
	1: string{"banana"}
	2: string{"cherry"}
	3: string{"durian"}
}
`)

	// method .values() with deleted node
	mustParseSchemaRunScriptAssertOutput(t,
		`
	`,
		`mytypes`,
		`
m = datalark.Map(_={'a': 'apple', 'b': 'banana', 'c': 'cherry'})
m.pop('a')
print(m.values())
`, `
list{
	0: string{"banana"}
	1: string{"cherry"}
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

func TestDeterministic(t *testing.T) {
	mustParseSchemaRunScriptAssertOutput(t,
		`
	`,
		`mytypes`,
		`
m = datalark.Map(_={})
m['a'] = 'apple'
m['b'] = 'banana'
m['c'] = 'cherry'
m.setdefault('d', 'durian')
print(m)
`, `
map{
	string{"a"}: string{"apple"}
	string{"b"}: string{"banana"}
	string{"c"}: string{"cherry"}
	string{"d"}: string{"durian"}
}
`)
}

func TestMethodSetdefault(t *testing.T) {
	mustParseSchemaRunScriptAssertOutput(t,
		`
	`,
		`mytypes`,
		`
m = datalark.Map(_={'a': 'apple', 'b': 'banana', 'c': 'cherry'})
m['b'] = 'berry'
m.pop('c')
print(m.setdefault('a'))
print(m.setdefault('a', 'apricot'))
print(m.setdefault('b'))
print(m.setdefault('c', 'cantaloupe'))
print(m.setdefault('d', 'durian'))
print(m.setdefault('d'))
print(m.setdefault('e'))
print(m.setdefault('e', 'elderberry'))
m['f'] = 'fig'
m['g'] = 'grape'
print(m.setdefault('f'))
print(m.setdefault('g', 'guava'))
print(m)
`, `
string{"apple"}
string{"apple"}
string{"berry"}
string{"cantaloupe"}
string{"durian"}
string{"durian"}
null
null
string{"fig"}
string{"grape"}
map{
	string{"a"}: string{"apple"}
	string{"b"}: string{"berry"}
	string{"c"}: string{"cantaloupe"}
	string{"d"}: string{"durian"}
	string{"e"}: null
	string{"f"}: string{"fig"}
	string{"g"}: string{"grape"}
}
`)
}

func TestMethodUpdate(t *testing.T) {
	mustParseSchemaRunScriptAssertOutput(t,
		`
	`,
		`mytypes`,
		`
m = datalark.Map(_={'a': 'apple', 'b': 'banana', 'c': 'cherry'})
m.update({'b': 'berry', 'd': 'durian'})
print(m)
`, `
map{
	string{"a"}: string{"apple"}
	string{"b"}: string{"berry"}
	string{"c"}: string{"cherry"}
	string{"d"}: string{"durian"}
}
`)
}

func TestMethodCopy(t *testing.T) {
	mustParseSchemaRunScriptAssertOutput(t,
		`
	`,
		`mytypes`,
		`
m = datalark.Map(_={'a': 'apple', 'b': 'banana', 'c': 'cherry'})
m['b'] = 'berry'
n = m.copy()
m['c'] = 'cantaloupe'
print(n)
`, `
map{
	string{"a"}: string{"apple"}
	string{"b"}: string{"berry"}
	string{"c"}: string{"cherry"}
}
`)
}

func TestMethodFromKeys(t *testing.T) {
	mustParseSchemaRunScriptAssertOutput(t,
		`
	`,
		`mytypes`,
		`
m = datalark.Map(_={'a': 'apple'})
n = m.fromkeys(['b', 'c', 'd'], 'fruit')
print(n)
`, `
map{
	string{"b"}: string{"fruit"}
	string{"c"}: string{"fruit"}
	string{"d"}: string{"fruit"}
}
`)
}

func TestMethodPopitem(t *testing.T) {
	mustParseSchemaRunScriptAssertOutput(t,
		`
	`,
		`mytypes`,
		`
m = datalark.Map(_={'a': 'apple', 'b': 'banana', 'c': 'cherry'})
m.popitem()
print(m)
`, `
map{
	string{"a"}: string{"apple"}
	string{"b"}: string{"banana"}
}
`)

	mustParseSchemaRunScriptAssertOutput(t,
		`
	`,
		`mytypes`,
		`
m = datalark.Map(_={'a': 'apple', 'b': 'banana', 'c': 'cherry'})
m['d'] = 'durian'
m.popitem()
print(m)
`, `
map{
	string{"a"}: string{"apple"}
	string{"b"}: string{"banana"}
	string{"c"}: string{"cherry"}
}
`)

	mustParseSchemaRunScriptAssertOutput(t,
		`
	`,
		`mytypes`,
		`
m = datalark.Map(_={'a': 'apple', 'c': 'cherry'})
m['b'] = 'banana'
m.popitem()
print(m)
`, `
map{
	string{"a"}: string{"apple"}
	string{"c"}: string{"cherry"}
}
`)

}
