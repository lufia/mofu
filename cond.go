package mofu

import "reflect"

type anyMatcher int

func (anyMatcher) canAccept(arg *typeval) bool { return true }
func (anyMatcher) equal(o condExpr) bool       { return o == Any }
func (anyMatcher) String() string              { return "<any>" }

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
	v1 := tv.val.Interface()
	v2 := arg.val.Interface()
	return reflect.DeepEqual(v1, v2)
}

func (tv *typeval) equal(o condExpr) bool {
	p, ok := o.(*typeval)
	return ok && tv.canAccept(p)
}

var _ condExpr = (*typeval)(nil)

func newTypeval(v reflect.Value) *typeval {
	return &typeval{v.Type(), v}
}
