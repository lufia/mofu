// Package mofu provides utilities to create a mock function to use in test code.
package mofu

import (
	"fmt"
	"iter"
	"reflect"
	"slices"
	"sync/atomic"
)

// Mock is a mock object for creating a mock function.
type Mock[T any] struct {
	fn reflect.Value

	conds []*Cond[T]
	dflt  *Cond[T]
}

// MockFor creates an empty mock object.
func MockFor[T any]() *Mock[T] {
	var fn T
	return MockOf(fn)
}

// MockOf creates an empty mock object.
//
// Fn is only used to specify the type of a mock function.
func MockOf[T any](fn T) *Mock[T] {
	v := reflect.ValueOf(fn)
	if v.Type().Kind() != reflect.Func {
		panic("fn must be a function")
	}
	m := &Mock[T]{
		fn: v,
	}
	m.dflt = &Cond[T]{
		m: m,
	}
	return m
}

// Cond represents a condition for returning values identified by the arguments.
type Cond[T any] struct {
	m       *Mock[T]
	pattern []condExpr
	retOnce [][]*typeval
}

type condExpr interface {
	canAccept(arg *typeval) bool
	equal(o condExpr) bool
}

// isCorrect reports whether args equals the expected argument pattern of c.
//
// The caller should guarantee the length of args equal to the length of args of T.
func (c *Cond[T]) isCorrect(args []*typeval) bool {
	for i, m := range c.pattern {
		if !m.canAccept(args[i]) {
			return false
		}
	}
	return true
}

func (c *Cond[T]) equalPattern(pattern []condExpr) bool {
	for i, m := range c.pattern {
		if !m.equal(pattern[i]) {
			return false
		}
	}
	return true
}

// checkTypeval checks whether v is assignable to type typ.
// If so, it returns typeval. Otherwise returns an error.
func checkTypeval(v any, typ reflect.Type) (*typeval, error) {
	switch typ.Kind() {
	case reflect.Interface:
		if v == nil {
			return newTypeval(reflect.Zero(typ)), nil
		}
		val := reflect.ValueOf(v)
		if !val.Type().Implements(typ) {
			return nil, fmt.Errorf("%s does not implement %s", val.Type(), typ)
		}
		return &typeval{typ, val}, nil
	case reflect.Pointer, reflect.Slice:
		// TODO(lufia): nilable kinds ... reflect.Chan, reflect.Func, reflect.Map, reflect.Slice
		if v == nil {
			return newTypeval(reflect.Zero(typ)), nil
		}
		val := reflect.ValueOf(v)
		if t := val.Type(); t != typ {
			return nil, fmt.Errorf("mismatched types %s and %s", typ, t)
		}
		return &typeval{typ, val}, nil
	default:
		if v == nil {
			return nil, fmt.Errorf("cannot use nil as %s value", typ)
		}
		val := reflect.ValueOf(v)
		if t := val.Type(); t != typ {
			return nil, fmt.Errorf("mismatched types %s and %s", typ, t)
		}
		return &typeval{typ, val}, nil
	}
}

func checkReturnValue(values []any, types []reflect.Type, isVariadic bool) ([]*typeval, error) {
	if len(values) == 0 && len(types) == 0 {
		return nil, nil
	}
	if isVariadic {
		types = flattenVariadicType(types, len(values))
	}
	if len(values) != len(types) {
		return nil, fmt.Errorf("number of args/results must match to the function signature")
	}
	a := make([]*typeval, len(values))
	for i, v := range values {
		p, err := checkTypeval(v, types[i])
		if err != nil {
			return nil, err
		}
		a[i] = p
	}
	return a, nil
}

func checkMatcherPattern(values []any, types []reflect.Type, isVariadic bool) ([]condExpr, error) {
	if len(values) == 0 && len(types) == 0 {
		return nil, nil
	}
	if isVariadic {
		types = flattenVariadicType(types, len(values))
	}
	if len(values) != len(types) {
		return nil, fmt.Errorf("number of args/results must match to the function signature: %v vs %v", types, values)
	}
	a := make([]condExpr, len(values))
	for i, v := range values {
		switch v := v.(type) {
		case condExpr:
			a[i] = v
		default:
			p, err := checkTypeval(v, types[i])
			if err != nil {
				return nil, err
			}
			a[i] = p
		}
	}
	return a, nil
}

