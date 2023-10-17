package pfssync

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path"
	"syscall"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/tarutil"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"golang.org/x/sync/errgroup"
)

// Downloader is the standard interface for a PFS downloader.
type Downloader interface {
	// Download a PFS file to a location on the local filesystem.
	Download(storageRoot string, file *pfs.File, opts ...DownloadOption) error
}

type downloader struct {
	pachClient *CacheClient
	pipes      map[string]struct{}
	eg         *errgroup.Group
	done       bool
}

// WithDownloader provides a scoped environment for a Downloader.
func WithDownloader(pachClient *CacheClient, cb func(Downloader) error) (retErr error) {
	d := &downloader{
		pachClient: pachClient,
		pipes:      make(map[string]struct{}),
		eg:         &errgroup.Group{},
	}
	defer func() {
		d.done = true
		if err := d.closePipes(); retErr == nil {
			retErr = err
		}
	}()
	return cb(d)
}

func (d *downloader) closePipes() (retErr error) {
	pipes := make(map[string]io.Closer)
	defer func() {
		for path, pipe := range pipes {
			if err := pipe.Close(); retErr == nil {
				retErr = errors.EnsureStack(err)
			}
			if err := os.Remove(path); retErr == nil {
				retErr = errors.EnsureStack(err)
			}
			fmt.Println("core-2002: closed pipe and removed", path)
		}
	}()
	// Open all the pipes to unblock the goroutines.
	for path := range d.pipes {
		f, err := os.OpenFile(path, syscall.O_NONBLOCK+os.O_RDONLY, os.ModeNamedPipe)
		if err != nil {
			return errors.EnsureStack(err)
		}
		pipes[path] = f
	}
	fmt.Println("core-2002: opened all pipes")
	return errors.EnsureStack(d.eg.Wait())
}

type downloadConfig struct {
	lazy, empty    bool
	headerCallback func(*tar.Header) error
}

// Download a PFS file to a location on the local filesystem.
func (d *downloader) Download(storageRoot string, file *pfs.File, opts ...DownloadOption) error {
	if err := os.MkdirAll(storageRoot, 0700); err != nil {
		return errors.EnsureStack(err)
	}
	dc := &downloadConfig{}
	for _, opt := range opts {
		opt(dc)
	}
	if dc.lazy || dc.empty {
		return d.downloadInfo(storageRoot, file, dc)
	}
	r, err := d.pachClient.GetFileTAR(file.Commit, file.Path)
	if err != nil {
		fmt.Println("core-2002: got error calling Download() for file:", file.Path, ":", err.Error())
		return err
	}
	fmt.Println("core-2002: downloaded file TAR", file.Path)
	if dc.headerCallback != nil {
		err := tarutil.Import(storageRoot, r, dc.headerCallback)
		if err == nil {
			fmt.Println("core-2002: imported to", storageRoot, "via header callback")
		}
		if err != nil {
			fmt.Println("core-2002: error importing via the headerCallback", err.Error())
		}
		return err
	}
	err = tarutil.Import(storageRoot, r)
	if err == nil {
		fmt.Println("core-2002: imported to", storageRoot)
	}
	return err
}

func (d *downloader) downloadInfo(storageRoot string, file *pfs.File, config *downloadConfig) error {
	return d.pachClient.WalkFile(file.Commit, file.Path, func(fi *pfs.FileInfo) error {
		if fi.FileType == pfs.FileType_DIR {
			return nil
		}
		fullPath := path.Join(storageRoot, fi.File.Path)
		if err := os.MkdirAll(path.Dir(fullPath), 0700); err != nil {
			return errors.EnsureStack(err)
		}
		if config.lazy {
			fmt.Println("core-2002: making pipe:", fullPath)
			return d.makePipe(fullPath, func(w io.Writer) error {
				r, err := d.pachClient.GetFileTAR(file.Commit, fi.File.Path)
				if err != nil {
					return err
				}
				err = tarutil.Iterate(r, func(f tarutil.File) error {
					if config.headerCallback != nil {
						hdr, err := f.Header()
						if err != nil {
							return errors.EnsureStack(err)
						}
						if err := config.headerCallback(hdr); err != nil {
							return err
						}
					}
					return errors.EnsureStack(f.Content(w))
				}, true)
				fmt.Println("core-2002: finished making pipe")
				return err
			})
		}
		f, err := os.Create(fullPath)
		if err != nil {
			return errors.EnsureStack(err)
		}
		fmt.Println("core-2002: downloaded info:", file.Datum, file.Path)
		return errors.EnsureStack(f.Close())
	})
}
