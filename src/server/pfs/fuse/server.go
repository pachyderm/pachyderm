package fuse

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	pathpkg "path"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/hanwen/go-fuse/v2/fs"
	gofuse "github.com/hanwen/go-fuse/v2/fuse"
	"github.com/sirupsen/logrus"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/config"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/progress"
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


Config()
--------
Returns the current cluster status and endpoint.


Config(pachd_address)
---------------------
Update the Go client and config to point at the new pachd_address endpoint
if it's valid. Returns the cluster status and endpoint of the cluster at
"pachd_address".


AuthLogin()
-----------
Login to auth service if auth is activated.


AuthLogout()
------------
Logout of auth service.


Plan
====

1. Implement List(), Get(), MountBranch(), UnmountBranch(), UnmountByName()
   Commits are implicit when filesystems are unmounted (matches current fuse mount behavior)

2. Implement CommitBranch() and stop implicitly committing data on unmounts

3. Implement MountCommit() to support looking at a specific previous state, read-only

*/

type ServerOptions struct {
	MountDir string
	// Unmount is a channel that will be closed when the filesystem has been
	// unmounted. It can be nil in which case it's ignored.
	Unmount chan struct{}
}

type ConfigRequest struct {
	PachdAddress string `json:"pachd_address"`
	ServerCas    string `json:"server_cas"`
}

type Request struct {
	Action string // default empty, set to "commit" if we want to commit (verb) a mounted branch
	Repo   string
	Branch string
	Commit string // "" for no commit (commit as noun)
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
	States map[string]*MountStateMachine
	// map from mount name onto mfc for that mount
	mfcs     map[string]*client.ModifyFileClient
	root     *loopbackRoot
	opts     *Options
	tmpDir   string
	target   string
	mu       sync.Mutex
	configMu sync.RWMutex
	Cleanup  chan struct{}
}

func (mm *MountManager) ListByRepos() (ListRepoResponse, error) {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	// fetch list of available repos & branches from pachyderm, and overlay that
	// with their mount states
	lr := map[string]RepoResponse{}
	repos, err := mm.Client.ListRepo()
	if err != nil {
		return lr, err
	}
	for _, repo := range repos {
		rr := RepoResponse{Name: repo.Repo.Name, Branches: map[string]BranchResponse{}}
		readAccess := true
		if repo.AuthInfo != nil {
			readAccess = hasRepoRead(repo.AuthInfo.Permissions)
			rr.Authorization = "none"
		}
		if readAccess {
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
				// Add all mounts associated with a repo/branch
				for _, msm := range mm.States {
					if msm.MountKey == k && msm.State != "unmounted" {
						err := msm.RefreshMountState()
						if err != nil {
							return lr, err
						}
						br.Mount = append(br.Mount, msm.MountState)
					}
				}
				if br.Mount == nil {
					br.Mount = append(br.Mount, MountState{State: "unmounted"})
				}
				rr.Branches[branch.Branch.Name] = br
			}
			if repo.AuthInfo == nil {
				rr.Authorization = "off"
			} else if hasRepoWrite(repo.AuthInfo.Permissions) {
				rr.Authorization = "write"
			} else {
				rr.Authorization = "read"
			}
		}
		lr[repo.Repo.Name] = rr
	}

	// TODO: also add any repos/branches that have been deleted from pachyderm
	// but are still mounted here :-O we should send them a special signal to
	// tell them they're "stranded" or "missing" probably?

	return lr, nil
}

