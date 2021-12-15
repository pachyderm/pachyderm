package fuse

import (
	"context"
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
	// Unmount is a channel that will be closed when the filesystem has been
	// unmounted. It can be nil in which case it's ignored.
	Unmount chan struct{}
}

type Request struct {
	Mount  bool // true for desired state == mounted, false for desired state == unmounted
	Repo   string
	Branch string
	Commit string // "" for no commit
	Name   string
	Mode   string // "ro", "rw"
}

type Response struct {
	Repo       string
	Branch     string
	Commit     string // "" for no commit
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
	lr := map[string]RepoResponse{}
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
				br.Mount = s.MountState
			} else {
				br.Mount = MountState{State: "unmounted"}
			}
			rr.Branches[branch.Branch.Name] = br
		}
		lr[repo.Repo.Name] = rr
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
		requests:  make(chan Request),
		responses: make(chan Response),
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

func (mm *MountManager) MountBranch(key MountKey, name, mode string) (Response, error) {
	// name: an optional name for the mount, corresponds to where on the
	// filesystem the repo actually gets mounted. e.g. /pfs/{name}
	mm.MaybeStartFsm(key)
	fmt.Println("Sending mount request...")
	mm.States[key].requests <- Request{
		Mount:  true,
		Repo:   key.Repo,
		Branch: key.Branch,
		Commit: key.Commit,
		Name:   name,
		Mode:   mode,
	}
	fmt.Println("sent mount request!")
	fmt.Println("reading mount response...")
	response := <-mm.States[key].responses
	fmt.Println("read mount response!")
	return response, response.Error
}

func (mm *MountManager) UnmountBranch(key MountKey, name string) (Response, error) {
	mm.MaybeStartFsm(key)
	fmt.Println("Sending unmount request...")
	mm.States[key].requests <- Request{
		Mount:  false,
		Repo:   key.Repo,
		Branch: key.Branch,
		Commit: key.Commit,
		Name:   name,
		Mode:   "",
	}
	fmt.Println("sent unmount request!")
	fmt.Println("reading unmount response...")
	response := <-mm.States[key].responses
	fmt.Println("read unmount response!")
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

func Server(c *client.APIClient, sopts *ServerOptions) error {
	// TODO: respect opts.Daemonize
	fmt.Printf("Dynamically mounting pfs to %s\n", sopts.MountDir)

	mountOpts := &Options{
		Write: true,
		Fuse: &fs.Options{
			MountOptions: gofuse.MountOptions{
				Debug:  false,
				FsName: "pfs",
				Name:   "pfs",
			},
		},
		RepoOptions: make(map[string]*RepoOptions),
		// thread this through for the tests
		Unmount: sopts.Unmount,
	}
	fmt.Printf("mountOpts: %+v\n", mountOpts)

	mm, err := NewMountManager(c, sopts.MountDir, mountOpts)
	if err != nil {
		return err
	}

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
	router.Methods("PUT").
		Queries("name", "{name}").
		Path("/repos/{key:.+}/_unmount").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		vs := mux.Vars(req)
		k, ok := vs["key"]
		if !ok {
			http.Error(w, "no key", http.StatusBadRequest)
			return
		}
		name, ok := vs["name"]
		if !ok {
			http.Error(w, "no name", http.StatusBadRequest)
			return
		}
		key, err := mountKeyFromString(k)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		_, err = mm.UnmountBranch(key, name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
	// TODO: implement _commit

	// TODO: switch http server for gRPC server and bind to a unix socket not a
	// TCP port (just for convenient manual testing with curl for now...)
	// TODO: make port and bind ip parameterizable
	srv := &http.Server{Addr: ":9002", Handler: router}

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
		srv.Shutdown(context.Background())
	}()

	return srv.ListenAndServe()
}

type MountState struct {
	Name       string   `json:"name"`       // where to mount it. written by client
	MountKey   MountKey `json:"mount_key"`  // what to mount. written by client
	Mode       string   `json:"mode"`       // "ro", "rw", or "" if unknown/unspecified. written by client
	State      string   `json:"state"`      // "unmounted", "mounting", "mounted", "pushing", "unmounted", "error". written by fsm
	Status     string   `json:"status"`     // human readable string with additional info wrt State, e.g. an error message for the error state. written by fsm
	Mountpoint string   `json:"mountpoint"` // where on the filesystem it's mounted. written by fsm. can also be derived from {MountDir}/{Name}
}

type MountStateMachine struct {
	MountState
	manager   *MountManager
	mu        sync.Mutex
	requests  chan Request
	responses chan Response
}

// TODO: switch to pach internal types if appropriate?
type ListResponse map[string]RepoResponse

type BranchResponse struct {
	Name  string     `json:"name"`
	Mount MountState `json:"mount"`
}

type RepoResponse struct {
	Name     string                    `json:"name"`
	Branches map[string]BranchResponse `json:"branches"`
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

		if req.Mount {
			// copy data from request into fields that are documented as being
			// written by the client (see MountState struct)
			m.MountState.Name = req.Name
			m.MountState.MountKey = MountKey{
				Repo:   req.Repo,
				Branch: req.Branch,
				Commit: req.Commit,
			}
			m.MountState.Mode = req.Mode
			return mountingState
		} else {
			m.responses <- Response{
				Repo:       m.MountKey.Repo,
				Branch:     m.MountKey.Branch,
				Commit:     m.MountKey.Commit,
				Name:       m.Name,
				MountState: m.MountState,
				Error:      fmt.Errorf("can't unmount when we're already unmounted"),
			}
			// stay unmounted
		}
	}
}

