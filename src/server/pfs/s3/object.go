package s3

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"

	"github.com/gogo/protobuf/types"
	"github.com/gorilla/mux"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
)

const initMultipartSource = `
<InitiateMultipartUploadResult xmlns="http://s3.amazonaws.com/doc/2006-03-01/">
	<Bucket>{{ .bucket }}</Bucket>
	<Key>{{ .key }}</Key>
	<UploadId>{{ .uploadId }}</UploadId>
</InitiateMultipartUploadResult>
`

type objectHandler struct {
	pc                    *client.APIClient
	multipartDir          string
	initMultipartTemplate xmlTemplate
}

func newObjectHandler(pc *client.APIClient, multipartDir string) objectHandler {
	return objectHandler{
		pc:                    pc,
		multipartDir:          multipartDir,
		initMultipartTemplate: newXmlTemplate(http.StatusOK, "init-multipart", initMultipartSource),
	}
}

func (h objectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	repo := vars["repo"]
	branch := vars["branch"]
	file := vars["file"]

	branchInfo, err := h.pc.InspectBranch(repo, branch)
	if err != nil {
		writeMaybeNotFound(w, r, err)
		return
	}

	if err := r.ParseForm(); err != nil {
		writeBadRequest(w, err)
		return
	}

	uploadId := r.FormValue("uploadId")

	if r.Method == http.MethodGet || r.Method == http.MethodHead {
		if uploadId != "" {
			h.listMultipart(w, r, branchInfo, file, uploadId)
		} else {
			h.get(w, r, branchInfo, file)
		}
	} else if r.Method == http.MethodPost {
		if _, ok := r.Form["uploads"]; ok {
			h.initMultipart(w, r, branchInfo, file)
		} else if uploadId != "" {
			h.completeMultipart(w, r, branchInfo, file, uploadId)
		} else {
			http.NotFound(w, r)
		}
	} else if r.Method == http.MethodPut {
		if uploadId != "" {
			h.uploadMultipart(w, r, branchInfo, file, uploadId)
		} else {
			h.put(w, r, branchInfo, file)
		}
	} else if r.Method == http.MethodDelete {
		if uploadId != "" {
			h.abortMultipart(w, r, branchInfo, file, uploadId)
		} else {
			h.delete(w, r, branchInfo, file)
		}
	} else {
		// method filtering on the mux router should prevent this
		panic("unreachable")
	}
}

func (h objectHandler) get(w http.ResponseWriter, r *http.Request, branchInfo *pfs.BranchInfo, file string) {
	if branchInfo.Head == nil {
		http.NotFound(w, r)
		return
	}

	fileInfo, err := h.pc.InspectFile(branchInfo.Branch.Repo.Name, branchInfo.Branch.Name, file)
	if err != nil {
		writeMaybeNotFound(w, r, err)
		return
	}

	timestamp, err := types.TimestampFromProto(fileInfo.Committed)
	if err != nil {
		writeServerError(w, err)
		return
	}

	reader, err := h.pc.GetFileReadSeeker(branchInfo.Branch.Repo.Name, branchInfo.Branch.Name, file)
	if err != nil {
		writeServerError(w, err)
		return
	}

	http.ServeContent(w, r, "", timestamp, reader)
}

func (h objectHandler) put(w http.ResponseWriter, r *http.Request, branchInfo *pfs.BranchInfo, file string) {
	expectedHash := r.Header.Get("Content-MD5")

	if expectedHash != "" {
		expectedHashBytes, err := base64.StdEncoding.DecodeString(expectedHash)
		if err != nil {
			writeBadRequest(w, fmt.Errorf("could not decode `Content-MD5`, as it is not base64-encoded"))
			return
		}

		hasher := md5.New()
		reader := io.TeeReader(r.Body, hasher)

		_, err = h.pc.PutFileOverwrite(branchInfo.Branch.Repo.Name, branchInfo.Branch.Name, file, reader, 0)
		if err != nil {
			writeServerError(w, err)
			return
		}

		actualHash := hasher.Sum(nil)
		if !bytes.Equal(expectedHashBytes, actualHash) {
			err = fmt.Errorf("content checksums differ; expected=%x, actual=%x", expectedHash, actualHash)
			writeBadRequest(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
		return
	}

	_, err := h.pc.PutFileOverwrite(branchInfo.Branch.Repo.Name, branchInfo.Branch.Name, file, r.Body, 0)
	if err != nil {
		writeServerError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h objectHandler) delete(w http.ResponseWriter, r *http.Request, branchInfo *pfs.BranchInfo, file string) {
	if branchInfo.Head == nil {
		http.NotFound(w, r)
		return
	}

	if err := h.pc.DeleteFile(branchInfo.Branch.Repo.Name, branchInfo.Branch.Name, file); err != nil {
		writeMaybeNotFound(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h objectHandler) initMultipart(w http.ResponseWriter, r *http.Request, branchInfo *pfs.BranchInfo, file string) {
	if h.multipartDir == "" {
		writeBadRequest(w, fmt.Errorf("multipart uploads disabled"))
		return
	}
}

func (h objectHandler) listMultipart(w http.ResponseWriter, r *http.Request, branchInfo *pfs.BranchInfo, file string, uploadId string) {
	if h.multipartDir == "" {
		writeBadRequest(w, fmt.Errorf("multipart uploads disabled"))
		return
	}
}

func (h objectHandler) uploadMultipart(w http.ResponseWriter, r *http.Request, branchInfo *pfs.BranchInfo, file string, uploadId string) {
	if h.multipartDir == "" {
		writeBadRequest(w, fmt.Errorf("multipart uploads disabled"))
		return
	}
}

func (h objectHandler) completeMultipart(w http.ResponseWriter, r *http.Request, branchInfo *pfs.BranchInfo, file string, uploadId string) {
	if h.multipartDir == "" {
		writeBadRequest(w, fmt.Errorf("multipart uploads disabled"))
		return
	}
}

func (h objectHandler) abortMultipart(w http.ResponseWriter, r *http.Request, branchInfo *pfs.BranchInfo, file string, uploadId string) {
	if h.multipartDir == "" {
		writeBadRequest(w, fmt.Errorf("multipart uploads disabled"))
		return
	}
}
