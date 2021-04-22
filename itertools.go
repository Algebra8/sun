package sun

import (
	"fmt"

	"go.starlark.net/starlark"
)

// Type that attempts to allow operations between numerics,
// i.e. float and int.
// The relationship between starlark.Int and starlark.Float
// in floatOrInt is an XOR one. That is, if f_ points to a
// starlark.Float then i_ must be a nil pointer and vice-versa.
type floatOrInt struct {
	f_ *starlark.Float
	i_ *starlark.Int
}

// Unpacker for float or int type. This allows int types and float
// types to interact with one another, e.g. count(0, 0.1).
func (p *floatOrInt) Unpack(v starlark.Value) error {
	errorMsg := "floatOrInt must have default initialization"

	switch v := v.(type) {
	case starlark.Int:
		if p.f_ != nil {
			return fmt.Errorf(errorMsg)
		}
		p.i_ = &v
		return nil
	case starlark.Float:
		if p.i_ != nil {
			return fmt.Errorf(errorMsg)
		}
		p.f_ = &v
		return nil
	}
	return fmt.Errorf("got %s, want float or int", v.Type())
}

// Adding between what may be a float or an int with what also
// may be a float or an int.
// Determining which is which is done by checking whether floatOrInt's
// starlark.Int or starlark.Float pointers are nil.
// This makes assigning to floatOrInt's values dangerous: if int is
// checked first and is not nil but the value was supposed to reflect
// a float, i.e. float is also not nil, then there will probably be an
// error downstream.
func (fi *floatOrInt) add(n floatOrInt) error {
	switch {
	// fi is int; n is int
	case fi.i_ != nil && n.i_ != nil:
		x := fi.i_.Add(*n.i_)
		fi.i_ = &x
		return nil
	// fi is int; n is float
	case fi.i_ != nil && n.f_ != nil:
		x := starlark.Float(float64(fi.i_.Float()) + float64(*n.f_))
		fi.f_ = &x
		// Note that care must be taken to erase the underlying
		// starlark.Int value of the receiver since it now represents
		// a starlark.Float
		fi.i_ = nil
		return nil
	// fi is float; n is int
	case fi.f_ != nil && n.i_ != nil:
		x := starlark.Float(float64(*fi.f_) + float64(n.i_.Float()))
		fi.f_ = &x
		return nil
	// fi is float; n is float
	case fi.f_ != nil && n.f_ != nil:
		x := starlark.Float(float64(*fi.f_) + float64(*n.f_))
		fi.f_ = &x
		return nil
	}
	return fmt.Errorf("error with addition: types are not int, float combos")
}

func (fi floatOrInt) string() string {
	switch {
	case fi.i_ != nil:
		return fi.i_.String()
	case fi.f_ != nil:
		return fi.f_.String()
	default:
		// This block should not be reached.
		// starlark's String() method is being replicated
		// so an error is not raised.
		return ""
	}
}

// Equality operator between floatOrInt and starlark's Int, Float
// and Golang's int.
func (fi *floatOrInt) eq(v interface{}) bool {
	switch v := v.(type) {
	case starlark.Int:
		if fi.i_ != nil && *fi.i_ == v {
			return true
		} else {
			return false
		}
	case starlark.Float:
		if fi.f_ != nil && *fi.f_ == v {
			return true
		} else {
			return false
		}
	case int:
		if fi.i_ == nil {
			return false
		}
		var x int
		if e := starlark.AsInt(*fi.i_, &x); e != nil {
			panic(e)
		}
		return x == v
	}

	return false
}

type countObject struct {
	cnt    floatOrInt
	step   floatOrInt
	frozen bool
}

func (co *countObject) String() string {
	// As with the cpython implementation, we don't display
	// step when it is an integer equal to 1 (default step value).
	if co.step.eq(1) {
		return fmt.Sprintf("count(%v)", co.cnt.string())
	}
	return fmt.Sprintf("count(%v, %v)", co.cnt.string(), co.step.string())
}

func (co *countObject) Type() string {
	return "itertools.count"
}

func (co *countObject) Freeze() {
	if !co.frozen {
		co.frozen = true
	}
}

func (co *countObject) Truth() starlark.Bool {
	return starlark.True
}

func (co *countObject) Hash() (uint32, error) {
	// TODO(algebra8): Implement inherited type object hash.
	return uint32(10), nil
}

func (co *countObject) Iterate() starlark.Iterator {
	return &countIter{co: co}
}

type countIter struct {
	co *countObject
}

func (c *countIter) Next(p *starlark.Value) bool {
	if c.co.frozen {
		return false
	}

	switch {
	case c.co.cnt.i_ != nil:
		*p = c.co.cnt.i_
	case c.co.cnt.f_ != nil:
		*p = c.co.cnt.f_
	}

	if e := c.co.cnt.add(c.co.step); e != nil {
		panic(e)
	}

	return true
}

func (c *countIter) Done() {}

func count_(
	thread *starlark.Thread,
	_ *starlark.Builtin,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
) (starlark.Value, error) {
	var (
		defaultStart            = starlark.MakeInt(0)
		defaultStep             = starlark.MakeInt(1)
		start        floatOrInt = floatOrInt{}
		step         floatOrInt = floatOrInt{}
	)

	if err := starlark.UnpackPositionalArgs(
		"count", args, kwargs, 0, &start, &step,
	); err != nil {
		return nil, fmt.Errorf(
			"Got %v but expected no args, or one or two valid numbers",
			args.String(),
		)
	}

	// Check if start or step require default values.
	if start.f_ == nil && start.i_ == nil {
		start.i_ = &defaultStart
	}
	if step.f_ == nil && step.i_ == nil {
		step.i_ = &defaultStep
	}

	return &countObject{cnt: start, step: step}, nil
}
