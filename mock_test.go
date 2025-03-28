package mofu_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/lufia/mofu"
)

func ExampleMock_Return() {
	m := mofu.MockOf(time.Now)
	m.Return(time.Date(2025, time.March, 20, 0, 0, 0, 0, time.UTC))
	now, _ := m.Make()
	fmt.Println(now().Format(time.DateTime))
	// Output: 2025-03-20 00:00:00
}

func ExampleMock_Match() {
	m := mofu.MockOf(io.ReadAll)
	m.Match(mofu.Any).Return([]byte("OK"), nil)
	readAll, _ := m.Make()
	b, _ := readAll(&bytes.Buffer{})
	fmt.Println(string(b)) // Output: OK
}

func ExampleMatcher_Return() {
	m := mofu.MockOf(os.ReadFile)
	m.Match("a.txt").Return([]byte("OK"), nil)
	m.Match("x.txt").Return(nil, errors.ErrUnsupported)
	readFile, _ := m.Make()
	s, _ := readFile("a.txt")
	fmt.Printf("%s\n", s)
	_, err := readFile("x.txt")
	fmt.Println(err)
	// Output: OK
	// unsupported operation
}
