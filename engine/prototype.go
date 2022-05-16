package datalarkengine

import (
	"fmt"
	"reflect"

	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/ipld/go-ipld-prime/schema"
	"go.starlark.net/starlark"
)

// Prototype wraps an IPLD `datamodel.NodePrototype`, and in starlark,
// is a `Callable` which acts like a constructor for that NodePrototype.
//
// There is only one Prototype type, and its behavior varies based on
// the `datamodel.NodePrototype` its bound to.
type Prototype struct {
	name string
	np   datamodel.NodePrototype
}

func NewPrototype(name string, np datamodel.NodePrototype) *Prototype {
	return &Prototype{name: name, np: np}
}

func (p *Prototype) TypeName() string {
	return p.name
}

func (p *Prototype) NodePrototype() datamodel.NodePrototype {
	return p.np
}

// -- starlark.Value -->

var _ starlark.Value = (*Prototype)(nil)

func (p *Prototype) Type() string {
	if npt, ok := p.np.(schema.TypedPrototype); ok {
		return fmt.Sprintf("datalark.Prototype<%s>", npt.Type().Name())
	}
	return fmt.Sprintf("datalark.Prototype")
}
func (p *Prototype) String() string {
	return fmt.Sprintf("<built-in function %s>", p.Type())
}
func (p *Prototype) Freeze() {}
func (p *Prototype) Truth() starlark.Bool {
	return true
}
func (p *Prototype) Hash() (uint32, error) {
	return 0, nil
}

// -- starlark.Callable -->

var _ starlark.Callable = (*Prototype)(nil)

func (p *Prototype) Name() string {
	return p.String()
}

func (p *Prototype) CallInternal(thread *starlark.Thread, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	// If we have a TypedPrototype, try the appropriate constructors for its typekind.
	if npt, ok := p.np.(schema.TypedPrototype); ok {
		// TODO(dustmop): I don't understand TypedPrototypes at the moment, investigate
		// and return to this code.
		switch npt.Type().TypeKind() {
		case schema.TypeKind_Struct:
			return ConstructStruct(npt, thread, args, kwargs)
		case schema.TypeKind_Map:
			return ConstructMap(npt, thread, args, kwargs)
		default:
			panic(fmt.Errorf("nyi: datalark.Prototype.CallInternal for typed nodes with typekind %s", npt.Type().TypeKind()))
		}
	}

	// Otherwise, determine the types of arguments passed to this call, and
	// dispatch as appropriate
	nb := p.np.NewBuilder()
	switch {
	case len(args) > 0 && len(kwargs) > 0:
		return starlark.None, fmt.Errorf("datalark.Prototype.__call__: can either use positional or keyword arguments, but not both")
	case len(args) == 1:
		val := args[0]
		if err := assembleVal(nb, val); err != nil {
			gotType := reflect.TypeOf(val).Name()
			return starlark.None, fmt.Errorf("cannot create %s from %v of type %s", p.TypeName(), val, gotType)
		}
		return ToValue(nb.Build())
	case len(kwargs) > 0:
		// TODO(dustmop): This code is usually correct, except in the case of
		// unions, since they treat kwargs differently. Need to fix that to
		// handle them properly.
		dict, err := buildDictFromKwargs(kwargs)
		if err != nil {
			return starlark.None, err
		}
		if err := assembleVal(nb, dict); err != nil {
			return starlark.None, fmt.Errorf("datalark.Prototype.__call__: %w", err)
		}
		return ToValue(nb.Build())
	}
	return starlark.None, fmt.Errorf("datalark.Prototype.__call__: must be called with a single positional argument, or with keyword arguments")
}

// FUTURE: We can choose to implement Attrs and GetAttr on this, if we want to expose the ability to introspect things or look at types from skylark!
