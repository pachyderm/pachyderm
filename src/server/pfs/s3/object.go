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
		newNotFoundError(r, err).write(w)
		return
	}
	if branchInfo.Head == nil {
		newNoSuchKeyError(r).write(w)
		return
	}

	fileInfo, err := h.pc.InspectFile(branchInfo.Branch.Repo.Name, branchInfo.Branch.Name, file)
	if err != nil {
		newNotFoundError(r, err).write(w)
		return
	}

	timestamp, err := types.TimestampFromProto(fileInfo.Committed)
	if err != nil {
		newInternalError(r, err).write(w)
		return
	}

	reader, err := h.pc.GetFileReadSeeker(branchInfo.Branch.Repo.Name, branchInfo.Branch.Name, file)
	if err != nil {
		newInternalError(r, err).write(w)
		return
	}

	http.ServeContent(w, r, "", timestamp, reader)
}

func (h *objectHandler) put(w http.ResponseWriter, r *http.Request) {
	repo, branch, file := objectArgs(r)
	branchInfo, err := h.pc.InspectBranch(repo, branch)
	if err != nil {
		newNotFoundError(r, err).write(w)
		return
	}

	withBodyReader(w, r, func(reader io.Reader) bool {
		_, err := h.pc.PutFileOverwrite(branchInfo.Branch.Repo.Name, branchInfo.Branch.Name, file, reader, 0)

		if err != nil {
			newInternalError(r, err).write(w)
			return false
		}

		return true
	})
}

func (h *objectHandler) del(w http.ResponseWriter, r *http.Request) {
	repo, branch, file := objectArgs(r)
	branchInfo, err := h.pc.InspectBranch(repo, branch)
	if err != nil {
		newNotFoundError(r, err).write(w)
		return
	}
	if branchInfo.Head == nil {
		newNoSuchKeyError(r).write(w)
		return
	}

	if err := h.pc.DeleteFile(branchInfo.Branch.Repo.Name, branchInfo.Branch.Name, file); err != nil {
		newNotFoundError(r, err).write(w)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