func mountingState(m *MountStateMachine) StateFn {
	// NB: this function is responsible for placing a response on m.responses
	// _in all cases_
	m.transitionedTo("mounting", "")
	// TODO: refactor this so we're not reaching into another struct's lock
	func() {
		m.manager.mu.Lock()
		defer m.manager.mu.Unlock()
		m.manager.root.repoOpts[m.MountState.Name] = &RepoOptions{
			Repo:   m.MountKey.Repo,
			Branch: m.MountKey.Repo,
			Write:  m.Mode == "rw",
		}
		fmt.Printf("*** MOUNTINGSTATE repoOpts: %+v\n", m.manager.root.repoOpts)
	}()
	// re-downloading the repos with an updated RepoOptions set will have the
	// effect of causing it to pop into existence
	err := m.manager.root.mkdirMountNames()
	m.responses <- Response{
		Repo:       m.MountKey.Repo,
		Branch:     m.MountKey.Branch,
		Commit:     m.MountKey.Commit,
		Name:       m.Name,
		MountState: m.MountState,
		Error:      err,
	}
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
		req := <-m.requests
		if req.Mount {
			// TODO: handle remount case (switching mode):
			// if mounted ro and switching to rw, just upgrade
			// if mounted rw and switching to ro, upload changes, then switch the mount type
		} else {
			return unmountingState
		}
		m.responses <- Response{}
	}
}

// match exact directory or file within directory but not directory
// which is a substring prefix
// e.g., for prefix abc
// delete abc
// delete abc/foo
// but do NOT delete abcde
func cleanByPrefixStrings(theMap map[string]string, prefix string) {
	for k := range theMap {
		if k == prefix || strings.HasPrefix(k, prefix+"/") {
			delete(theMap, k)
		}
	}
}

// same as above for fileState type maps. _sigh_ .oO { generics! }
func cleanByPrefixFileStates(theMap map[string]fileState, prefix string) {
	for k := range theMap {
		if k == prefix || strings.HasPrefix(k, prefix+"/") {
			delete(theMap, k)
		}
	}
}