func (mm *MountManager) ListByMounts() (ListMountResponse, error) {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	mr := ListMountResponse{
		Mounted:   map[string]MountState{},
		Unmounted: []MountState{},
	}
	seen := map[MountKey]bool{}
	for name, msm := range mm.States {
		if msm.State == "mounted" || msm.State == "unmounting" {
			err := msm.RefreshMountState()
			if err != nil {
				return mr, err
			}
			mr.Mounted[name] = msm.MountState
			seen[msm.MountKey] = true
		}
	}
	// Iterate through repos/branches because there will be some not present in
	// state machine
	repos, err := mm.Client.ListRepo()
	if err != nil {
		return mr, err
	}
	for _, repo := range repos {
		bs, err := mm.Client.ListBranch(repo.Repo.Name)
		if err != nil {
			return mr, err
		}
		for _, branch := range bs {
			mk := MountKey{Repo: repo.Repo.Name, Branch: branch.Branch.Name}
			// TODO: Might need to modify this when we start using Commit and
			// and compare only repo and branch names
			if _, ok := seen[mk]; !ok {
				mr.Unmounted = append(mr.Unmounted, MountState{MountKey: mk, State: "unmounted"})
			}
		}
	}

	return mr, nil
}

func NewMountStateMachine(name string, mm *MountManager) *MountStateMachine {
	return &MountStateMachine{
		MountState: MountState{
			Name: name,
		},
		manager:   mm,
		mu:        sync.Mutex{},
		requests:  make(chan Request),
		responses: make(chan Response),
	}
}

func (mm *MountManager) MaybeStartFsm(name string) {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	_, ok := mm.States[name]
	if !ok {
		// set minimal state and kick it off!
		m := NewMountStateMachine(name, mm)
		mm.States[name] = m
		go func() {
			m.Run()
			// when execution completes, remove ourselves from the map so that
			// we can infer from the key in the map that the state machine is
			// running
			mm.mu.Lock()
			defer mm.mu.Unlock()
			delete(mm.States, name)
		}()
	} // else: fsm was already running
}

func (mm *MountManager) MountBranch(key MountKey, name, mode string) (Response, error) {
	// name: an optional name for the mount, corresponds to where on the
	// filesystem the repo actually gets mounted. e.g. /pfs/{name}
	mm.MaybeStartFsm(name)
	mm.States[name].requests <- Request{
		Action: "mount",
		Repo:   key.Repo,
		Branch: key.Branch,
		Commit: key.Commit,
		Name:   name,
		Mode:   mode,
	}
	response := <-mm.States[name].responses
	return response, response.Error
}

func (mm *MountManager) UnmountBranch(key MountKey, name string) (Response, error) {
	mm.MaybeStartFsm(name)
	mm.States[name].requests <- Request{
		Action: "unmount",
		Repo:   key.Repo,
		Branch: key.Branch,
		Commit: key.Commit,
		Name:   name,
		Mode:   "",
	}
	response := <-mm.States[name].responses
	return response, response.Error
}

func (mm *MountManager) CommitBranch(key MountKey, name string) (Response, error) {
	mm.MaybeStartFsm(name)
	mm.States[name].requests <- Request{
		Action: "commit",
		Repo:   key.Repo,
		Branch: key.Branch,
		Commit: key.Commit,
		Name:   name,
		Mode:   "",
	}
	response := <-mm.States[name].responses
	return response, response.Error
}

