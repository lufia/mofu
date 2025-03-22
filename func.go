// Package mofu provides utilities to create a mock function to use in test code.
package mofu

import (
	"iter"
	"reflect"
	"sync/atomic"
)

// Mock is a mock object for creating a mock function.
type Mock[T any] struct {
	fn          reflect.Value
	matchers    []*Matcher[T]
	dfltMatcher *Matcher[T]
}

// MockFor returns an empty mock object.
func MockFor[T any]() *Mock[T] {
	var fn T
	return MockOf(fn)
}

// MockOf returns an empty mock object.
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
	m.dfltMatcher = &Matcher[T]{
		m: m,
	}
	return m
}

// Matcher represents a matcher for a mock function arguments.
type Matcher[T any] struct {
	m       *Mock[T]
	pattern []any
	ret     []retValue
}

type retValue struct {
	values []any
	repeat bool
}

// The caller should guarantee the length of args equal to the length of args of T.
func (c *Matcher[T]) match(args []reflect.Value) bool {
	for i, v1 := range args {
		v2 := reflect.ValueOf(c.pattern[i])
		if v1.Type() != v2.Type() {
			return false
		}
		if !v1.Equal(v2) {
			return false
		}
	}
	return true
}

// Return adds the return values that will return them from a mock function.
func (c *Matcher[T]) Return(results ...any) *Matcher[T] {
	t := c.m.fn.Type()
	if len(results) != t.NumOut() {
		panic("number of results must exactly match to the func's results")
	}
	t1 := collectTypes(valueTypes(results))
	signature := collectTypes(resultTypes{t})
	if !typesSatisfy(t1, signature) {
		panic("type differ")
	}
	c.ret = append(c.ret, retValue{results, true})
	return c
}

// Return is like [Matcher.Return] except this adds results to the default matcher.
func (m *Mock[T]) Return(results ...any) *Mock[T] {
	m.dfltMatcher.Return(results...)
	return m
}

// Match returns a [Matcher].
func (m *Mock[T]) Match(args ...any) *Matcher[T] {
	t := m.fn.Type()
	signature := collectTypes(argTypes{t})
	if c := m.lookupMatcher(toValues(args, signature)); c != nil {
		return c
	}

	// TODO(lufia): variadic parameter (t.IsValiadic() == true)
	if len(args) != t.NumIn() {
		panic("number of args must exactly match to the func's args")
	}
	t1 := collectTypes(valueTypes(args))
	if !typesSatisfy(t1, signature) {
		panic("type differ")
	}
	c := &Matcher[T]{m, args, nil}
	m.matchers = append(m.matchers, c)
	return c
}

// typesSatisfy returns whether each item 's type of valueTypes are same corresponding to signatureTypes's.
//
// If the length of valueTypes is not equal to the length of signatureTypes,
// typesSatisfy compares only items during 0 to len(signatureTypes).
func typesSatisfy(valueTypes, signatureTypes []reflect.Type) bool {
	// TODO(lufia): only s2 <= s1
	for i, t := range signatureTypes {
		switch t.Kind() {
		case reflect.Interface:
			if valueTypes[i] != nil && !valueTypes[i].Implements(t) {
				return false
			}
		case reflect.Slice:
			if valueTypes[i] != nil && valueTypes[i] != t {
				return false
			}
		default:
			if valueTypes[i] != t {
				return false
			}
		}
	}
	return true
}

// Make returns a mock function and its recorder.
func (m *Mock[T]) Make() (T, *Recorder[T]) {
	t := m.fn.Type()
	signature := collectTypes(resultTypes{t})
	var r Recorder[T]
	p := reflect.MakeFunc(m.fn.Type(), func(args []reflect.Value) []reflect.Value {
		r.params = append(r.params, args)
		c := m.lookupMatcher(args)
		if c == nil {
			c = m.dfltMatcher
		}
		if len(c.ret) == 0 {
			r.call.Add(1)
			return m.zeroReturn()
		}
		off := int(r.call.Add(1) - 1)
		if off >= len(c.ret) {
			off = len(c.ret) - 1
		}
		return toValues(c.ret[off].values, signature)
	})
	return p.Interface().(T), &r
}

func (m *Mock[T]) lookupMatcher(args []reflect.Value) *Matcher[T] {
	for _, c := range m.matchers {
		if c.match(args) {
			return c
		}
	}
	return nil
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

func toValues(values []any, signatureTypes []reflect.Type) []reflect.Value {
	a := make([]reflect.Value, len(values))
	for i, v := range values {
		if v == nil {
			a[i] = reflect.Zero(signatureTypes[i])
		} else {
			a[i] = reflect.ValueOf(v)
		}
	}
	return a
}
