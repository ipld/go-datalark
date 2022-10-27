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

func TestListExtend(t *testing.T) {
	mustParseSchemaRunScriptAssertOutput(t,
		`
	`,
		`mytypes`,
		`
ls = datalark.List(_=['a', 'b', 'c'])
ls.extend(['d', 'e', 'f'])
print(len(ls))
print(ls)
`, `
6
list{
	0: string{"a"}
	1: string{"b"}
	2: string{"c"}
	3: string{"d"}
	4: string{"e"}
	5: string{"f"}
}
`)
}

func TestListMethodIndex(t *testing.T) {
	mustParseSchemaRunScriptAssertOutput(t,
		`
	`,
		`mytypes`,
		`
ls = datalark.List(_=['a', 'b', 'a'])
ls.append('a')
ls.append('d')
print(ls.index('a'))
print(ls.index('b'))
print(ls.index('c'))
print(ls.index('d'))
`, `
int{0}
int{1}
int{-1}
int{4}
`)
}

func TestListMethodInsert(t *testing.T) {
	mustParseSchemaRunScriptAssertOutput(t,
		`
	`,
		`mytypes`,
		`
ls = datalark.List(_=['a', 'b', 'c'])
ls.insert(1, 'x')
print(ls)
`, `
list{
	0: string{"a"}
	1: string{"x"}
	2: string{"b"}
	3: string{"c"}
}
`)

	mustParseSchemaRunScriptAssertOutput(t,
		`
	`,
		`mytypes`,
		`
ls = datalark.List(_=['a', 'b', 'c'])
ls.insert(3, 'x')
print(ls)
`, `
list{
	0: string{"a"}
	1: string{"b"}
	2: string{"c"}
	3: string{"x"}
}
`)

	mustParseSchemaRunScriptAssertOutput(t,
		`
	`,
		`mytypes`,
		`
ls = datalark.List(_=['a', 'b', 'c'])
ls.append('d')
ls.append('e')
ls.insert(4, 'x')
print(ls)
`, `
list{
	0: string{"a"}
	1: string{"b"}
	2: string{"c"}
	3: string{"d"}
	4: string{"x"}
	5: string{"e"}
}
`)

	mustParseSchemaRunScriptAssertOutput(t,
		`
	`,
		`mytypes`,
		`
ls = datalark.List(_=[])
ls.insert(0, 'x')
print(ls)
`, `
list{
	0: string{"x"}
}
`)

	mustParseSchemaRunScriptAssertOutput(t,
		`
	`,
		`mytypes`,
		`
ls = datalark.List(_=[])
ls.append('a')
ls.insert(0, 'x')
print(ls)
`, `
list{
	0: string{"x"}
	1: string{"a"}
}
`)

	mustParseSchemaRunScriptAssertOutput(t,
		`
	`,
		`mytypes`,
		`
ls = datalark.List(_=[])
ls.append('a')
ls.insert(1, 'x')
print(ls)
`, `
list{
	0: string{"a"}
	1: string{"x"}
}
`)
}
