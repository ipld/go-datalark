package datalarkengine_test

import (
	"fmt"

	"github.com/ipld/go-ipld-prime/node/bindnode"
	"github.com/ipld/go-ipld-prime/schema"
	"go.starlark.net/starlark"

	"github.com/ipld/go-datalark"
)

func eval(src string, tsname string, npts []schema.TypedPrototype) {
	globals := starlark.StringDict{}
	datalark.InjectGlobals(globals, datalark.ObjOfConstructorsForPrimitives())
	globals[tsname] = datalark.ObjOfConstructorsForPrototypes(npts...)

	thread := &starlark.Thread{
		Name: "thethreadname",
		Print: func(thread *starlark.Thread, msg string) {
			//caller := thread.CallFrame(1)
			//fmt.Printf("%s: %s: %s\n", caller.Pos, caller.Name, msg)
			fmt.Printf("%s\n", msg)
		},
	}

	_, err := starlark.ExecFile(thread, "thefilename.star", src, globals)
	if err != nil {
		panic(err)
	}
}

func Example_hello() {
	eval(`
print(String)
print(String("yo"))
x = {"bz": "zoo"}
print(Map(hey="hai", zonk="wot", **x))
print(Map({String("fun"): "heeey"}))
`, "", nil)

	// Output:
	// <built-in function datalark.Prototype>
	// string{"yo"}
	// map{
	// 	string{"hey"}: string{"hai"}
	// 	string{"zonk"}: string{"wot"}
	// 	string{"bz"}: string{"zoo"}
	// }
	// map{
	// 	string{"fun"}: string{"heeey"}
	// }
}

func Example_structs() {
	ts := schema.MustTypeSystem(
		schema.SpawnString("String"),
		schema.SpawnStruct("FooBar", []schema.StructField{
			schema.SpawnStructField("foo", "String", false, false),
			schema.SpawnStructField("bar", "String", false, false),
		}, nil),
	)
	type FooBar struct{ Foo, Bar string }

	eval(`
#print(dir(ts))
print(ts.FooBar)
print(ts.FooBar(foo="hai", bar="wot"))
`, "ts", []schema.TypedPrototype{
		bindnode.Prototype((*FooBar)(nil), ts.TypeByName("FooBar")),
	})

	// Output:
	// <built-in function datalark.Prototype<FooBar>>
	// struct<FooBar>{
	// 	foo: string<String>{"hai"}
	// 	bar: string<String>{"wot"}
	// }
}

func Example_mapWithStructKeys() {
	ts := schema.MustTypeSystem(
		schema.SpawnString("String"),
		schema.SpawnStruct("FooBar", []schema.StructField{
			schema.SpawnStructField("foo", "String", false, false),
			schema.SpawnStructField("bar", "String", false, false),
		}, schema.SpawnStructRepresentationStringjoin(":")),
		schema.SpawnMap("Map__FooBar__String", "FooBar", "String", false),
	)
	type FooBar struct{ Foo, Bar string }
	type M struct {
		Keys   []FooBar
		Values map[FooBar]string
	}

	eval(`
#print(ts.Map__FooBar__String({"f:b": "wot"})) # I want this to work someday, but it's not quite that magic yet.
print(ts.Map__FooBar__String({ts.FooBar(foo="f", bar="b"): "wot"}))
`, "ts", []schema.TypedPrototype{
		bindnode.Prototype((*FooBar)(nil), ts.TypeByName("FooBar")),
		bindnode.Prototype((*M)(nil), ts.TypeByName("Map__FooBar__String")),
	})

	// Output:
	// map<Map__FooBar__String>{
	// 	struct<FooBar>{foo: string<String>{"f"}, bar: string<String>{"b"}}: string<String>{"wot"}
	// }
}
