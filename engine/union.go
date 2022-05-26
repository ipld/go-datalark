package datalarkengine

import (
	"fmt"

	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/ipld/go-ipld-prime/printer"
	"github.com/ipld/go-ipld-prime/schema"
	"go.starlark.net/starlark"
)

type unionValue struct {
	node datamodel.Node
}

var _ Value = (*unionValue)(nil)

func newUnionValue(node datamodel.Node) Value {
	return &unionValue{node}
}

func (v *unionValue) Node() datamodel.Node {
	return v.node
}
func (v *unionValue) Type() string {
	return fmt.Sprintf("datalark.Union<%T>", v.node.(schema.TypedNode).Type().Name())
}
func (v *unionValue) String() string {
	return printer.Sprint(v.node)
}
func (v *unionValue) Freeze() {}
func (v *unionValue) Truth() starlark.Bool {
	return true
}
func (v *unionValue) Hash() (uint32, error) {
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
