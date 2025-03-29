package mofu_test

import (
	"fmt"
	"time"

	"github.com/lufia/mofu"
)

func ExampleRecorder_Count() {
	m := mofu.MockOf(time.Sleep)
	sleep, r := m.Make()
	sleep(100)
	fmt.Println(r.Count())
}

func ExampleRecorder_Replay() {
	m := mofu.MockOf(time.Sleep)
	sleep, r := m.Make()
	sleep(100 * time.Millisecond)
	for do := range r.Replay() {
		do(func(d time.Duration) {
			fmt.Println(d)
		})
	}
	// Output: 100ms
}
