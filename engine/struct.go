package datalarkengine

import (
	"fmt"

	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/ipld/go-ipld-prime/printer"
	"github.com/ipld/go-ipld-prime/schema"
	"go.starlark.net/starlark"
)

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
