package datalarkengine

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/node/bindnode"
	"github.com/ipld/go-ipld-prime/schema"
	"go.starlark.net/starlark"

	"github.com/ipld/go-datalark/testutil"
)

// Style note: parameters to these functions are meant to read somewhat narratively:
//  generally, things come in order of what would be set up first.
//   So: schema or prototypes first, then the script (which will refer to them), and the expected output last.
//  (This gets less obvious, and more important to have a convention for, as things get stringier, such as in evalWithUltramagic!)

// eval makes a skylark interpreter with a bunch of hardcoded default values,
// and runs your script in it.  It's for making tests brief to write,
// since in most tests we don't care to vary most of the ways you could customize the interpreter.
//
// Specifically:
// untyped constructors will be available under "datalark."
// and if you provide any typedPrototypes, those constructors will be under "mytypes.".
func eval(t *testing.T, npts []schema.TypedPrototype, script string, expect string) {
	t.Helper()
	var buf bytes.Buffer
	err := eval_helper(&buf, npts, script)
	qt.Assert(t, err, qt.Equals, nil)
	qt.Assert(t, buf.String(), qt.Equals, testutil.Dedent(expect))
}

// evalWithMagic is like eval, but takes a whole schema and makes prototypes for you,
// using bindnode (which uses the magic of reflect.StructOf) rather than needing handrolled types.
func evalWithMagic(t *testing.T, ts *schema.TypeSystem, script string, expect string) {
	t.Helper()
	var npts []schema.TypedPrototype
	for _, typeInfo := range ts.GetTypes() {
		npts = append(npts, bindnode.Prototype(nil, typeInfo))
	}
	eval(t, npts, script, expect)
}

// evalWithUltramagic is like evalWithMagic, but also parses schema DSL for you.
func evalWithUltramagic(t *testing.T, schemaDsl string, script string, expect string) {
	typesystem, err := ipld.LoadSchema("<noname>", strings.NewReader(schemaDsl))
	if err != nil {
		t.Fatal(err)
	}
	t.Helper()
	evalWithMagic(t, typesystem, script, expect)
}

// like eval, but doesn't need a testing.T parameter,
// and just prints output rather than asserting on it,
// so you can use it in examples.
func evalExample(npts []schema.TypedPrototype, script string) {
	if err := eval_helper(os.Stdout, npts, script); err != nil {
		panic(err)
	}
}

func eval_helper(output io.Writer, npts []schema.TypedPrototype, script string) error {
	script = testutil.Dedent(script)

	globals := starlark.StringDict{}
	globals["datalark"] = ObjOfConstructorsForPrimitives()
	globals["mytypes"] = ObjOfConstructorsForPrototypes(npts...)

	thread := &starlark.Thread{
		Name: "thethreadname",
		Print: func(thread *starlark.Thread, msg string) {
			//caller := thread.CallFrame(1)
			//fmt.Printf("%s: %s: %s\n", caller.Pos, caller.Name, msg)
			fmt.Fprintf(output, "%s\n", msg)
		},
	}

	_, err := starlark.ExecFile(thread, "thefilename.star", script, globals)
	return err
}
