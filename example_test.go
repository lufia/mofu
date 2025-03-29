package mofu_test

import (
	"fmt"
	"slices"

	"github.com/lufia/mofu"
)

type Retriever func(key string) (string, error)

func (fn Retriever) Get(key string) (string, error) {
	return fn(key)
}

func SUT(r Retriever) {
	r.Get("key1")
	r.Get("key2")
}

func Example() {
	mock := mofu.MockFor[Retriever]()
	mock.Return("OK", nil)

	fn, r := mock.Make()
	SUT(fn)

	fmt.Println(r.Count())
	scene := slices.Collect(r.Replay())
	scene[0](func(key string) (string, error) {
		fmt.Println(key)
		return "", nil // not used
	})
	scene[1](func(key string) (string, error) {
		fmt.Println(key)
		return "", nil // not used
	})
	// Output: 2
	// key1
	// key2
}
