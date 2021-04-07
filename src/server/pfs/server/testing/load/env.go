package load

type Env struct {
	client    Client
	validator *Validator
}

func NewEnv(client Client, spec *CommitsSpec) *Env {
	if spec.ThroughputSpec != nil {
		client = NewThroughputLimitClient(client, spec.ThroughputSpec)
	}
	if spec.CancelSpec != nil {
		client = NewCancelClient(client, spec.CancelSpec)
	}
	var validator *Validator
	if spec.ValidatorSpec != nil {
		client, validator = NewValidator(client, spec.ValidatorSpec)
	}
	return &Env{
		client:    client,
		validator: validator,
	}
}

func (e *Env) Client() Client {
	return e.client
}

func (e *Env) Validate(repo, commit string) error {
	if e.validator == nil {
		return nil
	}
	return e.validator.Validate(e.Client(), repo, commit)
}
