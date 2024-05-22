package pfsload

import (
	"context"
	"math/rand"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/internal/task"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

type Env struct {
	client      Client
	taskDoer    task.Doer
	fileSources map[string]*FileSourceSpec
	validator   *Validator
	seed        int64
	authToken   string
}

func NewEnv(ctx context.Context, c pfs.APIClient, taskService task.Service, spec *CommitSpec, seed int64) (*Env, error) {
	client := NewPachClient(c)
	random := rand.New(rand.NewSource(seed))
	fileSources := make(map[string]*FileSourceSpec)
	for _, fileSource := range spec.FileSources {
		fileSources[fileSource.Name] = fileSource
	}
	var validator *Validator
	var err error
	if spec.Validator != nil {
		validator, err = NewValidator(spec.Validator, random)
	}
	if err != nil {
		return nil, err
	}
	authToken, err := auth.GetAuthToken(ctx)
	if err != nil && !auth.IsErrNotSignedIn(err) {
		return nil, err
	}
	return &Env{
		client:      client,
		taskDoer:    taskService.NewDoer(namespace, "", nil),
		fileSources: fileSources,
		validator:   validator,
		seed:        seed,
		authToken:   authToken,
	}, nil
}

func (e *Env) Client() Client {
	return e.client
}

func (e *Env) AuthToken() string {
	return e.authToken
}

func (e *Env) TaskDoer() task.Doer {
	return e.taskDoer
}

func (e *Env) FileSource(name string) *FileSourceSpec {
	return e.fileSources[name]
}

func (e *Env) Validator() *Validator {
	return e.validator
}

func (e *Env) Seed() int64 {
	e.seed++
	return e.seed
}
