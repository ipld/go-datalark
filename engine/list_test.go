package datalarkengine

import (
	"testing"
)

func TestListAppend(t *testing.T) {
	// create list and index it to get individual elements
	mustParseSchemaRunScriptAssertOutput(t,
		`
	`,
		`mytypes`,
		`
ls = datalark.List(_=['a', 'b', 'c'])
print(ls[0])
print(ls[1])
print(ls[2])
print(len(ls))
print(ls)
`, `
string{"a"}
string{"b"}
string{"c"}
3
list{
	0: string{"a"}
	1: string{"b"}
	2: string{"c"}
}
`)

	// append to list, length and indexing and printing all work
	mustParseSchemaRunScriptAssertOutput(t,
		`
	`,
		`mytypes`,
		`
ls = datalark.List(_=['a', 'b'])
ls.append('c')
ls.append('d')
print(ls[0])
print(ls[1])
print(ls[2])
print(ls[3])
print(len(ls))
print(ls)
`, `
string{"a"}
string{"b"}
string{"c"}
string{"d"}
4
list{
	0: string{"a"}
	1: string{"b"}
	2: string{"c"}
	3: string{"d"}
}
`)

}

func TestListMethodClear(t *testing.T) {
	mustParseSchemaRunScriptAssertOutput(t,
		`
	`,
		`mytypes`,
		`
ls = datalark.List(_=['a', 'b'])
print(len(ls))
ls.clear()
print(len(ls))
print(ls)
`, `
2
0
list{}
`)
}

func TestListMethodCopy(t *testing.T) {
	mustParseSchemaRunScriptAssertOutput(t,
		`
	`,
		`mytypes`,
		`
ls = datalark.List(_=['a', 'b'])
ls.append('c')
cs = ls.copy()
ls.append('d')
print(cs)
print(ls)
`, `
list{
	0: string{"a"}
	1: string{"b"}
	2: string{"c"}
}
list{
	0: string{"a"}
	1: string{"b"}
	2: string{"c"}
	3: string{"d"}
}
`)
}

func TestListMethodCount(t *testing.T) {
	mustParseSchemaRunScriptAssertOutput(t,
		`
	`,
		`mytypes`,
		`
ls = datalark.List(_=['a', 'b', 'a'])
print(ls.count('a'))
print(ls.count('b'))
print(ls.count('c'))
print(ls.count('d'))
ls.append('a')
ls.append('d')
print(ls.count('a'))
print(ls.count('b'))
print(ls.count('c'))
print(ls.count('d'))
`, `
int{2}
int{1}
int{0}
int{0}
int{3}
int{1}
int{0}
int{1}
`)
}