func (mm *MountManager) UnmountAll() error {
	for name, msm := range mm.States {
		if msm.State == "mounted" {
			//TODO: Add Commit field here once we support mounting specific commits
			_, err := mm.UnmountBranch(MountKey{Repo: msm.MountKey.Repo, Branch: msm.MountKey.Branch}, name)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func NewMountManager(c *client.APIClient, target string, opts *Options) (ret *MountManager, retErr error) {
	if err := opts.validate(c); err != nil {
		return nil, err
	}
	rootDir, err := ioutil.TempDir("", "pfs")
	if err != nil {
		return nil, errors.WithStack(err)
	}
	logrus.Infof("Creating %s", rootDir)
	if err := os.MkdirAll(rootDir, 0777); err != nil {
		return nil, errors.WithStack(err)
	}
	root, err := newLoopbackRoot(rootDir, target, c, opts)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	logrus.Infof("Loopback root at %s", rootDir)
	return &MountManager{
		Client:   c,
		States:   map[string]*MountStateMachine{},
		mfcs:     map[string]*client.ModifyFileClient{},
		root:     root,
		opts:     opts,
		target:   target,
		tmpDir:   rootDir,
		mu:       sync.Mutex{},
		configMu: sync.RWMutex{},
		Cleanup:  make(chan struct{}),
	}, nil
}

func CreateMount(c *client.APIClient, mountDir string) (*MountManager, error) {
	mountOpts := &Options{
		Write: true,
		Fuse: &fs.Options{
			MountOptions: gofuse.MountOptions{
				Debug:      false,
				FsName:     "pfs",
				Name:       "pfs",
				AllowOther: true,
			},
		},
		RepoOptions: make(map[string]*RepoOptions),
		// thread this through for the tests
		Unmount: make(chan struct{}),
	}
	mm, err := NewMountManager(c, mountDir, mountOpts)
	if err != nil {
		return nil, err
	}
	go mm.Start()
	return mm, nil
}

func (mm *MountManager) Start() {
	err := mm.Run()
	if err != nil {
		logrus.Infof("Error running mount manager: %s", err)
		os.Exit(1)
	}
}

func (mm *MountManager) Run() error {
	fuse := mm.opts.getFuse()
	server, err := fs.Mount(mm.target, mm.root, fuse)
	if err != nil {
		return errors.WithStack(err)
	}
	go func() {
		<-mm.opts.getUnmount()
		server.Unmount()
	}()
	server.Wait()
	defer mm.FinishAll()
	err = mm.uploadFiles("")
	if err != nil {
		return err
	}
	err = os.RemoveAll(mm.tmpDir)
	if err != nil {
		return nil
	}
	close(mm.Cleanup)
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

func Server(sopts *ServerOptions, existingClient *client.APIClient) error {
	logrus.Infof("Dynamically mounting pfs to %s", sopts.MountDir)

	// This variable points to the MountManager for each connected cluster.
	// Updated when the config is updated.
	var mm *MountManager = &MountManager{}
	if existingClient != nil {
		var err error
		mm, err = CreateMount(existingClient, sopts.MountDir)
		if err != nil {
			return err
		}
		logrus.Infof("Connected to %s", existingClient.GetAddress().Qualified())
	}
	router := mux.NewRouter()
	router.Methods("GET").Path("/repos").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		errMsg, webCode := initialChecks(mm, true)
		if errMsg != "" {
			http.Error(w, errMsg, webCode)
			return
		}

		l, err := mm.ListByRepos()
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
	router.Methods("GET").Path("/mounts").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		errMsg, webCode := initialChecks(mm, true)
		if errMsg != "" {
			http.Error(w, errMsg, webCode)
			return
		}

		l, err := mm.ListByMounts()
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
			errMsg, webCode := initialChecks(mm, true)
			if errMsg != "" {
				http.Error(w, errMsg, webCode)
				return
			}

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
			l, err := mm.ListByRepos()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if _, ok := l[key.Repo]; !ok {
				http.Error(w, "repo does not exist", http.StatusBadRequest)
				return
			}

			_, err = mm.MountBranch(key, name, mode)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			l, err = mm.ListByRepos()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			marshalled, err := jsonMarshal(l)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Write(marshalled)
		})
	router.Methods("PUT").
		Queries("name", "{name}").
		Path("/repos/{key:.+}/_unmount").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		errMsg, webCode := initialChecks(mm, true)
		if errMsg != "" {
			http.Error(w, errMsg, webCode)
			return
		}

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
		l, err := mm.ListByRepos()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		marshalled, err := jsonMarshal(l)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(marshalled)
	})
	router.Methods("PUT").
		Queries("name", "{name}").
		Path("/repos/{key:.+}/_commit").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		errMsg, webCode := initialChecks(mm, true)
		if errMsg != "" {
			http.Error(w, errMsg, webCode)
			return
		}

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
		_, err = mm.CommitBranch(key, name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		l, err := mm.ListByRepos()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		marshalled, err := jsonMarshal(l)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(marshalled)
	})
	router.Methods("PUT").Path("/repos/_unmount").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		errMsg, webCode := initialChecks(mm, true)
		if errMsg != "" {
			http.Error(w, errMsg, webCode)
			return
		}

		err := mm.UnmountAll()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		l, err := mm.ListByRepos()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		marshalled, err := jsonMarshal(l)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(marshalled)
	})
	router.Methods("GET").Path("/config").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		errMsg, webCode := initialChecks(mm, false)
		if errMsg != "" {
			http.Error(w, errMsg, webCode)
			return
		}

		mm.configMu.RLock()
		defer mm.configMu.RUnlock()

		r, err := getClusterStatus(mm.Client)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		marshalled, err := jsonMarshal(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(marshalled)
	})
	router.Methods("PUT").Path("/config").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		mm.configMu.Lock()
		defer mm.configMu.Unlock()

		var cfgReq ConfigRequest
		err := json.NewDecoder(req.Body).Decode(&cfgReq)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		pachdAddress, err := grpcutil.ParsePachdAddress(cfgReq.PachdAddress)
		if err != nil {
			http.Error(w, "either empty or poorly formatted cluster endpoint", http.StatusBadRequest)
			return
		}
		if isNewCluster(mm, pachdAddress) {
			newClient, err := getNewClient(&cfgReq)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if mm.Client != nil {
				close(mm.opts.getUnmount())
				<-mm.Cleanup
				mm.Client.Close()
			}
			logrus.Infof("Updating pachd_address to %s\n", pachdAddress.Qualified())
			mm, err = CreateMount(newClient, sopts.MountDir)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		clusterStatus, err := getClusterStatus(mm.Client)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		marshalled, err := jsonMarshal(clusterStatus)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(marshalled)
	})
	router.Methods("PUT").Path("/auth/_login").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		errMsg, webCode := initialChecks(mm, false)
		if errMsg != "" {
			http.Error(w, errMsg, webCode)
			return
		}

		authActive, _ := mm.Client.IsAuthActive()
		if !authActive {
			http.Error(w, "auth isn't activated on the cluster", http.StatusInternalServerError)
			return
		}

		loginInfo, err := mm.Client.GetOIDCLogin(mm.Client.Ctx(), &auth.GetOIDCLoginRequest{})
		if err != nil {
			http.Error(w, "no authentication providers configured", http.StatusInternalServerError)
			return
		}
		authUrl := loginInfo.LoginURL
		state := loginInfo.State

		r := map[string]string{"auth_url": authUrl}
		marshalled, err := jsonMarshal(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(marshalled)

		go func() {
			resp, err := mm.Client.Authenticate(mm.Client.Ctx(), &auth.AuthenticateRequest{OIDCState: state})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			config.WritePachTokenToConfig(resp.PachToken, false)
			mm.Client.SetAuthToken(resp.PachToken)
		}()
	})
	router.Methods("PUT").Path("/auth/_logout").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		errMsg, webCode := initialChecks(mm, true)
		if errMsg != "" {
			http.Error(w, errMsg, webCode)
			return
		}

		cfg, err := config.Read(false, false)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, context, err := cfg.ActiveContext(true)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		err = mm.UnmountAll()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		context.SessionToken = ""
		cfg.Write()
		mm.Client.SetAuthToken("")
	})
	router.Methods("GET").Path("/health").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		r := map[string]string{"status": "running"}
		marshalled, err := jsonMarshal(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(marshalled)
	})

	// TODO: switch http server for gRPC server and bind to a unix socket not a
	// TCP port (just for convenient manual testing with curl for now...)
	// TODO: make port and bind ip parameterizable
	srv := &http.Server{Addr: ":9002", Handler: router}

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt)
		select {
		case <-sigChan:
		case <-sopts.Unmount:
		}
		if mm.Client != nil {
			close(mm.opts.getUnmount())
			<-mm.Cleanup
		}
		srv.Shutdown(context.Background())
	}()

	return errors.EnsureStack(srv.ListenAndServe())
}

