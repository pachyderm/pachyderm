package pfsload

import (
	"context"
	"path"

	"github.com/pachyderm/pachyderm/v2/src/client"
)

type OperationsSpec struct {
	Count          int              `yaml:"count,omitempty"`
	OperationSpecs []*OperationSpec `yaml:"operation,omitempty"`
}

func Operations(env *Env, repo, branch, commit string, spec *OperationsSpec) error {
	for i := 0; i < spec.Count; i++ {
		if err := FuzzOperation(env, repo, branch, commit, spec.OperationSpecs); err != nil {
			return err
		}
	}
	return nil
}

// TODO: Add different types of operations.
type OperationSpec struct {
	PutFileSpec    *PutFileSpec    `yaml:"putFile,omitempty"`
	DeleteFileSpec *DeleteFileSpec `yaml:"deleteFile,omitempty"`
	Prob           float64         `yaml:"prob,omitempty"`
}

func Operation(env *Env, repo, branch, commit string, spec *OperationSpec) error {
	if spec.PutFileSpec != nil {
		return PutFile(env, repo, branch, commit, spec.PutFileSpec)
	}
	return DeleteFile(env, repo, branch, commit, spec.DeleteFileSpec)
}

type PutFileSpec struct {
	FilesSpec *FilesSpec `yaml:"files,omitempty"`
}

func PutFile(env *Env, repo, branch, commit string, spec *PutFileSpec) error {
	c := env.Client()
	files, err := Files(env, spec.FilesSpec)
	if err != nil {
		return err
	}
	return c.WithModifyFileClient(context.Background(), repo, branch, commit, func(mf client.ModifyFile) error {
		for _, file := range files {
			if err := mf.PutFile(file.Path(), file.Reader()); err != nil {
				return err
			}
		}
		return nil
	})
}

type DeleteFileSpec struct {
	Count         int     `yaml:"count,omitempty"`
	DirectoryProb float64 `yaml:"directoryProb,omitempty"`
}

func DeleteFile(env *Env, repo, branch, commit string, spec *DeleteFileSpec) error {
	c := env.Client()
	return c.WithModifyFileClient(context.Background(), repo, branch, commit, func(mf client.ModifyFile) error {
		for i := 0; i < spec.Count; i++ {
			p, err := nextDeletePath(env, spec)
			if err != nil {
				return err
			}
			if err := mf.DeleteFile(p); err != nil {
				return err
			}
		}
		return nil
	})
}

func nextDeletePath(env *Env, spec *DeleteFileSpec) (string, error) {
	validator := env.Validator()
	p, err := validator.RandomFile()
	if err != nil {
		return "", err
	}
	for shouldExecute(env.Rand(), spec.DirectoryProb) {
		p, _ = path.Split(p)
		if p == "" {
			break
		}
	}
	return p, nil
}
