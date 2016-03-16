package pkghttp

import (
	"net/http"
	"text/template"

	"go.pedge.io/pkg/tmpl"
)

type templater struct {
	delegate pkgtmpl.Templater
}

func newTemplater(delegate pkgtmpl.Templater) *templater {
	return &templater{delegate}
}

func (h *templater) WithFuncs(funcMap template.FuncMap) Templater {
	return newTemplater(h.delegate.WithFuncs(funcMap))
}

func (h *templater) Execute(responseWriter http.ResponseWriter, name string, data interface{}) {
	if err := h.delegate.Execute(responseWriter, name, data); err != nil {
		ErrorInternal(responseWriter, err)
	}
}