func initialChecks(mm *MountManager, authCheck bool) (string, int) {
	if mm.Client == nil {
		return "not connected to a cluster", http.StatusNotFound
	}
	if authCheck && isAuthOnAndUserUnauthenticated(mm.Client) {
		return "user unauthenticated", http.StatusUnauthorized
	}
	return "", 0
}

func isAuthOnAndUserUnauthenticated(c *client.APIClient) bool {
	active, _ := c.IsAuthActive()
	if !active {
		return false
	}
	_, err := c.WhoAmI(c.Ctx(), &auth.WhoAmIRequest{})
	return err != nil
}

func getClusterStatus(c *client.APIClient) (map[string]string, error) {
	clusterStatus := "INVALID"
	err := c.Health()
	if err == nil {
		authActive, err := c.IsAuthActive()
		if err != nil {
			return nil, err
		}

		if authActive {
			clusterStatus = "AUTH_ENABLED"
		} else {
			clusterStatus = "AUTH_DISABLED"
		}
	}

	pachdAddress := c.GetAddress().Qualified()

	return map[string]string{"cluster_status": clusterStatus, "pachd_address": pachdAddress}, nil
}

func isNewCluster(mm *MountManager, pachdAddress *grpcutil.PachdAddress) bool {
	if mm.Client == nil {
		return true
	}
	if reflect.DeepEqual(pachdAddress, mm.Client.GetAddress()) {
		logrus.Infof("New endpoint is same as current endpoint: %s, no change\n", pachdAddress.Qualified())
		return false
	}
	return true
}

