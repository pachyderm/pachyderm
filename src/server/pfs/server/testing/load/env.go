package load

type Env struct {
	client      Client
	validator   *Validator
	fileSources map[string]FileSource
}

func NewEnv(client Client, spec *CommitsSpec) *Env {
	if spec.ThroughputSpec != nil {
		client = NewThroughputLimitClient(client, spec.ThroughputSpec)
	}
	var validator *Validator
	if spec.ValidatorSpec != nil {
		client, validator = NewValidator(client, spec.ValidatorSpec)
	}
	if spec.CancelSpec != nil {
		client = NewCancelClient(client, spec.CancelSpec)
	}
	fileSources := make(map[string]FileSource)
	for _, spec := range spec.FileSourceSpecs {
		fileSources[spec.Name] = NewFileSource(spec)
	}
	return &Env{
		client:      client,
		validator:   validator,
		fileSources: fileSources,
	}
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
