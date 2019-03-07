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

	fileInfo, err := h.pc.InspectFile(branchInfo.Branch.Repo.Name, branchInfo.Branch.Name, file)
	if err != nil {
		notFoundError(w, r, err)
		return
	}

	timestamp, err := types.TimestampFromProto(fileInfo.Committed)
	if err != nil {
		internalError(w, r, err)
		return
	}

	reader, err := h.pc.GetFileReadSeeker(branchInfo.Branch.Repo.Name, branchInfo.Branch.Name, file)
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

	withBodyReader(w, r, func(reader io.Reader) bool {
		_, err := h.pc.PutFileOverwrite(branchInfo.Branch.Repo.Name, branchInfo.Branch.Name, file, reader, 0)

		if err != nil {
			internalError(w, r, err)
			return false
		}

		return true
	})
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

	if err := h.pc.DeleteFile(branchInfo.Branch.Repo.Name, branchInfo.Branch.Name, file); err != nil {
		notFoundError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
