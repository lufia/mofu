package mofu_test

import (
	"fmt"
	"slices"

	"github.com/lufia/mofu"
)

type Retriever func(key string) string

func (fn Retriever) Get(key string) string {
	return fn(key)
}

func Cook(r Retriever) {
	r.Get("key1")
	r.Get("key2")
}

func Example() {
	mock := mofu.MockFor[Retriever]()
	mock.Return("OK")

	fn, r := mock.Make()
	Cook(fn)

	fmt.Println(r.Count())
	scene := slices.Collect(r.Replay())
	scene[0](func(key string) string {
		fmt.Println(key)
		return "" // not used
	})
	scene[1](func(key string) string {
		fmt.Println(key)
		return "" // not used
	})
	// Output: 2
	// key1
	// key2
}
