package obj

import (
	"context"
	"io"
	"strings"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pacherr"
	"github.com/sirupsen/logrus"
)

var _ Client = &uniformClient{}

// uniformClient is for ensuring uniform behavior across all the object clients
type uniformClient struct {
	c Client
}

func newUniformClient(c Client) Client {
	return &uniformClient{c: c}
}

func (uc *uniformClient) Put(ctx context.Context, name string, r io.Reader) (retErr error) {
	defer func() {
		retErr = errors.EnsureStack(retErr)
	}()
	name = strings.Trim(name, "/")
	return uc.c.Put(ctx, name, r)
}

func (cc *uniformClient) Get(ctx context.Context, name string, w io.Writer) (retErr error) {
	defer func() {
		retErr = errors.EnsureStack(retErr)
	}()
	name = strings.Trim(name, "/")
	return cc.c.Get(ctx, name, w)
}

func (cc *uniformClient) Delete(ctx context.Context, name string) (retErr error) {
	defer func() {
		retErr = errors.EnsureStack(retErr)
	}()
	name = strings.Trim(name, "/")
	err := cc.c.Delete(ctx, name)
	if pacherr.IsNotExist(err) {
		err = nil
	}
	return err
}

func (cc *uniformClient) Walk(ctx context.Context, prefix string, fn func(name string) error) (retErr error) {
	defer func() {
		retErr = errors.EnsureStack(retErr)
	}()
	return cc.c.Walk(ctx, prefix, fn)
}

func (uc *uniformClient) Exists(ctx context.Context, p string) (_ bool, retErr error) {
	defer func() {
		retErr = errors.EnsureStack(retErr)
	}()
	exists, err := uc.c.Exists(ctx, p)
	if pacherr.IsNotExist(err) {
		logrus.Warn("obj.Client Exists returned not exist error")
		exists = false
		err = nil
	}
	return exists, err
}
