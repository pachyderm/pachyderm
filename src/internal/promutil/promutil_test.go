package promutil

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

type fakeCounter int

func (c *fakeCounter) Add(value float64) { *c = *c + fakeCounter(int(value)) }

func TestCountingIO(t *testing.T) {
	testdata := "1234567890"
	read, wrote := fakeCounter(0), fakeCounter(0)
	src, dest := strings.NewReader(testdata), new(bytes.Buffer)
	if _, err := io.Copy(&CountingWriter{Writer: dest, Counter: &wrote}, &CountingReader{Reader: src, Counter: &read}); err != nil {
		t.Fatal(err)
	}
	if got, want := int(read), len(testdata); got != want {
		t.Errorf("read:\n  got: %v\n want: %v", got, want)
	}
	if got, want := int(wrote), len(testdata); got != want {
		t.Errorf("wrote:\n  got: %v\n want: %v", got, want)
	}
	if got, want := dest.String(), testdata; got != want {
		t.Errorf("copy:\n  got: %v\n want: %v", got, want)
	}
}
