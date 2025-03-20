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
	r.Get("key")
}

func Example() {
	mock := mofu.For(Retriever(nil))
	mock.Return("OK")
	fn, r := mock.Make()
	Cook(fn)
	fmt.Println(r.Count()) // Output: 1
}
