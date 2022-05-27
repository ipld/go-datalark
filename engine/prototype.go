package datalarkengine

import (
	"fmt"
	"reflect"

	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/ipld/go-ipld-prime/node/basicnode"
	"github.com/ipld/go-ipld-prime/schema"
	"go.starlark.net/starlark"
)

// Prototype wraps an IPLD `datamodel.NodePrototype`, and in starlark,
// is a `Callable` which acts like a constructor for that NodePrototype.
//
// There is only one Prototype type, and its behavior varies based on
// the `datamodel.NodePrototype` its bound to.
type Prototype struct {
	name string
	np   datamodel.NodePrototype
}

func NewPrototype(name string, np datamodel.NodePrototype) *Prototype {
	return &Prototype{name: name, np: np}
}

func (p *Prototype) TypeName() string {
	return p.name
}

func (p *Prototype) NodePrototype() datamodel.NodePrototype {
	return p.np
}

// -- starlark.Value -->

var _ starlark.Value = (*Prototype)(nil)

func (p *Prototype) Type() string {
	if npt, ok := p.np.(schema.TypedPrototype); ok {
		return fmt.Sprintf("datalark.Prototype<%s>", npt.Type().Name())
	}
	return fmt.Sprintf("datalark.Prototype")
}
func (p *Prototype) String() string {
	return fmt.Sprintf("<built-in function %s>", p.Type())
}
func (p *Prototype) Freeze() {}
func (p *Prototype) Truth() starlark.Bool {
	return true
}
func (p *Prototype) Hash() (uint32, error) {
	return 0, nil
}

// -- starlark.Callable -->

var _ starlark.Callable = (*Prototype)(nil)

func (p *Prototype) Name() string {
	return p.String()
}

// ArgSeq represents a sequence of arguments passed into a function. The
// sequence may or may not also have a mapping from argument names to positions,
// as is the case for keyword args or for restructured args
type ArgSeq struct {
	vals   []starlark.Value
	// ckey is used to store compound keys, such as a typed map with a
	// struct for a key. It is somewhat of a hack, intended to fix the
	// test `Example_mapWithStructKeys`. Ideally it shouldn't be needed.
	ckey   []starlark.Value
	names  map[string]int
	scalar bool
}

func buildArgSeq(args starlark.Tuple, kwargs []starlark.Tuple) (*ArgSeq, error) {
	argseq := &ArgSeq{}
	switch {
	case len(args) > 0 && len(kwargs) > 0:
		return nil, fmt.Errorf("can use either positional or keyword arguments, but not both")
	case len(args) > 0:
		// positional args
		argseq.vals = make([]starlark.Value, len(args))
		for i, arg := range args {
			argseq.vals[i] = arg
		}
		if len(args) == 1 {
			argseq.scalar = true
		}
		return argseq, nil
	case len(kwargs) == 1 && asString(kwargs[0][0]) == "_":
		// restructuring as a list
		if list, ok := kwargs[0][1].(*starlark.List); ok {
			size := list.Len()
			argseq.vals = make([]starlark.Value, size)
			for i := 0; i < size; i++ {
				argseq.vals[i] = list.Index(i)
			}
			return argseq, nil
		}
		// restructuring as a dict
		if dict, ok := kwargs[0][1].(*starlark.Dict); ok {
			keys := dict.Keys()
			argseq.vals = make([]starlark.Value, len(keys))
			argseq.ckey = make([]starlark.Value, len(keys))
			argseq.names = make(map[string]int)
			for i := 0; i < len(keys); i++ {
				argseq.names[asString(keys[i])] = i
				val, _, err := dict.Get(keys[i])
				if err != nil {
					return nil, err
				}
				argseq.ckey[i] = keys[i]
				argseq.vals[i] = val
			}
			return argseq, nil
		}
		return nil, fmt.Errorf("restructuring must use a list or dict of arguments")
	case len(kwargs) > 0:
		// keyword args
		argseq.vals = make([]starlark.Value, len(kwargs))
		argseq.ckey = make([]starlark.Value, len(kwargs))
		argseq.names = make(map[string]int)
		for i := 0; i < len(kwargs); i++ {
			argseq.names[asString(kwargs[i][0])] = i
			argseq.ckey[i] = kwargs[i][0]
			argseq.vals[i] = kwargs[i][1]
		}
		return argseq, nil
	default:
		// TODO(dustmop): Missing case, args and kwargs both empty. Is
		// this always an error or is there an actual use case to support?
	}
	return nil, fmt.Errorf("TODO(dustmop): Not Implemented")
}

func asString(v starlark.Value) string {
	if str, ok := v.(starlark.String); ok {
		return string(str)
	}
	// Will stringify as a starlark value. If it were a string, quotes
	// would be added, so the above branch handles that specially.
	return v.String()
}

func getStructFieldNames(structObj *schema.TypeStruct) []string {
	fields := structObj.Fields()
	result := make([]string, 0, len(fields))
	for _, f := range fields {
		result = append(result, f.Name())
	}
	return result
}

func rangeUpTo(n int) []int {
	res := make([]int, n)
	for i := range res {
		res[i] = i
	}
	return res
}

