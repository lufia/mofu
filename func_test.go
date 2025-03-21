package mofu

import (
	"testing"

	"github.com/m-mizutani/gt"
)

func TestFor(t *testing.T) {
	m := For(func() {})
	gt.NotNil(t, m)
}

func TestFor_notFunc(t *testing.T) {
	defer func() {
		e := recover()
		gt.NotNil(t, e)
	}()
	For(0)
}

func TestMockReturn(t *testing.T) {
	t.Run("repeat a value", func(t *testing.T) {
		m := For(func() string { return "" })
		m.Return("OK")
		fn, r := m.Make()
		gt.Equal(t, fn(), "OK")
		gt.Equal(t, fn(), "OK")
		gt.Equal(t, r.Count(), 2)
	})
	t.Run("repeat last value", func(t *testing.T) {
		m := For(func() int { return 0 })
		fn, r := m.Return(1).Return(2).Make()
		gt.Equal(t, fn(), 1)
		gt.Equal(t, fn(), 2)
		gt.Equal(t, fn(), 2)
		gt.Equal(t, r.Count(), 3)
	})
	t.Run("no return", func(t *testing.T) {
		m := For(func() bool { return false })
		fn, r := m.Make()
		gt.Equal(t, fn(), false)
		gt.Equal(t, r.Count(), 1)
	})

	t.Run("the length is less than the result's", func(t *testing.T) {
		defer func() {
			e := recover()
			gt.NotNil(t, e)
		}()
		m := For(func() float64 { return 0.0 })
		m.Return()
	})
	t.Run("the length is greater than the result's", func(t *testing.T) {
		defer func() {
			e := recover()
			gt.NotNil(t, e)
		}()
		m := For(func() float64 { return 0.0 })
		m.Return(1.0, 2.0)
	})

	t.Run("the type is not equal to the result's", func(t *testing.T) {
		defer func() {
			e := recover()
			gt.NotNil(t, e)
		}()
		m := For(func() string { return "" })
		m.Return(30)
	})

	t.Run("replay", func(t *testing.T) {
		m := For(func(i int) {})
		fn, r := m.Make()
		fn(100)
		r.Replay(0, func(i int) {
			gt.Equal(t, i, 100)
		})
	})
	t.Run("replay but out of range", func(t *testing.T) {
		defer func() {
			e := recover()
			gt.NotNil(t, e)
		}()
		m := For(func(i int) {})
		_, r := m.Make()
		r.Replay(0, nil)
	})
}

func TestMockMatch(t *testing.T) {
	t.Run("match", func(t *testing.T) {
		m := For(func(i int) int { return 0 })
		m.Match(10).Return(100)
		fn, r := m.Make()
		gt.Equal(t, fn(10), 100)
		gt.Equal(t, r.Count(), 1)
	})
	t.Run("match existing pattern", func(t *testing.T) {
		m := For(func(i int) int { return 0 })
		m.Match(10).Return(100)
		m.Match(10).Return(101)
		fn, _ := m.Make()
		gt.Equal(t, fn(10), 100)
		gt.Equal(t, fn(10), 101)
	})
	t.Run("not match", func(t *testing.T) {
		m := For(func(i int) int { return 0 })
		m.Match(10).Return(100)
		m.Return(2)
		fn, _ := m.Make()
		gt.Equal(t, fn(0), 2)
	})

	t.Run("the length is less than the argument's", func(t *testing.T) {
		defer func() {
			e := recover()
			gt.NotNil(t, e)
		}()
		m := For(func(int) {})
		m.Match()
	})
	t.Run("the length is greater than the argument's", func(t *testing.T) {
		defer func() {
			e := recover()
			gt.NotNil(t, e)
		}()
		m := For(func(string) {})
		m.Match("a", "b")
	})

	t.Run("the type is not equal to the argument's", func(t *testing.T) {
		defer func() {
			e := recover()
			gt.NotNil(t, e)
		}()
		m := For(func(string) {})
		m.Match(30)
	})
}
