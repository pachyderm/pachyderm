package load

import (
	"bytes"
	"sort"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/server/pkg/tarutil"
)

type ValidatorSpec struct{}

type Validator struct {
	spec       *ValidatorSpec
	fileHashes map[string][]byte
}

func NewValidator(spec *ValidatorSpec) *Validator {
	return &Validator{
		spec:       spec,
		fileHashes: make(map[string][]byte),
	}
}

func (v *Validator) AddFiles(files []tarutil.File) error {
	for _, file := range files {
		hdr, err := file.Header()
		if err != nil {
			return err
		}
		h := pfs.NewHash()
		if err := file.Content(h); err != nil {
			return err
		}
		v.fileHashes[hdr.Name] = h.Sum(nil)
	}
	return nil
}

func (v *Validator) Validate(c *client.APIClient, repo, commit string) (retErr error) {
	var namesSorted []string
	for name := range v.fileHashes {
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
	r, err := GetTar(c, repo, commit, &GetTarSpec{})
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
		if !bytes.Equal(v.fileHashes[namesSorted[0]], h.Sum(nil)) {
			return errors.Errorf("file %v's header and/or content is incorrect", namesSorted[0])
		}
		namesSorted = namesSorted[1:]
		return nil
	})
}
