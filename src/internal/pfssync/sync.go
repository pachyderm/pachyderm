package pfssync

import (
	"archive/tar"
	"io"
	"os"
	"path"
	"path/filepath"
	"syscall"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/tarutil"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"golang.org/x/sync/errgroup"
)

// Importer is the standard interface for a PFS importer.
type Importer interface {
	// Import a PFS file to a location on the local filesystem.
	Import(storageRoot string, file *pfs.File, opts ...ImportOption) error
}

type importer struct {
	pachClient *client.APIClient
	pipes      map[string]struct{}
	eg         *errgroup.Group
	done       bool
}

// WithImporter provides a scoped environment for an Importer.
func WithImporter(pachClient *client.APIClient, cb func(Importer) error) (retErr error) {
	i := &importer{
		pachClient: pachClient,
		pipes:      make(map[string]struct{}),
		eg:         &errgroup.Group{},
	}
	defer func() {
		i.done = true
		if err := i.closePipes(); retErr == nil {
			retErr = err
		}
	}()
	return cb(i)
}

func (i *importer) closePipes() (retErr error) {
	pipes := make(map[string]io.Closer)
	defer func() {
		for path, pipe := range pipes {
			if err := pipe.Close(); retErr == nil {
				retErr = err
			}
			if err := os.Remove(path); retErr == nil {
				retErr = err
			}
		}
	}()
	// Open all the pipes to unblock the goroutines.
	for path := range i.pipes {
		f, err := os.OpenFile(path, syscall.O_NONBLOCK+os.O_RDONLY, os.ModeNamedPipe)
		if err != nil {
			return err
		}
		pipes[path] = f
	}
	return i.eg.Wait()
}

type importConfig struct {
	lazy, empty    bool
	headerCallback func(*tar.Header) error
}

// Import a PFS file to a location on the local filesystem.
func (i *importer) Import(storageRoot string, file *pfs.File, opts ...ImportOption) error {
	if err := os.MkdirAll(storageRoot, 0700); err != nil {
		return err
	}
	c := &importConfig{}
	for _, opt := range opts {
		opt(c)
	}
	if c.lazy || c.empty {
		return i.importInfo(storageRoot, file, c)
	}
	r, err := i.pachClient.GetTarFile(file.Commit.Repo.Name, file.Commit.ID, file.Path)
	if err != nil {
		return err
	}
	if c.headerCallback != nil {
		return tarutil.Import(storageRoot, r, c.headerCallback)
	}
	return tarutil.Import(storageRoot, r)
}

func (i *importer) importInfo(storageRoot string, file *pfs.File, config *importConfig) (retErr error) {
	repo := file.Commit.Repo.Name
	commit := file.Commit.ID
	return i.pachClient.WalkFile(repo, commit, file.Path, func(fi *pfs.FileInfo) error {
		basePath, err := filepath.Rel(path.Dir(file.Path), fi.File.Path)
		if err != nil {
			return err
		}
		fullPath := path.Join(storageRoot, basePath)
		if fi.FileType == pfs.FileType_DIR {
			return os.MkdirAll(fullPath, 0700)
		}
		if config.lazy {
			return i.makePipe(fullPath, func(w io.Writer) error {
				r, err := i.pachClient.GetTarFile(repo, commit, fi.File.Path)
				if err != nil {
					return err
				}
				return tarutil.Iterate(r, func(f tarutil.File) error {
					if config.headerCallback != nil {
						hdr, err := f.Header()
						if err != nil {
							return err
						}
						if err := config.headerCallback(hdr); err != nil {
							return err
						}
					}
					return f.Content(w)
				}, true)
			})
		}
		f, err := os.Create(fullPath)
		if err != nil {
			return err
		}
		return f.Close()
	})
}
