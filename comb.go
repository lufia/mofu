package mofu

import (
	"reflect"
	"slices"
)

// Selector is a method selector of the interface T.
type Selector[T any] struct {
	methods map[string]MockInterface
}

// MockInterface is an interface for wrapping [Mock].
type MockInterface interface {
	Name() string
	funcType() reflect.Type
	makeFunc() reflect.Value
}

// Name returns name of m or empty string if m is created by an anonymous function.
func (m *Mock[T]) Name() string {
	return m.name
}

func (m *Mock[T]) funcType() reflect.Type {
	return m.fn
}

func (m *Mock[T]) makeFunc() reflect.Value {
	fn, _ := m.Make()
	return reflect.ValueOf(fn)
}

// Implement returns a [Selector] constructed of mocks.
// Each mock must be created by [MockOf] with T.Method syntax.
func Implement[T any](mocks ...MockInterface) *Selector[T] {
	t := reflect.TypeFor[T]()
	if t.Kind() != reflect.Interface {
		panic("type parameter T must be an interface type")
	}
	methods := make(map[string]MockInterface)
	for _, m := range mocks {
		name := m.Name()
		if name == "" {
			panic("implementing the interface with an unnamed mock")
		}
		if m.funcType().NumIn() < 1 {
			panic("the mock must be created with Type.Method syntax")
		}
		methods[name] = m
	}
	return &Selector[T]{methods}
}

// Invoke returns a function that matches the name of fn and its signature from among s.
func Invoke[T, I any](s *Selector[I], fn T) T {
	v := reflect.ValueOf(fn)
	if v.Type().Kind() != reflect.Func {
		panic("fn must be a function")
	}
	name := funcName(v)
	m, ok := s.methods[name]
	if !ok {
		panic("no method")
	}
	iface := reflect.TypeFor[I]()
	if !equalMethodFunc(m.funcType(), v.Type(), iface) {
		panic("signatures are different")
	}
	return createFunc(m.makeFunc(), v.Type()).(T)
}

func equalMethodFunc(m, fn, iface reflect.Type) bool {
	// m have a receiver at the first arg but fn does not have a receiver.
	a1 := collectTypes(argTypes{m})
	if len(a1) == 0 {
		panic("not a method")
	}
	if !iface.Implements(a1[0]) {
		return false
	}
	a2 := collectTypes(argTypes{fn})
	if len(a1)-1 != len(a2) || !slices.Equal(a1[1:], a2) {
		return false
	}

	r1 := collectTypes(resultTypes{m})
	r2 := collectTypes(resultTypes{fn})
	if len(r1) != len(r2) || !slices.Equal(r1, r2) {
		return false
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
