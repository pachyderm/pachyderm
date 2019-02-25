package s3

import (
	"net/http"

	"github.com/pachyderm/pachyderm/src/client"
)

const listBucketsSource = `
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
	listTemplate xmlTemplate
}

func newRootHandler(pc *client.APIClient) rootHandler {
	return rootHandler{
		pc:           pc,
		listTemplate: newXmlTemplate(http.StatusOK, "list-buckets", listBucketsSource),
	}
}

func (h rootHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	buckets, err := h.pc.ListRepo()
	if err != nil {
		writeServerError(w, err)
		return
	}

	h.listTemplate.render(w, buckets)
}
