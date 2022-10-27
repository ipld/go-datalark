package datalarkengine

import (
	"errors"
	"fmt"

	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/ipld/go-ipld-prime/node/basicnode"
	"github.com/ipld/go-ipld-prime/printer"
	"go.starlark.net/starlark"
)

type listValue struct {
	node   datamodel.Node
	suffix []datamodel.Node
}

var (
	_ Value              = (*listValue)(nil)
	_ starlark.Indexable = (*listValue)(nil)
	_ starlark.Sequence  = (*listValue)(nil)
)

func newListValue(node datamodel.Node) Value {
	return &listValue{node, nil}
}

func (v *listValue) Node() datamodel.Node {
	v.applyChangesToNode()
	return v.node
}
func (v *listValue) Type() string {
	// TODO(dustmop): Can a list be a TypedNode? I believe so, it
	// is used for a homogeneous typed list.
	return fmt.Sprintf("datalark.List")
}
func (v *listValue) String() string {
	v.applyChangesToNode()
	return printer.Sprint(v.node)
}
func (v *listValue) Freeze() {}
func (v *listValue) Truth() starlark.Bool {
	return true
}
func (v *listValue) Hash() (uint32, error) {
	return 0, errors.New("TODO")
}

// NewList converts a starlark.List into a datalark.Value
func NewList(starList *starlark.List) (Value, error) {
	nb := basicnode.Prototype.List.NewBuilder()
	size := starList.Len()
	la, err := nb.BeginList(int64(size))
	if err != nil {
		return nil, err
	}
	for i := 0; i < size; i++ {
		item := starList.Index(i)
		if err := assembleFrom(la.AssembleValue(), item); err != nil {
			return nil, fmt.Errorf("cannot add %v of type %T", item, item)
		}
	}
	if err := la.Finish(); err != nil {
		return nil, err
	}
	return newListValue(nb.Build()), nil
}

// starlark.Sequence

func (v *listValue) Iterate() starlark.Iterator {
	panic(fmt.Errorf("TODO(dustmop): listValue.Iterate not implemented for %T", v))
}

func (v *listValue) Len() int {
	return int(v.node.Length()) + len(v.suffix)
}

// starlark.Indexable

func (v *listValue) Index(i int) starlark.Value {
	totalLen := int(v.node.Length()) + len(v.suffix)
	if i >= totalLen {
		panic(fmt.Errorf("index out of range, index = %d, len = %d", i, totalLen))
	}
	if i < int(v.node.Length()) {
		item, err := v.node.LookupByIndex(int64(i))
		if err != nil {
			panic(err)
		}
		return nodeToHost(item)
	}
	j := i - int(v.node.Length())
	return nodeToHost(v.suffix[j])
}

// starlark.HasAttrs : starlark.List

func (v *listValue) Attr(name string) (starlark.Value, error) {
	builtin, ok := listMethods[name]
	if !ok {
		return starlark.None, fmt.Errorf("attribute %s not found", name)
	}
	return builtin.BindReceiver(v), nil
}

func (v *listValue) AttrNames() []string {
	res := make([]string, 0, len(listMethods))
	for name := range listMethods {
		res = append(res, name)
	}
	return res
}

// utility

func (v *listValue) clear() {
	nb := v.node.Prototype().NewBuilder()
	la, _ := nb.BeginList(0)
	_ = la.Finish()
	v.node = nb.Build()
	v.suffix = nil
}

// methods

type listMethod func(*listValue, []starlark.Value) (starlark.Value, error)

var listMethods = map[string]*starlark.Builtin{
	"append":  NewListMethod("append", _listAppend, 1, 1),
	"clear":   NewListMethod("clear", _listClear, 0, 0),
	"copy":    NewListMethod("copy", _listCopy, 0, 0),
	"count":   NewListMethod("count", _listCount, 1, 1),
	"extend":  NewListMethod("extend", _listExtend, 1, 1),
	"index":   NewListMethod("index", _listIndex, 1, 1),
	"insert":  NewListMethod("insert", _listInsert, 2, 2),
	"remove":  NewListMethod("remove", _listRemove, 1, 1),
	"reverse": NewListMethod("reverse", _listReverse, 0, 0),
	"sort":    NewListMethod("sort", _listSort, 0, 2),
}

func NewListMethod(name string, meth listMethod, numNeed, numAllow int) *starlark.Builtin {
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
		mv := b.Receiver().(*listValue)
		return meth(mv, paramList)
	}
	return starlark.NewBuiltin(name, starlarkMethod)
}

func _listAppend(lv *listValue, args []starlark.Value) (starlark.Value, error) {
	hostItem, err := starToHost(args[0])
	if err != nil {
		return nil, err
	}
	lv.suffix = append(lv.suffix, hostItem.Node())
	return starlark.None, nil
}

func _listClear(lv *listValue, args []starlark.Value) (starlark.Value, error) {
	lv.clear()
	return starlark.None, nil
}

func _listCopy(lv *listValue, args []starlark.Value) (starlark.Value, error) {
	build := make([]datamodel.Node, len(lv.suffix))
	for i := 0; i < len(lv.suffix); i++ {
		build[i] = lv.suffix[i]
	}
	return &listValue{lv.node, build}, nil
}

