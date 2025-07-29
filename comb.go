package mofu

import (
	"reflect"

	"github.com/ovechkin-dm/go-dyno/pkg/dyno"
)

// selector is a method selector of the interface T.
type selector[T any] struct {
	methods map[string]*method
}

type method struct {
	m MockFunc
	r any // *Recorder[T]
	f reflect.Value
}

// MockFunc is an interface for wrapping [Mock].
type MockFunc interface {
	Name() string
	funcType() reflect.Type
	makeFunc() (reflect.Value, any)
}

// Name returns name of m or empty string if m is created by an anonymous function.
func (m *Mock[T]) Name() string {
	return m.name
}

func (m *Mock[T]) funcType() reflect.Type {
	return m.fn
}

func (m *Mock[T]) makeFunc() (reflect.Value, any) {
	fn, r := m.Make()
	return reflect.ValueOf(fn), r
}

// Implement implements the interface I. It is constructed of mocks.
// Each mock must be created by [MockOf] with I.Method syntax.
func Implement[I any](mocks ...MockFunc) I {
	iface, _ := ImplementInterface[I](mocks...)
	return iface
}

// Recorders is a collection of [Recorder] for the interface.
type Recorders[I any] struct {
	methods map[MockFunc]*method
}

// RecorderFor returns [Recorder] corresponds to the [Mock].
// It will panic if the method does not exist in the interface.
func RecorderFor[I, T any](r *Recorders[I], m *Mock[T]) *Recorder[T] {
	meth, ok := r.methods[m]
	if !ok {
		panic(m.Name() + ": method is not defined")
	}
	return meth.r.(*Recorder[T])
}

// Implement implements the interface I. It is constructed of mocks.
// Each mock must be created by [MockOf] with I.Method syntax.
func ImplementInterface[I any](mocks ...MockFunc) (I, *Recorders[I]) {
	t := reflect.TypeFor[I]()
	if t.Kind() != reflect.Interface {
		panic("type parameter I must be an interface type")
	}
	recorders := make(map[MockFunc]*method)
	methods := make(map[string]*method)
	for _, m := range mocks {
		name := m.Name()
		if name == "" {
			panic("implementing the interface with an unnamed mock")
		}
		if m.funcType().NumIn() < 1 {
			panic("the mock must be created with Type.Method syntax")
		}
		f, r := m.makeFunc()
		meth := &method{m, r, f}
		methods[name] = meth
		recorders[m] = meth
	}
	s := selector[I]{methods}
	iface, err := dyno.Dynamic[I](s.handleMethod)
	if err != nil {
		panic(err)
	}
	return iface, &Recorders[I]{recorders}
}

// handleMethod invokes a method that matches the name of fn and its signature from among s.
func (s *selector[T]) handleMethod(meth reflect.Method, args []reflect.Value) []reflect.Value {
	m, ok := s.methods[meth.Name]
	if !ok {
		panic("no method")
	}
	fn := m.f
	a := make([]reflect.Value, len(args)+1)
	a[0] = reflect.Zero(fn.Type().In(0))
	copy(a[1:], args)
	return fn.Call(a)
}
