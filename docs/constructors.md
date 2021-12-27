Datalark Constructor Functions
==============================

"Constructor functions" refers to functions that Datalark provides to the Starlark environment
which user code can then use to create [Datalark Values](./datalark-values.md).

Some constructors are simple -- there's a `String` function which, well, returns a string.

Some constructors are more complex -- especially the ones that are created to handle the types that *you* provide!



Constructor Conventions
-----------------------

Constructor functions have to figure out *two* things from the arguments that you give them:

- if the arguments are meant to be type-level views of data or representation-level views of the data;
- and then how to map that data into the type that's being constructed.

All of this has to be figured out from information which can come in from:

- positional arguments...
- or "kwargs"...
- or entire starlark dicts and lists as positional arguments...
- where data is sometimes typed already, and sometimes not...
- and a mixture of all of the above!
- Plus variations in the constructor function itself (sometimes there's more than one, for the same type)!

Sometimes how all this works out this is obvious.
Sometimes it's less so.

Consider the following examples, to see where the distinctions might be important:

#### Example 1

Given some types:

```ipldsch
type Foobar struct {
	foo Foo (rename "bar")
	bar Bar (rename "foo")
} representation map

type Foo string
type Bar string
```

And then a datalark constructor call:

```python
Foobar(foo="ooo", bar="aarrr")
Foobar(_={"foo":"ooo", "bar":"aarrr"})
```

Which data should result?
(Are the constructors going to prefer the representation, with its confusing renames of the fields?  Or use the type-level names of the fields?)

And are these two constructions even the same?

#### Example 2

Given some types:

```ipldsch
type FooOrBar union {
	| Foo "foo:"
	| Bar "bar:"
} representation stringprefix

type Foo string
type Bar string
```

And then a datalark constructor call:

```python
FooOrBar("foo:ooo")
FooOrBar(Foo="ooo")
```

Presumably both these constructions are equivalent.
But how does Datalark know how to parse these?

### The Decision Tree For Mode

Generally, constructors act with type-level behaviors (and field names) first,
and bend to representation-level behaviors if it seems appropriate.

The actual rules are a little more complex:

1. The explicit mode is used, if applicable.
2. The data kind indicates whether to use type-level or representation-level mode, if the typekind's kind and the representation strategy's kind are distinct.
3. The prevaling mode is used, if applicable.
4. The type-level mode is used.
5. If (and only if) the global "DWIM" mode is enabled for this module,
  and using the type-level mode failed because map keys didn't match, but would match for the representation mode,
  then the representation mode is tried.

"Explicit mode" means if you've said something like `FooBar.Repr(...)` or `FooBar.Typed(...)` instead of just `FooBar(...)`.
In these cases, of course, Datalark listens to you.

Data kind kicks in when it can, but can only be decisive in some situations.
For example, for unions with a representation strategy such as stringprefix,
it's perfectly clear that if you give the constructor a string, then that is representation-mode data.
In other scenarios -- such as if an argument is a map, and the type we're constructing is a struct with a map representation --
the data kind alone isn't enough information be distinctive, so we continue to the next rule.

"Prevailing mode" describes the situation of being in the middle of transforming a whole tree of data:
in that situation, one of these rules already applied at the base of the tree,
and we can just continue using that choice as we go deeper.

If none of those earlier rules was able to make the pick,
then we default assuming we should be working in terms of type-level information.

The fallback to representation mode, if the type-level structure didn't match the arguments,
isn't checked at all unless you ask for it, but is available as a last resort _if_ you enable it.
(This is off by default, because checking for it is expensive, and sometimes it's ambiguous.)
(TODO: not implemented yet; and we may review if it can be on by default when it is.)

### The Decision Tree for Positional vs Kwargs vs Restructuring

#### Positional Args

Simple scalars (ints, strings, etc) are constructed with single positional arguments.

Struct types can be constructed with positional arguments.
Each argument becomes the next field in the struct, per the type definition.
The number of arguments must match the number of fields in the struct,
except trailing optional fields may be omitted.
(However, you have several options when it comes to structs:
it may also be worth considering the use of kwargs, because those can be easier to read.)

Map types can be constructed with positional arguments.
Each argument in this case must be a map or starlark dict itself
(or interpretable as one -- so a struct can be used, for example!).

List types can be constructed with positional arguments.
Each argument becomes a member of the list.

Unions are (sometimes) allowed to use single positional arguments, too,
as a little bit of a shortcut:
for a union with a representation strategy with a string kind (e.g. like `stringprefix`),
we treat `FooOrBar("foo:ooo")` as valid and equivalent to `FooOrBar(_="foo:ooo")`.
(We do this because it's unambiguous (see Rule 2 in the section about Decision Tree for Mode),
and saves you a few keystrokes.)

#### Kwargs

Kwargs (short for "keyword args") refers to the starlark syntax of `somefunc(keyword=arg)`.

Struct types can be constructed with kwargs.
The keyword should match the field names in the struct's type definition.

Kwargs also (sometimes) work for creating a struct from its representation:
for example, for some type `type Foo struct { bar String (rename "b") }`,
the constructor `Foo.Repr(b: "bar")` is acceptable.
(Of course, rename directives can also use strings that aren't valid as kwargs in Starlark,
so this feature is only valid when such edge cases are avoided.)

Maps can be created with kwargs; `Map(foo: "bar")` is roughly equivalent to `{"foo": "bar"}`.

Kwargs should not be used if the values may contain underscores.
Kwargs keywords starting with an underscore are reserved;
Datalark sometimes uses them to engage other features.

#### Restructuring Args

Restructuring args is a convention we introduce in Datalark, which looks like `somefunc(_=arg)`.
(It's kwargs, where the keyword is the underscore character.)

"Restructuring" refers to taking some Starlark value --
whether it be something simple like a string, or something complex like nested dicts and lists --
and process the whole thing into an IPLD Node tree at once.

For example, "restructuring" means if you want an IPLD struct type,
you can create a Starlark dict with the same keys as the struct's field names,
hand that dict to the struct constructor function using the restructuring style,
and you'll get the desired outcome.

Restructuring also works deeply: to continue the above example,
if that struct contained another struct as a field,
and you put another corresponding dict inside your dict,
this should "just work".

Restructuring mode is available for pretty much everything,
and is engaged via use of a single special kwarg: underscore.
For example, a restructuring construction call might look like this:
`Foobar(_={"foo":"ooo", "bar":"aarrr"})`

#### Examples

- `List(1, 2, 3)` creates a list of length three.
- `List([1, 2, 3])` creates a list of length one, which contains a list of length three!  (Contrast with below.)
- `List(_=[1, 2, 3])` creates a list of length three.  (Contrast with above!)

Given some types:

```ipldsch
type Fun struct {
	fob FooOrBar
	zot String
}
type FooOrBar union {
	| Foo "foo:"
	| Bar "bar:"
} representation stringprefix
type Foo string
type Bar string
```

- `Fun(FooOrBar(foo="ooo"), "zot")` works just fine.
- `Fun("foo:ooo", "zot")` works too!  (When processing the first arg, we have a type expectation, and there's no prevailing mode, so Rule 2 can kick in, and thus we parse the the string into the union based on the type info and representation strategy that we have for it.)
- `Fun({"Foo":"ooo"}, "zot")` also works.  (Similar to above, but Rule 2 guided us a different way this time.)
- `Fun({"foo:":"ooo"}, "zot")` does NOT fly.  (Rule 2 on the first arg says to use type mode for the union, but then the dict key is from the representation level, so things don't line up, and this gets rejected.)
- `Fun(fob="foo:ooo", zot="zot")`, as you'd expect.
- `Fun(fob=FooOrBar(foo="ooo"), zot="zot")` works too.
- `Fun(_={"fob":{"Foo":"ooo"}, "zot":"zot"})` works!  This is the restructuring style in action.
- `Fun(_={"fob":"foo:ooo", "zot":"zot"})` works!  (There's a prevailing mode by the time we get to the union string, which would prefer to continue operating at type level, but the kind rule dominates it!)
- `Fun.Typed(_={"fob":"foo:ooo", "zot":"zot"})` does NOT fly.  (The explicit statement of representation mode is sticky all the way through.)
- `Fun.Typed(_={"fob":FooOrBar("foo:ooo"), "zot":"zot"})` works fine.  (Because `FooOrBar` is another construction call all its own, the rules restart -- the explicit mode from `Fun.Typed` doesn't apply anymore, so we're back to the kind rule working -- and then since we return a completely processed `FooBar` value, the restructuring that `Fun.Typed` is doing just accepts that value, regardless of its origin story, and moves along.)


Prototypes and Constructor Variations
-------------------------------------

The main constructor for a type is a function you can call -- but it's also a "prototype",
and you can get other constructors from it.

The most typical examples of this are `TypeName.Repr(...)` and `TypeName.Typed(...)`,
which let you explicitly construct a value using its representation or its type-level mode,
rather than trying to "figure it out" by best guess, as the default constructor does.
(TODO: not implemented yet.)

Some prototypes may also support access to type information via other attributes.
(TODO: not implemented yet.)

Some prototypes may also support additional constructor functions that have different styles:
for example, returning a builder object that lets one construct very large maps or lists incrementally,
or scan in large bytes sequences as streams.
(TODO: not implemented yet.)


Constructors per Kind
---------------------

Constructor functions have behaviors per their [kind](https://ipld.io/glossary/#kind).
In other words, the constructor function for a struct has certain styles of usage;
the constructor for a string has other styles of usage.
(Probably not a surprise!)

You'll find pages all the "`using_*.md`" files in this directory
which demonstrate in detail what calling conventions can be used for constructing various kinds of data.


