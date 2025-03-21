package mofu

import (
	"testing"

	"github.com/m-mizutani/gt"
)

func TestFor(t *testing.T) {
	m := MockFor[func()]()
	gt.NotNil(t, m)
}

func TestFor_notFunc(t *testing.T) {
	defer func() {
		e := recover()
		gt.NotNil(t, e)
	}()
	MockOf(0)
}

func TestMockReturn(t *testing.T) {
	t.Run("repeat a value", func(t *testing.T) {
		m := MockFor[func() string]()
		m.Return("OK")
		fn, r := m.Make()
		gt.Equal(t, fn(), "OK")
		gt.Equal(t, fn(), "OK")
		gt.Equal(t, r.Count(), 2)
	})
	t.Run("repeat last value", func(t *testing.T) {
		m := MockFor[func() int]()
		fn, r := m.Return(1).Return(2).Make()
		gt.Equal(t, fn(), 1)
		gt.Equal(t, fn(), 2)
		gt.Equal(t, fn(), 2)
		gt.Equal(t, r.Count(), 3)
	})
	t.Run("no return", func(t *testing.T) {
		m := MockFor[func() bool]()
		fn, r := m.Make()
		gt.Equal(t, fn(), false)
		gt.Equal(t, r.Count(), 1)
	})

	t.Run("the length is less than the result's", func(t *testing.T) {
		defer func() {
			e := recover()
			gt.NotNil(t, e)
		}()
		m := MockFor[func() float64]()
		m.Return()
	})
	t.Run("the length is greater than the result's", func(t *testing.T) {
		defer func() {
			e := recover()
			gt.NotNil(t, e)
		}()
		m := MockFor[func() float64]()
		m.Return(1.0, 2.0)
	})

	t.Run("the type is not equal to the result's", func(t *testing.T) {
		defer func() {
			e := recover()
			gt.NotNil(t, e)
		}()
		m := MockFor[func() string]()
		m.Return(30)
	})

	t.Run("replay", func(t *testing.T) {
		m := MockFor[func(int)]()
		fn, r := m.Make()
		fn(100)
		for do := range r.Replay() {
			do(func(i int) {
				gt.Equal(t, i, 100)
			})
		}
	})
}

func TestMockMatch(t *testing.T) {
	t.Run("match", func(t *testing.T) {
		m := MockFor[func(int) int]()
		m.Match(10).Return(100)
		fn, r := m.Make()
		gt.Equal(t, fn(10), 100)
		gt.Equal(t, r.Count(), 1)
	})
	t.Run("match existing pattern", func(t *testing.T) {
		m := MockFor[func(int) int]()
		m.Match(10).Return(100)
		m.Match(10).Return(101)
		fn, _ := m.Make()
		gt.Equal(t, fn(10), 100)
		gt.Equal(t, fn(10), 101)
	})
	t.Run("not match", func(t *testing.T) {
		m := MockFor[func(int) int]()
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
		m := MockFor[func(int)]()
		m.Match()
	})
	t.Run("the length is greater than the argument's", func(t *testing.T) {
		defer func() {
			e := recover()
			gt.NotNil(t, e)
		}()
		m := MockFor[func(string)]()
		m.Match("a", "b")
	})

	t.Run("the type is not equal to the argument's", func(t *testing.T) {
		defer func() {
			e := recover()
			gt.NotNil(t, e)
		}()
		m := MockFor[func(string)]()
		m.Match(30)
	})
}
