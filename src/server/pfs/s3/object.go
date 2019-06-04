package s3

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	pfsServer "github.com/pachyderm/pachyderm/src/server/pfs"
)

type objectHandler struct {
	pc   *client.APIClient
	view map[string]*pfs.Commit
}

func newObjectHandler(pc *client.APIClient, view map[string]*pfs.Commit) *objectHandler {
	return &objectHandler{pc: pc, view: view}
}

func (h *objectHandler) get(w http.ResponseWriter, r *http.Request) {
	repo, commit, file := objectArgs(w, r, h.view)
	ci, err := h.pc.InspectCommit(repo, commit)
	if err != nil {
		maybeNotFoundError(w, r, err)
		return
	}
	// commit returned by objectArgs might be a branch which might get
	// committed to while this function runs, this freezes it.
	commit = ci.Commit.ID
	if strings.HasSuffix(file, "/") {
		invalidFilePathError(w, r)
		return
	}

	fileInfo, err := h.pc.InspectFile(repo, commit, file)
	if err != nil {
		maybeNotFoundError(w, r, err)
		return
	}

	timestamp, err := types.TimestampFromProto(fileInfo.Committed)
	if err != nil {
		internalError(w, r, err)
		return
	}

	w.Header().Set("ETag", fmt.Sprintf("\"%x\"", fileInfo.Hash))
	reader, err := h.pc.GetFileReadSeeker(repo, commit, file)
	if err != nil {
		internalError(w, r, err)
		return
	}

	http.ServeContent(w, r, file, timestamp, reader)
}

func (h *objectHandler) put(w http.ResponseWriter, r *http.Request) {
	repo, commit, file := objectArgs(w, r, h.view)
	ci, err := h.pc.InspectCommit(repo, commit)
	// BranchNoHeadErr is acceptable because the put will create the head
	if err != nil && !pfsServer.IsBranchNoHeadErr(err) {
		maybeNotFoundError(w, r, err)
		return
	}
	// commit returned by objectArgs might be a branch which might get
	// committed to while this function runs, this freezes it.
	commit = ci.Commit.ID
	if strings.HasSuffix(file, "/") {
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
	success := false

	_, err = h.pc.PutFileOverwrite(repo, commit, file, reader, 0)
	if err != nil {
		internalError(w, r, err)
		return
	}

	defer func() {
		// try to clean up the file if an error occurred
		if !success {
			if err = h.pc.DeleteFile(repo, commit, file); err != nil {
				requestLogger(r).Errorf("could not cleanup file after an error: %v", err)
			}
		}
	}()

	actualHashBytes := hasher.Sum(nil)
	if expectedHashBytes != nil && !bytes.Equal(expectedHashBytes, actualHashBytes) {
		badDigestError(w, r)
		return
	}

	fileInfo, err := h.pc.InspectFile(repo, commit, file)
	if err != nil {
		internalError(w, r, err)
		return
	}

	success = true
	w.Header().Set("ETag", fmt.Sprintf("\"%x\"", fileInfo.Hash))
	w.WriteHeader(http.StatusOK)
}

func (h *objectHandler) del(w http.ResponseWriter, r *http.Request) {
	repo, commit, file := objectArgs(w, r, h.view)
	ci, err := h.pc.InspectCommit(repo, commit)
	if err != nil {
		maybeNotFoundError(w, r, err)
		return
	}
	// commit returned by objectArgs might be a branch which might get
	// committed to while this function runs, this freezes it.
	commit = ci.Commit.ID
	if strings.HasSuffix(file, "/") {
		invalidFilePathError(w, r)
		return
	}

	if err := h.pc.DeleteFile(repo, commit, file); err != nil {
		maybeNotFoundError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
