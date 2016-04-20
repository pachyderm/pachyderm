package obj

import (
	"io"

	"golang.org/x/net/context"
)

// Client is an interface to object storage.
type Client interface {
	// Writer returns a writer which writes to an object.
	// It should error if the object already exists or we don't have sufficient
	// permissions to write it.
	Writer(name string) (io.WriteCloser, error)
	// Reader returns a reader which reads from an object.
	// If `size == 0`, the reader should read from the offset till the end of the object.
	// It should error if the object doesn't exist or we don't have sufficient
	// permission to read it.
	Reader(name string, offset uint64, size uint64) (io.ReadCloser, error)
	// Delete deletes an object.
	// It should error if the object doesn't exist or we don't have sufficient
	// permission to delete it.
	Delete(name string) error
	// Walk calls `fn` with the names of objects which can be found under `prefix`.
	Walk(prefix string, fn func(name string) error) error
}

func NewGoogleClient(ctx context.Context, bucket string) (Client, error) {
	return newGoogleClient(ctx, bucket)
}

func NewAmazonClient(bucket string, id string, secret string, token string,
	region string) (Client, error) {
	return newAmazonClient(bucket, id, secret, token, region)
}
