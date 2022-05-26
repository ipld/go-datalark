Using Unions with Datalark
==========================

Union types are defined in IPLD Schemas, like this:

[testmark]:# (hello-unions/schema)
```ipldsch
type FooOrBar union {
       | Foo "foo"
       | Bar "bar"
} representation keyed

type Foo string
type Bar string
```

(This is only one example of a union type.
There are many different union representation strategies for unions;
See https://ipld.io/docs/schemas/ .)

In Datalark, you can construct and interact with unions in many ways.


Creating Union Values
---------------------

### Creating unions with kwargs

Unions can be created with kwargs:

[testmark]:# (hello-unions/hello-unions/create/script.various/kwargs)
```python
print(mytypes.FooOrBar(Foo="valueOfTypeFoo"))
```

(Note that a capitalized "Foo" is the kwarg -- because in unions, the member type name is considered the key,
when interacting with the union value at the type level.)

### Creating unions with restructuring

TODO: not yet supported.  Should look like the following:

[testmark]:# (hello-unions/hello-unions/create/script.various/restructuring)
```python
print(mytypes.FooOrBar(_={"Foo":"valueOfTypeFoo"}))
```

This is functionally equivalent to the kwargs style (although it is slightly more general,
because kwargs may be limited by starlark's syntax rules for the kwarg string).

### Simple union result

All the above syntaxes produce the same result:

[testmark]:# (hello-unions/hello-unions/create/output)
```text
union<FooOrBar>{string<Foo>{"valueOfTypeFoo"}}
```
