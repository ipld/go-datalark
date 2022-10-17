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

func nodeToHost(n datamodel.Node) Value {
	val, err := ToValue(n)
	if err != nil {
		panic(err)
	}
	return val
}

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
		return newBasicValue(n, datamodel.Kind_Null), nil
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

// assembleFrom assigns the incoming starlark Value to the node assembler
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
func assembleFrom(na datamodel.NodeAssembler, starVal starlark.Value) error {
	// if input value is already a hosted datalark Value, use its Node
	if hostVal, ok := starVal.(Value); ok {
		return na.AssignNode(hostVal.Node())
	}

	// try any of the starlark primitives we can recognize
	switch starObj := starVal.(type) {
	case starlark.Bool:
		return na.AssignBool(bool(starObj))
	case starlark.Int:
		i, ok := starObj.Int64()
		if !ok {
			return fmt.Errorf("could not convert %v to int64", starVal)
		}
		return na.AssignInt(i)
	case starlark.Float:
		return na.AssignFloat(float64(starObj))
	case starlark.String:
		return na.AssignString(string(starObj))
	case starlark.Bytes:
		return na.AssignBytes([]byte(starObj))
	case starlark.IterableMapping:
		size := -1
		if starSeq, ok := starObj.(starlark.Sequence); ok {
			// TODO(dustmop): If this conversion fails, size will be invalid
			size = starSeq.Len()
		}
		ma, err := na.BeginMap(int64(size))
		if err != nil {
			return err
		}
		starIter := starObj.Iterate()
		defer starIter.Done()
		var sval starlark.Value
		for starIter.Next(&sval) {
			if err := assembleFrom(ma.AssembleKey(), sval); err != nil {
				return err
			}
			sval, _, err := starObj.Get(sval)
			if err != nil {
				return err
			}
			if err := assembleFrom(ma.AssembleValue(), sval); err != nil {
				return err
			}
		}
		return ma.Finish()
	case starlark.Iterable:
		size := -1
		if seq, ok := starObj.(starlark.Sequence); ok {
			// TODO(dustmop): If this conversion fails, size will be invalid
			size = seq.Len()
		}
		la, err := na.BeginList(int64(size))
		if err != nil {
			return err
		}
		starIter := starObj.Iterate()
		defer starIter.Done()
		var sval starlark.Value
		for starIter.Next(&sval) {
			if err := assembleFrom(la.AssembleValue(), sval); err != nil {
				return err
			}
		}
		return la.Finish()
	}

	return fmt.Errorf("could not coerce %v of type %q into ipld datamodel", starVal, starVal.Type())
}

// convert a generic starlark.Value into a datalark.Value
func starToHost(val starlark.Value) (Value, error) {
	switch it := val.(type) {
	case Value:
		// already a datalark.Value, just return it
		return it, nil
	case starlark.NoneType:
		return NewNull(), nil
	case starlark.Bool:
		b := bool(it)
		return NewBool(b), nil
	case starlark.Int:
		n, ok := it.Int64()
		if !ok {
			return nil, fmt.Errorf("int64 out or range, could not convert: %v", val)
		}
		return NewInt(n), nil
	case starlark.Float:
		f := float64(it)
		return NewFloat(f), nil
	case starlark.String:
		return NewString(string(it)), nil
	case *starlark.List:
		return NewList(it)
	case starlark.Bytes:
		return NewBytes([]byte(string(it))), nil
	case *starlark.Dict:
		panic("TODO(dustmop): implement starToHost for dict")
	default:
		// Tuple, Set, Function, Builtin
		panic(fmt.Sprintf("unsupported type for starToHost: %T", val))
	}
}
