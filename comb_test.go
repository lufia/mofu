package mofu_test

import (
	"fmt"
	"io"

	"github.com/lufia/mofu"
)

type MockClient struct {
	sel *mofu.Selector
}

func (c MockClient) Read(p []byte) (int, error) {
	return mofu.Invoke(c.sel, c.Read)(p)
}

func (c MockClient) Close() error {
	return mofu.Invoke(c.sel, c.Close)()
}

func ExampleImplement() {
	read, _ := mofu.MockOf(io.Reader.Read).Return(0, io.EOF).Make()
	close, _ := mofu.MockOf(io.Closer.Close).Return(nil).Make()
	m := &MockClient{
		sel: mofu.Implement(read, close),
	}
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
