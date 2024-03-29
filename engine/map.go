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
	node     ipldmodel.Node
	add      map[string]ipldmodel.Node
	addNames []string
	del      map[string]struct{}
	replace  map[string]ipldmodel.Node
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
	return &mapValue{node, nil, nil, nil, nil}
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
	skey, ok := in.(starlark.String)
	if !ok {
		return starlark.None, false, fmt.Errorf("cannot index map using %v of type %T", in, in)
	}
	name := string(skey)

	// if key has been deleted, return nil early
	if _, ok := v.del[name]; ok {
		return nil, false, nil
	}

	// look in add, replace first
	if nval, ok := v.add[name]; ok {
		return nodeToHost(nval), true, nil
	}
	if nval, ok := v.replace[name]; ok {
		return nodeToHost(nval), true, nil
	}
	// look in the ipld node
	nval, err := v.node.LookupByString(name)
	if err != nil {
		return nil, false, err
	}
	return nodeToHost(nval), true, err
}

// starlark.Sequence

func (v *mapValue) Iterate() starlark.Iterator {
	panic(fmt.Errorf("TODO(dustmop): mapValue.Iterate not implemented for %T", v))
}

func (v *mapValue) Len() int {
	return int(v.node.Length()) + len(v.add) - len(v.del)
}

// utility methods

func (v *mapValue) clear() {
	nb := v.node.Prototype().NewBuilder()
	ma, _ := nb.BeginMap(0)
	_ = ma.Finish()
	v.node = nb.Build()
	v.add = nil
	v.addNames = nil
	v.replace = nil
	v.del = nil
}

func (v *mapValue) removeKey(skey starlark.String) starlark.Value {
	name := string(skey)

	if v.add != nil {
		if node, ok := v.add[name]; ok {
			// if key had been added, remove from the add map
			delete(v.add, name)
			v.addNames = removeFromSlice(v.addNames, name)
			return nodeToHost(node)
		}
	}
	if v.replace != nil {
		if node, ok := v.replace[name]; ok {
			// if key had been replaced, remove from the replace map, and add to delete
			delete(v.replace, name)
			if v.del == nil {
				v.del = make(map[string]struct{})
			}
			v.del[name] = struct{}{}
			return nodeToHost(node)
		}
	}
	if v.del != nil {
		if _, ok := v.del[name]; ok {
			// if key had been deleted, do nothing, just return
			return nil
		}
	}

	sval, found, _ := v.Get(skey)
	if found {
		// remove the key by marking it as deleted
		name := string(skey)
		if v.del == nil {
			v.del = make(map[string]struct{})
		}
		v.del[name] = struct{}{}
		return sval
	}

	// key not found, return nil and let caller handle it
	return nil
}

func (v *mapValue) lastInsertedKey() (string, bool) {
	if len(v.addNames) > 0 {
		return v.addNames[len(v.addNames)-1], true
	}

	hasKey := false
	lastKey := ""
	nodeMapIter := v.node.MapIterator()
	for !nodeMapIter.Done() {
		nkey, _, err := nodeMapIter.Next()
		if err != nil {
			continue
		}
		name, err := nkey.AsString()
		if err != nil {
			continue
		}
		lastKey = name
		hasKey = true
	}

	return lastKey, hasKey
}

// starlark.HasAttrs : starlark.Map

type mapMethod func(*mapValue, []starlark.Value) (starlark.Value, error)

