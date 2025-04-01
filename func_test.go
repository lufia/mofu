package mofu

import (
	"fmt"
	"testing"

	"github.com/m-mizutani/gt"
)

func TestMockFor(t *testing.T) {
	m := MockFor[func() int]()
	fn, r := m.Make()
	gt.Equal(t, fn(), 0)
	gt.Equal(t, r.Count(), 1)
}

func TestMockFor_notFunc(t *testing.T) {
	defer func() {
		e := recover()
		gt.NotNil(t, e)
	}()
	MockOf(0)
}

func TestMock_Return(t *testing.T) {
	t.Run("repeat a value", func(t *testing.T) {
		m := MockFor[func() string]()
		m.Return("OK")
		fn, r := m.Make()
		gt.Equal(t, fn(), "OK")
		gt.Equal(t, fn(), "OK")
		gt.Equal(t, r.Count(), 2)
	})
	t.Run("returns once values before the default value", func(t *testing.T) {
		m := MockFor[func() int]()
		fn, r := m.ReturnOnce(1).Return(2).Make()
		gt.Equal(t, fn(), 1)
		gt.Equal(t, fn(), 2)
		gt.Equal(t, fn(), 2)
		gt.Equal(t, r.Count(), 3)
	})
	t.Run("panic when Return is called twice", func(t *testing.T) {
		defer func() {
			e := recover()
			gt.NotNil(t, e)
		}()
		m := MockFor[func() int]()
		m.Return(2)
		m.Return(1)
	})
	t.Run("panic when Return is called after Panic", func(t *testing.T) {
		defer func() {
			e := recover()
			gt.NotNil(t, e)
		}()
		m := MockFor[func() int]()
		m.Panic("fake")
		m.Return(1)
	})
}

func TestMock_Panic(t *testing.T) {
	t.Run("repeat panic", func(t *testing.T) {
		m := MockFor[func() string]()
		m.Panic("hello")
		fn, r := m.Make()
		test := func() {
			defer func() {
				e := recover()
				gt.NotNil(t, e)
				gt.String(t, e.(string)).Equal("hello")
			}()
			fn()
		}
		test()
		test()
		gt.Equal(t, r.Count(), 2)
	})
	t.Run("returns once values before the default value", func(t *testing.T) {
		m := MockFor[func() int]()
		fn, r := m.PanicOnce("once").Panic("default").Make()
		test := func(s string) {
			defer func() {
				e := recover()
				gt.NotNil(t, e)
				gt.String(t, e.(string)).Equal(s)
			}()
			fn()
		}
		test("once")
		test("default")
		test("default")
		gt.Equal(t, r.Count(), 3)
	})
	t.Run("panic when Panic is called twice", func(t *testing.T) {
		defer func() {
			e := recover()
			gt.NotNil(t, e)
		}()
		m := MockFor[func()]()
		m.Panic("fake1")
		m.Panic("fake2")
	})
	t.Run("panic when Panic is called after Return", func(t *testing.T) {
		defer func() {
			e := recover()
			gt.NotNil(t, e)
		}()
		m := MockFor[func() int]()
		m.Return(1)
		m.Panic("fake")
	})
}

func TestMock_ReturnOnce(t *testing.T) {
	t.Run("returns values once", func(t *testing.T) {
		m := MockFor[func() int]()
		fn, r := m.ReturnOnce(1).Make()
		gt.Equal(t, fn(), 1)
		gt.Equal(t, fn(), 0)
		gt.Equal(t, r.Count(), 2)
	})
	t.Run("the length is less than the result's", func(t *testing.T) {
		defer func() {
			e := recover()
			gt.NotNil(t, e)
		}()
		m := MockFor[func() float64]()
		m.ReturnOnce()
	})
	t.Run("the length is greater than the result's", func(t *testing.T) {
		defer func() {
			e := recover()
			gt.NotNil(t, e)
		}()
		m := MockFor[func() float64]()
		m.ReturnOnce(1.0, 2.0)
	})

	t.Run("the type is not equal to the result's", func(t *testing.T) {
		defer func() {
			e := recover()
			gt.NotNil(t, e)
		}()
		m := MockFor[func() string]()
		m.ReturnOnce(30)
	})
}

func TestMock_PanicOnce(t *testing.T) {
	t.Run("panic once", func(t *testing.T) {
		m := MockFor[func() int]()
		fn, r := m.PanicOnce("hello").Make()
		test := func(s string) {
			defer func() {
				e := recover()
				gt.NotNil(t, e)
				gt.String(t, e.(string)).Equal(s)
			}()
			fn()
		}
		test("hello")
		gt.Equal(t, fn(), 0)
		gt.Equal(t, r.Count(), 2)
	})
}

func TestRecorder_Replay(t *testing.T) {
	t.Run("replay first record only", func(t *testing.T) {
		m := MockFor[func(int)]()
		fn, r := m.Make()
		fn(100)
		fn(200)
		for do := range r.Replay() {
			do(func(i int) {
				gt.Equal(t, i, 100)
			})
			break
		}
	})

	t.Run("replay all", func(t *testing.T) {
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

func TestMock_When(t *testing.T) {
	t.Run("match", func(t *testing.T) {
		m := MockFor[func(int) int]()
		m.When(10).ReturnOnce(100)
		fn, r := m.Make()
		gt.Equal(t, fn(10), 100)
		gt.Equal(t, r.Count(), 1)
	})
	t.Run("match existing pattern", func(t *testing.T) {
		m := MockFor[func(int) int]()
		m.When(10).ReturnOnce(100)
		m.When(10).ReturnOnce(101)
		fn, _ := m.Make()
		gt.Equal(t, fn(10), 100)
		gt.Equal(t, fn(10), 101)
	})
	t.Run("not match", func(t *testing.T) {
		m := MockFor[func(int) int]()
		m.When(10).ReturnOnce(100)
		m.ReturnOnce(2)
		fn, _ := m.Make()
		gt.Equal(t, fn(0), 2)
	})

	t.Run("the length is less than the argument's", func(t *testing.T) {
		defer func() {
			e := recover()
			gt.NotNil(t, e)
		}()
		m := MockFor[func(int)]()
		m.When()
	})
	t.Run("the length is greater than the argument's", func(t *testing.T) {
		defer func() {
			e := recover()
			gt.NotNil(t, e)
		}()
		m := MockFor[func(string)]()
		m.When("a", "b")
	})

	t.Run("the type is not equal to the argument's", func(t *testing.T) {
		defer func() {
			e := recover()
			gt.NotNil(t, e)
		}()
		m := MockFor[func(string)]()
		m.When(30)
	})

	t.Run("variadic arguments", func(t *testing.T) {
		m := MockOf(fmt.Sprint)
		m.When(1, 2).ReturnOnce("1 2")
		fn, _ := m.Make()
		gt.Equal(t, fn(1, 2), "1 2")
	})
}
