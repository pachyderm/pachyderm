package obj

import (
	"context"
	"io"
)

// Client is an interface to object storage.
type Client interface {
	// Put writes the data from r to an object at name
	// It should error if the object already exists or we don't have sufficient
	// permissions to write it.
	Put(ctx context.Context, name string, r io.Reader) error

	// Get writes the data for an object to w
	// If `size == 0`, the reader should read from the offset till the end of the object.
	// It should error if the object doesn't exist or we don't have sufficient
	// permission to read it.
	Get(ctx context.Context, name string, w io.Writer) error

	// Delete deletes an object.
	// It should error if the object doesn't exist or we don't have sufficient
	// permission to delete it.
	Delete(ctx context.Context, name string) error

	// Walk calls `fn` with the names of objects which can be found under `prefix`.
	Walk(ctx context.Context, prefix string, fn func(name string) error) error

	// Exists checks if a given object already exists
	Exists(ctx context.Context, name string) (bool, error)
}
