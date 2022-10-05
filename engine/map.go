package datalarkengine

import (
	"errors"
	"fmt"

	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/ipld/go-ipld-prime/node/basicnode"
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

type mapMethod func(*mapValue, starlark.Tuple, []starlark.Tuple) (starlark.Value, error)

var mapMethods = map[string]*starlark.Builtin{
	"clear":      NewMapMethod("clear", _mapClear, 0),
	"copy":       NewMapMethod("copy", _mapCopy, 0),
	"fromkeys":   NewMapMethod("fromkeys", _mapFromkeys, 2),
	"get":        NewMapMethod("get", _mapGet, 2),
	"items":      NewMapMethod("items", _mapItems, 0),
	"keys":       NewMapMethod("keys", _mapKeys, 0),
	"pop":        NewMapMethod("pop", _mapPop, 2),
	"popitem":    NewMapMethod("popitem", _mapPopitem, 0),
	"setdefault": NewMapMethod("setdefault", _mapSetdefault, 2),
	"update":     NewMapMethod("update", _mapUpdate, 1),
	"values":     NewMapMethod("values", _mapValues, 0),
}

func NewMapMethod(name string, meth mapMethod, numParam int) *starlark.Builtin {
	starlarkMethod := func(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		mv := b.Receiver().(*mapValue)
		return meth(mv, args, kwargs)
	}
	return starlark.NewBuiltin(name, starlarkMethod)
}

func _mapClear(mv *mapValue, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return starlark.None, nil
}

func _mapCopy(mv *mapValue, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return starlark.None, nil
}

func _mapFromkeys(mv *mapValue, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return starlark.None, nil
}

func _mapGet(mv *mapValue, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return starlark.None, nil
}

func appendTwoItemList(ls []starlark.Value, knode datamodel.Node, vnode datamodel.Node) ([]starlark.Value, error) {
	k, err := ToValue(knode)
	if err != nil {
		return nil, err
	}
	v, err := ToValue(vnode)
	if err != nil {
		return nil, err
	}
	newList, err := NewList(starlark.NewList([]starlark.Value{k, v}))
	if err != nil {
		return nil, err
	}
	return append(ls, newList), nil
}

func _mapItems(mv *mapValue, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var items []starlark.Value
	var err error

	miter := mv.node.MapIterator()
	for !miter.Done() {
		// get the key and convert to a string
		key, val, err := miter.Next()
		if err != nil {
			return starlark.None, err
		}
		keystr, err := key.AsString()
		if err != nil {
			return starlark.None, err
		}

		if repl, ok := mv.replace[keystr]; ok {
			items, err = appendTwoItemList(items, key, repl)
			if err != nil {
				return starlark.None, err
			}
			continue
		}
		items, err = appendTwoItemList(items, key, val)
		if err != nil {
			return starlark.None, err
		}
	}

	// add new keys and values to the new builder
	for keystr, val := range mv.added {
		items, err = appendTwoItemList(items, basicnode.NewString(keystr), val)
		if err != nil {
			return starlark.None, err
		}
	}

	return NewList(starlark.NewList(items))
}

func _mapKeys(mv *mapValue, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var items []starlark.Value

	miter := mv.node.MapIterator()
	for !miter.Done() {
		key, _, err := miter.Next()
		if err != nil {
			return starlark.None, err
		}
		items = append(items, MustToValue(key))
	}

	// add new keys and values to the new builder
	for keystr := range mv.added {
		items = append(items, MustToValue(basicnode.NewString(keystr)))
	}

	return NewList(starlark.NewList(items))
}

func _mapPop(mv *mapValue, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return starlark.None, nil
}

func _mapPopitem(mv *mapValue, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return starlark.None, nil
}

func _mapSetdefault(mv *mapValue, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return starlark.None, nil
}

func _mapUpdate(mv *mapValue, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return starlark.None, nil
}

func _mapValues(mv *mapValue, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var items []starlark.Value

	miter := mv.node.MapIterator()
	for !miter.Done() {
		// get the key and convert to a string
		key, val, err := miter.Next()
		if err != nil {
			return starlark.None, err
		}
		keystr, err := key.AsString()
		if err != nil {
			return starlark.None, err
		}

		if repl, ok := mv.replace[keystr]; ok {
			items = append(items, MustToValue(repl))
			continue
		}
		items = append(items, MustToValue(val))
	}

	// add new keys and values to the new builder
	for _, val := range mv.added {
		items = append(items, MustToValue(val))
	}

	return NewList(starlark.NewList(items))
}

func (v *mapValue) Attr(name string) (starlark.Value, error) {
	builtin, ok := mapMethods[name]
	if !ok {
		return starlark.None, fmt.Errorf("attribute %s not found", name)
	}
	return builtin.BindReceiver(v), nil
}

func (v *mapValue) AttrNames() []string {
	res := make([]string, 0, len(mapMethods))
	for name := range mapMethods {
		res = append(res, name)
	}
	return res
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
	// if there are no changes, just return fast
	if len(v.added) == 0 && len(v.replace) == 0 {
		return nil
	}

	// start building a new map node
	nb := v.node.Prototype().NewBuilder()
	size := v.Len()
	ma, err := nb.BeginMap(int64(size))
	if err != nil {
		return err
	}

	// iterate the contents of the previous map node
	miter := v.node.MapIterator()
	for !miter.Done() {
		// get the key and convert to a string
		key, val, err := miter.Next()
		if err != nil {
			return err
		}
		keystr, err := key.AsString()
		if err != nil {
			return err
		}

		// assign the string key to the new builder
		na := ma.AssembleKey()
		if err = na.AssignString(keystr); err != nil {
			return err
		}
		if repl, ok := v.replace[keystr]; ok {
			// if this key was replaced, use the replacement value
			na = ma.AssembleValue()
			if err = na.AssignNode(repl); err != nil {
				return err
			}
			continue
		}
		// otherwise copy the original value
		na = ma.AssembleValue()
		if err = na.AssignNode(val); err != nil {
			return err
		}
	}

	// add new keys and values to the new builder
	for keystr, val := range v.added {
		na := ma.AssembleKey()
		if err = na.AssignString(keystr); err != nil {
			return nil
		}
		na = ma.AssembleValue()
		if err = na.AssignNode(val); err != nil {
			return nil
		}
	}

	// finish up and clear the mutation maps
	err = ma.Finish()
	if err != nil {
		return err
	}
	v.node = nb.Build()
	v.added = make(map[string]datamodel.Node)
	v.replace = make(map[string]datamodel.Node)
	return nil
}
