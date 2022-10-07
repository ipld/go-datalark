package datalarkengine

import (
	"errors"
	"fmt"

	ipldmodel "github.com/ipld/go-ipld-prime/datamodel"
	"github.com/ipld/go-ipld-prime/node/basicnode"
	"github.com/ipld/go-ipld-prime/printer"
	"github.com/ipld/go-ipld-prime/schema"
	"go.starlark.net/starlark"
)

type mapValue struct {
	node    ipldmodel.Node
	add     map[string]ipldmodel.Node
	replace map[string]ipldmodel.Node
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

func newMapValue(node ipldmodel.Node) Value {
	return &mapValue{node, nil, nil}
}

func (v *mapValue) Node() ipldmodel.Node {
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
	return int(v.node.Length()) + len(v.add)
}

// utility methods

func (v *mapValue) clear() {
	nb := v.node.Prototype().NewBuilder()
	v.node = nb.Build()
	v.add = nil
	v.replace = nil
}

// starlark.HasAttrs : starlark.Map

type mapMethod func(*mapValue, []starlark.Value) (starlark.Value, error)

var mapMethods = map[string]*starlark.Builtin{
	"clear":      NewMapMethod("clear", _mapClear, 0, 0),
	"copy":       NewMapMethod("copy", _mapCopy, 0, 0),
	"fromkeys":   NewMapMethod("fromkeys", _mapFromkeys, 1, 2),
	"get":        NewMapMethod("get", _mapGet, 1, 2),
	"items":      NewMapMethod("items", _mapItems, 0, 0),
	"keys":       NewMapMethod("keys", _mapKeys, 0, 0),
	"pop":        NewMapMethod("pop", _mapPop, 1, 2),
	"popitem":    NewMapMethod("popitem", _mapPopitem, 0, 0),
	"setdefault": NewMapMethod("setdefault", _mapSetdefault, 1, 2),
	"update":     NewMapMethod("update", _mapUpdate, 1, 1),
	"values":     NewMapMethod("values", _mapValues, 0, 0),
}

func NewMapMethod(name string, meth mapMethod, numNeed, numAllow int) *starlark.Builtin {
	starlarkMethod := func(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var first, second starlark.Value
		err := starlark.UnpackArgs(b.Name(), args, nil, "first?", &first, "second?", &second)
		if err != nil {
			return nil, err
		}
		paramList := make([]starlark.Value, 0, 2)
		if first != nil {
			paramList = append(paramList, first)
		}
		if second != nil {
			paramList = append(paramList, second)
		}
		if len(paramList) < numNeed {
			return starlark.None, fmt.Errorf("need %d parameters, got %d", numNeed, len(paramList))
		}
		if len(paramList) > numAllow {
			return starlark.None, fmt.Errorf("allows %d parameters, got %d", numAllow, len(paramList))
		}
		mv := b.Receiver().(*mapValue)
		return meth(mv, paramList)
	}
	return starlark.NewBuiltin(name, starlarkMethod)
}

func _mapClear(mv *mapValue, args []starlark.Value) (starlark.Value, error) {
	mv.clear()
	return starlark.None, nil
}

func _mapCopy(mv *mapValue, args []starlark.Value) (starlark.Value, error) {
	return starlark.None, nil
}

func _mapFromkeys(mv *mapValue, args []starlark.Value) (starlark.Value, error) {
	return starlark.None, nil
}

func _mapGet(mv *mapValue, args []starlark.Value) (starlark.Value, error) {
	return starlark.None, nil
}

func appendTwoItemListAsHost(hostList []starlark.Value, none ipldmodel.Node, ntwo ipldmodel.Node) ([]starlark.Value, error) {
	h := nodeToHost(none)
	g := nodeToHost(ntwo)
	newHostList, err := NewList(starlark.NewList([]starlark.Value{h, g}))
	if err != nil {
		return nil, err
	}
	return append(hostList, newHostList), nil
}

func _mapItems(mv *mapValue, args []starlark.Value) (starlark.Value, error) {
	var hostItems []starlark.Value
	var err error

	nodeMapIter := mv.node.MapIterator()
	for !nodeMapIter.Done() {
		// get the key and convert to a string
		nkey, nval, err := nodeMapIter.Next()
		if err != nil {
			return starlark.None, err
		}
		gstrKey, err := nkey.AsString()
		if err != nil {
			return starlark.None, err
		}

		if nodeReplace, ok := mv.replace[gstrKey]; ok {
			hostItems, err = appendTwoItemListAsHost(hostItems, nkey, nodeReplace)
			if err != nil {
				return starlark.None, err
			}
			continue
		}
		hostItems, err = appendTwoItemListAsHost(hostItems, nkey, nval)
		if err != nil {
			return starlark.None, err
		}
	}

	// add new keys and values to the new builder
	for gstrKey, nval := range mv.add {
		hostItems, err = appendTwoItemListAsHost(hostItems, basicnode.NewString(gstrKey), nval)
		if err != nil {
			return starlark.None, err
		}
	}

	return NewList(starlark.NewList(hostItems))
}

func _mapKeys(mv *mapValue, args []starlark.Value) (starlark.Value, error) {
	var hostItems []starlark.Value

	nodeMapIter := mv.node.MapIterator()
	for !nodeMapIter.Done() {
		nkey, _, err := nodeMapIter.Next()
		if err != nil {
			return starlark.None, err
		}
		hostItems = append(hostItems, nodeToHost(nkey))
	}

	// add new keys and values to the new builder
	for gstrKey := range mv.add {
		hostItems = append(hostItems, nodeToHost(basicnode.NewString(gstrKey)))
	}

	// return as a datalark.Value(*datalark.List) with starlark.Value interface
	return NewList(starlark.NewList(hostItems))
}

func _mapPop(mv *mapValue, args []starlark.Value) (starlark.Value, error) {
	return starlark.None, nil
}

func _mapPopitem(mv *mapValue, args []starlark.Value) (starlark.Value, error) {
	return starlark.None, nil
}

func _mapSetdefault(mv *mapValue, args []starlark.Value) (starlark.Value, error) {
	return starlark.None, nil
}

func _mapUpdate(mv *mapValue, args []starlark.Value) (starlark.Value, error) {
	return starlark.None, nil
}

func _mapValues(mv *mapValue, args []starlark.Value) (starlark.Value, error) {
	// all content should be datalark.Node, but using a starlark.Value interface
	var hostItems []starlark.Value

	nodeMapIter := mv.node.MapIterator()
	for !nodeMapIter.Done() {
		// get the ipld key and convert it to a go-lang string
		nkey, nval, err := nodeMapIter.Next()
		if err != nil {
			return starlark.None, err
		}
		gstrKey, err := nkey.AsString()
		if err != nil {
			return starlark.None, err
		}

		// if the value has been replaced, use the replacement
		if nodeReplace, ok := mv.replace[gstrKey]; ok {
			hostItems = append(hostItems, nodeToHost(nodeReplace))
			continue
		}
		hostItems = append(hostItems, nodeToHost(nval))
	}

	// add new keys and values to the new builder
	for _, nodeAdd := range mv.add {
		hostItems = append(hostItems, nodeToHost(nodeAdd))
	}

	// return as a datalark.Value(*datalark.List) with starlark.Value interface
	return NewList(starlark.NewList(hostItems))
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

	dv, err := starToHost(val)
	if err != nil {
		return err
	}
	node := dv.Node()

	var name string
	name, _ = starlark.AsString(nameVal)
	exist, _ := v.node.LookupByString(name)
	if exist == nil {
		if v.add == nil {
			v.add = make(map[string]ipldmodel.Node)
		}
		v.add[name] = node
	} else {
		if v.replace == nil {
			v.replace = make(map[string]ipldmodel.Node)
		}
		v.replace[name] = node
	}
	return nil
}

func (v *mapValue) applyChangesToNode() error {
	// if there are no changes, just return fast
	if len(v.add) == 0 && len(v.replace) == 0 {
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
	for gstrKey, nodeAdd := range v.add {
		na := ma.AssembleKey()
		if err = na.AssignString(gstrKey); err != nil {
			return nil
		}
		na = ma.AssembleValue()
		if err = na.AssignNode(nodeAdd); err != nil {
			return nil
		}
	}

	// finish up and clear the mutation maps
	err = ma.Finish()
	if err != nil {
		return err
	}
	v.node = nb.Build()
	v.add = make(map[string]ipldmodel.Node)
	v.replace = make(map[string]ipldmodel.Node)
	return nil
}
