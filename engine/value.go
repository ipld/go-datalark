package datalarkengine

import (
	"fmt"

	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/ipld/go-ipld-prime/printer"
	"github.com/ipld/go-ipld-prime/node/basicnode"
	"github.com/ipld/go-ipld-prime/schema"
	"go.starlark.net/starlark"
)

type Value interface {
	starlark.Value
	Node() datamodel.Node
}

type basicValue struct {
	node datamodel.Node
	kind datamodel.Kind
}

var _ Value = (*basicValue)(nil)

func newBasicValue(node datamodel.Node, kind datamodel.Kind) Value {
	return &basicValue{node, kind}
}

func (v *basicValue) Node() datamodel.Node {
	return v.node
}

func (v *basicValue) Type() string {
	if typed, ok := v.node.(schema.TypedNode); ok {
		return fmt.Sprintf("datalark.%s<%T>", v.kind, typed.Type().Name())
	}
	return fmt.Sprintf("datalark.%s", v.kind)
}

func (v *basicValue) String() string {
	return printer.Sprint(v.node)
}

func (v *basicValue) Freeze() {}

func (v *basicValue) Truth() starlark.Bool {
	return true
}

func (v *basicValue) Hash() (uint32, error) {
	// TODO(dustmop): implement me
	return 0, nil
}

// Constructors for convenience

func NewString(text string) Value {
	nb := basicnode.Prototype.String.NewBuilder()
	nb.AssignString(text)
	return newBasicValue(nb.Build(), datamodel.Kind_String)
}
