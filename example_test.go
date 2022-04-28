package datalark_test

import (
	"fmt"

	"github.com/ipld/go-ipld-prime/node/bindnode"
	ipld "github.com/ipld/go-ipld-prime/schema"
	"go.starlark.net/starlark"

	"github.com/ipld/go-datalark"
	"github.com/ipld/go-datalark/testutil"
)

func Example_hello() {
	// Prepare things needed by a starlark interpreter.  (This is Starlark boilerplate!)
	thread := &starlark.Thread{
		Name: "thethreadname",
		Print: func(thread *starlark.Thread, msg string) {
			fmt.Printf("%s\n", msg)
		},
	}

	// Use datalark to make IPLD value constructors available to Starlark!
	globals := starlark.StringDict{}
	globals["datalark"] = datalark.ObjOfConstructorsForPrimitives()

	// Now here's our demo script:
	script := testutil.Dedent(`
		print(datalark.String("yo"))
	`)

	// Invoke the starlark interpreter!
	_, err := starlark.ExecFile(thread, "thefilename.star", script, globals)
	if err != nil {
		panic(err)
	}

	// Output:
	// string{"yo"}
}

func Example_helloTypes() {
	// In this example we'll use an IPLD Schema!
	typesystem, err := ipld.LoadSchema("<noname>", strings.NewReader(`
		type FooBar struct {
			foo String
			bar String
		}
	`))
	if err != nil {
		panic(err)
	}

	// And we'll bind it to this golang native type:
	type FooBar struct{ Foo, Bar string }

	// Prepare things needed by a starlark interpreter.  (This is Starlark boilerplate!)
	thread := &starlark.Thread{
		Name: "thethreadname",
		Print: func(thread *starlark.Thread, msg string) {
			fmt.Printf("%s\n", msg)
		},
	}

	// Use datalark to make IPLD value constructors available to Starlark!
	globals := starlark.StringDict{}
	globals["datalark"] = datalark.ObjOfConstructorsForPrimitives()
	globals["mytypes"] = datalark.ObjOfConstructorsForPrototypes(
		bindnode.Prototype((*FooBar)(nil), ts.TypeByName("FooBar")),
	)

	// Now here's our demo script:
	script := testutil.Dedent(`
		print(mytypes.FooBar)
		print(mytypes.FooBar(foo="helloooo", bar="world!"))
	`)

	// Invoke the starlark interpreter!
	_, err := starlark.ExecFile(thread, "thefilename.star", script, globals)
	if err != nil {
		panic(err)
	}

	// Output:
	// <built-in function datalark.Prototype<FooBar>>
	// struct<FooBar>{
	// 	foo: string<String>{"helloooo"}
	// 	bar: string<String>{"world!"}
	// }
}

func Example_helloGlobals() {
	// In this example, we do similar things to the other examples,
	// except we put our functions directly into the global namespace.
	// You may wish to do this to make it even easier to use
	// (but remember to weigh it against cluttering the namespace your users will experience!).

	// Prepare things needed by a starlark interpreter.  (This is Starlark boilerplate!)
	thread := &starlark.Thread{
		Name: "thethreadname",
		Print: func(thread *starlark.Thread, msg string) {
			fmt.Printf("%s\n", msg)
		},
	}

	// Use datalark to make IPLD value constructors available to Starlark!
	// Note the use of 'InjectGlobals' here -- this puts things into scope without any namespace,
	// as opposed to what we did in other examples, which let you choose a name in the globals to put everything under.
	globals := starlark.StringDict{}
	datalark.InjectGlobals(globals, datalark.ObjOfConstructorsForPrimitives())

	// Now here's our demo script:
	script := testutil.Dedent(`
		print(String("yo")) # look, no 'datalark.' prefix!
	`)

	// Invoke the starlark interpreter!
	_, err := starlark.ExecFile(thread, "thefilename.star", script, globals)
	if err != nil {
		panic(err)
	}

	// Output:
	// string{"yo"}
}
