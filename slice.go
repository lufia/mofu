package mofu

import (
	"iter"
	"reflect"
)

type typeSlice interface {
	N() int
	Get(i int) reflect.Type
}

func allTypes(a typeSlice) iter.Seq[reflect.Type] {
	return func(yield func(reflect.Type) bool) {
		for i := range a.N() {
			if !yield(a.Get(i)) {
				break
			}
		}
	}
}

type resultTypes struct {
	fn reflect.Type
}

func (t resultTypes) N() int                 { return t.fn.NumOut() }
func (t resultTypes) Get(i int) reflect.Type { return t.fn.Out(i) }

type argTypes struct {
	fn reflect.Type
}

func (t argTypes) N() int                 { return t.fn.NumIn() }
func (t argTypes) Get(i int) reflect.Type { return t.fn.In(i) }

type valueTypes []any

func (a valueTypes) N() int                 { return len(a) }
func (a valueTypes) Get(i int) reflect.Type { return reflect.TypeOf(a[i]) }
