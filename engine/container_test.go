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

// Test map construction using kwargs
func TestMapKwargs(t *testing.T) {
	stdout, err := runScript(nil, "", `
m = datalark.Map(a="apple", b="banana")
print(m)
`)
	if err != nil {
		t.Fatal(err)
	}
	assertEqual(t, stdout, `map{
	string{"a"}: string{"apple"}
	string{"b"}: string{"banana"}
}
`)
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
