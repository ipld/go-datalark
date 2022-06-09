package datalarkengine

import (
	"fmt"

	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/ipld/go-ipld-prime/node/basicnode"
	"github.com/ipld/go-ipld-prime/printer"
	"github.com/ipld/go-ipld-prime/schema"

	"go.starlark.net/starlark"
	"go.starlark.net/syntax"
)

type Value interface {
	starlark.Value
	Node() datamodel.Node
}

// basicValue is used to store basic (non-recursive) types like bool, int, float, string, etc
type basicValue struct {
	node datamodel.Node
	kind datamodel.Kind
}

var _ Value = (*basicValue)(nil)
var _ starlark.HasBinary = (*basicValue)(nil)

func newBasicValue(node datamodel.Node, kind datamodel.Kind) Value {
	if kind != datamodel.Kind_Null &&
		kind != datamodel.Kind_Bool &&
		kind != datamodel.Kind_Int &&
		kind != datamodel.Kind_Float &&
		kind != datamodel.Kind_String &&
		kind != datamodel.Kind_Bytes &&
		kind != datamodel.Kind_Link {
		panic(fmt.Sprintf("invalid kind for basic value: %v", kind))
	}
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

// constructors for convenience, each returns a datalark Value or panics upon error

// NewNull constructs a null Value
func NewNull() Value {
	nb := basicnode.Prototype.Any.NewBuilder()
	if err := nb.AssignNull(); err != nil {
		panic(err)
	}
	return newBasicValue(nb.Build(), datamodel.Kind_Null)
}

// NewBool constructs a bool Value
func NewBool(b bool) Value {
	nb := basicnode.Prototype.Bool.NewBuilder()
	if err := nb.AssignBool(b); err != nil {
		panic(err)
	}
	return newBasicValue(nb.Build(), datamodel.Kind_Bool)
}

// NewInt constructs a int Value
func NewInt(n int64) Value {
	nb := basicnode.Prototype.Int.NewBuilder()
	if err := nb.AssignInt(n); err != nil {
		panic(err)
	}
	return newBasicValue(nb.Build(), datamodel.Kind_Int)
}

// NewFloat constructs a float Value
func NewFloat(f float64) Value {
	nb := basicnode.Prototype.Float.NewBuilder()
	if err := nb.AssignFloat(f); err != nil {
		panic(err)
	}
	return newBasicValue(nb.Build(), datamodel.Kind_Float)
}

// NewString constructs a string Value
func NewString(text string) Value {
	nb := basicnode.Prototype.String.NewBuilder()
	if err := nb.AssignString(text); err != nil {
		panic(err)
	}
	return newBasicValue(nb.Build(), datamodel.Kind_String)
}

// NewBytes constructs a bytes Value
func NewBytes(d []byte) Value {
	nb := basicnode.Prototype.Bytes.NewBuilder()
	if err := nb.AssignBytes(d); err != nil {
		panic(err)
	}
	return newBasicValue(nb.Build(), datamodel.Kind_Bytes)
}

// NewBytes constructs a Link Value
func NewLink(x datamodel.Link) Value {
	nb := basicnode.Prototype.Link.NewBuilder()
	if err := nb.AssignLink(x); err != nil {
		panic(err)
	}
	return newBasicValue(nb.Build(), datamodel.Kind_Link)
}

func (v *basicValue) Binary(op syntax.Token, y starlark.Value, side starlark.Side) (starlark.Value, error) {
	if op == syntax.PLUS || op == syntax.MINUS || op == syntax.STAR || op == syntax.SLASH {
		if other, ok := y.(*basicValue); ok {
			return v.binaryBasicOp(op, other)
		}
		if _, ok := y.(starlark.Int); ok {
			num, err := starlark.AsInt32(y)
			if err != nil {
				return starlark.None, err
			}
			return v.binaryBasicOp(op, NewInt(int64(num)).(*basicValue))
		}
		if _, ok := y.(starlark.Float); ok {
			f, ok := starlark.AsFloat(y)
			if !ok {
				return starlark.None, fmt.Errorf("could not convert %s to float", y)
			}
			return v.binaryBasicOp(op, NewFloat(float64(f)).(*basicValue))
		}
	}
	return starlark.None, fmt.Errorf("cannot %T %s %T", v, op, y)
}

func (v *basicValue) binaryBasicOp(op syntax.Token, other *basicValue) (starlark.Value, error) {
	if v.kind == datamodel.Kind_Int && other.kind == datamodel.Kind_Int {
		left, err := v.node.AsInt()
		if err != nil {
			return starlark.None, err
		}
		rite, err := other.node.AsInt()
		if err != nil {
			return starlark.None, err
		}
		if op == syntax.PLUS {
			return NewInt(left + rite), nil
		} else if op == syntax.MINUS {
			return NewInt(left - rite), nil
		} else if op == syntax.STAR {
			return NewInt(left * rite), nil
		} else if op == syntax.SLASH {
			return NewInt(left / rite), nil
		}
	}
	if v.kind == datamodel.Kind_Float && other.kind == datamodel.Kind_Float {
		left, err := v.node.AsFloat()
		if err != nil {
			return starlark.None, err
		}
		rite, err := other.node.AsFloat()
		if err != nil {
			return starlark.None, err
		}
		if op == syntax.PLUS {
			return NewFloat(left + rite), nil
		} else if op == syntax.MINUS {
			return NewFloat(left - rite), nil
		} else if op == syntax.STAR {
			return NewFloat(left * rite), nil
		} else if op == syntax.SLASH {
			return NewFloat(left / rite), nil
		}
	}
	return starlark.None, fmt.Errorf("cannot apply op %s to %T and %T", op, v, other)
}
