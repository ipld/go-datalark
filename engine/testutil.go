package datalarkengine

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/ipfs/go-cid"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/datamodel"
	cidlink "github.com/ipld/go-ipld-prime/linking/cid"
	"github.com/ipld/go-ipld-prime/node/bindnode"
	"github.com/ipld/go-ipld-prime/schema"
	"go.starlark.net/starlark"

	"github.com/ipld/go-datalark/testutil"
)

// mustParseSchemaRunScriptAssertOutput parses a schema and runs a
// script, asserting that its output matches what is expected
func mustParseSchemaRunScriptAssertOutput(t *testing.T, schemaText, globalName, script, expect string) {
	if t != nil {
		t.Helper()
	}
	typesystem, err := ipld.LoadSchema("<noname>", strings.NewReader(schemaText))
	if err != nil {
		if t != nil {
			t.Fatal(err)
		}
		panic(err)
	}
	var defines []schema.TypedPrototype
	for _, typeInfo := range typesystem.GetTypes() {
		defines = append(defines, bindnode.Prototype(nil, typeInfo))
	}
	assertScriptOutput(t, defines, globalName, script, expect)
}

// assertScriptOutput evaluates a script with the given defintions bound to the
// name "mytypes", asserting that the output matches what is expected
func assertScriptOutput(t *testing.T, defines []schema.TypedPrototype, globalName, script, expect string) {
	t.Helper()

	output, err := runScript(defines, globalName, script)
	qt.Assert(t, err, qt.Equals, nil)
	qt.Assert(t, output, qt.Equals, testutil.Dedent(expect))
}

// mustExecExample evaluates the script with the given definitions bound to the
// given global name, and writes the output to stdout. Panics if an error occurs
func mustExecExample(t *testing.T, defines []schema.TypedPrototype, globalName, script string) {
	stdout, err := runScript(defines, globalName, script)
	if err != nil {
		if t != nil {
			t.Fatal(err)
		}
		panic(err)
	}
	fmt.Printf("%s", stdout)
}

// runScript evaluates the script with the given definitions bound to the given
// global name, and returns the output and error
func runScript(defines []schema.TypedPrototype, globalName, script string) (string, error) {
	var buf bytes.Buffer

	script = testutil.Dedent(script)

	globals := starlark.StringDict{}
	globals["datalark"] = PrimitiveConstructors()
	globals[globalName] = MakeConstructors(defines)

	thread := &starlark.Thread{
		Name: "thethreadname",
		Print: func(thread *starlark.Thread, msg string) {
			fmt.Fprintf(&buf, "%s\n", msg)
		},
	}

	_, err := starlark.ExecFile(thread, "thefilename.star", script, globals)
	return buf.String(), err
}

func mustRunScript(t *testing.T, defines []schema.TypedPrototype, globalName, script string) string {
	content, err := runScript(defines, globalName, script)
	if err != nil {
		if t != nil {
			t.Fatal(err)
		}
		panic(t)
	}
	return content
}

func newTestLink() datamodel.Link {
	// Example link from:
	// https://github.com/ipld/go-ipld-prime/blob/master/datamodel/equal_test.go
	someCid, _ := cid.Cast([]byte{1, 85, 0, 5, 0, 1, 2, 3, 4})
	return cidlink.Link{Cid: someCid}
}
