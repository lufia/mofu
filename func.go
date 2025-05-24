package mofu

import (
	"fmt"
	"iter"
	"reflect"
	"slices"
	"sync"
)

// Mock is a mock object for creating a mock function.
type Mock[T any] struct {
	fn   reflect.Type
	name string

	conds []*Cond[T]
	dflt  *Cond[T]
}

// MockFor creates an empty mock object.
func MockFor[T any]() *Mock[T] {
	t := reflect.TypeFor[T]()
	if t.Kind() != reflect.Func {
		panic("fn must be a function")
	}
	name := ""
	if t.NumMethod() == 1 {
		name = t.Method(0).Name
	}
	return createMock[T](t, name)
}

// MockOf creates an empty mock object.
//
// Fn is only used to specify the type of a mock function.
func MockOf[T any](fn T) *Mock[T] {
	v := reflect.ValueOf(fn)
	t := v.Type()
	if t.Kind() != reflect.Func {
		panic("fn must be a function")
	}
	return createMock[T](t, funcName(v))
}

func createMock[T any](t reflect.Type, name string) *Mock[T] {
	m := &Mock[T]{
		fn:   t,
		name: name,
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
	evalq   []evaluator
	dflt    evaluator
}

type condExpr interface {
	canAccept(arg *typeval) bool
	equal(o condExpr) bool
}

type evaluator interface {
	Eval(args []reflect.Value) []reflect.Value
}

type returnValues []*typeval

func (a returnValues) Eval(_ []reflect.Value) []reflect.Value {
	values := make([]reflect.Value, len(a))
	for i, v := range a {
		values[i] = v.val
	}
	return values
}

type panicObject struct {
	v any
}

func (o panicObject) Eval(_ []reflect.Value) []reflect.Value {
	panic(o.v)
}

type evalFunc[T any] struct { // T must be a func
	fn T
}

func (f evalFunc[T]) Eval(args []reflect.Value) []reflect.Value {
	fn := reflect.ValueOf(f.fn)
	return fn.Call(args)
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
	case reflect.Pointer, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func:
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

func checkReturnValue(values []any, types []reflect.Type, isVariadic bool) (returnValues, error) {
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
	return returnValues(a), nil
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

// ReturnOnce adds the return values to the eval queue of the mock function.
func (c *Cond[T]) ReturnOnce(results ...any) *Cond[T] {
	types := collectTypes(resultTypes{c.m.fn})
	a, err := checkReturnValue(results, types, false)
	if err != nil {
		panic(err)
	}
	c.evalq = append(c.evalq, a)
	return c
}

// ReturnOnceFunc adds fn to the eval queue of the mock function.
func (c *Cond[T]) ReturnOnceFunc(fn T) *Cond[T] {
	c.evalq = append(c.evalq, &evalFunc[T]{fn})
	return c
}

// PanicOnce adds panic(v) to the eval queue of the mock function.
func (c *Cond[T]) PanicOnce(v any) *Cond[T] {
	c.evalq = append(c.evalq, &panicObject{v})
	return c
}

// Return overwrites default behavior of the mock function with results.
// It panics if any of [Cond.ReturnFunc], [Cond.Panic] and this is called two or more times.
func (c *Cond[T]) Return(results ...any) *Cond[T] {
	if c.dflt != nil {
		panic("either Return or Panic called twice for a condition")
	}
	types := collectTypes(resultTypes{c.m.fn})
	a, err := checkReturnValue(results, types, false)
	if err != nil {
		panic(err)
	}
	c.dflt = a
	return c
}

// ReturnFunc overwrites default behavior of the mock function with fn.
// It panics if any of [Cond.Return], [Cond.Panic] and this is called two or more times.
func (c *Cond[T]) ReturnFunc(fn T) *Cond[T] {
	if c.dflt != nil {
		panic("either Return or Panic called twice for a condition")
	}
	c.dflt = &evalFunc[T]{fn}
	return c
}

// Panic overwrites default behavior of the mock function with panic(v).
// It panics if any of [Cond.Return], [Cond.ReturnFunc] and this is called two or more times.
func (c *Cond[T]) Panic(v any) *Cond[T] {
	if c.dflt != nil {
		panic("either Return or Panic called twice for a condition")
	}
	c.dflt = &panicObject{v}
	return c
}

// ReturnOnce is like [Cond.ReturnOnce] except this adds the return values to the default condition.
func (m *Mock[T]) ReturnOnce(results ...any) *Mock[T] {
	m.dflt.ReturnOnce(results...)
	return m
}

// ReturnOnceFunc is like [Cond.ReturnOnceFunc] expect this adds fn to the default condition.
func (m *Mock[T]) ReturnOnceFunc(fn T) *Mock[T] {
	m.dflt.ReturnOnceFunc(fn)
	return m
}

// PanicOnce is like [Cond.PanicOnce] except this adds panic(v) to the default condition.
func (m *Mock[T]) PanicOnce(v any) *Mock[T] {
	m.dflt.PanicOnce(v)
	return m
}

// Return is like [Cond.Return] except this overwrites to the default condition.
// It panics if either [Mock.Return] or [Mock.Panic] is called two or more times.
func (m *Mock[T]) Return(results ...any) *Mock[T] {
	m.dflt.Return(results...)
	return m
}

// ReturnFunc is like [Cond.ReturnFunc] except this overwrites to the default condition.
func (m *Mock[T]) ReturnFunc(fn T) *Mock[T] {
	m.dflt.ReturnFunc(fn)
	return m
}

// Panic is like [Cond.Panic] except this overwrites to the default condition.
// It panics if either [Mock.Return] or [Mock.Panic] is called two or more times.
func (m *Mock[T]) Panic(v any) *Mock[T] {
	m.dflt.Panic(v)
	return m
}

// When returns a [Cond].
func (m *Mock[T]) When(args ...any) *Cond[T] {
	types := collectTypes(argTypes{m.fn})
	pattern, err := checkMatcherPattern(args, types, m.fn.IsVariadic())
	if err != nil {
		panic(err)
	}
	return m.registerMatcher(pattern)
}

// Make returns a mock function and its recorder.
func (m *Mock[T]) Make() (T, *Recorder[T]) {
	var r Recorder[T]
	r.nused = make(map[*Cond[T]]int)
	p := reflect.MakeFunc(m.fn, func(args []reflect.Value) []reflect.Value {
		a := fromValues(args)
		if m.fn.IsVariadic() {
			a = flattenVariadic(a)
		}
		c := m.lookupCond(a)
		if c == nil {
			c = m.dflt
		}
		off := r.nused[c]

		r.Lock()
		r.params = append(r.params, args)
		r.nused[c]++
		r.call++
		r.Unlock()

		ret := c.dflt
		n := len(c.evalq)
		if off < n {
			ret = c.evalq[off]
		}
		if ret == nil {
			return m.zeroReturn()
		}
		return ret.Eval(args)
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

func (m *Mock[T]) lookupCond(args []*typeval) *Cond[T] {
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
	c := &Cond[T]{m, pattern, nil, nil}
	m.conds = append(m.conds, c)
	return c
}

func (m *Mock[T]) zeroReturn() []reflect.Value {
	n := m.fn.NumOut()
	a := make([]reflect.Value, n)
	for i := range a {
		a[i] = reflect.Zero(m.fn.Out(i))
	}
	return a
}

// Recorder records the statistics of a mock function.
type Recorder[T any] struct {
	sync.RWMutex

	call   int64
	nused  map[*Cond[T]]int
	params [][]reflect.Value
}

// Count returns the call count of the mock function.
func (r *Recorder[T]) Count() int64 {
	r.RLock()
	defer r.RUnlock()
	return r.call
}

// Replay returns an iterator over all call logs of an mock function.
// Each call reproduces its situation with function arguments.
func (r *Recorder[T]) Replay() iter.Seq[func(T)] {
	return func(yield func(func(T)) bool) {
		r.RLock()
		params := r.params
		r.RUnlock()
		for _, a := range params {
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
