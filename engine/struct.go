package datalarkengine

import (
	"fmt"

	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/ipld/go-ipld-prime/printer"
	"github.com/ipld/go-ipld-prime/schema"
	"go.starlark.net/starlark"
)

func ConstructStruct(npt schema.TypedPrototype, _ *starlark.Thread, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	// Parsing args for struct construction is *very* similar to for maps...
	//  Except structs also allow positional arguments; maps can't make sense of that.

	// Try parsing two different ways: either positional, or kwargs (but not both).
	nb := npt.NewBuilder()
	switch {
	case len(args) > 0 && len(kwargs) > 0:
		return starlark.None, fmt.Errorf("datalark.Struct: can either use positional or keyword arguments, but not both")

	case len(kwargs) > 0:
		err := buildMapFromKwargs(nb, kwargs)
		if err != nil {
			return starlark.None, err
		}

	case len(args) == 0:
		// Well, okay.  Hope the whole struct is optional fields though, or you're probably gonna get a schema validation error.
		ma, err := nb.BeginMap(0)
		if err != nil {
			return starlark.None, err
		}
		if err := ma.Finish(); err != nil {
			return starlark.None, err
		}

	case len(args) == 1:
		// TODO(dustmop): Validate that this is a dict, fail early otherwise
		// If there's one arg, and it's a starlark dict, 'assembleVal' will do the right thing and restructure that into us.
		if err := assembleVal(nb, args[0]); err != nil {
			return starlark.None, fmt.Errorf("datalark.Struct: %w", err)
		}

	case len(args) > 1:
		return starlark.None, fmt.Errorf("datalark.Struct: if using positional arguments, only one is expected: a dict which we can restructure to match this type")

	default:
		panic("unreachable")
	}
	return newStructValue(nb.Build()), nil
}

type structValue struct {
	node datamodel.Node
}

var _ Value = (*structValue)(nil)

func newStructValue(node datamodel.Node) Value {
	return &structValue{node}
}

func (v *structValue) Node() datamodel.Node {
	return v.node
}
func (v *structValue) Type() string {
	return fmt.Sprintf("datalark.Struct<%T>", v.node.(schema.TypedNode).Type().Name())
}
func (v *structValue) String() string {
	return printer.Sprint(v.node)
}
func (v *structValue) Freeze() {}
func (v *structValue) Truth() starlark.Bool {
	return true
}
func (v *structValue) Hash() (uint32, error) {
	// Riffing off Starlark's algorithm for Tuple, which is in turn riffing off Python.
	var x, mult uint32 = 0x345678, 1000003
	l := v.node.Length()
	for itr := v.node.MapIterator(); !itr.Done(); {
		_, v, err := itr.Next()
		if err != nil {
			return 0, err
		}
		w, err := ToValue(v)
		if err != nil {
			return 0, err
		}
		y, err := w.Hash()
		if err != nil {
			return 0, err
		}
		x = x ^ y*mult
		mult += 82520 + uint32(l+l)
	}
	return x, nil
}

func (v *structValue) Attr(name string) (starlark.Value, error) {
	// TODO: distinction between 'Attr' and 'Get'.  This can/should list functions, I think.  'Get' makes it unambiguous.  I think.
	// TODO: perhaps also add a "__constr__" or "__proto__" function to everything?
	n, err := v.node.LookupByString(name)
	if err != nil {
		return nil, err
	}
	return ToValue(n)
}

func (v *structValue) AttrNames() []string {
	names := make([]string, 0, v.node.Length())
	for itr := v.node.MapIterator(); !itr.Done(); {
		k, _, err := itr.Next()
		if err != nil {
			panic(fmt.Errorf("error while iterating: %w", err)) // should *really* not happen for structs, incidentally.
		}
		ks, _ := k.AsString()
		names = append(names, ks)
	}
	return names
}

func (v *structValue) SetField(name string, val starlark.Value) error {
	return fmt.Errorf("datalark values are immutable")
}

