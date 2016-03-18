package pkgtmpl // import "go.pedge.io/pkg/tmpl"

import (
	"io"
	"text/template"
)

// Templater handles templates.
type Templater interface {
	WithFuncs(funcMap template.FuncMap) Templater
	Execute(writer io.Writer, name string, data interface{}) error
}

// NewTemplater creates a new Templater.
func NewTemplater(baseDirPath string) Templater {
	return newTemplater(baseDirPath)
}