func unmountingState(m *MountStateMachine) StateFn {
	// NB: this function is responsible for placing a response on m.responses
	// _in all cases_
	m.transitionedTo("unmounting", "")
	// For now unmountingState is the only place where actual uploads to
	// Pachyderm happen. In the future, we want to split this out to a separate
	// committingState so that users can commit a repo while it's still mounted.

	// TODO XXX VERY IMPORTANT: pause/block filesystem operations during the
	// upload, otherwise we could get filesystem inconsistency! Need a sort of
	// lock which multiple fs operations can hold but only one "pauser" can.

	// upload any files whose paths start with where we're mounted
	err := m.manager.uploadFiles(m.Name)
	if err != nil {
		fmt.Printf("Error while uploading! %s", err)
		m.transitionedTo("error", err.Error())
		m.responses <- Response{
			Repo:       m.MountKey.Repo,
			Branch:     m.MountKey.Branch,
			Commit:     m.MountKey.Commit,
			Name:       m.Name,
			MountState: m.MountState,
			Error:      err,
		}
		return errorState
	}

	// close the mfc, uploading files, then delete it
	mfc, err := m.manager.mfc(m.Name)
	if err != nil {
		fmt.Printf("Error while getting mfc! %s", err)
		m.transitionedTo("error", err.Error())
		m.responses <- Response{
			Repo:       m.MountKey.Repo,
			Branch:     m.MountKey.Branch,
			Commit:     m.MountKey.Commit,
			Name:       m.Name,
			MountState: m.MountState,
			Error:      err,
		}
		return errorState
	}
	err = mfc.Close()
	if err != nil {
		fmt.Printf("Error while closing mfc! %s", err)
		m.transitionedTo("error", err.Error())
		m.responses <- Response{
			Repo:       m.MountKey.Repo,
			Branch:     m.MountKey.Branch,
			Commit:     m.MountKey.Commit,
			Name:       m.Name,
			MountState: m.MountState,
			Error:      err,
		}
		return errorState
	}

	// cleanup
	func() {
		m.manager.mu.Lock()
		defer m.manager.mu.Unlock()
		delete(m.manager.mfcs, m.Name)
	}()
	func() {
		m.manager.root.mu.Lock()
		defer m.manager.root.mu.Unlock()
		// forget what we knew about the mount
		delete(m.manager.root.repoOpts, m.MountState.Name)
		cleanByPrefixStrings(m.manager.root.branches, m.MountState.Name)
		cleanByPrefixStrings(m.manager.root.commits, m.MountState.Name)
		cleanByPrefixFileStates(m.manager.root.files, m.MountState.Name)
	}()

	// remove from loopback filesystem so that it actually disappears for the user
	cleanPath := m.manager.root.rootPath + "/" + m.Name
	fmt.Printf("Path is %s\n", cleanPath)

	err = os.RemoveAll(cleanPath)
	m.responses <- Response{
		Repo:       m.MountKey.Repo,
		Branch:     m.MountKey.Branch,
		Commit:     m.MountKey.Commit,
		Name:       m.Name,
		MountState: m.MountState,
		Error:      err,
	}
	if err != nil {
		fmt.Printf("Error while cleaning! %s", err)
		m.transitionedTo("error", err.Error())
		return errorState
	}

	return unmountedState
}

func errorState(m *MountStateMachine) StateFn {
	for {
		// eternally error on responses
		<-m.requests
		m.responses <- Response{
			Repo:       m.MountKey.Repo,
			Branch:     m.MountKey.Branch,
			Commit:     m.MountKey.Commit,
			Name:       m.Name,
			MountState: m.MountState,
			Error:      fmt.Errorf("in error state, last error message was: %s", m.Status),
		}
	}
}

// given the name of a _mount_, hand back a ModifyFileClient that's configured
// to upload to that repo.
func (mm *MountManager) mfc(name string) (*client.ModifyFileClient, error) {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	if mfc, ok := mm.mfcs[name]; ok {
		return mfc, nil
	}
	var repoName string
	opts, ok := mm.root.repoOpts[name]
	if !ok {
		// assume the repo name is the same as the mount name, e.g in the
		// pachctl mount (with no -r args) case where they all get mounted based
		// on their name
		repoName = name
	} else {
		repoName = opts.Repo
	}
	mfc, err := mm.Client.NewModifyFileClient(client.NewCommit(repoName, mm.root.branch(name), ""))
	if err != nil {
		return nil, err
	}
	mm.mfcs[name] = mfc
	return mfc, nil
}

func (mm *MountManager) uploadFiles(prefixFilter string) error {
	fmt.Println("Uploading files to Pachyderm...")
	// Rendering progress bars for thousands of files significantly slows down
	// throughput. Disabling progress bars takes throughput from 1MB/sec to
	// 200MB/sec on my system, when uploading 18K small files.
	progress.Disable()
	for path, state := range mm.root.files {
		fmt.Printf("Considering file %s\n", path)
		if !strings.HasPrefix(path, prefixFilter) {
			fmt.Printf("skip hasprefix\n")
			continue
		}
		if state != dirty {
			fmt.Printf("skip not dirty\n")
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
