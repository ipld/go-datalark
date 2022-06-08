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
