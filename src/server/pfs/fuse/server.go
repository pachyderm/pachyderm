package fuse

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	pathpkg "path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gorilla/mux"
	"github.com/hanwen/go-fuse/v2/fs"
	gofuse "github.com/hanwen/go-fuse/v2/fuse"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/progress"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
)

// TODO: split out into multiple files

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


MountBranch(repo, branch, name, mode)
-------------------------------------
Mount the given branch of the repo into the configured mount directory.

E.g. will result in `/pfs/{name}` being mounted.


MountCommit(repo, branch, commit, name)
---------------------------------------
TODO: implement this.

Mount the specified commit or global commit of the given branch of the given
repo into the configured directories.

Will result in `/pfs/{name}` being mounted.

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


UnmountByName(name)
-------------------
Unmount whatever is mounted at e.g. `/pfs/{name}`


CommitBranch(repo, branch, message)
-----------------------------------
TODO: implement this

Persist the current state of the given mount to a new commit in Pachyderm.

CommitBranch is not supported on repo-branch-commit mounts (as they are not
writeable?)


Plan
====

1. Implement List(), Get(), MountBranch(), UnmountBranch(), UnmountByName()
   Commits are implicit when filesystems are unmounted (matches current fuse mount behavior)

2. Implement CommitBranch() and stop implicitly committing data on unmounts

3. Implement MountCommit() to support looking at a specific previous state, read-only

*/

type ServerOptions struct {
	Daemonize bool
	MountDir  string
	Socket    string
	LogFile   string
}

type MountBranchRequest struct {
	Repo   string
	Branch string
	Name   string
	Mode   string // "ro", "rw"
}

type MountBranchResponse struct {
	Repo       string
	Branch     string
	Name       string
	MountState MountState
	Error      error
}

type MountManager struct {
	Client *client.APIClient
	// only put a value into the States map when we have a goroutine running for
	// it. i.e. when we try to mount it for the first time.
	States map[MountKey]*MountStateMachine
	mfcs   map[string]*client.ModifyFileClient
	root   *loopbackRoot
	opts   *Options
	tmpDir string
	target string
	mu     sync.Mutex
}

func (mm *MountManager) List() (ListResponse, error) {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	// fetch list of available repos & branches from pachyderm, and overlay that
	// with their mount states
	lr := ListResponse{Repos: map[string]RepoResponse{}}
	repos, err := mm.Client.ListRepo()
	if err != nil {
		return lr, err
	}
	fmt.Printf("repos: %+v\n", repos)
	for _, repo := range repos {
		rr := RepoResponse{Name: repo.Repo.Name, Branches: map[string]BranchResponse{}}
		bs, err := mm.Client.ListBranch(repo.Repo.Name)
		if err != nil {
			return lr, err
		}
		for _, branch := range bs {
			br := BranchResponse{Name: branch.Branch.Name}
			k := MountKey{
				Repo:   repo.Repo.Name,
				Branch: branch.Branch.Name,
				Commit: "",
			}
			s, ok := mm.States[k]
			if ok {
				br.State = s.MountState
			} else {
				br.State = MountState{State: "unmounted"}
			}
			rr.Branches[branch.Branch.Name] = br
		}
		lr.Repos[repo.Repo.Name] = rr
	}

	// TODO: also add any repos/branches that have been deleted from pachyderm
	// but are still mounted here :-O we should send them a special signal to
	// tell them they're "stranded" or "missing" probably?

	return lr, nil
}

func NewMountStateMachine(key MountKey, mm *MountManager) *MountStateMachine {
	return &MountStateMachine{
		MountState: MountState{
			MountKey: key,
		},
		manager:   mm,
		mu:        sync.Mutex{},
		requests:  make(chan MountBranchRequest),
		responses: make(chan MountBranchResponse),
	}
}

func (mm *MountManager) MaybeStartFsm(key MountKey) {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	_, ok := mm.States[key]
	if !ok {
		// set minimal state and kick it off!
		m := NewMountStateMachine(key, mm)
		mm.States[key] = m
		go func() {
			m.Run()
			// when execution completes, remove ourselves from the map so that
			// we can infer from the key in the map that the state machine is
			// running
			mm.mu.Lock()
			defer mm.mu.Unlock()
			delete(mm.States, key)
		}()
	} // else: fsm was already running
}

func (mm *MountManager) MountBranch(key MountKey, name, mode string) (MountBranchResponse, error) {
	// name: an optional name for the mount, corresponds to where on the
	// filesystem the repo actually gets mounted. e.g. /pfs/{name}
	mm.MaybeStartFsm(key)
	fmt.Println("Sending request...")
	mm.States[key].requests <- MountBranchRequest{
		Repo:   key.Repo,
		Branch: key.Branch,
		Name:   name,
		Mode:   mode,
	}
	fmt.Println("sent!")
	fmt.Println("reading response...")
	response := <-mm.States[key].responses
	fmt.Println("read!")
	return response, response.Error
}