func getNewClient(cfgReq *ConfigRequest) (*client.APIClient, error) {
	pachdAddress, _ := grpcutil.ParsePachdAddress(cfgReq.PachdAddress)
	// Check if new cluster endpoint is valid
	var options []client.Option
	if cfgReq.ServerCas != "" {
		pemBytes, err := base64.StdEncoding.DecodeString(cfgReq.ServerCas)
		if err != nil {
			return nil, errors.Wrap(err, "could not decode server CA certs")
		}
		options = append(options, client.WithAdditionalRootCAs(pemBytes))
	}
	options = append(options, client.WithDialTimeout(5*time.Second))
	newClient, err := client.NewFromPachdAddress(pachdAddress, options...)
	if err != nil {
		return nil, errors.Wrapf(err, "could not connect to %s", pachdAddress.Qualified())
	}
	// Update config file and cachedConfig
	err = updateConfig(cfgReq)
	if err != nil {
		return nil, errors.Wrap(err, "issue updating config")
	}

	return newClient, nil
}

func updateConfig(cfgReq *ConfigRequest) error {
	cfg, err := config.Read(false, false)
	if err != nil {
		return err
	}
	pachdAddress, _ := grpcutil.ParsePachdAddress(cfgReq.PachdAddress)
	cfg.V2.ActiveContext = "mount-server"
	cfg.V2.Contexts[cfg.V2.ActiveContext] = &config.Context{PachdAddress: pachdAddress.Qualified(), ServerCAs: cfgReq.ServerCas}
	err = cfg.Write()
	if err != nil {
		return err
	}

	return nil
}

func jsonMarshal(t interface{}) ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(t)
	return buffer.Bytes(), err
}

func hasRepoRead(permissions []auth.Permission) bool {
	for _, p := range permissions {
		if p == auth.Permission_REPO_READ {
			return true
		}
	}
	return false
}

func hasRepoWrite(permissions []auth.Permission) bool {
	for _, p := range permissions {
		if p == auth.Permission_REPO_WRITE {
			return true
		}
	}
	return false
}

