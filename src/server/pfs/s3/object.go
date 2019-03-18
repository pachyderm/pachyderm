package s3

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"fmt"
	"crypto/md5"
	"encoding/base64"

	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/src/client"
)

type objectHandler struct {
	pc *client.APIClient
}

func newObjectHandler(pc *client.APIClient) *objectHandler {
	return &objectHandler{pc: pc}
}

func (h *objectHandler) get(w http.ResponseWriter, r *http.Request) {
	repo, branch, file := objectArgs(w, r)
	branchInfo, err := h.pc.InspectBranch(repo, branch)
	if err != nil {
		notFoundError(w, r, err)
		return
	}
	if branchInfo.Head == nil {
		noSuchKeyError(w, r)
		return
	}
	if isMeta(file) || strings.HasSuffix(file, "/") {
		invalidFilePathError(w, r)
		return
	}

	fileInfo, err := h.pc.InspectFile(branchInfo.Branch.Repo.Name, branchInfo.Head.ID, file)
	if err != nil {
		notFoundError(w, r, err)
		return
	}

	timestamp, err := types.TimestampFromProto(fileInfo.Committed)
	if err != nil {
		internalError(w, r, err)
		return
	}

	meta, err := getMeta(h.pc, branchInfo.Branch.Repo.Name, branchInfo.Head.ID, file)
	if err != nil {
		internalError(w, r, err)
		return
	}
	if meta != nil {
		w.Header().Set("ETag", fmt.Sprintf("\"%s\"", meta.MD5))
	}

	reader, err := h.pc.GetFileReadSeeker(branchInfo.Branch.Repo.Name, branchInfo.Head.ID, file)
	if err != nil {
		internalError(w, r, err)
		return
	}

	http.ServeContent(w, r, file, timestamp, reader)
}

func (h *objectHandler) put(w http.ResponseWriter, r *http.Request) {
	repo, branch, file := objectArgs(w, r)
	branchInfo, err := h.pc.InspectBranch(repo, branch)
	if err != nil {
		notFoundError(w, r, err)
		return
	}
	if isMeta(file) || strings.HasSuffix(file, "/") {
		invalidFilePathError(w, r)
		return
	}

	expectedHash, ok := r.Header["Content-Md5"]
	var expectedHashBytes []uint8
	if ok && len(expectedHash) == 1 {
		expectedHashBytes, err = base64.StdEncoding.DecodeString(expectedHash[0])
		if err != nil || len(expectedHashBytes) != 16 {
			invalidDigestError(w, r)
			return
		}
	}

	hasher := md5.New()
	reader := io.TeeReader(r.Body, hasher)

	client, err := h.pc.NewPutFileClient()
	if err != nil {
		internalError(w, r, err)
		return
	}
	defer func() {
		if err := client.Close(); err != nil {
			requestLogger(r).Errorf("could not close put file client: %v", err)
		}
	}()

	_, err = client.PutFileOverwrite(branchInfo.Branch.Repo.Name, branchInfo.Branch.Name, file, reader, 0)
	if err != nil {
		internalError(w, r, err)
		return
	}

	errored := true
	defer func() {
		if errored {
			// try to clean up the file if an error occurred
			if err := h.pc.DeleteFile(branchInfo.Branch.Repo.Name, branchInfo.Branch.Name, file); err != nil {
				requestLogger(r).Errorf("could not cleanup file after an error: %v", err)
			}
		}
	}()

	actualHashBytes := hasher.Sum(nil)
	actualHash := fmt.Sprintf("%x", actualHashBytes)
	if expectedHashBytes != nil && !bytes.Equal(expectedHashBytes, actualHashBytes) {
		badDigestError(w, r)
		return
	}

	meta := ObjectMeta { MD5: actualHash }
	if err = putMeta(client, branchInfo.Branch.Repo.Name, branchInfo.Branch.Name, file, &meta); err != nil {
		internalError(w, r, err)
		return
	}

	w.Header().Set("ETag", fmt.Sprintf("\"%s\"", actualHash))
	w.WriteHeader(http.StatusOK)
	errored = false
}

func (h *objectHandler) del(w http.ResponseWriter, r *http.Request) {
	repo, branch, file := objectArgs(w, r)
	branchInfo, err := h.pc.InspectBranch(repo, branch)
	if err != nil {
		notFoundError(w, r, err)
		return
	}
	if branchInfo.Head == nil {
		noSuchKeyError(w, r)
		return
	}
	if isMeta(file) || strings.HasSuffix(file, "/") {
		invalidFilePathError(w, r)
		return
	}

	if err = delMeta(h.pc, branchInfo.Branch.Repo.Name, branchInfo.Branch.Name, file); err != nil {
		internalError(w, r, err)
		return
	}
	if err := h.pc.DeleteFile(branchInfo.Branch.Repo.Name, branchInfo.Branch.Name, file); err != nil {
		notFoundError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