func NewMountManager(c *client.APIClient, target string, opts *Options) (ret *MountManager, retErr error) {
	if err := opts.validate(c); err != nil {
		return nil, err
	}
	commits := make(map[string]string)
	for repo, branch := range opts.getBranches() {
		if uuid.IsUUIDWithoutDashes(branch) {
			commits[repo] = branch
		}
	}
	rootDir, err := ioutil.TempDir("", "pfs")
	if err != nil {
		return nil, errors.WithStack(err)
	}
	fmt.Printf("Creating %s\n", rootDir)
	if err := os.MkdirAll(rootDir, 0777); err != nil {
		return nil, err
	}
	root, err := newLoopbackRoot(rootDir, target, c, opts)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Loopback root at %s\n", rootDir)
	return &MountManager{
		Client: c,
		States: map[MountKey]*MountStateMachine{},
		mfcs:   map[string]*client.ModifyFileClient{},
		root:   root,
		opts:   opts,
		target: target,
		tmpDir: rootDir,
		mu:     sync.Mutex{},
	}, nil
}

func (mm *MountManager) Cleanup() error {
	return os.RemoveAll(mm.tmpDir)
}

func (mm *MountManager) Start() error {
	fuse := mm.opts.getFuse()
	fmt.Printf("Starting mount to %s, fuse=%+v\n", mm.target, fuse)
	server, err := fs.Mount(mm.target, mm.root, fuse)
	if err != nil {
		return errors.WithStack(err)
	}
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	go func() {
		select {
		case <-sigChan:
		case <-mm.opts.getUnmount():
		}
		server.Unmount()
	}()
	server.Wait()
	defer mm.FinishAll()
	err = mm.uploadFiles("")
	if err != nil {
		return err
	}
	return nil
}

func (mm *MountManager) FinishAll() (retErr error) {
	for _, mfc := range mm.mfcs {
		if err := mfc.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}
	return retErr
}

func Server(c *client.APIClient, opts *ServerOptions) error {
	// TODO: respect opts.Daemonize
	fmt.Printf("Dynamically mounting pfs to %s\n", opts.MountDir)

	mountOpts := &Options{
		Write: true,
		Fuse: &fs.Options{
			MountOptions: gofuse.MountOptions{
				Debug:  true,
				FsName: "pfs",
				Name:   "pfs",
			},
		},
		RepoOptions: make(map[string]*RepoOptions),
	}
	fmt.Printf("mountOpts: %+v\n", mountOpts)

	// in server mode, we allow empty mounts. normally, passing no repoOptions
	// makes fuse mount everything, but we actually want to start out with an
	// empty mount in the server case.
	//
	// TODO: make this not be a global!
	ALLOW_EMPTY = true

	mm, err := NewMountManager(c, opts.MountDir, mountOpts)
	if err != nil {
		return err
	}

	go func() {
		err := mm.Start()
		if err != nil {
			fmt.Printf("Error running mount manager: %s\n", err)
			os.Exit(1)
		}
		err = mm.Cleanup()
		if err != nil {
			fmt.Printf("Error cleaning up mount manager: %s\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}()

	router := mux.NewRouter()
	router.Methods("GET").Path("/repos").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		l, err := mm.List()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		marshalled, err := json.Marshal(l)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Write(marshalled)
	})
	router.Methods("PUT").
		Path("/repos/{key:.+}/_mount").
		Queries("mode", "{mode}").
		Queries("name", "{name}").
		HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			vs := mux.Vars(req)
			mode, ok := vs["mode"]
			if !ok {
				http.Error(w, "no mode", http.StatusBadRequest)
				return
			}
			name, ok := vs["name"]
			if !ok {
				http.Error(w, "no name", http.StatusBadRequest)
				return
			}
			k, ok := vs["key"]
			if !ok {
				http.Error(w, "no key", http.StatusBadRequest)
				return
			}
			key, err := mountKeyFromString(k)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if key.Commit != "" {
				http.Error(
					w,
					"don't support mounting commits yet",
					http.StatusBadRequest,
				)
				return
			}
			// TODO: use response (serialize it to the client, it's polite to hand
			// back the object you just modified in the API response)
			_, err = mm.MountBranch(key, name, mode)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		})
	router.Methods("PUT").Path("/repos/{key:.+}/_unmount").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
	})
	// TODO: implement _commit

	// TODO: switch http server for gRPC server and bind to a unix socket not a
	// TCP port (just for convenient manual testing with curl for now...)
	http.Handle("/", router)
	// TODO: make port and bind ip parameterizable
	return http.ListenAndServe(":9002", nil)
}

type MountState struct {
	Name       string
	MountKey   MountKey
	State      string // "unmounted", "mounting", "mounted", "pushing", "unmounted", "error"
	Status     string // human readable string with additional info wrt State, e.g. an error message for the error state
	Mode       string // "ro", "rw", or "" if unknown/unspecified
	Mountpoint string // where on the filesystem it's mounted
}

type MountStateMachine struct {
	MountState
	manager   *MountManager
	mu        sync.Mutex
	requests  chan MountBranchRequest
	responses chan MountBranchResponse
}

// TODO: switch to pach internal types if appropriate?
type ListResponse struct {
	Repos map[string]RepoResponse
}

