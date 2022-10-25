package fuse

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	pathpkg "path"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/hanwen/go-fuse/v2/fs"
	gofuse "github.com/hanwen/go-fuse/v2/fuse"
	"github.com/sirupsen/logrus"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/clientsdk"
	"github.com/pachyderm/pachyderm/v2/src/internal/config"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/progress"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
)

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

type MountRequest struct {
	Mounts []*MountInfo `json:"mounts"`
}

type UnmountRequest struct {
	Mounts []string `json:"mounts"`
}

type CommitRequest struct {
	Mount string `json:"mount"`
}

type MountDatumResponse struct {
	Id        string `json:"id"`
	Idx       int    `json:"idx"`
	NumDatums int    `json:"num_datums"`
}

type DatumsResponse struct {
	NumDatums int        `json:"num_datums"`
	Input     *pps.Input `json:"input"`
	CurrIdx   int        `json:"curr_idx"`
}

type MountInfo struct {
	Name   string   `json:"name"`
	Repo   string   `json:"repo"`
	Branch string   `json:"branch"`
	Commit string   `json:"commit"` // "" for no commit (commit as noun)
	Files  []string `json:"files"`
	Glob   string   `json:"glob"`
	Mode   string   `json:"mode"` // "ro", "rw"
}

type Request struct {
	*MountInfo
	Action string // default empty, set to "commit" if we want to commit (verb) a mounted branch
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

	Datums       []*pps.DatumInfo
	DatumInput   *pps.Input
	CurrDatumIdx int

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
		rr := RepoResponse{Repo: repo.Repo.Name}
		readAccess := true
		if repo.AuthInfo != nil {
			readAccess = hasRepoRead(repo.AuthInfo.Permissions)
			rr.Authorization = "none"
		}
		if readAccess {
			bis, err := mm.Client.ListBranch(repo.Repo.Name)
			if err != nil {
				return lr, err
			}
			for _, bi := range bis {
				rr.Branches = append(rr.Branches, bi.Branch.Name)
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
		Unmounted: map[string]RepoResponse{},
	}
	// Keep track of existing repos and branches to see if mm.States contains deleted repos/branches
	repoBranches := map[string]map[string]bool{}
	repos, err := mm.Client.ListRepo()
	if err != nil {
		return mr, err
	}
	for _, repo := range repos {
		repoBranches[repo.Repo.Name] = map[string]bool{}
		rr := RepoResponse{Repo: repo.Repo.Name}
		readAccess := true
		if repo.AuthInfo != nil {
			readAccess = hasRepoRead(repo.AuthInfo.Permissions)
			rr.Authorization = "none"
		}
		if readAccess {
			bis, err := mm.Client.ListBranch(repo.Repo.Name)
			if err != nil {
				// Repo was deleted between ListRepo and ListBranch RPCs
				if auth.IsErrNoRoleBinding(err) {
					continue
				}
				return mr, err
			}
			for _, bi := range bis {
				repoBranches[repo.Repo.Name][bi.Branch.Name] = true
				rr.Branches = append(rr.Branches, bi.Branch.Name)
			}
			if repo.AuthInfo == nil {
				rr.Authorization = "off"
			} else if hasRepoWrite(repo.AuthInfo.Permissions) {
				rr.Authorization = "write"
			} else {
				rr.Authorization = "read"
			}
		}
		mr.Unmounted[repo.Repo.Name] = rr
	}

	// Unmount mounted repos/branches that were deleted
	for name, msm := range mm.States {
		if msm.State == "mounted" {
			_, ok := repoBranches[msm.Repo]
			if !ok {
				mm.unmountDeletedRepos(name)
				continue
			}
			_, ok = repoBranches[msm.Repo][msm.Branch]
			if !ok && msm.Mode == "ro" {
				mm.unmountDeletedRepos(name)
				continue
			}
		}
	}

	// Get data on mounted commits
	for name, msm := range mm.States {
		if msm.State == "mounted" {
			if err := msm.RefreshMountState(); err != nil {
				return mr, err
			}
			ms := msm.MountState
			ms.Files = nil
			mr.Mounted[name] = ms
		}
	}

	return mr, nil
}

func (mm *MountManager) unmountDeletedRepos(name string) {
	mm.States[name].requests <- Request{
		Action: "unmount",
	}
	<-mm.States[name].responses
}

func NewMountStateMachine(name string, mm *MountManager) *MountStateMachine {
	return &MountStateMachine{
		MountState: MountState{
			MountInfo: MountInfo{
				Name: name,
			},
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

func (mm *MountManager) MountRepo(mi *MountInfo) (Response, error) {
	mm.MaybeStartFsm(mi.Name)
	mm.States[mi.Name].requests <- Request{
		Action:    "mount",
		MountInfo: mi,
	}
	response := <-mm.States[mi.Name].responses
	return response, response.Error
}

func (mm *MountManager) UnmountRepo(name string) (Response, error) {
	mm.MaybeStartFsm(name)
	mm.States[name].requests <- Request{
		Action: "unmount",
	}
	response := <-mm.States[name].responses
	return response, response.Error
}

func (mm *MountManager) CommitRepo(name string) (Response, error) {
	mm.MaybeStartFsm(name)
	mm.States[name].requests <- Request{
		Action: "commit",
	}
	response := <-mm.States[name].responses
	return response, response.Error
}

func (mm *MountManager) UnmountAll() error {
	for name, msm := range mm.States {
		if msm.State == "mounted" {
			//TODO: Add Commit field here once we support mounting specific commits
			if _, err := mm.UnmountRepo(name); err != nil {
				return err
			}
		}
	}
	delete(mm.root.repoOpts, "out")

	return nil
}

func NewMountManager(c *client.APIClient, target string, opts *Options) (ret *MountManager, retErr error) {
	if err := opts.validate(c); err != nil {
		return nil, err
	}
	rootDir, err := os.MkdirTemp("", "pfs")
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
	if err := mm.Run(); err != nil {
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
		server.Unmount() //nolint:errcheck
	}()
	server.Wait()

	defer mm.FinishAll() //nolint:errcheck
	if err := mm.uploadFiles(""); err != nil {
		return err
	}
	if err := os.RemoveAll(mm.tmpDir); err != nil {
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
		w.Write(marshalled) //nolint:errcheck
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
		w.Write(marshalled) //nolint:errcheck
	})
	router.Methods("PUT").Path("/_mount").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		errMsg, webCode := initialChecks(mm, true)
		if errMsg != "" {
			http.Error(w, errMsg, webCode)
			return
		}
		if len(mm.Datums) > 0 {
			http.Error(w, "can't mount repos while in mounted datum mode", http.StatusBadRequest)
			return
		}

		// Verify request payload
		var mreq MountRequest
		defer req.Body.Close()
		if err := json.NewDecoder(req.Body).Decode(&mreq); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		lr, err := mm.ListByRepos()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := verifyMountRequest(mreq.Mounts, lr); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		for _, mi := range mreq.Mounts {
			if _, err := mm.MountRepo(mi); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		lm, err := mm.ListByMounts()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		marshalled, err := jsonMarshal(lm)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(marshalled) //nolint:errcheck
	})
	router.Methods("PUT").Path("/_unmount").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		errMsg, webCode := initialChecks(mm, true)
		if errMsg != "" {
			http.Error(w, errMsg, webCode)
			return
		}
		if len(mm.Datums) > 0 {
			http.Error(w, "can't unmount repos while in mounted datum mode", http.StatusBadRequest)
			return
		}

		var umreq UnmountRequest
		defer req.Body.Close()
		if err := json.NewDecoder(req.Body).Decode(&umreq); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		for _, name := range umreq.Mounts {
			if _, err := mm.UnmountRepo(name); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		lm, err := mm.ListByMounts()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		marshalled, err := jsonMarshal(lm)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(marshalled) //nolint:errcheck
	})
	router.Methods("PUT").Path("/_commit").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		errMsg, webCode := initialChecks(mm, true)
		if errMsg != "" {
			http.Error(w, errMsg, webCode)
			return
		}

		var creq CommitRequest
		defer req.Body.Close()
		if err := json.NewDecoder(req.Body).Decode(&creq); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if _, err := mm.CommitRepo(creq.Mount); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		lm, err := mm.ListByMounts()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		marshalled, err := jsonMarshal(lm)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(marshalled) //nolint:errcheck
	})
	router.Methods("PUT").Path("/_unmount_all").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		errMsg, webCode := initialChecks(mm, true)
		if errMsg != "" {
			http.Error(w, errMsg, webCode)
			return
		}

		if err := mm.UnmountAll(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := removeOutDir(mm); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		lm, err := mm.ListByMounts()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		marshalled, err := jsonMarshal(lm)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(marshalled) //nolint:errcheck

		mm.Datums = []*pps.DatumInfo{}
		mm.DatumInput = &pps.Input{}
		mm.CurrDatumIdx = -1
	})
	router.Methods("PUT").Path("/_mount_datums").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		errMsg, webCode := initialChecks(mm, true)
		if errMsg != "" {
			http.Error(w, errMsg, webCode)
			return
		}

		defer req.Body.Close()
		pipelineBytes, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		pipelineReader, err := ppsutil.NewPipelineManifestReader(pipelineBytes)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		pipelineReq, err := pipelineReader.NextCreatePipelineRequest()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		mm.DatumInput = pipelineReq.Input
		mm.Datums, err = mm.Client.ListDatumInputAll(mm.DatumInput)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if len(mm.Datums) == 0 {
			http.Error(w, "no datums match the given input spec", http.StatusBadRequest)
			return
		}

		logrus.Infof("Mounting first datum (%s)", mm.Datums[0].Datum.ID)
		mis := datumToMounts(mm.Datums[0])
		if err := mm.UnmountAll(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		for _, mi := range mis {
			if _, err := mm.MountRepo(mi); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		createLocalOutDir(mm)
		mm.CurrDatumIdx = 0

		resp := MountDatumResponse{
			Id:        mm.Datums[0].Datum.ID,
			Idx:       0,
			NumDatums: len(mm.Datums),
		}
		marshalled, err := jsonMarshal(resp)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(marshalled) //nolint:errcheck
	})
	router.Methods("PUT").Path("/_show_datum").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		errMsg, webCode := initialChecks(mm, true)
		if errMsg != "" {
			http.Error(w, errMsg, webCode)
			return
		}
		if len(mm.Datums) == 0 {
			http.Error(w, "no datums mounted", http.StatusBadRequest)
			return
		}

		vs := req.URL.Query()
		idxStr := vs.Get("idx")
		id := vs.Get("id")
		if idxStr == "" && id == "" {
			http.Error(w, "need to specify either datum idx or id", http.StatusBadRequest)
			return
		}
		idx := 0
		if idxStr != "" {
			var err error
			idx, err = strconv.Atoi(idxStr)
			if err != nil {
				http.Error(w, "used a non-integer for datum index", http.StatusBadRequest)
				return
			}
		}

		var di *pps.DatumInfo
		if id != "" {
			foundDatum := false
			for idx, di = range mm.Datums {
				if di.Datum.ID == id {
					foundDatum = true
					break
				}
			}
			if !foundDatum {
				http.Error(w, "specify a valid datum id", http.StatusBadRequest)
				return
			}
		} else {
			di = mm.Datums[idx]
		}
		mis := datumToMounts(di)
		if err := mm.UnmountAll(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := removeOutDir(mm); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		for _, mi := range mis {
			if _, err := mm.MountRepo(mi); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		createLocalOutDir(mm)
		mm.CurrDatumIdx = idx

		resp := MountDatumResponse{
			Id:        di.Datum.ID,
			Idx:       idx,
			NumDatums: len(mm.Datums),
		}
		marshalled, err := jsonMarshal(resp)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(marshalled) //nolint:errcheck
	})
	router.Methods("GET").Path("/datums").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		errMsg, webCode := initialChecks(mm, true)
		if errMsg != "" {
			http.Error(w, errMsg, webCode)
			return
		}

		var resp DatumsResponse
		if len(mm.Datums) == 0 {
			resp = DatumsResponse{
				NumDatums: 0,
			}
		} else {
			resp = DatumsResponse{
				NumDatums: len(mm.Datums),
				Input:     mm.DatumInput,
				CurrIdx:   mm.CurrDatumIdx,
			}
		}
		marshalled, err := jsonMarshal(resp)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(marshalled) //nolint:errcheck
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
		w.Write(marshalled) //nolint:errcheck
	})
	router.Methods("PUT").Path("/config").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		mm.configMu.Lock()
		defer mm.configMu.Unlock()

		var cfgReq ConfigRequest
		defer req.Body.Close()
		if err := json.NewDecoder(req.Body).Decode(&cfgReq); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		pachdAddress, err := grpcutil.ParsePachdAddress(cfgReq.PachdAddress)
		if err != nil {
			http.Error(w, "either empty or poorly formatted cluster endpoint", http.StatusBadRequest)
			return
		}
		if isNewCluster(mm, pachdAddress) {
			newClient, err := getNewClient(cfgReq)
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
			if mm, err = CreateMount(newClient, sopts.MountDir); err != nil {
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
		w.Write(marshalled) //nolint:errcheck
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
		w.Write(marshalled) //nolint:errcheck

		go func() {
			resp, err := mm.Client.Authenticate(mm.Client.Ctx(), &auth.AuthenticateRequest{OIDCState: state})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			config.WritePachTokenToConfig(resp.PachToken, false) //nolint:errcheck
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
		if err := mm.UnmountAll(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		context.SessionToken = ""
		cfg.Write() //nolint:errcheck
		mm.Client.SetAuthToken("")
	})
	router.Methods("GET").Path("/health").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		r := map[string]string{"status": "running"}
		marshalled, err := jsonMarshal(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(marshalled) //nolint:errcheck
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
		srv.Shutdown(context.Background()) //nolint:errcheck
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
	if err := c.Health(); err == nil {
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

func getNewClient(cfgReq ConfigRequest) (*client.APIClient, error) {
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
	if err = updateConfig(cfgReq); err != nil {
		return nil, errors.Wrap(err, "issue updating config")
	}

	return newClient, nil
}

func updateConfig(cfgReq ConfigRequest) error {
	cfg, err := config.Read(false, false)
	if err != nil {
		return err
	}
	pachdAddress, _ := grpcutil.ParsePachdAddress(cfgReq.PachdAddress)
	cfg.V2.ActiveContext = "mount-server"
	cfg.V2.Contexts[cfg.V2.ActiveContext] = &config.Context{PachdAddress: pachdAddress.Qualified(), ServerCAs: cfgReq.ServerCas}
	if err = cfg.Write(); err != nil {
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

func verifyMountRequest(mis []*MountInfo, lr ListRepoResponse) error {
	for _, mi := range mis {
		if mi.Name == "" {
			return errors.Errorf("no name specified in request %+v", mi)
		}
		if mi.Repo == "" {
			return errors.Errorf("no repo specified in request %+v", mi)
		}
		if mi.Branch == "" {
			mi.Branch = "master"
		}
		if mi.Commit != "" {
			// TODO: case of same commit id on diff branches
			return errors.Errorf("don't support mounting commits yet in request %+v", mi)
		}
		if mi.Files != nil && mi.Glob != "" {
			return errors.Errorf("can't specify both files and glob pattern in request %+v", mi)
		}
		if mi.Mode == "" {
			mi.Mode = "ro"
		}
		if _, ok := lr[mi.Repo]; !ok {
			return errors.Errorf("repo does not exist")
		}
	}

	return nil
}

func datumToMounts(d *pps.DatumInfo) []*MountInfo {
	mounts := map[string]*MountInfo{}
	files := map[string]map[string]bool{}
	for _, fi := range d.Data {
		repo := fi.File.Commit.Branch.Repo.Name
		branch := fi.File.Commit.Branch.Name
		// TODO: add commit here
		mount := repo
		if branch != "master" {
			mount = mount + "_" + branch
		}
		if _, ok := files[mount]; !ok {
			files[mount] = map[string]bool{}
		}

		if mi, ok := mounts[mount]; ok {
			if _, ok := files[mount][fi.File.Path]; !ok {
				mi.Files = append(mi.Files, fi.File.Path)
				mounts[mount] = mi
			}
		} else {
			mi := &MountInfo{
				Name:   mount,
				Repo:   repo,
				Branch: branch,
				Files:  []string{fi.File.Path},
				Mode:   "ro",
			}
			mounts[mount] = mi
		}
		files[mount][fi.File.Path] = true
	}

	mis := []*MountInfo{}
	for _, mi := range mounts {
		mis = append(mis, mi)
	}
	return mis
}

func removeOutDir(mm *MountManager) error {
	cleanPath := mm.root.rootPath + "/out"
	return errors.EnsureStack(os.RemoveAll(cleanPath))
}

func createLocalOutDir(mm *MountManager) {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	mm.root.repoOpts["out"] = &RepoOptions{
		Name:  "out",
		File:  &pfs.File{Commit: &pfs.Commit{Branch: &pfs.Branch{Repo: &pfs.Repo{Name: "out"}}}},
		Write: true}
}

type MountState struct {
	MountInfo         // mount details. written by client
	State      string `json:"state"`      // "unmounted", "mounting", "mounted", "pushing", "unmounted", "error". written by fsm
	Status     string `json:"status"`     // human readable string with additional info wrt State, e.g. an error message for the error state. written by fsm
	Mountpoint string `json:"mountpoint"` // where on the filesystem it's mounted. written by fsm. can also be derived from {MountDir}/{Name}
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
	// TODO: Update when supporting projects in notebooks
	branchInfo, err := m.manager.Client.InspectBranch(m.Repo, m.Branch)
	if err != nil {
		return err
	}

	// set the latest commit on the branch in our LatestCommit
	m.LatestCommit = branchInfo.Head.ID

	// calculate how many commits behind LatestCommit ActualMountedCommit is
	listClient, err := m.manager.Client.PfsAPIClient.ListCommit(m.manager.Client.Ctx(), &pfs.ListCommitRequest{
		Repo: branchInfo.Branch.Repo,
		To:   branchInfo.Head,
		All:  true,
	})
	if err != nil {
		return grpcutil.ScrubGRPC(err)
	}
	commitInfos, err := clientsdk.ListCommit(listClient)
	if err != nil {
		return err
	}
	// reverse slice
	for i, j := 0, len(commitInfos)-1; i < j; i, j = i+1, j-1 {
		commitInfos[i], commitInfos[j] = commitInfos[j], commitInfos[i]
	}

	// iterate over commits in branch, counting how many are behind LatestCommit
	logrus.Infof("mount: %s", m.Name)
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

type RepoResponse struct {
	Repo          string   `json:"repo"`
	Branches      []string `json:"branches"`
	Authorization string   `json:"authorization"` // "off", "none", "read", "write"
	// TODO: Commits map[string]CommitResponse
}

type ListMountResponse struct {
	Mounted   map[string]MountState   `json:"mounted"`
	Unmounted map[string]RepoResponse `json:"unmounted"`
}

type GetResponse RepoResponse

// state machines in the style of Rob Pike's Lexical Scanning
// http://www.timschmelmer.com/2013/08/state-machines-with-state-functions-in.html

// setup

type StateFn func(*MountStateMachine) StateFn

func (m *MountStateMachine) transitionedTo(state, status string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	logrus.Infof("[%s] (%s, %s, %s) %s -> %s", m.Name, m.Repo, m.Branch, m.Commit, m.State, state)
	m.manager.root.setState(m.Name, state)
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
			m.Name = req.Name
			m.Repo = req.Repo
			m.Branch = req.Branch
			m.Commit = req.Commit
			m.Files = req.Files
			m.Mode = req.Mode
			return mountingState
		case "commit":
			m.responses <- Response{
				MountState: m.MountState,
				Error:      fmt.Errorf("bad request: can't commit when unmounted"),
			}
			return unmountedState
		case "unmount":
			m.responses <- Response{
				MountState: m.MountState,
				Error:      fmt.Errorf("can't unmount when we're already unmounted"),
			}
			// stay unmounted
			return unmountedState
		default:
			m.responses <- Response{
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
		m.manager.root.repoOpts[m.Name] = &RepoOptions{
			Name:     m.Name,
			File:     client.NewFile(m.Repo, m.Branch, "", ""),
			Subpaths: m.Files,
			Write:    m.Mode == "rw",
		}
		m.manager.root.branches[m.Name] = m.Branch
	}()
	// re-downloading the repos with an updated RepoOptions set will have the
	// effect of causing it to pop into existence
	err := m.manager.root.mkdirMountNames()
	m.responses <- Response{
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
			if req.Repo != m.Repo || req.Branch != m.Branch {
				m.responses <- Response{
					Repo:       req.Repo,
					Branch:     req.Branch,
					Commit:     req.Commit,
					Name:       req.Name,
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
				Repo:       req.Repo,
				Branch:     req.Branch,
				Commit:     req.Commit,
				Name:       req.Name,
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
		if err := m.manager.uploadFiles(m.Name); err != nil {
			return err
		}
		// close the mfc, uploading files, then delete it
		mfc, err := m.manager.mfc(m.Name)
		if err != nil {
			return err
		}
		if err = mfc.Close(); err != nil {
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

	if err := m.maybeUploadFiles(); err != nil {
		logrus.Infof("Error while uploading! %s", err)
		m.transitionedTo("error", err.Error())
		m.responses <- Response{
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

	// TODO XXX VERY IMPORTANT: pause/block filesystem operations during the
	// upload, otherwise we could get filesystem inconsistency! Need a sort of
	// lock which multiple fs operations can hold but only one "pauser" can.

	if err := m.maybeUploadFiles(); err != nil {
		logrus.Infof("Error while uploading! %s", err)
		m.transitionedTo("error", err.Error())
		m.responses <- Response{
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
	logrus.Infof("Removing path %s", cleanPath)

	err := os.RemoveAll(cleanPath)
	m.responses <- Response{
		MountState: m.MountState,
		Error:      err,
	}
	if err != nil {
		logrus.Infof("Error while cleaning! %s", err)
		m.transitionedTo("error", err.Error())
		return errorState
	}
	m.manager.mu.Lock()
	defer m.manager.mu.Unlock()
	delete(m.manager.States, m.Name)

	return unmountedState
}

func errorState(m *MountStateMachine) StateFn {
	for {
		// eternally error on responses
		<-m.requests
		m.responses <- Response{
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
