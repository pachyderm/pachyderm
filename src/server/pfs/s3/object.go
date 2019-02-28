package s3

import (
	"io"
	"net/http"

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
	repo, branch, file := objectArgs(r)
	branchInfo, err := h.pc.InspectBranch(repo, branch)
	if err != nil {
		writeMaybeNotFound(w, r, err)
		return
	}
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

func (h *objectHandler) put(w http.ResponseWriter, r *http.Request) {
	repo, branch, file := objectArgs(r)
	branchInfo, err := h.pc.InspectBranch(repo, branch)
	if err != nil {
		writeMaybeNotFound(w, r, err)
		return
	}

	success, err := withBodyReader(r, func(reader io.Reader) bool {
		_, err := h.pc.PutFileOverwrite(branchInfo.Branch.Repo.Name, branchInfo.Branch.Name, file, reader, 0)

		if err != nil {
			writeServerError(w, err)
			return false
		}

		return true
	})

	// if there's no error but the operation is not successful, we've already
	// written a response to the client
	if err != nil {
		writeBadRequest(w, err)
	} else if success {
		w.WriteHeader(http.StatusOK)
	}
}

func (h *objectHandler) del(w http.ResponseWriter, r *http.Request) {
	repo, branch, file := objectArgs(r)
	branchInfo, err := h.pc.InspectBranch(repo, branch)
	if err != nil {
		writeMaybeNotFound(w, r, err)
		return
	}
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
