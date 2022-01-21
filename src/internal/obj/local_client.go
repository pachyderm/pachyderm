package obj

import (
	"bytes"
	"context"
	"encoding/base64"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pacherr"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"github.com/sirupsen/logrus"
)

// NewLocalClient returns a Client that stores data on the local file system
func NewLocalClient(rootDir string) (Client, error) {
	c, err := newFSClient(rootDir)
	if err != nil {
		return nil, err
	}
	if monkeyTest {
		c = &monkeyClient{c}
	}
	return newUniformClient(c), nil
}

type fsClient struct {
	rootDir string
}

func newFSClient(rootDir string) (Client, error) {
	c := &fsClient{
		rootDir: filepath.Clean(rootDir),
	}
	if c.rootDir == "" || c.rootDir == "/" || c.rootDir == "." {
		panic("you probably didn't want to set the local client's root path to " + c.rootDir)
	}
	if err := c.init(); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *fsClient) Put(ctx context.Context, name string, r io.Reader) (retErr error) {
	staging := c.stagingPathFor(name)
	final := c.finalPathFor(name)
	f, err := os.Create(staging)
	if err != nil {
		return errors.EnsureStack(err)
	}
	defer c.closeFile(&retErr, f)
	defer c.removeFile(&retErr, staging)
	if _, err := io.Copy(f, r); err != nil {
		return errors.EnsureStack(err)
	}
	if err := f.Close(); err != nil {
		return errors.EnsureStack(err)
	}
	return errors.EnsureStack(os.Rename(staging, final))
}

func (c *fsClient) Get(ctx context.Context, name string, w io.Writer) (retErr error) {
	defer func() { retErr = c.transformError(retErr, name) }()
	f, err := os.Open(c.finalPathFor(name))
	if err != nil {
		return errors.EnsureStack(err)
	}
	defer c.closeFile(&retErr, f)
	_, err = io.Copy(w, f)
	return errors.EnsureStack(err)
}

func (c *fsClient) Delete(ctx context.Context, name string) error {
	err := os.Remove(c.finalPathFor(name))
	if os.IsNotExist(err) {
		err = nil
	}
	return err
}

func (c *fsClient) Exists(ctx context.Context, name string) (bool, error) {
	_, err := os.Stat(c.finalPathFor(name))
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, errors.EnsureStack(err)
	}
	return true, nil
}

func (c *fsClient) Walk(ctx context.Context, prefix string, cb func(string) error) error {
	dirEnts, err := os.ReadDir(filepath.Join(c.rootDir, "objects"))
	if err != nil {
		return errors.EnsureStack(err)
	}
	enc := base64.URLEncoding
	for _, dirEnt := range dirEnts {
		// TODO: There is a better way to do this (encode the prefix rather than decode all and filter)
		// but I'm not really sure how to do that with the filesystem API.
		// We would need to seek to a certain dir entry.
		name, err := enc.DecodeString(dirEnt.Name())
		if err != nil {
			return errors.Wrapf(err, "parsing object name")
		}
		if bytes.HasPrefix(name, []byte(prefix)) {
			if err := cb(string(name)); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *fsClient) BucketURL() ObjectStoreURL {
	return ObjectStoreURL{
		Scheme: "local",
		Bucket: strings.ReplaceAll(filepath.ToSlash(c.rootDir), "/", "."),
	}
}

func (c *fsClient) stagingPathFor(name string) string {
	return filepath.Join(c.rootDir, "staging", uuid.NewWithoutDashes())
}

func (c *fsClient) finalPathFor(name string) string {
	enc := base64.URLEncoding
	return filepath.Join(c.rootDir, "objects", enc.EncodeToString([]byte(name)))
}

func (c *fsClient) init() error {
	if err := os.RemoveAll(filepath.Join(c.rootDir, "staging")); err != nil {
		return errors.EnsureStack(err)
	}
	if err := os.MkdirAll(filepath.Join(c.rootDir, "staging"), 0755); err != nil {
		return errors.EnsureStack(err)
	}
	if err := os.MkdirAll(filepath.Join(c.rootDir, "objects"), 0755); err != nil {
		return errors.EnsureStack(err)
	}
	logrus.Infof("successfully initialized fs-backed object store at %s", c.rootDir)
	return nil
}

func (c *fsClient) transformError(err error, name string) error {
	if err == nil {
		return nil
	}
	if os.IsNotExist(err) || strings.HasSuffix(err.Error(), ": no such file or directory") {
		return pacherr.NewNotExist(c.BucketURL().String(), name)
	}
	return err
}

func (c *fsClient) closeFile(retErr *error, f *os.File) {
	err := f.Close()
	if err != nil && !strings.Contains(err.Error(), "file already closed") {
		if retErr == nil {
			*retErr = err
		} else {
			logrus.Errorf("error closing file: %v", err)
		}
	}
}

func (c *fsClient) removeFile(retErr *error, p string) {
	err := os.Remove(p)
	if os.IsNotExist(err) {
		err = nil
	}
	if err != nil {
		if retErr == nil {
			*retErr = err
		} else {
			logrus.Errorf("error deleting file: %v", err)
		}
	}
}
