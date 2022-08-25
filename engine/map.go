package datalarkengine

import (
	"errors"
	"fmt"

	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/ipld/go-ipld-prime/printer"
	"github.com/ipld/go-ipld-prime/schema"
	"go.starlark.net/starlark"
)

type mapValue struct {
	node datamodel.Node
}

var _ Value = (*mapValue)(nil)
var _ starlark.Mapping = (*mapValue)(nil)
var _ starlark.Sequence = (*mapValue)(nil)
var _ starlark.HasAttrs = (*mapValue)(nil)

func newMapValue(node datamodel.Node) Value {
	return &mapValue{node}
}

func (v *mapValue) Node() datamodel.Node {
	return v.node
}
func (v *mapValue) Type() string {
	if tn, ok := v.node.(schema.TypedNode); ok {
		return fmt.Sprintf("datalark.Map<%T>", tn.Type().Name())
	}
	return fmt.Sprintf("datalark.Map")
}
func (v *mapValue) String() string {
	return printer.Sprint(v.node)
}
func (v *mapValue) Freeze() {}
func (v *mapValue) Truth() starlark.Bool {
	return true
}
func (v *mapValue) Hash() (uint32, error) {
	return 0, errors.New("TODO")
}

// Get returns a value from a map, implementing starlark.Mapping
// example:
//
//   d = {'a': 'apple', 'b': 'banana'}
//   d['a'] # calls d.Get('a')
//
func (v *mapValue) Get(in starlark.Value) (out starlark.Value, found bool, err error) {
	keyStr, ok := in.(starlark.String)
	if !ok {
		return starlark.None, false, fmt.Errorf("cannot index map using %v of type %T", in, in)
	}
	key := string(keyStr)
	n, err := v.node.LookupByString(key)
	if err != nil {
		return nil, false, err
	}
	val, err := ToValue(n)
	return val, true, err
}

// starlark.Sequence

func (v *mapValue) Iterate() starlark.Iterator {
	panic(fmt.Errorf("TODO(dustmop): mapValue.Iterate not implemented for %T", v))
}

func (v *mapValue) Len() int {
	return int(v.node.Length())
}

// starlark.HasAttrs : starlark.Map

var mapMethods = []string{"clear", "copy", "fromkeys", "get", "items", "keys", "pop", "popitem", "setdefault", "update", "values"}

func (v *mapValue) Attr(name string) (starlark.Value, error) {
	// convert map to a starlark.Dict. not efficient, because it makes a copy
	dictVal := starlark.NewDict(v.Len())
	iter := v.node.MapIterator()
	for !iter.Done() {
		k, v, err := iter.Next()
		if err != nil {
			return starlark.None, err
		}
		key, err := ToValue(k)
		if err != nil {
			return starlark.None, err
		}
		val, err := ToValue(v)
		if err != nil {
			return starlark.None, err
		}
		err = dictVal.SetKey(key, val)
		if err != nil {
			return starlark.None, err
		}
	}
	method := func(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		// get actual method from underlying starlark.Dict
		method, err := dictVal.Attr(name)
		if err != nil {
			return starlark.None, err
		}
		// call the method, and convert the result to a datalark.Value
		res, err := starlark.Call(thread, method, args, kwargs)
		if err != nil {
			return starlark.None, err
		}
		return starlarkToDatalarkValue(res)
	}
	return starlark.NewBuiltin(name, method), nil
}

func (v *mapValue) AttrNames() []string {
	return mapMethods
}
