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
	node datamodel.Node
}

var _ Value = (*listValue)(nil)
var _ starlark.Sequence = (*listValue)(nil)

func newListValue(node datamodel.Node) Value {
	return &listValue{node}
}

func (v *listValue) Node() datamodel.Node {
	return v.node
}
func (v *listValue) Type() string {
	// TODO(dustmop): Can a list be a TypedNode? I believe so, it
	// is used for a homogeneous typed list.
	return fmt.Sprintf("datalark.List")
}
func (v *listValue) String() string {
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
	return int(v.node.Length())
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

// methods

type listMethod func(*listValue, []starlark.Value) (starlark.Value, error)

var listMethods = map[string]*starlark.Builtin{
	"append":  NewListMethod("append", _listAppend, 1, 1), // element
	"clear":   NewListMethod("clear", _listClear, 0, 0),
	"copy":    NewListMethod("copy", _listCopy, 0, 0),
	"count":   NewListMethod("count", _listCount, 1, 1),   // value
	"extend":  NewListMethod("extend", _listExtend, 1, 1), // iterable
	"index":   NewListMethod("index", _listIndex, 1, 1),   // element
	"insert":  NewListMethod("insert", _listInsert, 2, 2), // pos, element
	"remove":  NewListMethod("remove", _listRemove, 1, 1), // element
	"reverse": NewListMethod("reverse", _listReverse, 0, 0),
	"sort":    NewListMethod("sort", _listSort, 0, 2), // ?reverse, ?key
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