func unifyTraversalOrder(argseq* ArgSeq, fieldNames []string) []int {
	if argseq.names == nil {
		return nil
	}
	res := make([]int, len(fieldNames))
	for i, name := range fieldNames {
		// TODO(dustmop): Handle optional / nullable values, better errors
		// Otherwise, map each name from arg to fields. The int in each
		// position of `res` tells where to find the value in `argseq.vals`
		// For example:
		//   args   (b='banana', c='cherry', a='apple')
		//   fields (a, b, c)
		// res = [2, 0, 1]
		pos := argseq.names[name]
		res[i] = pos
	}
	return res
}

func (p *Prototype) CallInternal(thread *starlark.Thread, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	// Prototype is being called with some starlark values. Determine what
	// the incoming arguments are, and use them to construct an ArgSeq.
	argseq, err := buildArgSeq(args, kwargs)
	if err != nil {
		return starlark.None, err
	}
	// construct the prototype's desired type using the ArgSeq
	return constructNewValue(p, argseq)
}

func constructNewValue(p *Prototype, argseq *ArgSeq) (starlark.Value, error) {
	nb := p.np.NewBuilder()
	if tp, ok := p.np.(schema.TypedPrototype); ok {
		// construct a Typed value, such as a type-specific map or union or struct

		ma, err := nb.BeginMap(int64(len(argseq.vals)))
		if err != nil {
			return starlark.None, err
		}

		switch it := tp.Type().(type) {
		case *schema.TypeMap:
			// construct a typed map
			for i, v := range argseq.vals {
				compoundKey := argseq.ckey[i]
				err := assembleVal(ma.AssembleKey(), compoundKey)
				if err != nil {
					return starlark.None, err
				}
				err = assembleVal(ma.AssembleValue(), v)
				if err != nil {
					return starlark.None, err
				}
			}
		case *schema.TypeUnion:
			// construct a union
			if len(argseq.vals) != 1 {
				return starlark.None, fmt.Errorf("union must be given a map with only 1 key")
			}
			for n, i := range argseq.names {
				err = assembleVal(ma.AssembleKey(), starlark.String(n))
				if err != nil {
					return starlark.None, err
				}
				err = assembleVal(ma.AssembleValue(), argseq.vals[i])
				if err != nil {
					return starlark.None, err
				}
			}
			if err := ma.Finish(); err != nil {
				return starlark.None, err
			}

		case *schema.TypeStruct:
			// construct a struct
			fieldNames := getStructFieldNames(it)

			// determine the order to apply the arguments
			argOrder := unifyTraversalOrder(argseq, fieldNames)
			if argOrder == nil {
				argOrder = rangeUpTo(len(argseq.vals))
			}

			// apply each argument by using its value to assemble a field
			for i, j := range argOrder {
				err := assembleVal(ma.AssembleKey(), starlark.String(fieldNames[i]))
				if err != nil {
					return starlark.None, err
				}
				err = assembleVal(ma.AssembleValue(), argseq.vals[j])
				if err != nil {
					return starlark.None, err
				}
			}
		}

		if err := ma.Finish(); err != nil {
			return starlark.None, err
		}

		return ToValue(nb.Build())
	}

	switch p.np.(type) {
	case basicnode.Prototype__Bool, basicnode.Prototype__Int, basicnode.Prototype__Float, basicnode.Prototype__String, basicnode.Prototype__Bytes:
		// scalar value being constucted
		if !argseq.scalar {
			return starlark.None, fmt.Errorf("wrong arguments for scalar constructor")
		}
		val := argseq.vals[0]
		if err := assembleVal(nb, val); err != nil {
			gotType := reflect.TypeOf(val).Name()
			return starlark.None, fmt.Errorf("cannot create %s from %v of type %s", p.TypeName(), val, gotType)
		}

	case basicnode.Prototype__List:
		size := len(argseq.vals)
		// list value being constructed
		// TODO(dustmop): What if a dict or kwargs are provided? Is that an
		// error, or are the key names just ignored? Figure it out and
		// add a test case.
		la, err := nb.BeginList(int64(size))
		if err != nil {
			return starlark.None, err
		}
		for _, val := range argseq.vals {
			if err := assembleVal(la.AssembleValue(), val); err != nil {
				gotType := reflect.TypeOf(val).Name()
				return starlark.None, fmt.Errorf("cannot create %s from %v of type %s", p.TypeName(), val, gotType)
			}
		}
		if err := la.Finish(); err != nil {
			return starlark.None, err
		}

	case basicnode.Prototype__Map:
		// untyped map being constructed
		if argseq.names == nil {
			return starlark.None, fmt.Errorf("no names for arguments")
		}
		ma, err := nb.BeginMap(int64(len(argseq.vals)))
		if err != nil {
			return starlark.None, err
		}
		for n, i := range argseq.names {
			v := argseq.vals[i]
			if err := assembleVal(ma.AssembleKey(), starlark.String(n)); err != nil {
				return starlark.None, err
			}
			if err := assembleVal(ma.AssembleValue(), v); err != nil {
				return starlark.None, err
			}
		}
		if err := ma.Finish(); err != nil {
			return starlark.None, err
		}

	default:
		return starlark.None, fmt.Errorf("unknown type %T", p.np)
	}

	return ToValue(nb.Build())
}
