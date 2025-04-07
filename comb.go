package mofu

import (
	"reflect"

	"github.com/ovechkin-dm/go-dyno/pkg/dyno"
)

// selector is a method selector of the interface T.
type selector[T any] struct {
	methods map[string]MockFunc
}

// MockFunc is an interface for wrapping [Mock].
type MockFunc interface {
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

// Implement implements the interface T. It is constructed of mocks.
// Each mock must be created by [MockOf] with T.Method syntax.
func Implement[T any](mocks ...MockFunc) T {
	t := reflect.TypeFor[T]()
	if t.Kind() != reflect.Interface {
		panic("type parameter T must be an interface type")
	}
	methods := make(map[string]MockFunc)
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
	s := selector[T]{methods}
	iface, err := dyno.Dynamic[T](s.handleMethod)
	if err != nil {
		panic(err)
	}
	return iface
}

// handleMethod invokes a method that matches the name of fn and its signature from among s.
func (s *selector[T]) handleMethod(meth reflect.Method, args []reflect.Value) []reflect.Value {
	m, ok := s.methods[meth.Name]
	if !ok {
		panic("no method")
	}
	fn := m.makeFunc()
	a := make([]reflect.Value, len(args)+1)
	a[0] = reflect.Zero(fn.Type().In(0))
	copy(a[1:], args)
	return fn.Call(a)
}
