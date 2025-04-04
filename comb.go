package mofu

import (
	"reflect"
)

type Selector struct {
	candidate []any
}

func Implement(a ...any) *Selector {
	return &Selector{a}
}

func Invoke[T any](s *Selector, fn T) T {
	fv := reflect.ValueOf(fn)
	for _, p := range s.candidate {
		v := reflect.ValueOf(p)
		if equalMethodFunc(v.Type(), fv.Type()) {
			return createFunc(v, fv.Type()).(T)
		}
	}
	panic("no method")
}

func equalMethodFunc(m, fn reflect.Type) bool {
	a1 := collectTypes(argTypes{m}) // m have a receiver at the first arg
	a2 := collectTypes(argTypes{fn})
	if len(a1)-1 != len(a2) || !equalTypes(a1[1:], a2) {
		return false
	}

	r1 := collectTypes(resultTypes{m})
	r2 := collectTypes(resultTypes{fn})
	if len(r1) != len(r2) || !equalTypes(r1, r2) {
		return false
	}
	return true
}

func equalTypes(a1, a2 []reflect.Type) bool {
	for i, t1 := range a1 {
		if t1 != a2[i] {
			return false
		}
	}
	return true
}

func createFunc(v reflect.Value, t reflect.Type) any {
	m := reflect.MakeFunc(t, func(a []reflect.Value) []reflect.Value {
		args := make([]reflect.Value, len(a)+1)
		args[0] = reflect.Zero(v.Type().In(0))
		copy(args[1:], a)
		return v.Call(args)
	})
	return m.Interface()
}
