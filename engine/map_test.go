package datalarkengine

import (
	"testing"
)

func TestMapBasic(t *testing.T) {
	// test map<string,string>
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

	// test map<string,int>
	stdout, err = runScript(nil, "", `
m = datalark.Map(_={"a": 123})
print(m)
`)
	if err != nil {
		t.Fatal(err)
	}
	assertEqual(t, stdout, `map{
	string{"a"}: int{123}
}
`)

}
