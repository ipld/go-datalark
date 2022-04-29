package datalarkengine

import (
	"fmt"
	"go.starlark.net/starlark"
	"github.com/ipld/go-ipld-prime/datamodel"
)

type Value interface {
	starlark.Value
	Kind() datamodel.Kind
}

// Null

var _ Value = (*NullValue)(nil)

type NullValue struct {
}

func (v *NullValue) Kind() datamodel.Kind {
	return datamodel.Kind_Null
}

func (v *NullValue) Type() string {
	return "Null"
}

func (v *NullValue) Freeze() {}

func (v *NullValue) Truth() starlark.Bool {
	return false
}

func (v *NullValue) String() string {
	return "null"
}

func (v *NullValue) Hash() (uint32, error) {
	// TODO(dustmop): IMPLEMENT ME
	return 0, nil
}

// Boolean

var _ Value = (*BoolValue)(nil)

type BoolValue struct {
	b bool
}

func (v *BoolValue) Kind() datamodel.Kind {
	return datamodel.Kind_Bool
}

func (v *BoolValue) Type() string {
	return "Bool"
}

func (v *BoolValue) Freeze() {}

func (v *BoolValue) Truth() starlark.Bool {
	return starlark.Bool(v.b)
}

func (v *BoolValue) String() string {
	if v.b {
		return "true"
	}
	return "false"
}

func (v *BoolValue) Hash() (uint32, error) {
	// TODO(dustmop): IMPLEMENT ME
	return 0, nil
}

// Integer

var _ Value = (*IntegerValue)(nil)

type IntegerValue struct {
	i int
}

func (v *IntegerValue) Kind() datamodel.Kind {
	return datamodel.Kind_Int
}

func (v *IntegerValue) Type() string {
	return "Integer"
}

func (v *IntegerValue) Freeze() {}

func (v *IntegerValue) Truth() starlark.Bool {
	return v.i != 0
}

func (v *IntegerValue) String() string {
	return fmt.Sprintf("%d", v.i)
}

func (v *IntegerValue) Hash() (uint32, error) {
	// TODO(dustmop): IMPLEMENT ME
	return 0, nil
}

// Float

var _ Value = (*FloatValue)(nil)

type FloatValue struct {
	f float32
}

func (v *FloatValue) Kind() datamodel.Kind {
	return datamodel.Kind_Float
}

func (v *FloatValue) Type() string {
	return "Float"
}

func (v *FloatValue) Freeze() {}

func (v *FloatValue) Truth() starlark.Bool {
	return v.f != 0.0
}

func (v *FloatValue) String() string {
	return fmt.Sprintf("%f", v.f)
}

func (v *FloatValue) Hash() (uint32, error) {
	// TODO(dustmop): IMPLEMENT ME
	return 0, nil
}

// String

var _ Value = (*StringValue)(nil)

type StringValue struct {
	s string
}

func (v *StringValue) Kind() datamodel.Kind {
	return datamodel.Kind_String
}

func (v *StringValue) Type() string {
	return "String"
}

func (v *StringValue) Freeze() {}

func (v *StringValue) Truth() starlark.Bool {
	return true
}

func (v *StringValue) String() string {
	return v.s
}

func (v *StringValue) Hash() (uint32, error) {
	// TODO(dustmop): IMPLEMENT ME
	return 0, nil
}

// Bytes

var _ Value = (*BytesValue)(nil)

type BytesValue struct {
	bs []byte
}

func (v *BytesValue) Kind() datamodel.Kind {
	return datamodel.Kind_Bytes
}

func (v *BytesValue) Type() string {
	return "Bytes"
}

func (v *BytesValue) Freeze() {}

func (v *BytesValue) Truth() starlark.Bool {
	return true
}

func (v *BytesValue) String() string {
	return fmt.Sprintf("%v", v.bs)
}

func (v *BytesValue) Hash() (uint32, error) {
	// TODO(dustmop): IMPLEMENT ME
	return 0, nil
}

// Link

var _ Value = (*LinkValue)(nil)

type LinkValue struct {
	// TODO(dustmop): Don't know what the shape of a link is yet.
	link string
}

func (v *LinkValue) Kind() datamodel.Kind {
	return datamodel.Kind_Link
}

func (v *LinkValue) Type() string {
	return "Link"
}

func (v *LinkValue) Freeze() {}

func (v *LinkValue) Truth() starlark.Bool {
	return true
}

func (v *LinkValue) String() string {
	return v.link
}

func (v *LinkValue) Hash() (uint32, error) {
	// TODO(dustmop): IMPLEMENT ME
	return 0, nil
}

