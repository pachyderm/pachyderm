package fuse

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/mux"
	"github.com/pachyderm/pachyderm/v2/src/client"
)

/*

A long-running server which dynamically manages FUSE mounts in the given
directory, exposing a gRPC API on a UNIX socket.

API
===

List()
------

Returns a complete set of repos in the system, their branches and commits, and
for each repo any branch or commit that is currently mounted, the status of that
mount on the system.


Get(repo)
---------
Return the subset of the output of List that is specific to the given repo.


MountBranch(repo, branch, mode)
-------------------------------
Mount the given branch of the repo into the configured versioned and unversioned
mount directories.

Will result in `/_pfs/{repo}@{branch}` being mounted, and `/pfs/{repo}` being
configured as a bind-mount to point to the same location.


MountCommit(repo, branch, commit)
---------------------------------
TODO: implement this.

Mount the specified commit or global commit of the given branch of the given
repo into the configured directories.

Will result in `/_pfs/{repo}@{branch}:{commit}` being mounted, and `/pfs/{repo}`
being configured as a bind-mount to point to the same location.

If the commit is open, the mount may be writeable until it's closed.
(TODO: Maybe commits should only be mounted read-only to simplify?)

If the commit is closed, the reader can assume the data is immutable.

Note that repo-branch-commit mounts are disjoint from repo-branch mounts.


UnmountBranch(repo, branch)
---------------------------
Unmount the given repo-branch pair. If the repo-branch is not currently mounted,
returns an error.


UnmountCommit(repo, branch, commit)
-----------------------------------
Unmount the given repo-branch-commit pair. If the repo-branch-commit is not
currently mounted, returns an error.

Note that repo-branch-commit mounts are disjoint from repo-branch mounts.


CommitBranch(repo, branch, message)
-----------------------------------
TODO: implement this

Persist the current state of the given mount to a new commit in Pachyderm.

CommitBranch is not supported on repo-branch-commit mounts (as they are not
writeable?)


Plan
====

1. Implement List(), Get(), MountBranch(), UnmountBranch()
   Commits are implicit when filesystems are unmounted (matches current fuse mount behavior)

2. Implement CommitBranch() and stop implicitly committing data on unmounts

3. Implement MountCommit() to support looking at a specific previous state, read-only

*/

type ServerOptions struct {
	Daemonize            bool
	VersionedMountsDir   string
	UnversionedMountsDir string
	Socket               string
	LogFile              string
}

type MountBranchRequest struct {
	Repo   string
	Branch string
	Mode   string // "ro", "rw"
}

type MountBranchResponse struct {
	Repo       string
	Branch     string
	MountState MountState
}

type MountManager struct {
	States        map[MountKey]MountState
	RequestChans  map[MountKey]chan MountBranchRequest
	ResponseChans map[MountKey]chan MountBranchResponse
	mu            sync.Mutex
}

func (mm *MountManager) Run() error {
	return nil
}

func (mm *MountManager) MountBranch() error {

}

func Server(c *client.APIClient, opts *ServerOptions) error {
	// TODO: respect opts.Daemonize
	mm := MountManager{}
	mm.Run()

	router := mux.NewRouter()
	router.Methods("GET").Path("/repos").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		mm.mu.Lock()
		defer mm.mu.Unlock()
		// XXX how will json.Marshal deal with our map keys being structs? We
		// might need to flatten this into a map[string]MountStates here using
		// key.String()
		marshalled, err := json.Marshal(mm.States)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		w.Write(marshalled)
	})
	router.Methods("PUT").Path("/repos/{key:.+}/_mount/?{mode}").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		vs := mux.Vars(req)
		key := mountKeyFromString(vs["key"])
		mm.MountBranch(key, key.Repo, key.Branch)

	})
	router.Methods("PUT").Path("/repos/{key:.+}/_unmount").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
	})
	// TODO: implement _commit

	// TODO: switch http server for gRPC server and bind to a unix socket not a
	// TCP port (just for convenient manual testing with curl for now...)
	http.Handle("/", router)
	// TODO: make port and bind ip parameterizable
	return http.ListenAndServe(":9001", nil)
}

type MountState struct {
	State  string // "unmounted", "mounting", "mounted", "pushing", "unmounted", "error"
	Status string // human readable string with additional info wrt State, e.g. an error message for the error state
	Mode   string // "ro", "rw", or "" if unknown/unspecified
}

type BranchResponse struct {
	Name  string
	Mount MountState // will be MountResponse{State: "unmounted"} if not mounted
}

// TODO: switch to pach internal types if appropriate?
type RepoResponse struct {
	Name     string
	Branches []BranchResponse
	// TODO: Commits []CommitResponse
}

type GetResponse RepoResponse

type ListResponse struct {
	Repos []RepoResponse
}

type MountKey struct {
	Repo   string
	Branch string
	Commit string
}

func String(m *MountKey) string {
	// m.Commit is optional
	if m.Commit == "" {
		return fmt.Sprintf("%s/%s", m.Repo, m.Branch)
	} else {
		return fmt.Sprintf("%s/%s/%s", m.Repo, m.Branch, m.Commit)
	}
}

func mountKeyFromString(key string) (MountKey, error) {
	shrapnel := strings.Split(key, "/")
	if len(shrapnel) == 3 {
		return MountKey{
			Repo:   shrapnel[0],
			Branch: shrapnel[1],
			Commit: shrapnel[2],
		}, nil
	} else if len(shrapnel) == 2 {
		return MountKey{
			Repo:   shrapnel[0],
			Branch: shrapnel[1],
		}, nil
	} else {
		return MountKey{}, fmt.Errorf(
			"Wrong number of '/'s in key %+v, got %d expected 2 or 3", key, len(shrapnel),
		)
	}
}
