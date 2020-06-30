package sync

import (
	"archive/tar"
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
)

func PullV2(pachClient *client.APIClient, file *pfs.File, storageRoot string) error {
	r, err := pachClient.GetTarV2(file.Commit.Repo.Name, file.Commit.ID, file.Path)
	if err != nil {
		return err
	}
	return TarToLocal(file, storageRoot, r)
}

// TODO: refactor these helper functions into a utility package.
func TarToLocal(file *pfs.File, storageRoot string, r io.Reader) error {
	tr := tar.NewReader(r)
	for {
		hdr, err := tr.Next()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		basePath, err := filepath.Rel(file.Path, hdr.Name)
		if err != nil {
			return err
		}
		// TODO: Use the tar header metadata.
		fullPath := path.Join(storageRoot, basePath)
		if hdr.Typeflag == tar.TypeDir {
			if err := os.MkdirAll(fullPath, 0700); err != nil {
				return err
			}
			continue
		}
		if err := WriteFile(fullPath, tr); err != nil {
			return err
		}
	}
}

func WriteFile(filePath string, r io.Reader) (retErr error) {
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer func() {
		if err := f.Close(); retErr == nil {
			retErr = err
		}
	}()
	_, err = io.Copy(f, r)
	return err
}

func LocalToTar(storageRoot string, w io.Writer) (retErr error) {
	tw := tar.NewWriter(w)
	defer func() {
		if err := tw.Close(); retErr == nil {
			retErr = err
		}
	}()
	return filepath.Walk(storageRoot, func(file string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if file == storageRoot {
			return nil
		}
		// TODO: link name?
		hdr, err := tar.FileInfoHeader(fi, "")
		if err != nil {
			return err
		}
		hdr.Name, err = filepath.Rel(storageRoot, file)
		if err != nil {
			return err
		}
		// TODO: Hack to match glob, need consistent canonicalization.
		hdr.Name = filepath.Join("/", hdr.Name)
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		if hdr.Typeflag == tar.TypeDir {
			return nil
		}
		f, err := os.Open(file)
		if err != nil {
			return err
		}
		defer func() {
			if err := f.Close(); retErr == nil {
				retErr = err
			}
		}()
		_, err = io.Copy(tw, f)
		return err
	})
}
