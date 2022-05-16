package datalarkengine

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/node/bindnode"
	"github.com/ipld/go-ipld-prime/schema"
)

func TestMapBasic(t *testing.T) {
	// test map<string,string>
	stdout, err := runScript(nil, "", `
		m = datalark.Map({"a": "apple"})
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
		m = datalark.Map({"a": 123})
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

// Test that a map in a struct gives the correct number of fields
func TestMapInStructFields(t *testing.T) {
	// Parse schema
	typesystem, err := ipld.LoadSchema("<noname>", strings.NewReader(`
		type LookupTable struct {
			name String
			data Map[String]String
		}
	`))
	if err != nil {
		panic(err)
	}

	// Sanity check the Type
	ourType := typesystem.TypeByName("LookupTable")
	assertEqual(t, ourType.Name(), "LookupTable")
	assertEqual(t, ourType.TypeKind().String(), "struct")

	// Validate that the number of fields matches what we exepct (2 fields)
	if ts, ok := ourType.(*schema.TypeStruct); ok {
		actualFields := fieldNames(ts)
		expectFields := []string{"name", "data"}
		// TODO(dustmop): This ends up with actualFields having 4 fields:
		// they are called "name", "data", "[", and "]"
		if diff := cmp.Diff(expectFields, actualFields); diff != "" {
			t.Errorf("fields mismatch (-want +got):\n%s", diff)
		}
	}

	type LookupTable struct {
		Name string
		Data map[string]string
	}

	// TODO(dustmop): This panics because the type believes it has 4 fields:
	// they are called "name", "data", "[", and "]"
	// So the check in go-ipld-prime/node/bindnode/infer.go in function
	// `verifyCompatibility` fails, that is `goType.NumField() != len(schemaFields)`
	bindings := []schema.TypedPrototype{
		bindnode.Prototype((*LookupTable)(nil), typesystem.TypeByName("LookupTable")),
	}
	_ = bindings

	// TODO(dustmop): Finish test
}

func fieldNames(ts *schema.TypeStruct) []string {
	fields := ts.Fields()
	result := []string{}
	for _, f := range fields {
		result = append(result, f.Name())
	}
	return result
}