func _listCount(lv *listValue, args []starlark.Value) (starlark.Value, error) {
	var elem starlark.Value
	err := starlark.UnpackArgs("count", args, nil, "elem", &elem)
	if err != nil {
		return nil, err
	}
	hostElem, err := starToHost(elem)
	if err != nil {
		return nil, err
	}
	count := 0
	nodeFind := hostElem.Node()
	iter := lv.node.ListIterator()
	for !iter.Done() {
		_, nodeItem, err := iter.Next()
		if err != nil {
			return nil, err
		}
		if datamodel.DeepEqual(nodeItem, nodeFind) {
			count++
		}
	}
	for _, nodeItem := range lv.suffix {
		if datamodel.DeepEqual(nodeItem, nodeFind) {
			count++
		}
	}
	return NewInt(int64(count)), nil
}

func _listExtend(lv *listValue, args []starlark.Value) (starlark.Value, error) {
	var svals starlark.Value
	if err := starlark.UnpackPositionalArgs("extend", args, nil, 1, &svals); err != nil {
		return starlark.None, err
	}

	siterable, ok := svals.(starlark.Iterable)
	if !ok {
		return nil, fmt.Errorf("list.extend requires an iterable")
	}
	starIter := siterable.Iterate()

	var starElem starlark.Value
	for starIter.Next(&starElem) {
		hostItem, err := starToHost(starElem)
		if err != nil {
			return nil, err
		}
		lv.suffix = append(lv.suffix, hostItem.Node())
	}

	return starlark.None, nil
}

func _listIndex(lv *listValue, args []starlark.Value) (starlark.Value, error) {
	var elem starlark.Value
	err := starlark.UnpackArgs("count", args, nil, "elem", &elem)
	if err != nil {
		return nil, err
	}
	hostElem, err := starToHost(elem)
	if err != nil {
		return nil, err
	}
	nodeFind := hostElem.Node()
	iter := lv.node.ListIterator()
	for !iter.Done() {
		i, nodeItem, err := iter.Next()
		if err != nil {
			return nil, err
		}
		if datamodel.DeepEqual(nodeItem, nodeFind) {
			return NewInt(i), nil
		}
	}
	for i, nodeItem := range lv.suffix {
		if datamodel.DeepEqual(nodeItem, nodeFind) {
			return NewInt(int64(i) + lv.node.Length()), nil
		}
	}
	return NewInt(-1), nil
}

func _listInsert(lv *listValue, args []starlark.Value) (starlark.Value, error) {
	var sindex starlark.Int
	var selem starlark.Value
	if err := starlark.UnpackPositionalArgs("insert", args, nil, 2, &sindex, &selem); err != nil {
		return starlark.None, err
	}

	index, ok := sindex.Int64()
	if !ok {
		return nil, fmt.Errorf("insert index invalid: %v", sindex)
	}

	if index < lv.node.Length() {
		// if index is within the already built ipld.Node, split the
		// node into prior elements, and remaining elements
		node, remain, err := lv.splitNodeAtIndex(index)
		if err != nil {
			return nil, err
		}

		lv.node = node
		lv.suffix = append(remain, lv.suffix...)
	}

	// going to insert by considering only the suffix slice
	afterIndex := int(index - lv.node.Length())
	hostItem, err := starToHost(selem)
	if err != nil {
		return nil, err
	}

	// rebuild the suffix, inserting the element when appropriate
	newSuffix := make([]datamodel.Node, 0, len(lv.suffix)+1)
	for i, nodeElem := range lv.suffix {
		if i == afterIndex {
			newSuffix = append(newSuffix, hostItem.Node())
		}
		newSuffix = append(newSuffix, nodeElem)
	}
	if afterIndex == len(lv.suffix) {
		newSuffix = append(newSuffix, hostItem.Node())
	}

	lv.suffix = newSuffix
	return starlark.None, nil
}

func _listRemove(lv *listValue, args []starlark.Value) (starlark.Value, error) {
	return nil, nil
}

func _listReverse(lv *listValue, args []starlark.Value) (starlark.Value, error) {
	return nil, nil
}

func _listSort(lv *listValue, args []starlark.Value) (starlark.Value, error) {
	return nil, nil
}

func (v *listValue) splitNodeAtIndex(splitIndex int64) (datamodel.Node, []datamodel.Node, error) {
	nb := basicnode.Prototype.List.NewBuilder()
	la, err := nb.BeginList(splitIndex)
	if err != nil {
		return nil, nil, err
	}
	remain := make([]datamodel.Node, 0)

	iter := v.node.ListIterator()
	for !iter.Done() {
		i, nodeItem, err := iter.Next()
		if err != nil {
			return nil, nil, err
		}
		if i < splitIndex {
			if err := la.AssembleValue().AssignNode(nodeItem); err != nil {
				return nil, nil, err
			}
			continue
		}
		remain = append(remain, nodeItem)
	}

	if err := la.Finish(); err != nil {
		return nil, nil, err
	}

	return nb.Build(), remain, nil
}

func (v *listValue) applyChangesToNode() error {
	if len(v.suffix) == 0 {
		return nil
	}

	nb := basicnode.Prototype.List.NewBuilder()
	size := int(v.node.Length()) + len(v.suffix)
	la, err := nb.BeginList(int64(size))
	if err != nil {
		return err
	}

	iter := v.node.ListIterator()
	for !iter.Done() {
		_, nodeItem, err := iter.Next()
		if err != nil {
			return err
		}
		if err := la.AssembleValue().AssignNode(nodeItem); err != nil {
			return err
		}
	}

	for _, nodeItem := range v.suffix {
		if err := la.AssembleValue().AssignNode(nodeItem); err != nil {
			return err
		}
	}

	if err := la.Finish(); err != nil {
		return err
	}

	v.node = nb.Build()
	v.suffix = nil
	return nil
}
