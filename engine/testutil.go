package datalarkengine

import (
	"fmt"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/ipld/go-ipld-prime/schema"
	"go.starlark.net/starlark"

	"github.com/ipld/go-datalark/testutil"
)

// eval makes a skylark interpreter with a bunch of hardcoded default values,
// and runs your script in it.  It's for making tests brief to write,
// since in most tests we don't care to vary most of the ways you could customize the interpreter.
//
// Specifically:
// untyped constructors will be available under "datalark."
// and if you provide any typedPrototypes, those constructors will be under "mytypes.".
func eval(t *testing.T, script string, npts []schema.TypedPrototype) {
	t.Helper()
	err := eval_helper(script, npts)
	qt.Assert(t, err, qt.Equals, nil)
}

// like eval, but doesn't need a testing.T parameter, so you can use it in examples.
func evalExample(script string, npts []schema.TypedPrototype) {
	if err := eval_helper(script, npts); err != nil {
		panic(err)
	}
}

func eval_helper(script string, npts []schema.TypedPrototype) error {
	script = testutil.Dedent(script)

	globals := starlark.StringDict{}
	globals["datalark"] = ObjOfConstructorsForPrimitives()
	globals["mytypes"] = ObjOfConstructorsForPrototypes(npts...)
	// TODO consider adding an 'emit' function which appends to a buffer we'll test against.
	// Then we have print logging free for wtf'ing.

	thread := &starlark.Thread{
		Name: "thethreadname",
		Print: func(thread *starlark.Thread, msg string) {
			//caller := thread.CallFrame(1)
			//fmt.Printf("%s: %s: %s\n", caller.Pos, caller.Name, msg)
			fmt.Printf("%s\n", msg)
		},
	}

	_, err := starlark.ExecFile(thread, "thefilename.star", script, globals)
	return err
}
