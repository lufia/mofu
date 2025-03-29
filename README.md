# mofu

Mofu provides utilities to create a mock function, like as *jest.fn*, to use in test code without any interfaces.

[![GoDev][godev-image]][godev-url]
[![Actions Status][actions-image]][actions-url]

## Usage

```go
func SUT(f func() int) {
	// ...
}
```

To test above *SUT*, you can create a mock function with *Mock.Make*.

```go
import (
	"testing"

	"github.com/lufia/mofu"
)

func TestFunc(t *testing.T) {
	m := mofu.MockFor[func() int]()
	m.ReturnOnce(1)
	fn, r := m.Make()
	SUT(fn)
	if n := r.Count(); n != 1 {
		t.Errorf("fn have been called %d times; want 1", n)
	}
}
```

*ReturnOnce* can stock multiple return values int the mock. If the return values reached to empty, the mock function returns default return values (initially zero values).

There is also *Return* method. This method can update default return values of the mock.

## Condition

If you'd like to switch return values by the function arguments, you can use *When* method.

```go
m := MockFor[func(string) int]()
m.When("foo").ReturnOnce(1)
m.When("bar").ReturnOnce(2)
fn, _ := m.Make()
fn("foo") // 1
fn("bar") // 2
fn("baz") // 0
```

[godev-image]: https://pkg.go.dev/badge/github.com/lufia/mofu
[godev-url]: https://pkg.go.dev/github.com/lufia/mofu
[actions-image]: https://github.com/lufia/mofu/actions/workflows/test.yml/badge.svg
[actions-url]: https://github.com/lufia/mofu/actions/workflow/test.yml
