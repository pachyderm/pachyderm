package pfsload

import (
	"math/rand"
)

type Env struct {
	client      Client
	validator   *Validator
	fileSources map[string]FileSource
	random      *rand.Rand
}

func NewEnv(client Client, spec *CommitsSpec, seed int64) (*Env, error) {
	var err error
	random := rand.New(rand.NewSource(seed))
	if spec.ThroughputSpec != nil {
		client, err = NewThroughputLimitClient(client, spec.ThroughputSpec, random)
	}
	if err != nil {
		return nil, err
	}
	var validator *Validator
	if spec.ValidatorSpec != nil {
		client, validator, err = NewValidator(client, spec.ValidatorSpec, random)
	}
	if err != nil {
		return nil, err
	}
	if spec.CancelSpec != nil {
		client, err = NewCancelClient(client, spec.CancelSpec, random)
	}
	if err != nil {
		return nil, err
	}
	fileSources := make(map[string]FileSource)
	for _, spec := range spec.FileSourceSpecs {
		fileSources[spec.Name] = NewFileSource(spec, random)
	}
	return &Env{
		client:      client,
		validator:   validator,
		fileSources: fileSources,
		random:      random,
	}, nil
}

func (e *Env) Client() Client {
	return e.client
}

func (e *Env) Validator() *Validator {
	return e.validator
}

func (e *Env) FileSource(name string) FileSource {
	return e.fileSources[name]
}

func (e *Env) Rand() *rand.Rand {
	return e.random
}
