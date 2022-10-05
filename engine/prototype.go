package datalarkengine

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/ipld/go-ipld-prime/datamodel"
	"github.com/ipld/go-ipld-prime/node/basicnode"
	"github.com/ipld/go-ipld-prime/schema"
	"go.starlark.net/starlark"
)

// Mode of construction, Typed or Repr or default (both)
type Mode int

const (
	AnyMode   Mode = 0
	TypedMode      = 1
	ReprMode       = 2
)

// Prototype wraps an IPLD `datamodel.NodePrototype`, and in starlark,
// is a `Callable` which acts like a constructor for that NodePrototype.
//
// There is only one Prototype type, and its behavior varies based on
// the `datamodel.NodePrototype` its bound to.
type Prototype struct {
	name string
	np   datamodel.NodePrototype
	mode Mode
}

func NewPrototype(name string, np datamodel.NodePrototype) *Prototype {
	return &Prototype{name: name, np: np, mode: AnyMode}
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

func (p *Prototype) Attr(name string) (starlark.Value, error) {
	if name == "Typed" {
		return &Prototype{name: p.name, np: p.np, mode: TypedMode}, nil
	} else if name == "Repr" {
		return &Prototype{name: p.name, np: p.np, mode: ReprMode}, nil
	}
	return starlark.None, nil
}

func (p *Prototype) AttrNames() []string {
	return []string{"Repr", "Typed"}
}

// ArgSeq represents a sequence of arguments passed into a function. The
// sequence may or may not also have a mapping from argument names to positions,
// as is the case for keyword args or for restructured args
type ArgSeq struct {
	vals []starlark.Value
	// ckey is used to store compound keys, such as a typed map with a
	// struct for a key. It is somewhat of a hack, intended to fix the
	// test `Example_mapWithStructKeys`. Ideally it shouldn't be needed, or
	// it should be generalized.
	ckey []starlark.Value
	// names is the ordered list of named keys
	names []string
	// scalar is whether the arguments is a single scalar value
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
			argseq.names = make([]string, len(keys))
			for i := 0; i < len(keys); i++ {
				argseq.names[i] = asString(keys[i])
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
		argseq.names = make([]string, len(kwargs))
		for i := 0; i < len(kwargs); i++ {
			argseq.names[i] = asString(kwargs[i][0])
			argseq.ckey[i] = kwargs[i][0]
			argseq.vals[i] = kwargs[i][1]
		}
		return argseq, nil
	default:
		// args and kwargs both empty, empty argseq already built
		return argseq, nil
	}
}

// IsSingleString returns whether the args are a single string value
func (args *ArgSeq) IsSingleString() bool {
	if len(args.names) == 0 && len(args.vals) == 1 {
		_, ok := args.vals[0].(starlark.String)
		return ok
	}
	return false
}

func asString(v starlark.Value) string {
	if str, ok := v.(starlark.String); ok {
		return string(str)
	}
	// Will stringify as a starlark value. If it were a string, quotes
	// would be added, so the above branch handles that specially.
	return v.String()
}

// how many arguments are allowed vs needed
type requireInfo struct {
	allowed int
	needed  int
}

func (ri *requireInfo) ensureValidNumFields(fieldNames []starlark.Value, argseq *ArgSeq) error {
	if ri.needed <= len(argseq.vals) && len(argseq.vals) <= ri.allowed {
		return nil
	}
	fieldList := make([]string, len(fieldNames))
	for i, val := range fieldNames {
		// TODO(dustmop): handle non-string values here
		fieldList[i] = string(val.(starlark.String))
	}
	return fmt.Errorf("expected %d values (%s), only got %d", len(fieldNames), strings.Join(fieldList, ","), len(argseq.vals))
}

// get the struct's field names and info about how many arguments are allowed vs needed
func getStructFieldInfo(structObj *schema.TypeStruct) ([]starlark.Value, *requireInfo) {
	needed := 0
	allowed := 0
	fields := structObj.Fields()
	result := make([]starlark.Value, 0, len(fields))
	for _, f := range fields {
		result = append(result, starlark.String(f.Name()))
		if f.IsOptional() {
			allowed++
		} else {
			allowed++
			needed++
		}
	}
	return result, &requireInfo{allowed: allowed, needed: needed}
}

func findMemberMatch(unionObj *schema.TypeUnion, val starlark.Value) starlark.String {
	typeName := val.Type()
	if v, ok := val.(Value); ok {
		if tn, ok := v.Node().(schema.TypedNode); ok {
			// TODO: match Repr as well
			typeName = tn.Type().Name()
		}
	}
	typeName = strings.ToLower(typeName)

	members := unionObj.Members()
	for _, m := range members {
		if strings.ToLower(m.Name()) == typeName {
			return starlark.String(m.Name())
		}
	}

	// not found
	return starlark.String("")
}

func rangeUpTo(n int) []int {
	res := make([]int, n)
	for i := range res {
		res[i] = i
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

// construct a new value with type matching the prototype, using the args for its state
func constructNewValue(p *Prototype, argseq *ArgSeq) (starlark.Value, error) {
	if tp, ok := p.np.(schema.TypedPrototype); ok {
		// construct a Typed value, such as a type-specific map or union or struct
		nb := p.np.NewBuilder()

		// state for how to construct each possible type
		var fieldNames []starlark.Value
		var ri *requireInfo

		switch it := tp.Type().(type) {
		case *schema.TypeMap:
			// typed map might be using complex keys
			fieldNames = argseq.ckey

		case *schema.TypeUnion:
			switch len(argseq.names) {
			case 0:
				// TODO: What if len(argseq.vals) == 0
				fieldNames = make([]starlark.Value, 1)
				fieldNames[0] = findMemberMatch(it, argseq.vals[0])
			case 1:
				fieldNames = []starlark.Value{starlark.String(argseq.names[0])}
			default:
				// union should only have 1 key in its map
				return starlark.None, fmt.Errorf("union must be given a map with only 1 key")
			}

		case *schema.TypeStruct:
			// struct has field names in its type
			fieldNames, ri = getStructFieldInfo(it)
			// if names were given for the arguments, use them for construction
			if argseq.names != nil {
				if err := ri.ensureValidNumFields(fieldNames, argseq); err != nil {
					return starlark.None, err
				}
				// TODO(dustmop): validate them against the struct
				fieldNames = make([]starlark.Value, len(argseq.names))
				for i, name := range argseq.names {
					fieldNames[i] = starlark.String(name)
				}
			}

		default:
			return starlark.None, fmt.Errorf("unknown type: %T", it)
		}

		// maybe can be constructed via data-kind representation agreement
		if p.mode == AnyMode || p.mode == ReprMode {
			if val, err := constructFromStringRepresentation(tp, argseq); err == nil {
				return val, nil
			}
			// ignore error because it was only the first attempt
		}

		if ri == nil {
			ri = &requireInfo{allowed: len(fieldNames), needed: len(fieldNames)}
		}

		// maybe construct using type agreement
		if p.mode == AnyMode || p.mode == TypedMode {
			val, err := constructUsingFieldsValues(nb, fieldNames, ri, argseq)
			if err == nil {
				return val, nil
			} else if p.mode == TypedMode {
				return starlark.None, err
			}
			// ignore error because there is one approach left to try
			err = nil
		}

		// TODO(dustmop): Is reqInfo supported by representation? Add a test.
		return constructAsRepresentation(tp, fieldNames, ri, argseq)
	}

	return constructBasicValue(p, argseq)
}

func constructBasicValue(p *Prototype, argseq *ArgSeq) (starlark.Value, error) {
	nb := p.np.NewBuilder()

	switch p.np.(type) {
	case basicnode.Prototype__Bool, basicnode.Prototype__Int, basicnode.Prototype__Float, basicnode.Prototype__String, basicnode.Prototype__Bytes:
		// scalar value being constucted
		if !argseq.scalar {
			return starlark.None, fmt.Errorf("wrong arguments for scalar constructor")
		}
		val := argseq.vals[0]
		if err := assembleFrom(nb, val); err != nil {
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
			if err := assembleFrom(la.AssembleValue(), val); err != nil {
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
		for i, n := range argseq.names {
			if err := assembleFrom(ma.AssembleKey(), starlark.String(n)); err != nil {
				return starlark.None, err
			}
			if err := assembleFrom(ma.AssembleValue(), argseq.vals[i]); err != nil {
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

func constructFromStringRepresentation(tp schema.TypedPrototype, argseq *ArgSeq) (starlark.Value, error) {
	// a single string representation form, such as `Alpha("beta:1")` to assign
	// the value "1" to the field "beta" of "Alpha". this is handled by the assembler
	if !argseq.IsSingleString() {
		return starlark.None, fmt.Errorf("arguments are not a single string")
	}
	nb := tp.Representation().NewBuilder()
	if err := assembleFrom(nb, argseq.vals[0]); err != nil {
		return starlark.None, err
	}
	return ToValue(nb.Build())
}

func constructAsRepresentation(tp schema.TypedPrototype, fieldNames []starlark.Value, ri *requireInfo, argseq *ArgSeq) (starlark.Value, error) {
	return constructUsingFieldsValues(tp.Representation().NewBuilder(), fieldNames, ri, argseq)
}

// assemble the node as a map of fields and values
func constructUsingFieldsValues(nb datamodel.NodeBuilder, fieldNames []starlark.Value, ri *requireInfo, argseq *ArgSeq) (starlark.Value, error) {
	if err := ri.ensureValidNumFields(fieldNames, argseq); err != nil {
		return starlark.None, err
	}
	ma, err := nb.BeginMap(int64(len(argseq.vals)))
	if err != nil {
		return starlark.None, err
	}
	for i := range fieldNames {
		if i >= len(argseq.vals) {
			break
		}
		if err := assembleFrom(ma.AssembleKey(), fieldNames[i]); err != nil {
			return starlark.None, err
		}
		if err := assembleParameter(ma, argseq.vals[i], false); err != nil {
			return starlark.None, err
		}
	}
	if err := ma.Finish(); err != nil {
		return starlark.None, err
	}

	return ToValue(nb.Build())
}

func assembleParameter(ma datamodel.MapAssembler, val starlark.Value, allowRepr bool) error {
	na := ma.AssembleValue()
	err := assembleFrom(na, val)
	// if `err` is non-nil, it may get reused below
	if err == nil {
		return nil
	}

	if v, ok := val.(Value); ok {
		if tn, ok := v.Node().(schema.TypedNode); ok {
			// try assembling using the representation
			// TODO(dustmop): Should this only be attempted if `ma` is from a
			// representation assembler?
			// TODO(dustmop): Should this block be run before the former block
			// that just tries to use `assembleFrom`? Probably harmless to run
			// that first, since it ignores any error.
			if tn.Type().RepresentationBehavior() == datamodel.Kind_String {
				str, err := tn.Representation().AsString()
				if err != nil {
					return err
				}
				return na.AssignString(str)
			}
		}
	}

	if !allowRepr {
		// reusing the `err` value from above
		return err
	}

	// if representation is allowed, get the assembler's prototype...
	tp, ok := na.Prototype().(schema.TypedPrototype)
	if !ok {
		// reusing the `err` value from above
		return err
	}

	// then take the assembler's representation and try using that
	builder := tp.Representation().NewBuilder()
	if err := assembleFrom(builder, val); err != nil {
		return err
	}
	return na.AssignNode(builder.Build())
}
