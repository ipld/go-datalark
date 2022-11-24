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

[testmark]:# (hello-unions/script.various/create-by-kwargs)
```python
print(mytypes.FooOrBar(Foo="valueOfTypeFoo"))
```

(Note that that the keyword here is _capitalized_ "Foo"!
It's the type name!
This corresponds to how IPLD Schemas handle unions:
when interacting with a union value at the type level,
it's treated as a map, where the member type name is the key.)


### Creating unions with restructuring

Unions can be created using the restructuring style:

[testmark]:# (hello-unions/script.various/create-by-restructuring)
```python
print(mytypes.FooOrBar(_={"Foo":"valueOfTypeFoo"}))
```

This is functionally equivalent to the kwargs style (although it is slightly more general,
because kwargs may be limited by starlark's syntax rules for the kwarg string).

Both the above syntaxes produce the same result:

[testmark]:# (hello-unions/output)
```text
union<FooOrBar>{string<Foo>{"valueOfTypeFoo"}}
```


### Creating unions with positional arguments

Some unions can be constructed with positional arguments.
Here is a union whose members have distinct types:

[testmark]:# (positional-unions/schema)
```ipldsch
type NameOrNum union {
       | String string
       | Int    int
} representation kinded
```

This can be constructed using a positional arg:

[testmark]:# (positional-unions/script)
```text
print(mytypes.NameOrNum("value"))
```

Which produces this output:

[testmark]:# (positional-unions/output)
```text
union<NameOrNum>{string<String>{"value"}}
```

Note that this kind of usage only works for some unions, and depends on the union's type declaration!
In this case, it works because the union is a _kinded_ union... meaning we can look at positional argument,
and just from whether it's a number or a string, we can decide which of the union's member types it uses.


### Creating stringprefix union values

Stringjoin unions can be constructed by use of strings as a positional argument, too.

[testmark]:# (stringprefix-unions/schema)
```ipldsch
type String2 string
type FooOrBar union {
       | String  "a:"
       | String2 "b:"
} representation stringprefix
```

[testmark]:# (stringprefix-unions/script.various/create-by-string)
```text
print(mytypes.FooOrBar("b:zyx"))
```

They can also still be constructed using restructuring style and the type-level view,
which would look like this:

[testmark]:# (stringprefix-unions/script.various/create-by-restructuring)
```text
print(mytypes.FooOrBar(_={"String2":"zyx"}))
```

Both of these approaches to construction produce the same value:

[testmark]:# (stringprefix-unions/output)
```text
union<FooOrBar>{string<String2>{"zyx"}}
```


### Creating maps with complex keys

Sometimes union values can be created implicitly.
This occurs such as when a typed map has keys that are a union type!

This is well supported in a variety of ways
(including working fine via restructuring constructions, where the values have previously been held in a starlark dict, which didn't force types to occur).

Consider the following schema (which has the types "String" and "String2" just to make the point):

[testmark]:# (implicit-union-via-mapkey/schema)
```ipldsch
type String2 string
type FooOrBar union {
       | String  "a:"
       | String2 "b:"
} representation stringprefix
type FancyMap {FooOrBar:String}
```

A value of `FooOrBar` can be created in any of the following ways:

[testmark]:# (implicit-union-via-mapkey/script)
```text
print({mytypes.FooOrBar("b:zyx"): "heck"})
print("---")
print(mytypes.FancyMap(_={mytypes.FooOrBar("b:zyx"): "heck"}))
print("---")
print(mytypes.FancyMap(_={"b:zyx": "heck"}))
```

Notice in the third style there, no union constructor is visible at all!
It happens implicitly during the restructuring into the typed map,
which naturally forces typing onto the keys, as well.

The output of the above script looks like this:

[testmark]:# (implicit-union-via-mapkey/output)
```text
{union<FooOrBar>{string<String2>{"zyx"}}: "heck"}
---
map<FancyMap>{
	union<FooOrBar>{string<String2>{"zyx"}}: string<String>{"heck"}
}
---
map<FancyMap>{
	union<FooOrBar>{string<String2>{"zyx"}}: string<String>{"heck"}
}
```


### Deeper compositions with unions

We can create more and more complex schemas using unions somewhere inside,
and still have a great deal of flexibility in creating them,
including remarkably implicitly.

Consider this schema, where a struct contains a typed map that has a union type in the keys:

[testmark]:# (deeper-compositions-of-unions/schema)
```ipldsch
type String2 string
type FooOrBar union {
       | String  "a:"
       | String2 "b:"
} representation stringprefix
type Fluster struct {
	theMap {FooOrBar:String}
}
```

We can use a kwargs constructor for the struct, and the map can be a starlark dict
that's subjected to restructuring, and thus our union value is first seen as just a string:

[testmark]:# (deeper-compositions-of-unions/script)
```text
print(mytypes.Fluster(theMap={"b:zyx": "heck"}))
```

And that produces this value (note the union, and that its member is the `String2` type):

[testmark]:# (deeper-compositions-of-unions/output)
```text
struct<Fluster>{
	theMap: map<Map__FooOrBar__String>{
		union<FooOrBar>{string<String2>{"zyx"}}: string<String>{"heck"}
	}
}
```
