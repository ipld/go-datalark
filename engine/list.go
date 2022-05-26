package datalarkengine

import (
	"errors"
	"fmt"

	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/ipld/go-ipld-prime/printer"
	"go.starlark.net/starlark"
)

type listValue struct {
	node datamodel.Node
}

var _ Value = (*listValue)(nil)

func newListValue(node datamodel.Node) Value {
	return &listValue{node}
}

func (v *listValue) Node() datamodel.Node {
	return v.node
}
func (v *listValue) Type() string {
	// TODO(dustmop): Can a list be a TypedNode? I believe so, it
	// is used for a homogeneous typed list.
	return fmt.Sprintf("datalark.List")
}
func (v *listValue) String() string {
	return printer.Sprint(v.node)
}
func (v *listValue) Freeze() {}
func (v *listValue) Truth() starlark.Bool {
	return true
}
func (v *listValue) Hash() (uint32, error) {
	return 0, errors.New("TODO")
}
