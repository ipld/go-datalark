/*
	datalark makes IPLD data legible to, and constructable in, starlark.

	Given an IPLD Schema (and optionally, a list of types to focus on),
	datalark can generate a set of starlark constructor functions for those types.
	These functions should generally DWIM ("do what I mean"):
	for structs, they accept kwargs corresponding to the field names, etc.
	Some functions get clever: for example, for structs with stringy representations (stringjoin, etc),
	the representation form can be used as an argument to the constructor instead of the kwargs form,
	and the construction will "DWIM" with that information and parse it in the appropriate way.

	Standard datamodel data is also always legible,
	and a set of functions for creating it can also be obtained from the datalark package.

	All IPLD data exposed to starlark always acts as if it is "frozen", in starlark parlance.
	This should be unsurprising, since IPLD is already oriented around immutability.

	datalark can be used on natural golang structs by combining it with the
	go-ipld-prime/node/bindnode package.
	This may make it an interesting alternative to github.com/starlight-go/starlight
	(although admittedly more complicated; it's probably only worth it if you
	also already value some of the features of IPLD Schemas).

	Future objectives for this package include the ability to provide a function to starlark
	which will accept an IPLD Schema document and a type name as parameters,
	and will return a constructor for that type.
	(Not yet implemented.)
*/
package datalark

import (
	"github.com/ipld/go-ipld-prime/schema"
	"go.starlark.net/starlark"

	"github.com/ipld/go-datalark/engine"
)

// InjectGlobals mutates a starlark.StringDict to contain the values in the given Object.
// It will panic if keys that aren't starlark.String are encountered, if iterators error, etc.
func InjectGlobals(globals starlark.StringDict, obj *datalarkengine.Object) {
	datalarkengine.InjectGlobals(globals, obj)
}

// PrimitiveConstrutors returns an Object containing constructor functions
// for all the IPLD Data Model kinds -- strings, maps, etc -- as those names, in TitleCase.
func PrimitiveConstructors() *datalarkengine.Object {
	return datalarkengine.PrimitiveConstructors()
}

// MakeConstructors returns an Object containing constructor functions for IPLD typed
// nodes, based on the list of schema.TypedPrototype provided, and using the names
// of each of those prototype's types as the keys.
func MakeConstructors(prototypes []schema.TypedPrototype) *datalarkengine.Object {
	return datalarkengine.MakeConstructors(prototypes)
}
