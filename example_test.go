package mofu_test

import (
	"fmt"

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
	r.Replay(0, func(key string) string {
		fmt.Println(key)
		return "" // not used
	})
	r.Replay(1, func(key string) string {
		fmt.Println(key)
		return "" // not used
	})
	// Output: 2
	// key1
	// key2
}
