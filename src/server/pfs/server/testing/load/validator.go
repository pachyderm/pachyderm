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

func (vc *validatorClient) WithModifyFileClient(ctx context.Context, repo, commit string, cb func(client.ModifyFileClient) error) error {
	var files map[string][]byte
	if err := vc.Client.WithModifyFileClient(ctx, repo, commit, func(mfc client.ModifyFileClient) error {
		vmfc := &validatorModifyFileClient{
			ModifyFileClient: mfc,
			files:            make(map[string][]byte),
		}
		if err := cb(vmfc); err != nil {
			return err
		}
		files = vmfc.files
		return nil
	}); err != nil {
		return err
	}
	vc.validator.PutFiles(files)
	return nil
}

type validatorModifyFileClient struct {
	client.ModifyFileClient
	files map[string][]byte
}

func (vmfc *validatorModifyFileClient) PutFile(path string, r io.Reader, opts ...client.PutFileOption) error {
	h := pfs.NewHash()
	if err := vmfc.ModifyFileClient.PutFile(path, io.TeeReader(r, h), opts...); err != nil {
		return err
	}
	vmfc.files[path] = h.Sum(nil)
	return nil
}
