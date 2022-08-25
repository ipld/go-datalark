/*
	datalarkengine contains all the low-level binding logic.

	Perhaps somewhat surprisingly, it includes even wrapper types for the more primitive kinds (like string).
	This is important (rather than just converting them directly to starlark's values)
	because we may want things like IPLD type information (or even just NodePrototype) to be retained,
	as well as sometimes wanting the original pointer to be retained for efficiency reasons.
*/
package datalarkengine

import (
	"fmt"

	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/ipld/go-ipld-prime/schema"
	"go.starlark.net/starlark"
)

func ToValue(n datamodel.Node) (Value, error) {
	if nt, ok := n.(schema.TypedNode); ok {
		switch nt.Type().TypeKind() {
		case schema.TypeKind_Struct:
			return newStructValue(n), nil
		case schema.TypeKind_Union:
			return newUnionValue(n), nil
		case schema.TypeKind_Enum:
			panic("IMPLEMENT ME!")
		}
	}
	switch n.Kind() {
	case datamodel.Kind_Map:
		return newMapValue(n), nil
	case datamodel.Kind_List:
		return newListValue(n), nil
	case datamodel.Kind_Null:
		panic("IMPLEMENT ME!")
	case datamodel.Kind_Bool:
		return newBasicValue(n, datamodel.Kind_Bool), nil
	case datamodel.Kind_Int:
		return newBasicValue(n, datamodel.Kind_Int), nil
	case datamodel.Kind_Float:
		return newBasicValue(n, datamodel.Kind_Float), nil
	case datamodel.Kind_String:
		return newBasicValue(n, datamodel.Kind_String), nil
	case datamodel.Kind_Bytes:
		return newBasicValue(n, datamodel.Kind_Bytes), nil
	case datamodel.Kind_Link:
		panic("IMPLEMENT ME!")
	case datamodel.Kind_Invalid:
		panic("invalid!")
	default:
		panic("unreachable")
	}
}

// assembleVal assigns the incoming starlark Value to the node assembler
//
// Attempt to put the starlark Value into the ipld NodeAssembler.
// If we see it's one of our own wrapped types, yank it back out and use AssignNode.
// If it's a starlark string, take that and use AssignString.
// Other starlark primitives, similarly.
// Dicts and lists are also handled.
//
// This makes some attempt to be nice to foreign/user-defined "types" in starlark as well;
// in particular, anything implementing `starlark.IterableMapping` will be converted into map-building assignments,
// and anything implementing just `starlark.Iterable` (and not `starlark.IterableMapping`) will be converted into list-building assignments.
// However, there is no support for primitives unless they're one of the concrete types from the starlark package;
// starlark doesn't have a concept of a data model where you can ask what "kind" something is,
// so if it's not *literally* one of the concrete types that we can match on, well, we're outta luck.
func assembleVal(na datamodel.NodeAssembler, sval starlark.Value) error {
	// if the incoming value is already a datalark Value, use its Node
	if v, ok := sval.(Value); ok {
		return na.AssignNode(v.Node())
	}

	// Try any of the starlark primitives we can recognize.
	// TODO(dustmop): Add explicit type-checking to each case. The nodeAssmebler
	// must have compatible type with the incoming value to assign, otherwise
	// the error comes from the nodeAssembler and may not accurately describe the
	// problem
	switch s2 := sval.(type) {
	case starlark.Bool:
		return na.AssignBool(bool(s2))
	case starlark.Int:
		i, ok := s2.Int64()
		if !ok {
			return fmt.Errorf("could not convert %v to int64", sval)
		}
		return na.AssignInt(i)
	case starlark.Float:
		return na.AssignFloat(float64(s2))
	case starlark.String:
		return na.AssignString(string(s2))
	case starlark.Bytes:
		return na.AssignBytes([]byte(s2))
	case starlark.IterableMapping:
		size := -1
		if seq, ok := s2.(starlark.Sequence); ok {
			// TODO(dustmop): If this conversion fails, size will be invalid
			size = seq.Len()
		}
		ma, err := na.BeginMap(int64(size))
		if err != nil {
			return err
		}
		itr := s2.Iterate()
		defer itr.Done()
		var k starlark.Value
		for itr.Next(&k) {
			if err := assembleVal(ma.AssembleKey(), k); err != nil {
				return err
			}
			v, _, err := s2.Get(k)
			if err != nil {
				return err
			}
			if err := assembleVal(ma.AssembleValue(), v); err != nil {
				return err
			}
		}
		return ma.Finish()
	case starlark.Iterable:
		size := -1
		if seq, ok := s2.(starlark.Sequence); ok {
			// TODO(dustmop): If this conversion fails, size will be invalid
			size = seq.Len()
		}
		la, err := na.BeginList(int64(size))
		if err != nil {
			return err
		}
		itr := s2.Iterate()
		defer itr.Done()
		var v starlark.Value
		for itr.Next(&v) {
			if err := assembleVal(la.AssembleValue(), v); err != nil {
				return err
			}
		}
		return la.Finish()
	}

	return fmt.Errorf("could not coerce %v of type %q into ipld datamodel", sval, sval.Type())
}

// convert a generic starlark.Value into a datalark.Value
func starlarkToDatalarkValue(val starlark.Value) (Value, error) {
	switch it := val.(type) {
	case starlark.Int:
		n, ok := it.Int64()
		if !ok {
			return nil, fmt.Errorf("int64 out or range, could not convert: %v", val)
		}
		return NewInt(n), nil
	case starlark.String:
		return NewString(string(it)), nil
	case *starlark.List:
		return NewList(it)
	}
	panic(fmt.Sprintf("TODO(dustmop): implement starlarkToDatalarkValue for %T", val))
}
