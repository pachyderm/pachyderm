package s3

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"encoding/json"
	"fmt"

	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/sirupsen/logrus"
)

type ObjectMeta struct {
	md5 string `json:"md5"`
}

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
	if strings.HasSuffix(file, ".s3g.json") {
		invalidFilePathError(w, r)
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

	metaReader, err := h.pc.GetFileReader(branchInfo.Branch.Repo.Name, branchInfo.Branch.Name, fmt.Sprintf("%s.s3g.json", file), 0, 0)
	if err != nil {
		if !fileNotFoundMatcher.MatchString(err.Error()) {
			internalError(w, r, err)
			return
		}
	} else {
		meta := new(ObjectMeta)
		if err = json.NewDecoder(metaReader).Decode(meta); err != nil {
			if !fileNotFoundMatcher.MatchString(err.Error()) {
				internalError(w, r, err)
				return
			}
		} else {
			w.Header().Set("ETag", meta.md5)
		}
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
	if strings.HasSuffix(file, ".s3g.json") {
		invalidFilePathError(w, r)
		return
	}

	commit, err := h.pc.StartCommit(branchInfo.Branch.Repo.Name, branchInfo.Branch.Name)
	if err != nil {
		internalError(w, r, err)
		return
	}

	finished := withBodyReader(w, r, func(reader io.Reader, hash []uint8) bool {
		_, err = h.pc.PutFileOverwrite(branchInfo.Branch.Repo.Name, commit.ID, file, reader, 0)
		if err != nil {
			internalError(w, r, err)
			return false
		}

		if hash != nil {
			meta := ObjectMeta { 
				md5: fmt.Sprintf("%x", hash),
			}
			metaBytes, err := json.Marshal(meta)
			if err != nil {
				panic(err)
			}
			metaReader := bytes.NewReader(metaBytes)
			metaFile := fmt.Sprintf("%s.s3g.json", file)
			_, err = h.pc.PutFileOverwrite(branchInfo.Branch.Repo.Name, commit.ID, metaFile, metaReader, 0)
			if err != nil {
				internalError(w, r, err)
				return false
			}
		}

		return true
	})

	if finished {
		err = h.pc.FinishCommit(branchInfo.Branch.Repo.Name, commit.ID)
		if err != nil {
			logrus.Errorf("s3gateway: could not finish commit: %v", err)
		}
	} else {
		// TODO: is this right?
		err = h.pc.DeleteCommit(branchInfo.Branch.Repo.Name, commit.ID)
		if err != nil {
			logrus.Errorf("s3gateway: could not delete commit: %v", err)
		}
	}
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
	if strings.HasSuffix(file, ".s3g.json") {
		invalidFilePathError(w, r)
		return
	}

	commit, err := h.pc.StartCommit(branchInfo.Branch.Repo.Name, branchInfo.Branch.Name)
	if err != nil {
		internalError(w, r, err)
		return
	}

	isFinished := false
	defer func() {
		if !isFinished {
			if err := h.pc.FinishCommit(branchInfo.Branch.Repo.Name, commit.ID); err != nil {
				logrus.Errorf("s3gateway: could not finish commit: %v", err)
			}
		}
		
	}()

	if err := h.pc.DeleteFile(branchInfo.Branch.Repo.Name, commit.ID, file); err != nil {
		notFoundError(w, r, err)
		return
	}

	if err = h.pc.DeleteFile(branchInfo.Branch.Repo.Name, commit.ID, fmt.Sprintf("%s.s3g.json", file)); err != nil {
		// ignore errors related to the metadata file not being found, since
		// it may validly not exist
		if !fileNotFoundMatcher.MatchString(err.Error()) {
			internalError(w, r, err)
			return
		}
	}

	isFinished = true
	if err := h.pc.FinishCommit(branchInfo.Branch.Repo.Name, commit.ID); err != nil {
		internalError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
