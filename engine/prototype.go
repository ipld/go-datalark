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
		// restructing
		dict, ok := kwargs[0][1].(*starlark.Dict)
		if !ok {
			return nil, fmt.Errorf("restructing must use a dict of arguments")
		}
		keys := dict.Keys()
		argseq.vals = make([]starlark.Value, len(keys))
		argseq.names = make(map[string]int)
		for i := 0; i < len(keys); i++ {
			argseq.names[asString(keys[i])] = i
			val, _, err := dict.Get(keys[i])
			if err != nil {
				return nil, err
			}
			argseq.vals[i] = val
		}
		return argseq, nil
	case len(kwargs) > 0:
		// keyword args
		argseq.vals = make([]starlark.Value, len(kwargs))
		argseq.names = make(map[string]int)
		for i := 0; i < len(kwargs); i++ {
			argseq.names[asString(kwargs[i][0])] = i
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

func isScalar(p *Prototype) bool {
	switch p.np.(type) {
	case basicnode.Prototype__Bool, basicnode.Prototype__Int, basicnode.Prototype__Float, basicnode.Prototype__String, basicnode.Prototype__Bytes:
		return true
	}
	return false
}

func isMap(p *Prototype) bool {
	switch p.np.(type) {
	case basicnode.Prototype__Map:
		return true
	}
	return false
}

func getStructFields(p *Prototype) [][]string {
	if npt, ok := p.np.(schema.TypedPrototype); ok {
		structObj, ok := npt.Type().(*schema.TypeStruct)
		if !ok {
			return nil
		}
		fields := structObj.Fields()
		result := make([][]string, 0, len(fields))
		for _, f := range fields {
			pair := []string{f.Name(), f.Type().Name()}
			result = append(result, pair)
		}
		return result
	}
	return nil
}

func (p *Prototype) CallInternal(thread *starlark.Thread, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	// Prototype is being called with some starlark values. Determine what
	// the incoming arguments are, and use that to figure out how to match
	// them to the constructor's parameters.

	// Convert them to an ArgSeq
	argseq, err := buildArgSeq(args, kwargs)
	if err != nil {
		return starlark.None, err
	}

	// Scalar values are easy
	if isScalar(p) {
		if !argseq.scalar {
			return starlark.None, fmt.Errorf("TODO: better error")
		}
		nb := p.np.NewBuilder()
		val := argseq.vals[0]
		if err := assembleVal(nb, val); err != nil {
			gotType := reflect.TypeOf(val).Name()
			return starlark.None, fmt.Errorf("cannot create %s from %v of type %s", p.TypeName(), val, gotType)
		}
		return ToValue(nb.Build())
	}

	// Handle constructing an untyped map
	if isMap(p) {
		if argseq.names == nil {
			// TODO(dustmop): Better error message
			return starlark.None, fmt.Errorf("no names for arguments")
		}
		nb := p.np.NewBuilder()
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
		return ToValue(nb.Build())
	}

	// Handle constructing a struct
	fieldPairs := getStructFields(p)
	if fieldPairs != nil {
		npt, _ := p.np.(schema.TypedPrototype)
		nb := npt.NewBuilder()
		ma, err := nb.BeginMap(int64(len(argseq.vals)))
		if err != nil {
			return starlark.None, err
		}
		// TODO(dustmop): Instead, unify the argseq's names against the field names
		// Handle things like "foo:bar" for stringpairs representation, etc.
		for i, v := range argseq.vals {
			fieldName := fieldPairs[i][0]
			err := assembleVal(ma.AssembleKey(), starlark.String(fieldName))
			if err != nil {
				return starlark.None, err
			}
			err = assembleVal(ma.AssembleValue(), v)
			if err != nil {
				// TODO(dustmop): accumulate errors instead
				return starlark.None, err
			}
		}
		if err := ma.Finish(); err != nil {
			return starlark.None, err
		}
		return ToValue(nb.Build())
	}

	return starlark.None, nil
}