var mapMethods = map[string]*starlark.Builtin{
	"clear":      NewMapMethod("clear", mapMethodClear, 0, 0),
	"copy":       NewMapMethod("copy", mapMethodCopy, 0, 0),
	"fromkeys":   NewMapMethod("fromkeys", mapMethodFromkeys, 1, 2),
	"get":        NewMapMethod("get", mapMethodGet, 1, 2),
	"items":      NewMapMethod("items", mapMethodItems, 0, 0),
	"keys":       NewMapMethod("keys", mapMethodKeys, 0, 0),
	"pop":        NewMapMethod("pop", mapMethodPop, 1, 2),
	"popitem":    NewMapMethod("popitem", mapMethodPopitem, 0, 0),
	"setdefault": NewMapMethod("setdefault", mapMethodSetdefault, 1, 2),
	"update":     NewMapMethod("update", mapMethodUpdate, 1, 1),
	"values":     NewMapMethod("values", mapMethodValues, 0, 0),
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

func mapMethodClear(mv *mapValue, args []starlark.Value) (starlark.Value, error) {
	mv.clear()
	return starlark.None, nil
}

func mapMethodCopy(mv *mapValue, args []starlark.Value) (starlark.Value, error) {
	build := &mapValue{}
	build.node = mv.node
	if mv.add != nil {
		build.add = make(map[string]ipldmodel.Node, len(mv.add))
		build.addNames = make([]string, 0, len(mv.addNames))
		for name, nval := range mv.add {
			build.add[name] = nval
			build.addNames = append(build.addNames, name)
		}
	}
	if mv.replace != nil {
		build.replace = make(map[string]ipldmodel.Node, len(mv.replace))
		for name, nval := range mv.replace {
			build.replace[name] = nval
		}
	}
	if mv.del != nil {
		build.del = make(map[string]struct{}, len(mv.del))
		for name := range mv.del {
			build.del[name] = struct{}{}
		}
	}
	return build, nil
}

func mapMethodFromkeys(mv *mapValue, args []starlark.Value) (starlark.Value, error) {
	var skeys, svalue starlark.Value
	if err := starlark.UnpackPositionalArgs("fromkeys", args, nil, 1, &skeys, &svalue); err != nil {
		return starlark.None, err
	}

	starKeys, ok := skeys.(starlark.Iterable)
	if !ok {
		return nil, fmt.Errorf("map.update requires an iterable")
	}

	// get the default value as a datalark kind
	if svalue == nil {
		svalue = starlark.None
	}
	hostVal, err := starToHost(svalue)
	if err != nil {
		return nil, err
	}

	// start building a new map node
	nb := mv.node.Prototype().NewBuilder()
	ma, err := nb.BeginMap(0)
	if err != nil {
		return nil, err
	}

	// iterate the list of keys
	starIter := starKeys.Iterate()
	defer starIter.Done()

	var skey starlark.Value
	for starIter.Next(&skey) {
		key, ok := starlark.AsString(skey)
		if !ok {
			return nil, fmt.Errorf("could not convert key to string: %v", skey)
		}
		// construct each key value pair in the new map
		na := ma.AssembleKey()
		if err = na.AssignString(key); err != nil {
			return nil, err
		}
		na = ma.AssembleValue()
		if err = na.AssignNode(hostVal.Node()); err != nil {
			return nil, err
		}
	}
	err = ma.Finish()
	if err != nil {
		return nil, err
	}
	return newMapValue(nb.Build()), nil
}

func mapMethodGet(mv *mapValue, args []starlark.Value) (starlark.Value, error) {
	var skey, sdefault starlark.Value
	if err := starlark.UnpackPositionalArgs("get", args, nil, 1, &skey, &sdefault); err != nil {
		return starlark.None, err
	}
	// lookup value, method Get handles add,replace,del
	sval, found, err := mv.Get(skey)
	if found {
		return sval, err
	}
	// if not found, return the default param if one is given
	if sdefault != nil {
		return starToHost(sdefault)
	}
	return starlark.None, nil
}

func mapMethodItems(mv *mapValue, args []starlark.Value) (starlark.Value, error) {
	var hostItems []starlark.Value
	var err error

	nodeMapIter := mv.node.MapIterator()
	for !nodeMapIter.Done() {
		// get the key and convert to a string
		nkey, nval, err := nodeMapIter.Next()
		if err != nil {
			return starlark.None, err
		}
		name, err := nkey.AsString()
		if err != nil {
			return starlark.None, err
		}

		if _, ok := mv.del[name]; ok {
			continue
		}
		if nodeReplace, ok := mv.replace[name]; ok {
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
	for _, name := range mv.addNames {
		nval := mv.add[name]
		hostItems, err = appendTwoItemListAsHost(hostItems, basicnode.NewString(name), nval)
		if err != nil {
			return starlark.None, err
		}
	}

	return NewList(starlark.NewList(hostItems))
}

func mapMethodKeys(mv *mapValue, args []starlark.Value) (starlark.Value, error) {
	var hostItems []starlark.Value

	nodeMapIter := mv.node.MapIterator()
	for !nodeMapIter.Done() {
		nkey, _, err := nodeMapIter.Next()
		if err != nil {
			return starlark.None, err
		}
		name, err := nkey.AsString()
		if err != nil {
			return starlark.None, err
		}
		if _, ok := mv.del[name]; ok {
			continue
		}
		hostItems = append(hostItems, nodeToHost(nkey))
	}

	// add new keys and values to the new builder
	for _, name := range mv.addNames {
		hostItems = append(hostItems, nodeToHost(basicnode.NewString(name)))
	}

	// return as a datalark.Value(*datalark.List) with starlark.Value interface
	return NewList(starlark.NewList(hostItems))
}

func mapMethodPop(mv *mapValue, args []starlark.Value) (starlark.Value, error) {
	var skey starlark.String
	var sdefault starlark.Value
	if err := starlark.UnpackPositionalArgs("pop", args, nil, 1, &skey, &sdefault); err != nil {
		return starlark.None, err
	}
	sval := mv.removeKey(skey)
	if sval != nil {
		return sval, nil
	}
	if sdefault != nil {
		return sdefault, nil
	}
	return nil, fmt.Errorf("error, not found: %s", skey)
}

func mapMethodPopitem(mv *mapValue, args []starlark.Value) (starlark.Value, error) {
	name, hasKey := mv.lastInsertedKey()
	if !hasKey {
		return starlark.None, fmt.Errorf("error, not found: %s", name)
	}
	return mv.removeKey(starlark.String(name)), nil
}

func mapMethodSetdefault(mv *mapValue, args []starlark.Value) (starlark.Value, error) {
	var skey starlark.String
	var svalue starlark.Value
	if err := starlark.UnpackPositionalArgs("setdefault", args, nil, 1, &skey, &svalue); err != nil {
		return starlark.None, err
	}

	// if value exists, return it
	sval, found, err := mv.Get(skey)
	if found {
		return starToHost(sval)
	}
	if svalue == nil {
		svalue = starlark.None
	}
	// insert the default value
	err = mv.SetKey(skey, svalue)
	if err != nil {
		return starlark.None, err
	}
	// return it
	return starToHost(svalue)
}

func mapMethodUpdate(mv *mapValue, args []starlark.Value) (starlark.Value, error) {
	starObj, ok := args[0].(starlark.IterableMapping)
	if !ok {
		return nil, fmt.Errorf("map.update requires an iterable mapping")
	}

	starIter := starObj.Iterate()
	defer starIter.Done()

	var skey starlark.Value
	for starIter.Next(&skey) {
		sval, _, err := starObj.Get(skey)
		if err != nil {
			return nil, err
		}
		err = mv.SetKey(skey, sval)
		if err != nil {
			return nil, err
		}
	}

	return starlark.None, nil
}

func mapMethodValues(mv *mapValue, args []starlark.Value) (starlark.Value, error) {
	// all content should be datalark.Node, but using a starlark.Value interface
	var hostItems []starlark.Value

	nodeMapIter := mv.node.MapIterator()
	for !nodeMapIter.Done() {
		// get the ipld key and convert it to a go-lang string
		nkey, nval, err := nodeMapIter.Next()
		if err != nil {
			return starlark.None, err
		}
		name, err := nkey.AsString()
		if err != nil {
			return starlark.None, err
		}

		// if the value has been deleted, skip it
		if _, ok := mv.del[name]; ok {
			continue
		}
		// if the value has been replaced, use the replacement
		if nodeReplace, ok := mv.replace[name]; ok {
			hostItems = append(hostItems, nodeToHost(nodeReplace))
			continue
		}
		hostItems = append(hostItems, nodeToHost(nval))
	}

	// add new keys and values to the new builder
	for _, name := range mv.addNames {
		nodeAdd := mv.add[name]
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
func (v *mapValue) SetKey(starName, starVal starlark.Value) error {
	hval, err := starToHost(starVal)
	if err != nil {
		return err
	}
	node := hval.Node()

	var name string
	name, _ = starlark.AsString(starName)

	if v.add != nil {
		if _, ok := v.add[name]; ok {
			v.add[name] = node
			return nil
		}
	}
	if v.replace != nil {
		if _, ok := v.replace[name]; ok {
			v.replace[name] = node
			return nil
		}
	}
	if v.del != nil {
		if _, ok := v.del[name]; ok {
			delete(v.del, name)
		}
	}

	exist, _ := v.node.LookupByString(name)
	if exist == nil {
		if v.add == nil {
			v.add = make(map[string]ipldmodel.Node)
		}
		v.add[name] = node
		v.addNames = append(v.addNames, name)
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
	if len(v.add) == 0 && len(v.replace) == 0 && len(v.del) == 0 {
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
	nodeMapIter := v.node.MapIterator()
	for !nodeMapIter.Done() {
		// get the key and convert to a string
		nkey, nval, err := nodeMapIter.Next()
		if err != nil {
			return err
		}
		name, err := nkey.AsString()
		if err != nil {
			return err
		}

		// if this key has been deleted, skip it
		if _, ok := v.del[name]; ok {
			continue
		}

		// assign the string key to the new builder
		na := ma.AssembleKey()
		if err = na.AssignString(name); err != nil {
			return err
		}
		if nodeReplace, ok := v.replace[name]; ok {
			// if this key was replaced, use the replacement value
			na = ma.AssembleValue()
			if err = na.AssignNode(nodeReplace); err != nil {
				return err
			}
			continue
		}
		// otherwise copy the original value
		na = ma.AssembleValue()
		if err = na.AssignNode(nval); err != nil {
			return err
		}
	}

	// add new keys and values to the new builder
	for _, name := range v.addNames {
		nodeAdd := v.add[name]
		na := ma.AssembleKey()
		if err = na.AssignString(name); err != nil {
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
	v.add = nil
	v.addNames = nil
	v.replace = nil
	v.del = nil
	return nil
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

func removeFromSlice(subject []string, needle string) []string {
	for i, val := range subject {
		if val == needle {
			return append(subject[:i], subject[i+1:]...)
		}
	}
	return subject
}
