package datalarkengine

import (
	"github.com/ipld/go-ipld-prime/node/bindnode"
	"github.com/ipld/go-ipld-prime/schema"
)

func Example_structs() {
	ts := schema.MustTypeSystem(
		schema.SpawnString("String"),
		schema.SpawnStruct("FooBar", []schema.StructField{
			schema.SpawnStructField("foo", "String", false, false),
			schema.SpawnStructField("bar", "String", false, false),
		}, nil),
	)
	type FooBar struct{ Foo, Bar string }

	evalExample(`
		print(mytypes.FooBar)
		print(mytypes.FooBar(foo="hai", bar="wot"))
		x = {"foo": "z"}
		x["bar"] = "å!"
		print(mytypes.FooBar(**x))
	`, []schema.TypedPrototype{
		bindnode.Prototype((*FooBar)(nil), ts.TypeByName("FooBar")),
	})

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

	evalExample(`
		#print(mytypes.Map__FooBar__String({"f:b": "wot"})) # I want this to work someday, but it's not quite that magic yet.
		print(mytypes.Map__FooBar__String({mytypes.FooBar(foo="f", bar="b"): "wot"}))
	`, []schema.TypedPrototype{
		bindnode.Prototype((*FooBar)(nil), ts.TypeByName("FooBar")),
		bindnode.Prototype((*M)(nil), ts.TypeByName("Map__FooBar__String")),
	})

	// Output:
	// map<Map__FooBar__String>{
	// 	struct<FooBar>{foo: string<String>{"f"}, bar: string<String>{"b"}}: string<String>{"wot"}
	// }
}
