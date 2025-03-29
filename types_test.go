package mofu

import (
	"testing"

	"github.com/m-mizutani/gt"
)

func TestWhen_nilable(t *testing.T) {
	t.Run("slice", func(t *testing.T) {
		m := MockFor[func([]string) int]()
		m.When(nil).ReturnOnce(1)
		fn, _ := m.Make()
		gt.Equal(t, fn(nil), 1)
	})
	t.Run("map", func(t *testing.T) {
		m := MockFor[func(map[string]bool) int]()
		m.When(nil).ReturnOnce(1)
		fn, _ := m.Make()
		gt.Equal(t, fn(nil), 1)
	})
	t.Run("chan", func(t *testing.T) {
		m := MockFor[func(chan string) int]()
		m.When(nil).ReturnOnce(1)
		fn, _ := m.Make()
		gt.Equal(t, fn(nil), 1)
	})
	t.Run("func", func(t *testing.T) {
		m := MockFor[func(func()) int]()
		m.When(nil).ReturnOnce(1)
		fn, _ := m.Make()
		gt.Equal(t, fn(nil), 1)
	})
}

func TestReturn_zeroValue(t *testing.T) {
	t.Run("slice", func(t *testing.T) {
		m := MockFor[func() []string]()
		fn, _ := m.Make()
		gt.Nil(t, fn())
	})
	t.Run("map", func(t *testing.T) {
		m := MockFor[func() map[string]bool]()
		fn, _ := m.Make()
		gt.Nil(t, fn())
	})
	t.Run("chan", func(t *testing.T) {
		m := MockFor[func() chan int]()
		fn, _ := m.Make()
		gt.Nil(t, fn())
	})
	t.Run("func", func(t *testing.T) {
		m := MockFor[func() func()]()
		fn, _ := m.Make()
		// gt.Nil does not support func
		gt.Equal(t, fn(), nil)
	})
}
