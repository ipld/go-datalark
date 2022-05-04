package datalarkengine

import (
	"fmt"

	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/ipld/go-ipld-prime/printer"
	"github.com/ipld/go-ipld-prime/schema"
	"go.starlark.net/starlark"
)

func ConstructString(np datamodel.NodePrototype, _ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var val string
	if err := starlark.UnpackPositionalArgs("datalark.String", args, kwargs, 1, &val); err != nil {
		return nil, err
	}

	nb := np.NewBuilder()
	if err := nb.AssignString(val); err != nil {
		return nil, err
	}
	return ToValue(nb.Build())
}

type String1 struct {
	val datamodel.Node
}

func NewString1(p datamodel.NodePrototype, s string) *String1 {
	nb := p.NewBuilder()
	nb.AssignString(s)
	n := nb.Build()
	return &String1{n}
}

func (g *String1) Node() datamodel.Node {
	return g.val
}
func (g *String1) Type() string {
	if tn, ok := g.val.(schema.TypedNode); ok {
		return fmt.Sprintf("datalark.String<%T>", tn.Type().Name())
	}
	return fmt.Sprintf("datalark.String")
}
func (g *String1) String() string {
	return printer.Sprint(g.val)
}
func (g *String1) Freeze() {}
func (g *String1) Truth() starlark.Bool {
	return true
}
func (g *String1) Hash() (uint32, error) {
	s, _ := g.val.AsString()
	return starlark.String(s).Hash()
}
