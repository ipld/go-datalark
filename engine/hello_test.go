package datalarkengine

import (
	"strings"

	"github.com/ipld/go-ipld-prime"
	"github.com/ipld/go-ipld-prime/node/bindnode"
	"github.com/ipld/go-ipld-prime/schema"
)

func Example_structs() {
	// Start with a schema.
	typesystem, err := ipld.LoadSchema("<noname>", strings.NewReader(`
		type FooBar struct {
			foo String
			bar String
		}
	`))
	if err != nil {
		panic(err)
	}

	// These are the golang types we'll bind it to.
	type FooBar struct{ Foo, Bar string }

	// These are the bindings we'll export to starlark.
	bindings := []schema.TypedPrototype{
		bindnode.Prototype((*FooBar)(nil), typesystem.TypeByName("FooBar")),
	}

	// Here's a script running on them:
	mustExecExample(nil, bindings, "mytypes", `
		print(mytypes.FooBar)
		print(mytypes.FooBar(foo="hai", bar="wot"))
		x = {"foo": "z"}
		x["bar"] = "å!"
		print(mytypes.FooBar(**x))
	`)

	// Output:
	// <built-in function datalark.Prototype<FooBar>>
	// struct<FooBar>{
	// 	foo: string<String>{"hai"}
	// 	bar: string<String>{"wot"}
	// }
	// struct<FooBar>{
	//	foo: string<String>{"z"}
	//	bar: string<String>{"å!"}
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

	mustExecExample(nil, []schema.TypedPrototype{
		bindnode.Prototype((*FooBar)(nil), ts.TypeByName("FooBar")),
		bindnode.Prototype((*M)(nil), ts.TypeByName("Map__FooBar__String")),
	},
		"mytypes",
		`
		#print(mytypes.Map__FooBar__String({"f:b": "wot"})) # I want this to work someday, but it's not quite that magic yet.
		print(mytypes.Map__FooBar__String({mytypes.FooBar(foo="f", bar="b"): "wot"}))
	`)

	// Output:
	// map<Map__FooBar__String>{
	// 	struct<FooBar>{foo: string<String>{"f"}, bar: string<String>{"b"}}: string<String>{"wot"}
	// }
}
