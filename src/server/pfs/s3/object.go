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

type objectHandler struct {
	pc *client.APIClient
}

func newObjectHandler(pc *client.APIClient) objectHandler {
	return objectHandler{
		pc: pc,
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

	if r.Method == http.MethodGet || r.Method == http.MethodHead {
		h.getObject(w, r, branchInfo, file)
	} else if r.Method == http.MethodPut {
		h.putObject(w, r, branchInfo, file)
	} else if r.Method == http.MethodDelete {
		h.deleteObject(w, r, branchInfo, file)
	} else {
		// method filtering on the mux router should prevent this
		panic("unreachable")
	}
}

func (h objectHandler) getObject(w http.ResponseWriter, r *http.Request, branchInfo *pfs.BranchInfo, file string) {
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

func (h objectHandler) putObject(w http.ResponseWriter, r *http.Request, branchInfo *pfs.BranchInfo, file string) {
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

func (h objectHandler) deleteObject(w http.ResponseWriter, r *http.Request, branchInfo *pfs.BranchInfo, file string) {
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