func flattenVariadicType(types []reflect.Type, n int) []reflect.Type {
	if len(types) > n {
		return types
	}
	a := make([]reflect.Type, n)
	last := types[len(types)-1]
	copy(a, types[:len(types)-1])
	d := n - len(types) + 1
	copy(a[len(a)-n:], slices.Repeat([]reflect.Type{last.Elem()}, d))
	return a
}

// ReturnOnce adds the return values that will return them from a mock function.
func (c *Cond[T]) ReturnOnce(results ...any) *Cond[T] {
	t := c.m.fn.Type()
	types := collectTypes(resultTypes{t})
	a, err := checkReturnValue(results, types, false)
	if err != nil {
		panic(err)
	}
	c.retOnce = append(c.retOnce, a)
	return c
}

// ReturnOnce is like [Cond.Return] except this adds results to the default matcher.
func (m *Mock[T]) ReturnOnce(results ...any) *Mock[T] {
	m.dflt.ReturnOnce(results...)
	return m
}

// When returns a [Cond].
func (m *Mock[T]) When(args ...any) *Cond[T] {
	t := m.fn.Type()
	types := collectTypes(argTypes{t})
	pattern, err := checkMatcherPattern(args, types, t.IsVariadic())
	if err != nil {
		panic(err)
	}
	return m.registerMatcher(pattern)
}

// Make returns a mock function and its recorder.
func (m *Mock[T]) Make() (T, *Recorder[T]) {
	var r Recorder[T]
	p := reflect.MakeFunc(m.fn.Type(), func(args []reflect.Value) []reflect.Value {
		r.params = append(r.params, args)
		a := fromValues(args)
		if m.fn.Type().IsVariadic() {
			a = flattenVariadic(a)
		}
		c := m.lookupMatcher(a)
		if c == nil {
			c = m.dflt
		}
		if len(c.retOnce) == 0 {
			r.call.Add(1)
			return m.zeroReturn()
		}
		off := r.call.Add(1) - 1
		n := int64(len(c.retOnce))
		if off >= n {
			off = n - 1
		}
		return toValues(c.retOnce[off])
	})
	return p.Interface().(T), &r
}

func fromValues(values []reflect.Value) []*typeval {
	a := make([]*typeval, len(values))
	for i, v := range values {
		a[i] = newTypeval(v)
	}
	return a
}

func flattenVariadic(a []*typeval) []*typeval {
	if len(a) == 0 {
		return nil
	}
	last := a[len(a)-1].val
	s := make([]*typeval, len(a)-1+last.Len())
	n := len(a) - 1
	copy(s, a[:n])
	for i := range last.Len() {
		s[n+i] = newTypeval(last.Index(i))
	}
	return s
}

func toValues(a []*typeval) []reflect.Value {
	values := make([]reflect.Value, len(a))
	for i, v := range a {
		values[i] = v.val
	}
	return values
}

func (m *Mock[T]) lookupMatcher(args []*typeval) *Cond[T] {
	for _, c := range m.conds {
		if c.isCorrect(args) {
			return c
		}
	}
	return nil
}

func (m *Mock[T]) registerMatcher(pattern []condExpr) *Cond[T] {
	for _, c := range m.conds {
		if c.equalPattern(pattern) {
			return c
		}
	}
	c := &Cond[T]{m, pattern, nil}
	m.conds = append(m.conds, c)
	return c
}

func (m *Mock[T]) zeroReturn() []reflect.Value {
	t := m.fn.Type()
	n := t.NumOut()
	a := make([]reflect.Value, n)
	for i := range a {
		a[i] = reflect.Zero(t.Out(i))
	}
	return a
}

// Recorder records the statistics of a mock function.
type Recorder[T any] struct {
	call   atomic.Int64
	fn     T
	params [][]reflect.Value
}

// Count returns the call count of the mock function.
func (r *Recorder[T]) Count() int64 {
	return r.call.Load()
}

// Replay returns an iterator over all call logs of an mock function.
// Each call reproduces its situation with function arguments.
func (r *Recorder[T]) Replay() iter.Seq[func(T)] {
	return func(yield func(func(T)) bool) {
		for _, a := range r.params {
			do := func(fn T) {
				v := reflect.ValueOf(fn)
				v.Call(a)
			}
			if !yield(do) {
				break
			}
		}
	}
}
