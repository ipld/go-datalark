Using Structs with Datalark
===========================

Struct types are defined in IPLD Schemas, like this:

[testmark]:# (hello-structs/schema)
```ipldsch
type FooBar struct {
	foo String
	bar String
}
```

In Datalark, you can construct and interact with structs in many ways.


Creating Struct Values
----------------------

### Creating simple structs

First, let's look at creating simple structs, like the "FooBar" struct defined earlier.

Creating structs with keyword args:

[testmark]:# (hello-structs/create/script.various/kwargs)
```python
print(mytypes.FooBar(foo="one", bar="two"))
```

Creating structs with object literals:

[testmark]:# (hello-structs/create/script.various/objliteral)
```python
print(mytypes.FooBar(_={
	"foo": "one",
	"bar": "two",
}))
```

Or creating structs with other more complex starlark syntaxes:

[testmark]:# (hello-structs/create/script.various/complex)
```python
x = {"foo": "one"}
x["bar"] = "two"
print(mytypes.FooBar(**x))
```

All of the above syntaxes do the same things,
and so those print calls will emit the same result:

[testmark]:# (hello-structs/create/output)
```text
struct<FooBar>{
	foo: string<String>{"one"}
	bar: string<String>{"two"}
}
```

### Creating nested structs

Let's look at what kind of syntax is available for creating nested structures.

We'll need some (slightly) more complex types:

[testmark]:# (nested-structs/schema)
```ipldsch
type Frob struct {
	foo String
	bar String
	baz Baz
}
type Baz struct {
	bop String
}
```

Using another plain starlark object still works.
In fact, it works even if the restructuring has to go deep --
see how this creates the deeper struct out of a nested literal:

[testmark]:# (nested-structs/create/script.various/objliteral)
```python
print(mytypes.Frob(_={
	"foo": "oof",
	"bar": "rab",
	"baz": {"bop": "pob"},
}))
```

Of course, you can also assemble things one at a time,
constructing each type yourself, and composing them:

[testmark]:# (nested-structs/create/script.various/steps)
```python
x = mytypes.Baz(bop="pob")
y = mytypes.Frob(_={"foo":"oof", "bar":"rab", "baz":x})
print(y)
```

The result, either way you do it, is this:

[testmark]:# (nested-structs/create/output)
```text
struct<Frob>{
	foo: string<String>{"oof"}
	bar: string<String>{"rab"}
	baz: struct<Baz>{
		bop: string<String>{"pob"}
	}
}
```
