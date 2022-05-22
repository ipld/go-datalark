package datalarkengine

import (
	"fmt"
	"reflect"

	"github.com/ipld/go-ipld-prime/datamodel"
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
	case len(kwargs) == 1 && kwargs[0][0].String() == "_":
		// restructing
		dict, ok := kwargs[0][1].(*starlark.Dict)
		if !ok {
			return nil, fmt.Errorf("restructing must use a dict of arguments")
		}
		keys := dict.Keys()
		argseq.vals = make([]starlark.Value, len(keys))
		argseq.names = make(map[string]int)
		for i := 0; i < len(keys); i++ {
			argseq.names[keys[i].String()] = i
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
			argseq.names[kwargs[i][0].String()] = i
			argseq.vals[i] = kwargs[i][1]
		}
		return argseq, nil
	default:
		// TODO(dustmop): Missing case, args and kwargs both empty. Is
		// this always an error or is there an actual use case to support?
	}
	return nil, fmt.Errorf("TODO(dustmop): Not Implemented")
}

func isScalar(p *Prototype) bool {
	// TODO(dustmop): This is bad, no need to look at stringified name from
	// the Prototype
	name := p.name
	if name == "Bool" || name == "Int" || name == "Float" || name == "String" {
		return true
	}
	return false
}

func getFieldNames(p *Prototype) []string {
	if npt, ok := p.np.(schema.TypedPrototype); ok {
		_ = npt
		// TODO(dustmop): Retrieve field names from the Type
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

	_ = getFieldNames(p)
	// TODO: What we want to be able to do here is to iterate the
	// argseq, matching each item in the sequence to fields in the
	// target node being constructed.
	return starlark.None, nil
}
