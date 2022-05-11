package datalarkengine

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/node/bindnode"
	"github.com/ipld/go-ipld-prime/schema"
	"go.starlark.net/starlark"

	"github.com/ipld/go-datalark/testutil"
)

// assertSchemaAndScriptOutput parses a schema and runs a script, asserting that
// its output matches what is expected
func assertSchemaAndScriptOutput(t *testing.T, schemaText, globalName, script, expect string) {
	t.Helper()
	typesystem, err := ipld.LoadSchema("<noname>", strings.NewReader(schemaText))
	if err != nil {
		t.Fatal(err)
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
func mustExecExample(defines []schema.TypedPrototype, globalName, script string) {
	stdout, err := runScript(defines, globalName, script)
	if err != nil {
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
