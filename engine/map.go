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
	node    datamodel.Node
	added   map[string]datamodel.Node
	replace map[string]datamodel.Node
}

// compile-time interface assertions
var (
	_ Value              = (*mapValue)(nil)
	_ starlark.Value     = (*mapValue)(nil)
	_ starlark.Mapping   = (*mapValue)(nil)
	_ starlark.Sequence  = (*mapValue)(nil)
	_ starlark.HasSetKey = (*mapValue)(nil)
	_ starlark.HasAttrs  = (*mapValue)(nil)
)

func newMapValue(node datamodel.Node) Value {
	return &mapValue{node, nil, nil}
}

func (v *mapValue) Node() datamodel.Node {
	v.applyChangesToNode()
	return v.node
}
func (v *mapValue) Type() string {
	if tn, ok := v.node.(schema.TypedNode); ok {
		return fmt.Sprintf("datalark.Map<%T>", tn.Type().Name())
	}
	return fmt.Sprintf("datalark.Map")
}
func (v *mapValue) String() string {
	v.applyChangesToNode()
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
	return int(v.node.Length()) + len(v.added)
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

// starlark.HasSetKey

// SetKey assigns a value to a map at the given key
func (v *mapValue) SetKey(nameVal, val starlark.Value) error {

	dv, err := starlarkToDatalarkValue(val)
	if err != nil {
		return err
	}
	node := dv.Node()

	var name string
	name, _ = starlark.AsString(nameVal)
	exist, _ := v.node.LookupByString(name)
	if exist == nil {
		if v.added == nil {
			v.added = make(map[string]datamodel.Node)
		}
		v.added[name] = node
	} else {
		if v.replace == nil {
			v.replace = make(map[string]datamodel.Node)
		}
		v.replace[name] = node
	}
	return nil
}

func (v *mapValue) applyChangesToNode() error {
	// TODO: If added and replace are both empty, return fast
	nb := v.node.Prototype().NewBuilder()
	size := v.Len()
	ma, err := nb.BeginMap(int64(size))
	if err != nil {
		return err
	}

	//
	miter := v.node.MapIterator()
	for !miter.Done() {
		key, val, err := miter.Next()
		if err != nil {
			return err
		}
		keystr, err := key.AsString()
		if err != nil {
			return err
		}
		na := ma.AssembleKey()
		na.AssignString(keystr)
		if repl, ok := v.replace[keystr]; ok {
			na = ma.AssembleValue()
			na.AssignNode(repl)
			continue
		}
		na = ma.AssembleValue()
		na.AssignNode(val)
	}

	for keystr, val := range v.added {
		na := ma.AssembleKey()
		na.AssignString(keystr)
		na = ma.AssembleValue()
		na.AssignNode(val)
	}

	err = ma.Finish()
	if err != nil {
		return err
	}
	v.node = nb.Build()
	v.added = make(map[string]datamodel.Node)
	v.replace = make(map[string]datamodel.Node)
	return nil
}
