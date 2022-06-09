Complex Compositions with Datalark
==================================

Datalark together with IPLD Schemas can describe very complex structures of information.
Datalark can also create these datastructures using _either_ the "type-level"
_or_ the "representation-level" approach to the data, which offers a lot of flexibility.


Example of Kitchen Sink Complexity
----------------------------------

To see the full breadth of flexibility all at once,
let's take a quite exciting schema:

[testmark]:# (kitchensink/schema)
```ipldsch
type Alpha struct {
	beta Beta (rename "b")
} representation map

type Beta union {
	| Gamma "gamma:"
	| Delta "delta:"
} representation stringprefix

type Gamma string

type Delta struct {
	x String
	y String
} representation stringjoin {
	join ","
}
```

In this schema, we have:

- struct types and union types, and several of them, nested
- exciting representation strategies (e.g. stringprefix, stringjoin)
- and even nesting of things with those exciting representation strategies!

This will make for some very fun examples.

Most interestingly, there's going to be several different ways we can approach
creating values matching this structure, because we can use either the type-level
or representation-level approach, and we can mix and match them freely
if we regard it as convenient to do so!

### Kitchen Sink Example Value

In the following examples of datalark,
we'll try to produce the following value, printed in debug format as:

[testmark]:# (kitchensink/val1/output)
```text
struct<Alpha>{
	beta: union<Beta>{struct<Delta>{
		x: "1"
		y: "2"
	}}
}
```

### Kitchen Sink constructed by Type Level (Implicitly)

Perhaps the clearest way to construct this value is using the default constructors,
and giving them arguments that match the type-level structure,
and using another constructor function explicitly for each object:

[testmark]:# (skipme/kitchensink/val1/script.various/typelevel)
```python
print(mytypes.Alpha(
	# The following line is a positional argument to a struct constructor:
	mytypes.Beta(
		# The following line is a single argument to a union constructor
		#  (where the type info is all the info needed):
		mytypes.Delta(
			# We could use either positional or kwargs here, and chose kwargs:
			x="1",
			y="2",
		)
	)
))
```

(TODO: currently fails within union constructor: needs to support single positional argument if it's already typed)


### Kitchen Sink constructed by Representation (Implicitly)

Alternatively, we can use one of the representational constructors.
In this example, we start doing so at the `Beta` type,
and are able to have the default constructor switch into representation parsing mode,
because the kind of parameter we give it (a string, in this example) clearly matches the representation strategy.

[testmark]:# (skipme/kitchensink/val1/script.various/representation)
```python
print(mytypes.Alpha(
	mytypes.Beta(
		"gamma:1,2"
	)
))
```

(TODO: currently fails within union constructor: needs to support Rule 2.)

Note that the entire `Delta` value is contained within the same string literal!
The substring "`gamma:`" is consumed during the `Beta` constructor,
and the remainder of the string is handed off to `Delta` consturctor logic
(without seeing that call explicitly in the datalark syntax).

This kind of construction is using quite a lot of powerful features,
and may be somewhat confusing to understand on the first encounter.
However, as you can see, it also allows for remarkably concise expressions.

See [The Decision Tree For Mode](consturctors.md#the-decision-tree-for-mode) docs for more info on the rules demonstrated here.


### Kitchen Sink constructed by Representation (Even Shorter)

The above example can be even more brief by using even fewer constructor calls,
and letting datalark "figure it out":

[testmark]:# (skipme/kitchensink/val1/script.various/representation-shorter)
```python
print(mytypes.Alpha("gamma:1,2"))
```

(TODO: currently fails within union constructor: needs to support Rule 2.)

We only really need the very first type to be indicated by a constructor call;
from there, we see a positional argument (and know what type to expect for that);
and then can apply the same logic as in the previous example (because the value is a string,
it clearly isn't the type-level value; so, use the representation-level processing).

See [The Decision Tree For Mode](consturctors.md#the-decision-tree-for-mode) docs for more info on the rules demonstrated here.
