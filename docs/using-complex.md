Complex Compositions with Datalark
==================================

Datalark together with IPLD Schemas can describe very complex structures of information.
Datalark can also create these datastructures using _either_ the "type-level"
_or_ the "representation-level" approach to the data, which offers a lot of flexibility.
This flexiblity can be used to write terser structures, or emphasize clarity,
as the author sees fit.


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
		"delta:1,2"
	)
))
```

(TODO: currently fails within union constructor: needs to support Rule 2.)

Note that the entire `Delta` value is contained within the same string literal!
The substring "`delta:`" is consumed during the `Beta` constructor
(remember, `Beta` is the union type -- so first, it's figuring out which member the union will be occupied with),
and the remainder of the string ("`1,2`") is handed off to `Delta` constructor logic
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
print(mytypes.Alpha("delta:1,2"))
```

(TODO: currently fails within union constructor: needs to support Rule 2.)

We only really need the very first type to be indicated by a constructor call;
from there, we see a positional argument (and know what type to expect for that);
and then can apply the same logic as in the previous example (because the value is a string,
it clearly isn't the type-level value; so, use the representation-level processing).

See [The Decision Tree For Mode](consturctors.md#the-decision-tree-for-mode) docs for more info on the rules demonstrated here.


### Kitchen Sink constructed by Mix of Levels (Explicitly)

In this example, we use a bunch of explicit variations of constructor functions.

This is more verbose than any of the other examples,
but allows the author to be extremely clear about their intentions.

[testmark]:# (skipme/kitchensink/val1/script.various/explicit-mixed-level)
```python
# First of all, let's explicitly use a representation mode constructor, for fun:
print(mytypes.Alpha.Repr(
	# This constructor call is going to use restructuring style.
	# Note that the map key is "b" rather than "beta", because we're at representation level.
	_={"b":
		# At this moment, the "explicit mode" is still repr-level.
		# However, using another constructor call gives us an opportunity to switch,
		# and we'll even use an explict constructor mode to do so:
		mytypes.Beta.Typed(_={
			# Note the map key here is "Delta" (the type name),
			#  not "delta:" (which would be the representation-level discriminator string).
			"Delta": {"x": "1", "y": "2"},
			# Note that we used the type-level representation for the Delta value;
			# in this case we couldn't have used "1,2" and counted on Rule 2 to fix it up for us,
			# because we're still within a context where we explicitly said we're using typed mode.
		})
	}
))
```

(TODO: the syntax for explicit constructor level is a proposal and subject to review.)

There's little reason to do this in this example data.
However, you can imagine how in some circumstances, the explicitness might be important:
for example, if a struct with a map representation strategy has type-level field names that differ
from the keys used in the map representation (e.g. the "rename" feature is used),
then explicitly stating you want to use one level or the other may be important.


### Kitchen Sink constructed by Mix of Levels and Mixed Explicitness

We can use explicit constructors in some positions, and eschew them in others.

When we do this, the explicitly set mode continues for all the data processed
by that constructor function (and can be reset by using another constructor).

[testmark]:# (skipme/kitchensink/val1/script.various/mixed-explicitness)
```python
# The first parts of this example are the same as the previous.
# We start with an explicitly representation mode constructor:
print(mytypes.Alpha.Repr(
	_={"b":
		# At this moment, the "explicit mode" is still repr-level.
		# However, using another constructor call gives us an opportunity to switch.
		# Here we use an explicitly type-level constructor:
		mytypes.Beta.Typed(_={
			# Note the map key here is "Delta" (the type name),
			#  not "delta:" (which would be the representation-level discriminator string).
			"Delta": {"x": "1", "y": "2"}
			# Note that we also used the type-level representation for the Delta value;
			# in this case we could NOT have used "1,2" and counted on Rule 2 to fix it up for us,
			# because we're still within a context where we explicitly said we're using typed mode.
			# Explicit modes are Rule 1, so Rule 2 cannot be applied.
		})
	}
))
```

(TODO: the syntax for explicit constructor level is a proposal and subject to review.)

### Kitchen Sink constructed by Mix of Levels and Occasional Explicitness

Let's explore one more variation on the prior two examples:
what happens if we nest constructors that *aren't* explicit about their level?

[testmark]:# (skipme/kitchensink/val1/script.various/occasional-explicitness)
```python
# The first parts of this example are the same as the previous.
# We start with an explicitly representation mode constructor:
print(mytypes.Alpha.Repr(
	_={"b":
		# Note that this next constructor is the default "DWIM" constructor,
		#  *not* one setting an explicit level.
		mytypes.Beta(_={
			# Note the map key here is "Delta" (the type name),
			#  not "delta:" (which would be the representation-level discriminator string).
			"Delta": "1,2"
			# ... and yet we're also able to use "1,2" (the representation-level value for a Delta)
			# and count on Rule 2 to fix it up for us.
			# We had a prevailing mode of type-level (Rule 3), but not an explict mode (Rule 1);
			# therefore Rule 2 (kinds can force a level change) could apply.
		})
	}
))
```

(TODO: the syntax for explicit constructor level is a proposal and subject to review.)

In short: whenever you use a constructor function, the explicitness is reset.

This was a pretty complex example.
The reasoning behind what works and why is in comments in the code, but also,
see [The Decision Tree For Mode](consturctors.md#the-decision-tree-for-mode) docs for more info on the rules demonstrated here.
