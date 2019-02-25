package s3

// Go doesn't have built-in support for XML templating, but it does have HTML
// templating. This jerry-rigs HTML templating to provide XML templating.

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/gogo/protobuf/types"
	"github.com/sirupsen/logrus"
)

// Put at the beginning of every rendered template. This is not put in the
// template itself because HTML templating escapes the first question mark.
const xmlIntro = `<?xml version="1.0" encoding="UTF-8"?>`

type xmlTemplate struct {
	code int
	tmpl *template.Template
}

func newXmlTemplate(code int, name string, source string) xmlTemplate {
	funcMap := template.FuncMap{
		"formatTime": formatTime,
	}

	tmpl := template.Must(template.New(name).
		Funcs(funcMap).
		Parse(source))

	return xmlTemplate{
		code: code,
		tmpl: tmpl,
	}
}

func (x xmlTemplate) render(w http.ResponseWriter, payload interface{}) {
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(x.code)
	fmt.Fprintln(w, xmlIntro)

	if err := x.tmpl.Execute(w, payload); err != nil {
		// log the error, but don't send anything back to the client because:
		// 1) most likely there was an error writing, implying that the socket
		//    is flaky or broken and that further writing may fail as well
		// 2) we've already written an OK status
		logrus.Infof("s3gateway: xml templating: execute failed: %v", err)
	}
}

func formatTime(timestamp *types.Timestamp) string {
	return timestamp.String()
}