type MountState struct {
	Name       string   `json:"name"`       // where to mount it. written by client
	MountKey   MountKey `json:"mount_key"`  // what to mount. written by client
	Mode       string   `json:"mode"`       // "ro", "rw", or "" if unknown/unspecified. written by client
	State      string   `json:"state"`      // "unmounted", "mounting", "mounted", "pushing", "unmounted", "error". written by fsm
	Status     string   `json:"status"`     // human readable string with additional info wrt State, e.g. an error message for the error state. written by fsm
	Mountpoint string   `json:"mountpoint"` // where on the filesystem it's mounted. written by fsm. can also be derived from {MountDir}/{Name}
	// the following are used by the "refresh" button feature in the jupyter plugin
	ActualMountedCommit  string `json:"actual_mounted_commit"`   // the actual commit that was mounted at mount time. written by fsm
	LatestCommit         string `json:"latest_commit"`           // the latest available commit on the branch, last time RefreshMountState() was called. written by fsm
	HowManyCommitsBehind int    `json:"how_many_commits_behind"` // how many commits are behind the latest commit on the branch. written by fsm
}

type MountStateMachine struct {
	MountState
	manager   *MountManager
	mu        sync.Mutex
	requests  chan Request
	responses chan Response
}

func (m *MountStateMachine) RefreshMountState() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	logrus.Infof("starting RefreshMountState")

	if m.State != "mounted" {
		return nil
	}

	// get commit from loopbackRoot
	commit, ok := m.manager.root.commits[m.Name]
	if !ok {
		// Don't have anything to calculate HowManyCommitsBehind from
		m.Status = "unable to load current commit"
		return nil
	}

	m.ActualMountedCommit = commit

	// Get the latest commit on the branch
	branchInfo, err := m.manager.Client.InspectBranch(m.MountKey.Repo, m.MountKey.Branch)
	if err != nil {
		return err
	}

	// set the latest commit on the branch in our LatestCommit
	m.LatestCommit = branchInfo.Head.ID

	// calculate how many commits behind LatestCommit ActualMountedCommit is
	commitInfos, err := m.manager.Client.ListCommit(branchInfo.Branch.Repo, branchInfo.Head, nil, 0)
	if err != nil {
		return err
	}
	// reverse slice
	for i, j := 0, len(commitInfos)-1; i < j; i, j = i+1, j-1 {
		commitInfos[i], commitInfos[j] = commitInfos[j], commitInfos[i]
	}

	// iterate over commits in branch, counting how many are behind LatestCommit
	indexOfCurrentCommit := -1
	for i, commitInfo := range commitInfos {
		logrus.Infof("%d: commitInfo.Commit.ID: %s, m.ActualMountedCommit: %s", i, commitInfo.Commit.ID, m.ActualMountedCommit)
		if commitInfo.Commit.ID == m.ActualMountedCommit {
			indexOfCurrentCommit = i
			break
		}
	}
	indexOfLatestCommit := len(commitInfos) - 1

	/*

		Example:

		0
		1
		2
		3 <- current
		4
		5 <- latest

		indexOfCurrentCommit = 3
		indexOfLatestCommit = 5
		indexOfLatestCommit - indexOfCurrentCommit = 2

	*/

	m.HowManyCommitsBehind = indexOfLatestCommit - indexOfCurrentCommit

	first8chars := func(s string) string {
		if len(s) < 8 {
			return s
		}
		return s[:8]
	}
	if m.HowManyCommitsBehind > 0 {
		m.Status = fmt.Sprintf(
			"%d commits behind latest; current = %s (%d'th), latest = %s (%d'th)",
			m.HowManyCommitsBehind,
			first8chars(m.ActualMountedCommit),
			indexOfCurrentCommit,
			first8chars(m.LatestCommit),
			indexOfLatestCommit,
		)
	} else {
		m.Status = fmt.Sprintf(
			"up to date on latest commit %s",
			first8chars(m.LatestCommit),
		)
	}

	return nil
}