type BranchResponse struct {
	Name  string
	State MountState
}

type RepoResponse struct {
	Name     string
	Branches map[string]BranchResponse
	// TODO: Commits map[string]CommitResponse
}

type GetResponse RepoResponse

type MountKey struct {
	Repo   string
	Branch string
	Commit string
}

func (m *MountKey) String() string {
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
			"wrong number of '/'s in key %+v, got %d expected 2 or 3", key, len(shrapnel),
		)
	}
}

// state machines in the style of Rob Pike's Lexical Scanning
// http://www.timschmelmer.com/2013/08/state-machines-with-state-functions-in.html

// setup

type StateFn func(*MountStateMachine) StateFn

func (m *MountStateMachine) transitionedTo(state, status string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	fmt.Printf("[%s] %s -> %s\n", m.MountKey, m.State, state)
	m.State = state
	m.Status = status
}

func (m *MountStateMachine) Run() {
	for state := discoveringState; state != nil; {
		state = state(m)
	}
	m.transitionedTo("gone", "")
}

// state functions

func discoveringState(m *MountStateMachine) StateFn {
	m.transitionedTo("discovering", "")
	// TODO: inspect the state of the system, and if we're half-unmounted, maybe
	// do cleanup before declaring us _truly_ unmounted
	return unmountedState
}

func unmountedState(m *MountStateMachine) StateFn {
	m.transitionedTo("unmounted", "")
	// TODO: listen on our request chan, mount filesystems, and respond
	for {
		req := <-m.requests

		fmt.Printf("Read req: %+v\n", req)
		m.responses <- MountBranchResponse{
			Repo:       m.MountKey.Repo,
			Branch:     m.MountKey.Branch,
			Name:       m.Name,
			MountState: m.MountState,
			Error:      nil,
		}
		return mountingState
		// TODO: actually mount that shit!
	}
}

func mountingState(m *MountStateMachine) StateFn {
	m.transitionedTo("mounting", "")
	// TODO: refactor this so we're not reaching into another struct's lock
	func() {
		m.manager.mu.Lock()
		defer m.manager.mu.Unlock()
		m.manager.root.repoOpts[m.MountKey.Repo] = &RepoOptions{
			Branch: m.MountKey.Repo,
			Write:  m.Mode == "rw",
		}
	}()
	// re-downloading the repos with an updated RepoOptions set will have the
	// effect of causing it to pop into existence, in theory

	// TODO: Test this!
	err := m.manager.root.downloadRepos()
	if err != nil {
		m.transitionedTo("error", err.Error())
		return errorState
	}
	return mountedState
}

func mountedState(m *MountStateMachine) StateFn {
	m.transitionedTo("mounted", "")
	for {
		// TODO: check request type. unmount requests can be satisfied by going
		// into unmounting. mount requests for already mounted repos should
		// go back into mounting, as they can remount a fs as ro/rw.
		//
		// XXX TODO: need to start respecting name rather than always mounting
		// repos at their own name in /pfs
		<-m.requests
		m.responses <- MountBranchResponse{}
	}
}

func errorState(m *MountStateMachine) StateFn {
	for {
		// eternally error on responses
		<-m.requests
		m.responses <- MountBranchResponse{
			Repo:       m.MountKey.Repo,
			Branch:     m.MountKey.Branch,
			Name:       m.Name,
			MountState: m.MountState,
			Error:      fmt.Errorf("in error state, last error message was: %s", m.Status),
		}
	}
}

func (mm *MountManager) mfc(repo string) (*client.ModifyFileClient, error) {
	if mfc, ok := mm.mfcs[repo]; ok {
		return mfc, nil
	}
	mfc, err := mm.Client.NewModifyFileClient(client.NewCommit(repo, mm.root.branch(repo), ""))
	if err != nil {
		return nil, err
	}
	mm.mfcs[repo] = mfc
	return mfc, nil
}

func (mm *MountManager) uploadFiles(prefixFilter string) error {
	fmt.Println("Uploading files to Pachyderm...")
	// Rendering progress bars for thousands of files significantly slows down
	// throughput. Disabling progress bars takes throughput from 1MB/sec to
	// 200MB/sec on my system, when uploading 18K small files.
	progress.Disable()
	for path, state := range mm.root.files {
		if !strings.HasPrefix(path, prefixFilter) {
			continue
		}
		if state != dirty {
			continue
		}
		parts := strings.Split(path, "/")
		mfc, err := mm.mfc(parts[0])
		if err != nil {
			return err
		}
		if err := func() (retErr error) {
			f, err := progress.Open(filepath.Join(mm.root.rootPath, path))
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					return mfc.DeleteFile(pathpkg.Join(parts[1:]...))
				}
				return errors.WithStack(err)
			}
			defer func() {
				if err := f.Close(); err != nil && retErr == nil {
					retErr = errors.WithStack(err)
				}
			}()
			return mfc.PutFile(pathpkg.Join(parts[1:]...), f)
		}(); err != nil {
			return err
		}
	}
	fmt.Println("Done!")
	return nil
}
