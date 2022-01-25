package pfsload

import (
	"bytes"
	"context"
	"io"
	"math/rand"
	"strings"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	"github.com/pachyderm/pachyderm/v2/src/internal/tarutil"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	pfsserver "github.com/pachyderm/pachyderm/v2/src/server/pfs"
)

type FrequencySpec struct {
	Count int `yaml:"count,omitempty"`
	Prob  int `yaml:"prob,omitempty"`
	count int
}

type ValidatorSpec struct {
	FrequencySpec *FrequencySpec `yaml:"frequency,omitempty"`
}

type Validator struct {
	spec   *ValidatorSpec
	buffer *fileset.Buffer
	random *rand.Rand
	files  []string
}

func NewValidator(client Client, spec *ValidatorSpec, random *rand.Rand) (Client, *Validator, error) {
	if spec.FrequencySpec != nil {
		if err := validateProb(spec.FrequencySpec.Prob); err != nil {
			return nil, nil, err
		}
	}
	v := &Validator{
		spec:   spec,
		buffer: fileset.NewBuffer(),
		random: random,
	}
	return &validatorClient{
		Client:    client,
		validator: v,
	}, v, nil
}

// TODO: The performance is better than it was, but long sequences of alternating add / delete can still be bad.
func (v *Validator) RandomFile() (string, error) {
	if v.files == nil {
		v.files = []string{}
		if err := v.buffer.WalkAdditive(func(p, _ string, r io.Reader) error {
			v.files = append(v.files, p)
			return nil
		}); err != nil {
			return "", err
		}
		v.random.Shuffle(len(v.files), func(i, j int) {
			v.files[i], v.files[j] = v.files[j], v.files[i]
		})
	}
	if len(v.files) > 0 {
		file := v.files[0]
		v.files = v.files[1:]
		return file, nil
	}
	return "no-op-delete", nil
}

type file struct {
	name string
	hash []byte
}

func (v *Validator) Validate(client Client, commit *pfs.Commit) (retErr error) {
	if v.spec.FrequencySpec != nil {
		freq := v.spec.FrequencySpec
		if freq.Count > 0 {
			freq.count++
			if freq.Count > freq.count {
				return nil
			}
			freq.count = 0
		} else if !shouldExecute(v.random, freq.Prob) {
			return nil
		}
	}
	var files []*file
	if err := v.buffer.WalkAdditive(func(p, tag string, r io.Reader) error {
		buf := &bytes.Buffer{}
		if _, err := io.Copy(buf, r); err != nil {
			return errors.EnsureStack(err)
		}
		files = append(files, &file{
			name: p,
			hash: buf.Bytes(),
		})
		return nil
	}); err != nil {
		return err
	}
	err := client.WaitCommitSet(commit.ID, func(ci *pfs.CommitInfo) error {
		if ci.Commit.Branch.Repo.Type != pfs.UserRepoType {
			return nil
		}
		return validate(client, ci.Commit, files)
	})
	return errors.EnsureStack(err)
}

func validate(client Client, commit *pfs.Commit, files []*file) (retErr error) {
	defer func() {
		if retErr == nil {
			if len(files) != 0 {
				retErr = errors.Errorf("got back less files than expected")
			}
		}
	}()
	r, err := client.GetFileTAR(client.Ctx(), commit, "**")
	if err != nil {
		return errors.EnsureStack(err)
	}
	err = tarutil.Iterate(r, func(file tarutil.File) error {
		if len(files) == 0 {
			return errors.Errorf("got back more files than expected")
		}
		hdr, err := file.Header()
		if err != nil {
			return errors.EnsureStack(err)
		}
		if strings.HasSuffix(hdr.Name, "/") {
			return nil
		}
		h := pfs.NewHash()
		if err := file.Content(h); err != nil {
			return errors.EnsureStack(err)
		}
		if !bytes.Equal(files[0].hash, h.Sum(nil)) {
			return errors.Errorf("file %v's content is incorrect (actual file path: %v)", files[0].name, hdr.Name)
		}
		files = files[1:]
		return nil
	}, true)
	if pfsserver.IsFileNotFoundErr(err) && len(files) == 0 {
		return nil
	}
	return err
}

type validatorClient struct {
	Client
	validator *Validator
}

func (vc *validatorClient) WithModifyFileClient(ctx context.Context, commit *pfs.Commit, cb func(client.ModifyFile) error) error {
	err := vc.Client.WithModifyFileClient(ctx, commit, func(mf client.ModifyFile) (retErr error) {
		vmfc := &validatorModifyFileClient{
			ModifyFile: mf,
			buffer:     fileset.NewBuffer(),
		}
		if err := cb(vmfc); err != nil {
			return err
		}
		for _, p := range vmfc.deletes {
			vc.validator.buffer.Delete(p, fileset.DefaultFileDatum)
		}
		return vmfc.buffer.WalkAdditive(func(p, tag string, r io.Reader) error {
			vc.validator.files = nil
			w := vc.validator.buffer.Add(p, tag)
			_, err := io.Copy(w, r)
			return errors.EnsureStack(err)
		})
	})
	return errors.EnsureStack(err)
}

type validatorModifyFileClient struct {
	client.ModifyFile
	buffer  *fileset.Buffer
	deletes []string
}

func (vmfc *validatorModifyFileClient) PutFile(path string, r io.Reader, opts ...client.PutFileOption) error {
	h := pfs.NewHash()
	if err := vmfc.ModifyFile.PutFile(path, io.TeeReader(r, h), opts...); err != nil {
		return errors.EnsureStack(err)
	}
	vmfc.buffer.Delete(path, fileset.DefaultFileDatum)
	w := vmfc.buffer.Add(path, fileset.DefaultFileDatum)
	_, err := io.Copy(w, bytes.NewReader(h.Sum(nil)))
	return errors.EnsureStack(err)
}

func (vmfc *validatorModifyFileClient) DeleteFile(path string, opts ...client.DeleteFileOption) error {
	if err := vmfc.ModifyFile.DeleteFile(path, opts...); err != nil {
		return errors.EnsureStack(err)
	}
	vmfc.deletes = append(vmfc.deletes, path)
	vmfc.buffer.Delete(path, fileset.DefaultFileDatum)
	return nil
}
