package mofu

import (
	"reflect"
)

type anyMatcher int

func (anyMatcher) canAccept(arg *typeval) bool { return true }
func (anyMatcher) equal(o condExpr) bool       { return o == Any }

var _ condExpr = (anyMatcher)(0)

const (
	Any = anyMatcher(0)
)

type typeval struct {
	typ reflect.Type
	val reflect.Value
}

func (tv *typeval) canAccept(arg *typeval) bool {
	if tv.typ != arg.typ {
		return false
	}
	if tv.val.Comparable() {
		return tv.val.Equal(arg.val)
	}
	if tv.typ.Kind() == reflect.Slice {
		return tv.val.UnsafePointer() == arg.val.UnsafePointer()
	}
	if tv.typ.Kind() == reflect.Map {
		return tv.val.UnsafePointer() == arg.val.UnsafePointer()
	}
	if tv.typ.Kind() == reflect.Func {
		return tv.val.UnsafePointer() == arg.val.UnsafePointer()
	}
	return false
}

func (tv *typeval) equal(o condExpr) bool {
	p, ok := o.(*typeval)
	return ok && tv.canAccept(p)
}

var _ condExpr = (*typeval)(nil)

func newTypeval(v reflect.Value) *typeval {
	return &typeval{v.Type(), v}
}
