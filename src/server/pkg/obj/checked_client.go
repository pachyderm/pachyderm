package obj

import (
	"context"
	"io"

	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
)

type checkedWriteCloser struct {
	io.WriteCloser
}

func newCheckedWriteCloser(wc io.WriteCloser) io.WriteCloser {
	return &checkedWriteCloser{
		WriteCloser: wc,
	}
}

func (cwc *checkedWriteCloser) Write(data []byte) (_ int, retErr error) {
	defer func() {
		retErr = errors.WithStack(retErr)
	}()
	return cwc.WriteCloser.Write(data)
}

func (cwc *checkedWriteCloser) Close() (retErr error) {
	defer func() {
		retErr = errors.WithStack(retErr)
	}()
	return cwc.WriteCloser.Close()
}

type checkedReadCloser struct {
	io.ReadCloser
	size  uint64
	count uint64
}

func newCheckedReadCloser(size uint64, rc io.ReadCloser) io.ReadCloser {
	return &checkedReadCloser{
		ReadCloser: rc,
		size:       size,
	}
}

func (crc *checkedReadCloser) Read(p []byte) (_ int, retErr error) {
	defer func() {
		if retErr != io.EOF {
			retErr = errors.WithStack(retErr)
		}
	}()
	if crc.size == 0 {
		return crc.ReadCloser.Read(p)
	}
	count, err := crc.ReadCloser.Read(p)
	crc.count += uint64(count)
	if err != nil {
		if errors.Is(err, io.EOF) && crc.count != crc.size {
			return count, errors.Errorf("read stream ended after the wrong length, expected: %d, actual: %d", crc.size, crc.count)
		} else if crc.count > crc.size {
			return count, errors.Wrapf(err, "read stream errored but also read more bytes than requested, expected: %d, actual: %d", crc.size, crc.count)
		}
		return count, err
	} else if crc.count > crc.size {
		return count, errors.Errorf("read stream read more bytes than requested, expected: %d, actual: %d", crc.size, crc.count)
	}
	return count, nil
}

func (crc *checkedReadCloser) Close() (retErr error) {
	defer func() {
		retErr = errors.WithStack(retErr)
	}()
	return crc.ReadCloser.Close()
}

var _ Client = &checkedClient{}

type checkedClient struct {
	Client
}

func newCheckedClient(c Client) Client {
	if c == nil {
		return nil
	}
	return &checkedClient{Client: c}
}

func (cc *checkedClient) Writer(ctx context.Context, name string) (_ io.WriteCloser, retErr error) {
	defer func() {
		retErr = errors.WithStack(retErr)
	}()
	wc, err := cc.Client.Writer(ctx, name)
	if err != nil {
		return nil, err
	}
	return newCheckedWriteCloser(wc), nil

}

func (cc *checkedClient) Reader(ctx context.Context, name string, offset uint64, size uint64) (_ io.ReadCloser, retErr error) {
	defer func() {
		retErr = errors.WithStack(retErr)
	}()
	rc, err := cc.Client.Reader(ctx, name, offset, size)
	if err != nil {
		// Enforce that clients return either an error or a reader, not both
		if rc != nil {
			return nil, errors.Wrap(err, "object client Reader() returned a reader and an error")
		}
		return nil, err
	}
	// Enforce that clients return either an error or a reader, not neither
	if rc == nil {
		return nil, errors.New("object client Reader() returned no reader and no error")
	}
	return newCheckedReadCloser(size, rc), nil
}

func (cc *checkedClient) Delete(ctx context.Context, name string) (retErr error) {
	defer func() {
		retErr = errors.WithStack(retErr)
	}()
	return cc.Client.Delete(ctx, name)
}

func (cc *checkedClient) Walk(ctx context.Context, prefix string, fn func(name string) error) (retErr error) {
	defer func() {
		retErr = errors.WithStack(retErr)
	}()
	return cc.Client.Walk(ctx, prefix, fn)
}
