package templates

type Template struct {
	SpecLanguageVersion int         `json:"spec-language-version"`
	Params              []ParamSpec `json:"params"`
	Templates           []string
}

type ParamSpec struct {
	Name        string          `json:"name"`
	Description *string         `json:"description,omitempty"`
	Validation  *ValidationSpec `json:"validation,omitempty"`
}

type ValidationSpec struct {
	Regex *string `json:"regex,omitempty"`

	// TODO(msteffen):
	// Range is something like "[1-5],-1"
	// Range *string `json:"range,omitempty"`
	// Type must currently be one of "float", "int", "string"
	// Type *string `json:"type,omitempty"`
}

// template YAML:
params:
  - name: value
	  validation:
		  regex: '[0-9]+'

// user YAML:
template: github.com/pachyderm/kubeflow
iterations: 5
