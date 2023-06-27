// Package kv provides a Key-Value Store interface and a few implementations.
package kv

import (
	"bytes"
	"context"

	"github.com/pachyderm/pachyderm/v2/src/internal/stream"
)

// ValueCallback is the type of functions used to access values
// It is never ok for the callback to retain the data.
// If it were ok, then the callee would just return the data.
type ValueCallback = func([]byte) error

type Getter interface {
	// Get looks up the value that corresponds to key and write the value into buf.
	// If buf is too small for the value then io.ErrShortBuffer is returned.
	Get(ctx context.Context, key []byte, buf []byte) (int, error)
}

type Putter interface {
	// Put creates an entry mapping key to value, overwriting any previous mapping.
	Put(ctx context.Context, key, value []byte) error
}

type GetPut interface {
	Getter
	Putter
}

type Deleter interface {
	// Delete removes the entry at key.
	// If there is no entry Delete returns nil
	Delete(ctx context.Context, key []byte) error
}

type KeyIterable interface {
	// NewKeyIterator returns a new iterator which will cover span.
	NewKeyIterator(span Span) stream.Iterator[[]byte]
}

// Store is a key-value store
type Store interface {
	Getter
	Putter
	Deleter
	Exists(ctx context.Context, key []byte) (bool, error)

	KeyIterable
}

// Span is a range of bytes from Begin inclusive to End exclusive
// As a special case if End == nil, then the span has no upper bound.
type Span struct {
	Begin []byte
	End   []byte
}

// Contains returns true if the Span contains k
func (s Span) Contains(k []byte) bool {
	if bytes.Compare(s.Begin, k) > 0 {
		return false
	}
	if s.End != nil && bytes.Compare(s.End, k) <= 0 {
		return false
	}
	return true
}

func SpanFromPrefix(prefix []byte) Span {
	return Span{
		Begin: prefix,
		End:   PrefixEnd(prefix),
	}
}

// PrefixEnd returns the key > all the keys with prefix p, but < any other key
func PrefixEnd(prefix []byte) []byte {
	if len(prefix) == 0 {
		return nil
	}
	var end []byte
	for i := len(prefix) - 1; i >= 0; i-- {
		c := prefix[i]
		if c < 0xff {
			end = make([]byte, i+1)
			copy(end, prefix)
			end[i] = c + 1
			break
		}
	}
	return end
}

// KeyAfter returns a byte slice ordered immediately after x lexicographically
// the motivating use case is iteration.
func KeyAfter(x []byte) []byte {
	y := append([]byte{}, x...)
	y = append(y, 0x00)
	return y
}
