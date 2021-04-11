package load

import (
	"bytes"
	"context"
	"io"
	"sort"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/tarutil"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

type ValidatorSpec struct{}

type Validator struct {
	spec  *ValidatorSpec
	files map[string][]byte
}

func NewValidator(client Client, spec *ValidatorSpec) (Client, *Validator) {
	v := &Validator{
		spec:  spec,
		files: make(map[string][]byte),
	}
	return &validatorClient{
		Client:    client,
		validator: v,
	}, v
}

func (v *Validator) PutFiles(files map[string][]byte) {
	for path, hash := range files {
		v.files[path] = hash
	}
}

func (v *Validator) DeleteFiles(files map[string]struct{}) {
	for path := range files {
		delete(v.files, path)
	}
}

func (v *Validator) RandomFile() string {
	for path := range v.files {
		return path
	}
	return "no-op-delete"
}

func (v *Validator) Validate(client Client, repo, commit string) (retErr error) {
	var namesSorted []string
	for name := range v.files {
		namesSorted = append(namesSorted, name)
	}
	sort.Strings(namesSorted)
	defer func() {
		if retErr == nil {
			if len(namesSorted) != 0 {
				retErr = errors.Errorf("got back less files than expected")
			}
		}
	}()
	r, err := client.GetFileTar(context.Background(), repo, commit, "**")
	if err != nil {
		return err
	}
	return tarutil.Iterate(r, func(file tarutil.File) error {
		if len(namesSorted) == 0 {
			return errors.Errorf("got back more files than expected")
		}
		hdr, err := file.Header()
		if err != nil {
			return err
		}
		if hdr.Name == "/" && namesSorted[0] == "/" {
			namesSorted = namesSorted[1:]
			return nil
		}
		h := pfs.NewHash()
		if err := file.Content(h); err != nil {
			return err
		}
		if !bytes.Equal(v.files[namesSorted[0]], h.Sum(nil)) {
			return errors.Errorf("file %v's content is incorrect", namesSorted[0])
		}
		namesSorted = namesSorted[1:]
		return nil
	})
}

type validatorClient struct {
	Client
	validator *Validator
}

func (vc *validatorClient) WithModifyFileClient(ctx context.Context, repo, commit string, cb func(client.ModifyFile) error) error {
	var putFiles map[string][]byte
	var deleteFiles map[string]struct{}
	if err := vc.Client.WithModifyFileClient(ctx, repo, commit, func(mf client.ModifyFile) error {
		vmfc := &validatorModifyFileClient{
			ModifyFile:  mf,
			putFiles:    make(map[string][]byte),
			deleteFiles: make(map[string]struct{}),
		}
		if err := cb(vmfc); err != nil {
			return err
		}
		putFiles = vmfc.putFiles
		deleteFiles = vmfc.deleteFiles
		return nil
	}); err != nil {
		return err
	}
	vc.validator.PutFiles(putFiles)
	vc.validator.DeleteFiles(deleteFiles)
	return nil
}

type validatorModifyFileClient struct {
	client.ModifyFile
	putFiles    map[string][]byte
	deleteFiles map[string]struct{}
}

func (vmfc *validatorModifyFileClient) PutFile(path string, r io.Reader, opts ...client.PutFileOption) error {
	h := pfs.NewHash()
	if err := vmfc.ModifyFile.PutFile(path, io.TeeReader(r, h), opts...); err != nil {
		return err
	}
	vmfc.putFiles[path] = h.Sum(nil)
	return nil
}

func (vmfc *validatorModifyFileClient) DeleteFile(path string, opts ...client.DeleteFileOption) error {
	if err := vmfc.ModifyFile.DeleteFile(path, opts...); err != nil {
		return err
	}
	vmfc.deleteFiles[path] = struct{}{}
	return nil
}
