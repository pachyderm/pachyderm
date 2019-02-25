package s3

import (
	"net/http"
	"text/template"

	"github.com/pachyderm/pachyderm/src/client"
)

const listBucketsSource = `<?xml version="1.0" encoding="UTF-8"?>
<ListAllMyBucketsResult xmlns="http://s3.amazonaws.com/doc/2006-03-01">
    <Owner>
    	<ID>000000000000000000000000000000</ID>
    	<DisplayName>pachyderm</DisplayName>
    </Owner>
    <Buckets>
        {{ range . }}
            <Bucket>
                <Name>{{ .Repo.Name }}</Name>
                <CreationDate>{{ formatTime .Created }}</CreationDate>
            </Bucket>
        {{ end }}
    </Buckets>
</ListAllMyBucketsResult>`

type rootHandler struct {
	pc           *client.APIClient
	listTemplate *template.Template
}

// TODO: support xml escaping
func newRootHandler(pc *client.APIClient) rootHandler {
	funcMap := template.FuncMap{
		"formatTime": formatTime,
	}

	listTemplate := template.Must(template.New("list-buckets").
		Funcs(funcMap).
		Parse(listBucketsSource))

	return rootHandler{
		pc:           pc,
		listTemplate: listTemplate,
	}
}

func (h rootHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	buckets, err := h.pc.ListRepo()
	if err != nil {
		writeServerError(w, err)
		return
	}

	if err = h.listTemplate.Execute(w, buckets); err != nil {
		writeServerError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/xml")
}
