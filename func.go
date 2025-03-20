// Package mofu provides utilities to create a mock function to use in test code.
package mofu

import (
	"reflect"
	"slices"
	"sync/atomic"
)

// Mock is a mock object for creating a mock function.
type Mock[T any] struct {
	fn  reflect.Value
	ret []retValue
}

type retValue struct {
	values []any
	repeat bool
}

// For returns an empty mock object.
//
// Fn is only used to specify the type of a mock function.
func For[T any](fn T) *Mock[T] {
	v := reflect.ValueOf(fn)
	if v.Type().Kind() != reflect.Func {
		panic("fn must be a function")
	}
	return &Mock[T]{
		fn: v,
	}
}

// Return adds the return values that will return them from a mock function.
func (m *Mock[T]) Return(results ...any) *Mock[T] {
	t := m.fn.Type()
	if len(results) != t.NumOut() {
		panic("number of results must exactly match to the func's results")
	}
	values := make([]any, len(results))
	for i, r := range results {
		rt := reflect.TypeOf(r)
		ft := t.Out(i)
		if rt != ft {
			panic("type differ")
		}
		values[i] = r
	}
	m.ret = append(m.ret, retValue{values, true})
	return m
}

// Make returns a mock function and its recorder.
func (m *Mock[T]) Make() (T, *Recorder[T]) {
	ret := slices.Clone(m.ret)
	var r Recorder[T]
	p := reflect.MakeFunc(m.fn.Type(), func(args []reflect.Value) []reflect.Value {
		if len(ret) == 0 {
			r.call.Add(1)
			return m.zeroReturn()
		}
		t := m.fn.Type()
		n := t.NumOut()
		a := make([]reflect.Value, n)
		off := int(r.call.Add(1) - 1)
		if off >= len(ret) {
			off = len(ret) - 1
		}
		retValues := ret[off].values
		for i, v := range retValues {
			a[i] = reflect.ValueOf(v)
		}
		return a
	})
	return p.Interface().(T), &r
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
	call atomic.Int32
	fn   T
}

// Count returns the call count of the mock function.
func (r *Recorder[T]) Count() int {
	return int(r.call.Load())
}
