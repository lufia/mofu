package mofu_test

import (
	"fmt"
	"io"

	"github.com/lufia/mofu"
)

func ExampleImplement() {
	read := mofu.MockOf(io.Reader.Read).Return(0, io.EOF)
	close := mofu.MockOf(io.Closer.Close).Return(nil)
	m := mofu.Implement[io.ReadCloser](read, close)
	fmt.Println(Consume(m)) // Output: EOF
}

func Consume(r io.ReadCloser) error {
	defer r.Close()
	buf := make([]byte, 1<<8)
	if _, err := r.Read(buf); err != nil {
		return err
	}
	return nil
}
