package datalarkengine

import (
	_ "fmt"
)

type Value interface {
	Type() string
}

var _ Value = (*NullValue)(nil)

type NullValue struct {
}

func (v *NullValue) Type() string {
	return "Null"
}

var _ Value = (*BooleanValue)(nil)

type BooleanValue struct {
	b bool
}

func (v *BooleanValue) Type() string {
	return "Boolean"
}

var _ Value = (*IntegerValue)(nil)

type IntegerValue struct {
	i int
}

func (v *IntegerValue) Type() string {
	return "Integer"
}