// TODO: switch to pach internal types if appropriate?
type ListRepoResponse map[string]RepoResponse

type BranchResponse struct {
	Name  string       `json:"name"`
	Mount []MountState `json:"mount"`
}

type RepoResponse struct {
	Name          string                    `json:"name"`
	Branches      map[string]BranchResponse `json:"branches"`
	Authorization string                    `json:"authorization"` // "off", "none", "read", "write"
	// TODO: Commits map[string]CommitResponse
}

type ListMountResponse struct {
	Mounted   map[string]MountState `json:"mounted"`
	Unmounted []MountState          `json:"unmounted"`
}

type GetResponse RepoResponse

type MountKey struct {
	Repo   string `json:"repo"`
	Branch string `json:"branch"`
	Commit string `json:"commit"`
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
		return MountKey{}, errors.WithStack(fmt.Errorf(
			"wrong number of '/'s in key %+v, got %d expected 2 or 3", key, len(shrapnel),
		))
	}
}

// state machines in the style of Rob Pike's Lexical Scanning
// http://www.timschmelmer.com/2013/08/state-machines-with-state-functions-in.html

// setup

type StateFn func(*MountStateMachine) StateFn

func (m *MountStateMachine) transitionedTo(state, status string) {

	// before we lock, as this fn takes the same lock.
	m.manager.root.setState(m.Name, state)

	m.mu.Lock()
	defer m.mu.Unlock()
	logrus.Infof("[%s] (%s) %s -> %s", m.Name, m.MountKey, m.State, state)
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
		switch req.Action {
		case "mount":
			// check user permissions on repo
			repoInfo, err := m.manager.Client.InspectRepo(req.Repo)
			if err != nil {
				m.responses <- Response{
					Repo:       req.Repo,
					Branch:     req.Branch,
					Commit:     req.Commit,
					Name:       req.Name,
					MountState: m.MountState,
					Error:      err,
				}
				return unmountedState
			}
			if repoInfo.AuthInfo != nil {
				if !hasRepoRead(repoInfo.AuthInfo.Permissions) {
					m.responses <- Response{
						Repo:       req.Repo,
						Branch:     req.Branch,
						Commit:     req.Commit,
						Name:       req.Name,
						MountState: m.MountState,
						Error:      fmt.Errorf("user doesn't have read permission on repo %s", req.Repo),
					}
					return unmountedState
				}

				if req.Mode == "rw" && !hasRepoWrite(repoInfo.AuthInfo.Permissions) {
					m.responses <- Response{
						Repo:       req.Repo,
						Branch:     req.Branch,
						Commit:     req.Commit,
						Name:       req.Name,
						MountState: m.MountState,
						Error:      fmt.Errorf("can't create writable mount since user doesn't have write access on repo %s", req.Repo),
					}
					return unmountedState
				}
			}

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
		case "commit":
			m.responses <- Response{
				Repo:       m.MountKey.Repo,
				Branch:     m.MountKey.Branch,
				Commit:     m.MountKey.Commit,
				Name:       m.Name,
				MountState: m.MountState,
				Error:      fmt.Errorf("bad request: can't commit when unmounted"),
			}
			return unmountedState
		case "unmount":
			m.responses <- Response{
				Repo:       m.MountKey.Repo,
				Branch:     m.MountKey.Branch,
				Commit:     m.MountKey.Commit,
				Name:       m.Name,
				MountState: m.MountState,
				Error:      fmt.Errorf("can't unmount when we're already unmounted"),
			}
			// stay unmounted
			return unmountedState
		default:
			m.responses <- Response{
				Repo:       m.MountKey.Repo,
				Branch:     m.MountKey.Branch,
				Commit:     m.MountKey.Commit,
				Name:       m.Name,
				MountState: m.MountState,
				Error:      fmt.Errorf("bad request: unsupported/unknown action %s while unmounted", req.Action),
			}
			return unmountedState
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
			Name:  m.Name,
			File:  client.NewFile(m.MountKey.Repo, m.MountKey.Branch, "", ""),
			Write: m.Mode == "rw",
		}
		m.manager.root.branches[m.Name] = m.MountKey.Branch
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
		switch req.Action {
		case "mount":
			// This check is necessary to make sure that if the incoming request
			// is trying to mount a repo at mount_name and mount_name is already
			// being used, the request repo and branch match the repo and branch
			// already associated with mount_name. It's essentially a safety
			// check for when we get to the remounting case below.
			if req.Repo != m.MountKey.Repo || req.Branch != m.MountKey.Branch {
				m.responses <- Response{
					Repo:       m.MountKey.Repo,
					Branch:     m.MountKey.Branch,
					Commit:     m.MountKey.Commit,
					Name:       m.Name,
					MountState: m.MountState,
					Error:      fmt.Errorf("mount at '%s' already in use", m.Name),
				}
			} else {
				m.responses <- Response{}
			}
			// TODO: handle remount case (switching mode):
			// if mounted ro and switching to rw, just upgrade
			// if mounted rw and switching to ro, upload changes, then switch the mount type
		case "unmount":
			return unmountingState
		case "commit":
			return committingState
		default:
			m.responses <- Response{
				Repo:       m.MountKey.Repo,
				Branch:     m.MountKey.Branch,
				Commit:     m.MountKey.Commit,
				Name:       m.Name,
				MountState: m.MountState,
				Error:      fmt.Errorf("bad request: unsupported/unknown action %s while mounted", req.Action),
			}
			return unmountedState
		}
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

func (m *MountStateMachine) maybeUploadFiles() error {
	// TODO XXX VERY IMPORTANT: pause/block filesystem operations during the
	// upload, otherwise we could get filesystem inconsistency! Need a sort of
	// lock which multiple fs operations can hold but only one "pauser" can.

	// Only upload files for writeable filesystems
	if m.Mode == "rw" {
		// upload any files whose paths start with where we're mounted
		err := m.manager.uploadFiles(m.Name)
		if err != nil {
			return err
		}
		// close the mfc, uploading files, then delete it
		mfc, err := m.manager.mfc(m.Name)
		if err != nil {
			return err
		}
		err = mfc.Close()
		if err != nil {
			return err
		}
		// cleanup mfc - a new one will be created on-demand
		func() {
			m.manager.mu.Lock()
			defer m.manager.mu.Unlock()
			delete(m.manager.mfcs, m.Name)
		}()
	}
	return nil
}

func committingState(m *MountStateMachine) StateFn {
	// NB: this function is responsible for placing a response on m.responses
	// _in all cases_
	m.transitionedTo("committing", "")

	err := m.maybeUploadFiles()
	if err != nil {
		logrus.Infof("Error while uploading! %s", err)
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

	// TODO: fill in more fields?
	m.responses <- Response{}
	// we always go back from committingState back into mountedState
	return mountedState
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

	err := m.maybeUploadFiles()
	if err != nil {
		logrus.Infof("Error while uploading! %s", err)
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
	logrus.Infof("Path is %s", cleanPath)

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
		logrus.Infof("Error while cleaning! %s", err)
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
		repoName = opts.File.Commit.Branch.Repo.Name
	}
	mfc, err := mm.Client.NewModifyFileClient(client.NewCommit(repoName, mm.root.branch(name), ""))
	if err != nil {
		return nil, err
	}
	mm.mfcs[name] = mfc
	return mfc, nil
}

func (mm *MountManager) uploadFiles(prefixFilter string) error {
	logrus.Info("Uploading files to Pachyderm...")
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
	logrus.Info("Done!")
	return nil
}
