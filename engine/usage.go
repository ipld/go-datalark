package datalarkengine

import (
	"github.com/ipld/go-ipld-prime/node/basicnode"
	"github.com/ipld/go-ipld-prime/schema"
	"go.starlark.net/starlark"
)

// See docs on datalark.InjectGlobals.
// Typically you should prefer using functions in the datalark package,
// rather than their equivalents in the datalarkengine package.
func InjectGlobals(globals starlark.StringDict, obj *Object) {
	// Technically this would work on any 'starlark.IterableMapping', but I don't think that makes the function more useful, and would make it *less* self-documenting.
	itr := obj.Iterate()
	defer itr.Done()
	var k starlark.Value
	for itr.Next(&k) {
		v, _, err := obj.Get(k)
		if err != nil {
			panic(err)
		}
		globals[string(k.(starlark.String))] = v
	}
}

// PrimitiveConstructors returns the constructors for primitive types as an Object
func PrimitiveConstructors() *Object {
	obj := NewObject(7)
	obj.SetKey(starlark.String("Map"), &Prototype{"Map", basicnode.Prototype.Map})
	obj.SetKey(starlark.String("List"), &Prototype{"List", basicnode.Prototype.List})
	obj.SetKey(starlark.String("Bool"), &Prototype{"Bool", basicnode.Prototype.Bool})
	obj.SetKey(starlark.String("Int"), &Prototype{"Int", basicnode.Prototype.Int})
	obj.SetKey(starlark.String("Float"), &Prototype{"Float", basicnode.Prototype.Float})
	obj.SetKey(starlark.String("String"), &Prototype{"String", basicnode.Prototype.String})
	obj.SetKey(starlark.String("Bytes"), &Prototype{"Bytes", basicnode.Prototype.Bytes})
	obj.Freeze()
	return obj
}

// MakeConstructors returns the constructors for the given prototypes as an Object
func MakeConstructors(prototypes []schema.TypedPrototype) *Object {
	obj := NewObject(len(prototypes))
	for _, npt := range prototypes {
		obj.SetKey(starlark.String(npt.Type().Name()), &Prototype{npt.Type().Name(), npt})
	}
	obj.Freeze()
	return obj
}
