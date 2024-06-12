package obj

import (
	"context"
	"fmt"
	"io"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"gocloud.dev/gcerrors"
)

// TestStorage is a defensive method for checking to make sure that storage is
// properly configured.
func TestStorage(ctx context.Context, bucket *Bucket) error {
	testObj := "test/" + uuid.NewWithoutDashes()
	data := []byte("test")
	if err := func() (retErr error) {
		w, err := bucket.NewWriter(ctx, testObj, nil)
		if err != nil {
			return errors.Wrapf(err, "create writer")
		}
		defer func() {
			retErr = errors.Join(retErr, errors.Wrap(w.Close(), "close writer"))
		}()
		if _, err = w.Write(data); err != nil {
			return errors.Wrapf(err, "write data")
		}
		return nil
	}(); err != nil {
		return errors.Wrapf(err, "unable to write to object storage")
	}
	if err := func() (retErr error) {
		r, err := bucket.NewReader(ctx, testObj, nil)
		if err != nil {
			return errors.Wrapf(err, "create reader")
		}
		content, err := io.ReadAll(r)
		if err != nil {
			return errors.Wrapf(err, "read")
		}
		defer func() {
			retErr = errors.Join(retErr, errors.Wrap(r.Close(), "close reader"))
		}()
		if string(content) != "test" {
			return errors.New(
				fmt.Sprintf("data written and data read do not match: written: bytes read: %s", string(content)))
		}
		return nil
	}(); err != nil {
		return errors.Wrapf(err, "unable to read from object storage")
	}
	if err := bucket.Delete(ctx, testObj); err != nil {
		return errors.Wrapf(err, "unable to delete from object storage")
	}
	// Try reading a non-existent object to make sure our IsNotExist function
	// works.
	_, err := bucket.NewReader(ctx, uuid.NewWithoutDashes(), nil)
	if gcerrors.Code(err) != gcerrors.NotFound {
		return errors.Wrapf(err, "storage is unable to discern NotExist errors, should count as NotExist")
	}
	return nil
}
