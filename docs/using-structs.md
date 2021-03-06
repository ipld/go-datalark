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


Creating structs with positional arguments:

[testmark]:# (hello-structs/create/script.various/positional)
```python
print(mytypes.FooBar("one", "two"))
```


Creating structs with keyword args:

[testmark]:# (hello-structs/create/script.various/kwargs)
```python
print(mytypes.FooBar(foo="one", bar="two"))
```


When using keyword args, order does not matter:

[testmark]:# (hello-structs/create/script.various/kwargs-order)
```python
print(mytypes.FooBar(bar="two", foo="one"))
```


Creating structs by restructuring objects:

[testmark]:# (hello-structs/create/script.various/objliteral)
```python
print(mytypes.FooBar(_={
	"foo": "one",
	"bar": "two",
}))
```


Or creating structs by applying a dict into keyword args:

[testmark]:# (hello-structs/create/script.various/apply-dict)
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

Using Struct Values
----------------------

Struct values can access their fields using the familiar doc notation used by regular python code.

[testmark]:# (access-structs/schema)
```ipldsch
type FooBar struct {
	foo String
	bar String
}
```

[testmark]:# (access-structs/access/script.various/use-field)
```python
obj = mytypes.FooBar(foo='abc', bar='def')
print(obj.foo)
```

As expected, this outputs the value of the struct field

[testmark]:# (access-structs/access/output)
```text
string<String>{"abc"}
```