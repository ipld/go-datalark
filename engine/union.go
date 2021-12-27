package datalarkengine

import (
	"fmt"

	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/ipld/go-ipld-prime/schema"
	"go.starlark.net/starlark"
)

// ConstructUnion takes a schema.TypedPrototype which must be of typekind 'union',
// creates a builder from it, unpacks the starlark arugments into it, and returns the resulting IPLD Node.
//
// Several styles are supported (some only for unions with certain representation strategies):
//
//   1. A single positional argument can be used, if the value is already typed,
//       e.g. `(yourtypes.MemberTypeName("value"))`.
//   2. Keyword args can be used, using a member type name as the key:
//       e.g. `(MemberTypeName="value")`.
//        If the value is not already typed, it will be fed to a type-level NodeBuilder for that type.
//   3. Keyword args can be used, using keys from the representation:
//       e.g. `(descriminantlabel="value")`.
//       This works regardless of the representation strategy.
//        If the value is not already typed, it will be fed to a representation-level NodeBuilder for that type.
//   4. An object for restructuring can be used, using a member type name as the key:
//       e.g. `({"MemberTypeName":"value"})`.
//        If the value is not already typed, it will be fed to a type-level NodeBuilder for that type.
//   5. An object for restructuring can be used, and will be handled however the representation strategy handles it:
//       e.g. `({"descriminantlabel":"value"})` for a keyed union,
//       or `({"tag":"descriminantlabel", "content":"value"}) for an envelope union.
//        If the value is not already typed, it will be fed to a representation-level NodeBuilder for that type.
//   6. For unions which have a representation strategy that works with strings,
//      a single positional argument that's an untyped string can be used:
//       e.g. `("discriminatorprefix:value")`.
//        (In fact, this is just another case of Style 5, but is worth noting.)
//
// The styles bind from top to bottom: e.g. Style 1 applies first if it can, and then down the list.
// In cases which may be ambiguous, such as between Style 4 and 5 if the union is keyed and the discriminant labels and the member type names happen to match,
// then Style 4 would be chosen.  (This might be consequential because it also sets the expectation for how the rest of a complex value might be processed: as the type-level, or the representation-level view.)
//
// Reminder: if you want to use restructuring styles, but a mixture of type-level and representation-level values with a large structure,
// you can do this within starlark: you can always use a new constructor function invocation to restart which level you're on.
//
func ConstructUnion(npt schema.TypedPrototype, _ *starlark.Thread, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	// Processing for each style in a short function, for code legibility.
	style1 := func(npt schema.TypedPrototype, kwarg starlark.Tuple) (starlark.Value, error) {
		nb := npt.NewBuilder()
		ma, err := nb.BeginMap(1)
		if err != nil {
			panic(fmt.Errorf("dishonest union implementation?!: %w", err))
		}
		_, err = ma.AssembleEntry(string(kwarg.Index(0).(starlark.String)))
		if err != nil {
			return starlark.None, fmt.Errorf("invalid arg to construction: must use a keyword ")
			// FIXME you hardly every want to use the type names.  they're capitalized and look weird.  you often want the repr behavior here.  but how distinguish??
			// Can we just check both tables and "dtrt"?
			// If so, have to do it before creating a buildler -- builder is already a commitment to a level.
			// ... if the value is deep, it requires picking a level and sticking to it too.
			//    I guess we attain that naturally by using assembleVal instead of recursing at the level starlark would.  Is that good?
			//  Yes, actually, I think it is.  I think that's the principle discovery.  We can read tealeaves about level immediately in the start of construction, but after that you lock in.
			//     ... except using a type again, even in a deep structure that gets restructured, should still dominate.  assembleVal probably doesn't do that yet.
			//         It does, actually!  but only on the assumption that the AssignNode receiving it knows how to grok either repr or type level nodes.  might be bugs in those zones to discover and fix; test heavily.
			// The other key insight I now realize is: yeah, the default constructor should do yolo-figure-it-out.
			//    If you ask it for EITHER type-level or repr-level things, it should hem itself in accordingly.
			//    In other words: the default shouldn't be one or the other.  The default is its own yolo mode thing.  There's three modes.  Not two.
		}
		return starlark.None, nil
	}
	style2 := func(npt schema.TypedPrototype, val Value) (starlark.Value, error) {
		panic("TODO")
		// Clear enough.
	}
	style3 := func(npt schema.TypedPrototype, pair starlark.IterableMapping) (starlark.Value, error) {
		panic("TODO")
		// Also Provokes the level question -- same thing as earlier though, really.
		// The amount of code that htis and style1 share should be nearly 100%.
	}
	style4 := func(npt schema.TypedPrototype, val starlark.String) (starlark.Value, error) {
		panic("TODO")
		// Clear enough.
	}
	// Discerning which style we're trying to fit into, below.
	switch {
	case len(args) > 0 && len(kwargs) > 0:
		return starlark.None, fmt.Errorf("datalark.Union<%s>: construction can have several forms but all either use positional or keyword arguments, not both", npt.Type().Name())
	case len(kwargs) > 0:
		if len(kwargs) > 1 {
			return starlark.None, fmt.Errorf("datalark.Union<%s>: construction using kwargs means we expect the member name as keyword and can only accept one argument, got %d", npt.Type().Name(), len(kwargs))
		}
		return style1(npt, kwargs[0])
	case len(args) > 0:
		if len(args) > 1 {
			return starlark.None, fmt.Errorf("datalark.Union<%s>: construction using positional args has several forms but all can only accept one argument, got %d", npt.Type().Name(), len(kwargs))
		}
		if dlval, ok := args[0].(Value); ok {
			return style2(npt, dlval)
		}
		if mapish, ok := args[0].(starlark.IterableMapping); ok {
			return style3(npt, mapish)
		}
		if npt.Type().RepresentationBehavior() != datamodel.Kind_String {
			return starlark.None, fmt.Errorf("datalark.Union<%s>: construction using positional args can only accept typed values or maps for restructuring, got something else", npt.Type().Name())
		}
		if str, ok := args[0].(starlark.String); ok {
			return style4(npt, str)
		}
		return starlark.None, fmt.Errorf("datalark.Union<%s>: construction using positional args can only accept typed values, maps for restructuring, or strings; got something else", npt.Type().Name())
	default:
		panic("unreachable")
	}
}
