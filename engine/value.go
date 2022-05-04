package datalarkengine

import (
	"fmt"

	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/ipld/go-ipld-prime/printer"
	"github.com/ipld/go-ipld-prime/schema"
	"go.starlark.net/starlark"
)

type Value interface {
	starlark.Value
	Node() datamodel.Node
}

type value struct {
	node datamodel.Node
	kind datamodel.Kind
}

func newValue(node datamodel.Node, kind datamodel.Kind) Value {
	return &value{node, kind}
}


func (v *value) Node() datamodel.Node {
	return v.node
}

func (v *value) Type() string {
	if typed, ok := v.node.(schema.TypedNode); ok {
		return fmt.Sprintf("datalark.%s<%T>", v.kind, typed.Type().Name())
	}
	return fmt.Sprintf("datalark.%s", v.kind)
}

func (v *value) String() string {
	return printer.Sprint(v.node)
}

func (v *value) Freeze() {}

func (v *value) Truth() starlark.Bool {
	return true
}

func (v *value) Hash() (uint32, error) {
	// TODO(dustmop): implement me
	return 0, nil
}
