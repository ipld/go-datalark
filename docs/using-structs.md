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
print(mytypes.FooBar({
	"foo": "one",
	"bar": "two",
}))
```

Or creating structs with other more complex starlark syntaxes:

[testmark]:# (hello-structs/create/script.various/complex)
```python
x = {"foo": "z"}
x["bar"] = "å!"
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
