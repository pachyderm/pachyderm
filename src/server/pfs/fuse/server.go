package fuse

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
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
	"go.uber.org/zap"
	"golang.org/x/exp/slices"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/config"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/progress"
	"github.com/pachyderm/pachyderm/v2/src/internal/serde"
	"github.com/pachyderm/pachyderm/v2/src/internal/signals"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
)

// TODO: Get all datums for now (value of 0). Once ListDatum pagination is efficient,
// specify page size. #ListDatumPagination
const NumDatumsPerPage = 0 // The number of datums requested per ListDatum call

const FuseServerPort = ":9002"

type ServerOptions struct {
	// MountDir is the directory where the mount will be created (the most
	// conventional and typically most useful value for this field is '/pfs',
	// causing mount-server to lay out files in the same way as Pachyderm's worker
	// binary. Code that correctly reads and writes data in this configuration can
	// then be run inside a Pachyderm pipeline without modification).
	MountDir string
	// Unmount is a channel that will be closed when the filesystem has been
	// unmounted. It can be nil, in which case it's ignored.
	Unmount chan struct{}
	// AllowOther, if true, configures mount-server to give read and write
	// permissions to "other" on the mount point ('chmod u+rw /pfs', effectively)
	AllowOther bool
	// SockPath, if set, configures mount-server to serve over a Unix socket at
	// the path it contains.
	SockPath string
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
	// We'll return this as empty string until #ListDatumPagination since we don't know the
	// ID for datums we artificially constructed in CreateDatums()
	Id                string `json:"id"`
	Idx               int    `json:"idx"`
	NumDatums         int    `json:"num_datums"`
	AllDatumsReceived bool   `json:"all_datums_received"`
}

type DatumsResponse struct {
	NumDatums         int        `json:"num_datums"`
	Input             *pps.Input `json:"input"`
	Idx               int        `json:"idx"`
	AllDatumsReceived bool       `json:"all_datums_received"`
}

type MountInfo struct {
	Name    string   `json:"name"`
	Project string   `json:"project"`
	Repo    string   `json:"repo"`
	Branch  string   `json:"branch"`
	Commit  string   `json:"commit"` // "" for no commit (commit as noun)
	Paths   []string `json:"paths"`
	Mode    string   `json:"mode"` // "ro", "rw"
}

type Request struct {
	*MountInfo
	Action string // default empty, set to "commit" if we want to commit (verb) a mounted branch
}

type Response struct {
	Project    string
	Repo       string
	Branch     string
	Commit     string // "" for no commit
	Name       string
	MountState MountState
	Error      error
}

type DatumState struct {
	Datums              []*pps.DatumInfo
	PaginationMarker    string
	DatumInput          *pps.Input
	DatumInputsToMounts map[string][]string
	DatumIdx            int
	AllDatumsReceived   bool
}

type MountManager struct {
	Client *client.APIClient
	// only put a value into the States map when we have a goroutine running for
	// it. i.e. when we try to mount it for the first time.
	States map[string]*MountStateMachine
	DatumState
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
	// fetch list of available repos & branches from pachyderm, and overlay that
	// with their mount states
	lr := ListRepoResponse{}
	repos, err := mm.Client.ListRepo()
	if err != nil {
		log.Info(pctx.TODO(), "Error listing repos", zap.Error(err))
		return lr, err
	}

	for _, repo := range repos {
		projectName := repo.Repo.GetProject().GetName()
		repoName := repo.Repo.Name
		rr := RepoResponse{Repo: repoName, Project: projectName}

		readAccess := true
		if repo.AuthInfo != nil {
			readAccess = hasRepoRead(repo.AuthInfo.Permissions)
			rr.Authorization = "none"
		}
		if readAccess {
			bis, err := mm.Client.ListBranch(projectName, repoName)
			if err != nil {
				// Repo was deleted between ListRepo and ListBranch RPCs
				if auth.IsErrNoRoleBinding(err) {
					continue
				}
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
		lr = append(lr, rr)
	}

	return lr, nil
}

func (mm *MountManager) ListByMounts() (ListMountResponse, error) {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	mr := ListMountResponse{
		Mounted:   []MountState{},
		Unmounted: []RepoResponse{},
	}

	for name, msm := range mm.States {
		if msm.State == "mounted" {
			// Check if a mounted repo/branch was deleted to remove it from state.
			var exists bool
			if msm.Mode == "ro" {
				exists, _ = msm.MountInfo.verifyProjectRepoBranchCommitExist(mm.Client)
			} else {
				exists, _ = msm.MountInfo.verifyProjectRepoExist(mm.Client)
			}
			if !exists {
				mm.unmountDeletedRepos(name)
				continue
			}

			if err := msm.RefreshMountState(); err != nil {
				return mr, err
			}
			ms := msm.MountState
			mr.Mounted = append(mr.Mounted, ms)
		}
	}

	lr, err := mm.ListByRepos()
	if err != nil {
		return mr, err
	}
	mr.Unmounted = lr

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
			if _, err := mm.UnmountRepo(name); err != nil {
				return err
			}
		}
	}
	return removeOutDir(mm)
}

func NewMountManager(c *client.APIClient, target string, opts *Options) (ret *MountManager, retErr error) {
	if err := opts.validate(c); err != nil {
		return nil, err
	}
	rootDir, err := os.MkdirTemp("", "pfs")
	if err != nil {
		return nil, errors.WithStack(err)
	}
	log.Info(pctx.TODO(), "Creating root directory", zap.String("path", rootDir))
	if err := os.MkdirAll(rootDir, 0777); err != nil {
		return nil, errors.WithStack(err)
	}
	root, err := newLoopbackRoot(rootDir, target, c, opts)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	log.Info(pctx.TODO(), "Loopback root created", zap.Stringer("path", root))
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

func CreateMount(c *client.APIClient, mountDir string, allowOther bool) (*MountManager, error) {
	mountOpts := &Options{
		Write: true,
		Fuse: &fs.Options{
			MountOptions: gofuse.MountOptions{
				Debug:      false,
				FsName:     "pfs",
				Name:       "pfs",
				AllowOther: allowOther,
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
		log.Info(pctx.TODO(), "Error running mount manager", zap.Error(err))
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
	log.Info(pctx.TODO(), "Dynamically mounting pfs", zap.String("mountDir", sopts.MountDir))

	// This variable points to the MountManager for each connected cluster.
	// Updated when the config is updated.
	var mm *MountManager = &MountManager{}
	if existingClient != nil {
		var err error
		mm, err = CreateMount(existingClient, sopts.MountDir, sopts.AllowOther)
		if err != nil {
			return err
		}
		log.Info(pctx.TODO(), "Connected to existing client", zap.String("address", existingClient.GetAddress().Qualified()))
	}
	router := mux.NewRouter()
	router.Methods("GET").Path("/repos").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		errMsg, webCode := initialChecks(mm, true)
		if errMsg != "" {
			http.Error(w, errMsg, webCode)
			return
		}

		reposList, err := mm.ListByRepos()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		marshalled, err := json.Marshal(reposList)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Write(marshalled) //nolint:errcheck
	})
	router.Methods("GET").Path("/projects").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		errMsg, webCode := initialChecks(mm, true)
		if errMsg != "" {
			http.Error(w, errMsg, webCode)
			return
		}

		projectsList, err := mm.Client.ListProject()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		marshalled, err := json.Marshal(projectsList)
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

		mountsList, err := mm.ListByMounts()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		marshalled, err := json.Marshal(mountsList)
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
		if err := mm.verifyMountRequest(mreq.Mounts); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		for _, mi := range mreq.Mounts {
			if _, err := mm.MountRepo(mi); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		mountsList, err := mm.ListByMounts()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		marshalled, err := jsonMarshal(mountsList)
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

		mountsList, err := mm.ListByMounts()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		marshalled, err := jsonMarshal(mountsList)
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

		mountsList, err := mm.ListByMounts()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		marshalled, err := jsonMarshal(mountsList)
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
		// Reset DatumState in case datums were mounted
		mm.resetDatumState()
		if err := mm.UnmountAll(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		mountsList, err := mm.ListByMounts()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		marshalled, err := jsonMarshal(mountsList)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(marshalled) //nolint:errcheck
	})
	router.Methods("PUT").Path("/datums/_mount").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		errMsg, webCode := initialChecks(mm, true)
		if errMsg != "" {
			http.Error(w, errMsg, webCode)
			return
		}
		if err := mm.UnmountAll(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// Reset DatumState in case datums were mounted
		mm.resetDatumState()
		defer req.Body.Close()
		pipelineReader, err := ppsutil.NewPipelineManifestReader(req.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("error creating pipelineReader from request: %v", err), http.StatusBadRequest)
			return
		}
		// TODO(INT-1006): This is a bit of a hack: the body is a tagged
		// pipeline input spec, not a full pipeline.  In order for
		// parsing to succeed, the next line disables spec validation.
		pipelineReader.DisableValidation()
		pipelineReq, err := pipelineReader.NextCreatePipelineRequest()
		if err != nil {
			http.Error(w, fmt.Sprintf("error reading next CreatePipeline request: %v", err), http.StatusBadRequest)
			return
		}
		var simpleInput bool
		if simpleInput, err = mm.processInput(pipelineReq.Input); err != nil {
			http.Error(w, fmt.Sprintf("error converting datum inputs to mounts: %v", err), http.StatusBadRequest)
			return
		}
		if err := mm.GetDatums(simpleInput); err != nil {
			http.Error(w, fmt.Sprintf("error listing spec's datums: %v", err), http.StatusInternalServerError)
			return
		}
		func() {
			mm.mu.Lock()
			defer mm.mu.Unlock()
			mm.DatumIdx = 0
		}()
		di := mm.Datums[mm.DatumIdx]
		log.Info(pctx.TODO(), "Mounting first datum")
		mis := mm.datumToMounts(di)
		for _, mi := range mis {
			if _, err := mm.MountRepo(mi); err != nil {
				errorMsg := fmt.Sprintf("error mounting %s/%s@%s:(%s): %v", mi.Project, mi.Repo, mi.Branch, strings.Join(mi.Paths, ","), err)
				http.Error(w, errorMsg, http.StatusInternalServerError)
				return
			}
		}
		createLocalOutDir(mm)

		resp := MountDatumResponse{
			Idx:               mm.DatumIdx,
			NumDatums:         len(mm.Datums),
			AllDatumsReceived: mm.AllDatumsReceived,
		}
		marshalled, err := jsonMarshal(resp)
		if err != nil {
			http.Error(w, fmt.Sprintf("error marshalling response: %v", err), http.StatusInternalServerError)
			return
		}
		w.Write(marshalled) //nolint:errcheck
	})
	router.Methods("PUT").Path("/datums/_prev").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		errMsg, webCode := initialChecks(mm, true)
		if errMsg != "" {
			http.Error(w, errMsg, webCode)
			return
		}
		if len(mm.Datums) == 0 {
			http.Error(w, "no datums mounted", http.StatusBadRequest)
			return
		}
		if mm.DatumIdx == 0 {
			http.Error(w, "already at first datum", http.StatusBadRequest)
			return
		}

		if err := mm.UnmountAll(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		func() {
			mm.mu.Lock()
			defer mm.mu.Unlock()
			mm.DatumIdx--
		}()
		di := mm.Datums[mm.DatumIdx]
		mis := mm.datumToMounts(di)
		for _, mi := range mis {
			if _, err := mm.MountRepo(mi); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		createLocalOutDir(mm)

		resp := MountDatumResponse{
			Idx:               mm.DatumIdx,
			NumDatums:         len(mm.Datums),
			AllDatumsReceived: mm.AllDatumsReceived,
		}
		marshalled, err := jsonMarshal(resp)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(marshalled) //nolint:errcheck
	})
	router.Methods("PUT").Path("/datums/_next").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		errMsg, webCode := initialChecks(mm, true)
		if errMsg != "" {
			http.Error(w, errMsg, webCode)
			return
		}
		if len(mm.Datums) == 0 {
			http.Error(w, "no datums mounted", http.StatusBadRequest)
			return
		}
		if !mm.AllDatumsReceived && mm.DatumIdx == len(mm.Datums)-1 {
			if err := mm.getMoreDatums(); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		if mm.AllDatumsReceived && mm.DatumIdx == len(mm.Datums)-1 {
			http.Error(w, "already at last datum", http.StatusBadRequest)
			return
		}

		if err := mm.UnmountAll(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		func() {
			mm.mu.Lock()
			defer mm.mu.Unlock()
			mm.DatumIdx++
		}()
		di := mm.Datums[mm.DatumIdx]
		mis := mm.datumToMounts(di)
		for _, mi := range mis {
			if _, err := mm.MountRepo(mi); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		createLocalOutDir(mm)

		resp := MountDatumResponse{
			Idx:               mm.DatumIdx,
			NumDatums:         len(mm.Datums),
			AllDatumsReceived: mm.AllDatumsReceived,
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

		resp := &DatumsResponse{}
		if len(mm.Datums) != 0 {
			resp = &DatumsResponse{
				NumDatums:         len(mm.Datums),
				Input:             mm.DatumInput,
				Idx:               mm.DatumIdx,
				AllDatumsReceived: mm.AllDatumsReceived,
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
	router.Methods("PUT").Path("/config").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		mm.configMu.Lock()
		defer mm.configMu.Unlock()

		var cfgReq ConfigRequest
		defer req.Body.Close()
		if err := json.NewDecoder(req.Body).Decode(&cfgReq); err != nil {
			http.Error(w, fmt.Sprintf("error decoding request: %v", err), http.StatusBadRequest)
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
				http.Error(w, fmt.Sprintf("error creating new client: %v", err), http.StatusInternalServerError)
				return
			}
			if mm.Client != nil {
				close(mm.opts.getUnmount())
				<-mm.Cleanup
				mm.Client.Close()
			}
			log.Info(pctx.TODO(), "Updating pachd_address", zap.String("address", pachdAddress.Qualified()))
			if mm, err = CreateMount(newClient, sopts.MountDir, sopts.AllowOther); err != nil {
				http.Error(w, fmt.Sprintf("error establishing mount with new pachd address: %v", err), http.StatusInternalServerError)
				return
			}
		}

		clusterStatus, err := getClusterStatus(mm.Client)
		if err != nil {
			http.Error(w, fmt.Sprintf("error retrieving cluster status: %v", err), http.StatusInternalServerError)
			return
		}
		marshalled, err := jsonMarshal(clusterStatus)
		if err != nil {
			http.Error(w, fmt.Sprintf("error marshalling cluster status for response: %v", err), http.StatusInternalServerError)
			return
		}
		w.Write(marshalled) //nolint:errcheck
	})
	// We have split login into two endpoints, both of which must be called to complete the login flow.
	// /auth/_login gets the OIDC login, which the user should be redirected to for login. Following this,
	// /auth/_login_token should be called to retrieve the session token so that the caller may use it as well.
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
		authUrl := loginInfo.LoginUrl
		state := loginInfo.State

		r := map[string]string{"auth_url": authUrl, "oidc_state": state}
		marshalled, err := jsonMarshal(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(marshalled) //nolint:errcheck
	})
	// takes oidc state as arg
	// returns pachyderm config for the mount-server
	router.Methods("PUT").Path("/auth/_login_token").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		defer req.Body.Close()
		data := make(map[string]string) // Maps "oidc" to the OIDC state
		if err := json.NewDecoder(req.Body).Decode(&data); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		resp, err := mm.Client.Authenticate(mm.Client.Ctx(), &auth.AuthenticateRequest{OidcState: data["oidc"]})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		config.WritePachTokenToConfig(resp.PachToken, false) //nolint:errcheck
		mm.Client.SetAuthToken(resp.PachToken)

		cfg, err := config.Read(false, false)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json, err := serde.EncodeJSON(cfg, serde.WithIndent(2))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(json) //nolint:errcheck
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
		health := map[string]string{"status": "running"}
		marshalled, err := jsonMarshal(health)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(marshalled) //nolint:errcheck
	})

	// Using unix domain socket
	var sockPath string
	var srv *http.Server
	if len(sopts.SockPath) > 0 {
		sockPath = sopts.SockPath
		srv = &http.Server{Handler: router}
	} else {
		srv = &http.Server{Addr: FuseServerPort, Handler: router}
	}

	log.AddLoggerToHTTPServer(pctx.TODO(), "http", srv)

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, signals.TerminationSignals...)
		select {
		case <-sigChan:
			os.Remove(sockPath)
		case <-sopts.Unmount:
		}
		if mm.Client != nil {
			close(mm.opts.getUnmount())
			<-mm.Cleanup
		}
		srv.Shutdown(context.Background()) //nolint:errcheck
	}()

	var socket net.Listener
	var err1 error
	var err2 error
	var err error
	if len(sopts.SockPath) > 0 {
		socket, err1 = net.Listen("unix", sockPath)
		err2 = srv.Serve(socket)
		err = errors.Join(err1, err2)
	} else {
		err = srv.ListenAndServe()
	}

	return errors.EnsureStack(err)
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
		log.Info(pctx.TODO(), "New endpoint is same as current endpoint", zap.String("address", pachdAddress.Qualified()))
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
	newClient, err := client.NewFromPachdAddress(pctx.TODO(), pachdAddress, options...)
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
	cfg.V2.Contexts[cfg.V2.ActiveContext] = &config.Context{PachdAddress: pachdAddress.Qualified(), ServerCas: cfgReq.ServerCas}
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
	return slices.Contains(permissions, auth.Permission_REPO_READ) &&
		slices.Contains(permissions, auth.Permission_REPO_LIST_COMMIT) &&
		slices.Contains(permissions, auth.Permission_REPO_LIST_BRANCH) &&
		slices.Contains(permissions, auth.Permission_REPO_LIST_FILE)
}

func hasRepoWrite(permissions []auth.Permission) bool {
	return slices.Contains(permissions, auth.Permission_REPO_WRITE)
}

func (mi *MountInfo) verifyProjectExists(client *client.APIClient) (bool, error) {
	if _, err := client.InspectProject(mi.Project); err != nil {
		return false, err
	}
	return true, nil
}

func (mi *MountInfo) verifyProjectRepoExist(client *client.APIClient) (bool, error) {
	if _, err := mi.verifyProjectExists(client); err != nil {
		return false, err
	}
	if _, err := client.InspectRepo(mi.Project, mi.Repo); err != nil {
		return false, err
	}
	return true, nil
}

func (mi *MountInfo) verifyProjectRepoBranchCommitExist(client *client.APIClient) (bool, error) {
	if _, err := mi.verifyProjectRepoExist(client); err != nil {
		return false, err
	}
	commitInfo, err := client.InspectCommit(mi.Project, mi.Repo, mi.Branch, mi.Commit)
	if err != nil {
		return false, err
	}
	if commitInfo.Commit.Branch != nil {
		mi.Branch = commitInfo.Commit.Branch.Name
	}
	mi.Commit = commitInfo.Commit.Id
	return true, nil
}

func (mm *MountManager) verifyMountRequest(mis []*MountInfo) error {
	for _, mi := range mis {
		if mi.Name == "" {
			return errors.Wrapf(errors.New("no name specified"), "mount request %+v", mi)
		}
		if mi.Project == "" {
			mi.Project = pfs.DefaultProjectName
		}
		if mi.Repo == "" {
			return errors.Wrapf(errors.New("no repo specified"), "mount request %+v", mi)
		}
		if len(mi.Paths) == 0 {
			mi.Paths = []string{"/"}
		}
		// In read-only mode, a commit or branch must exist to mount. In read-write mode, it
		// is not necessary for the branch to exist, as it will be created.
		if mi.Mode == "" || mi.Mode == "ro" {
			mi.Mode = "ro"
			if mi.Commit != "" && !uuid.IsUUIDWithoutDashes(mi.Commit) {
				return errors.Wrapf(errors.New("invalid commit ID"), "mount request %+v", mi)
			}
			if mi.Branch == "" && mi.Commit == "" {
				mi.Branch = "master"
			}
			if exists, err := mi.verifyProjectRepoBranchCommitExist(mm.Client); !exists {
				return errors.Wrapf(err, "mount request %+v", mi)
			}
		} else if mi.Mode == "rw" {
			if exists, err := mi.verifyProjectRepoExist(mm.Client); !exists {
				return errors.Wrapf(err, "mount request %+v", mi)
			}
			if mi.Branch == "" {
				mi.Branch = "master"
			}
			mi.Commit = "" // Disallow mounting a non-branch head commit in rw mode to respect commit provenance
		} else {
			return errors.Wrapf(errors.New("mount mode can only be 'ro' or 'rw'"), "mount request %+v", mi)
		}
	}
	return nil
}

// Visit each entry in the Input spec and set default values if unassigned
// Create a map from input to mount name(s) for use in mounting datums
// Return whether the input has a cron, group by, or join for later use to
// determine whether to use GlobFile or ListDatum #ListDatumPagination
func (mm *MountManager) processInput(datumInput *pps.Input) (bool, error) {
	if datumInput == nil {
		return true, errors.New("datum input is not specified")
	}

	datumInputsToMounts := map[string][]string{} // Maps input to mount name(s)
	simpleInput := true
	if err := pps.VisitInput(datumInput, func(input *pps.Input) error {
		if input.Pfs == nil {
			if input.Cron != nil || input.Join != nil || input.Group != nil {
				simpleInput = false
			}
			return nil
		}

		if input.Pfs.Project == "" {
			input.Pfs.Project = pfs.DefaultProjectName
		}
		if input.Pfs.Commit != "" {
			return errors.New("cannot specify commit in mounting datums Input")
		}
		if input.Pfs.Branch == "" {
			input.Pfs.Branch = "master"
		}
		if input.Pfs.Name == "" {
			input.Pfs.Name = input.Pfs.Project + "_" + input.Pfs.Repo
			if input.Pfs.Branch != "master" {
				input.Pfs.Name = input.Pfs.Name + "_" + input.Pfs.Branch
			}
		}
		// In the case a branch is created off another branch, listing datums returns files from
		// the original branch a commit is on. We should index on that original branch instead.
		bi, err := mm.Client.InspectBranch(input.Pfs.Project, input.Pfs.Repo, input.Pfs.Branch)
		if err != nil {
			return err
		}
		pfsInput := client.NewBranch(input.Pfs.Project, input.Pfs.Repo, bi.Head.Branch.Name).String()
		// In the case where a cross is done on the same repo, we will need to map FileInfo's from
		// the same repo to different user-specified mount names.
		if mountNames, ok := datumInputsToMounts[pfsInput]; ok {
			datumInputsToMounts[pfsInput] = append(mountNames, input.Pfs.Name)
		} else {
			datumInputsToMounts[pfsInput] = []string{input.Pfs.Name}
		}
		return nil
	}); err != nil {
		return true, err
	}
	func() {
		mm.mu.Lock()
		defer mm.mu.Unlock()
		mm.DatumInput = datumInput
		mm.DatumInputsToMounts = datumInputsToMounts
	}()
	return simpleInput, nil
}

func (mm *MountManager) GetNextXDatums(numDatums int, paginationMarker string) ([]*pps.DatumInfo, error) {
	req := &pps.ListDatumRequest{
		Input:            mm.DatumInput,
		Number:           int64(numDatums),
		PaginationMarker: paginationMarker,
	}

	ctx, cf := context.WithCancel(mm.Client.Ctx())
	defer cf()
	client, err := mm.Client.PpsAPIClient.ListDatum(ctx, req)
	if err != nil {
		return nil, err
	}
	var dis []*pps.DatumInfo
	for {
		di, err := client.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return dis, nil
			}
			return nil, err
		}
		dis = append(dis, di)
	}
}

func (mm *MountManager) GetDatums(simpleInput bool) (retErr error) {
	defer func() {
		retErr = grpcutil.ScrubGRPC(retErr)
	}()
	if simpleInput {
		return mm.CreateDatums()
	}
	// If input spec has Join or Group, fall back to ListDatum
	datums, err := mm.GetNextXDatums(NumDatumsPerPage, "")
	if err != nil {
		return err
	}
	if len(datums) == 0 {
		return errors.New("no datums to mount")
	}
	func() {
		mm.mu.Lock()
		defer mm.mu.Unlock()
		mm.Datums = datums
		mm.PaginationMarker = datums[len(datums)-1].Datum.Id
		if len(datums) < NumDatumsPerPage || NumDatumsPerPage == 0 {
			mm.AllDatumsReceived = true
		}
	}()
	return nil
}

// TODO: For input specs that only have cross, union, or PFS inputs, create datums manually
// using GlobFile for now because it's faster than ListDatum. Once ListDatum pagination is
// efficient, we can remove this. #ListDatumPagination
func (mm *MountManager) CreateDatums() error {
	datumsAtLevel := make(map[int][][]*pps.DatumInfo)
	if err := visitInput(mm.DatumInput, 0, func(input *pps.Input, level int) error {
		if input.Pfs != nil {
			// Glob file to get datums associated with this PFS input
			commit := client.NewCommit(input.Pfs.Project, input.Pfs.Repo, input.Pfs.Branch, "")
			datums := []*pps.DatumInfo{}
			if err := mm.Client.GlobFile(commit, input.Pfs.Glob, func(fi *pfs.FileInfo) error {
				datums = append(datums, &pps.DatumInfo{Data: []*pfs.FileInfo{fi}})
				return nil
			}); err != nil {
				return err
			}
			datumsAtLevel[level] = append(datumsAtLevel[level], datums)
		}
		if input.Union != nil {
			// Union all the children PFS inputs datums at level+1
			datums := []*pps.DatumInfo{}
			for _, pfsInputDatums := range datumsAtLevel[level+1] {
				datums = append(datums, pfsInputDatums...)
			}
			datumsAtLevel[level] = append(datumsAtLevel[level], datums)
			delete(datumsAtLevel, level+1)
		}
		if input.Cross != nil {
			// Cross all the children PFS inputs datums at level+1
			datums := crossDatums(datumsAtLevel[level+1])
			datumsAtLevel[level] = append(datumsAtLevel[level], datums)
			delete(datumsAtLevel, level+1)
		}
		return nil
	}); err != nil {
		return err
	}

	datums := datumsAtLevel[0][0]
	if len(datums) == 0 {
		return errors.New("no datums to mount")
	}
	func() {
		mm.mu.Lock()
		defer mm.mu.Unlock()
		mm.Datums = datums
		mm.AllDatumsReceived = true
	}()
	return nil
}

func (mm *MountManager) datumToMounts(d *pps.DatumInfo) []*MountInfo {
	datumInputsToMounts := getCopyOfMapping(mm.DatumInputsToMounts)
	mountToMi := map[string]*MountInfo{}
	for _, fi := range d.Data {
		project := fi.File.Commit.Branch.Repo.GetProject().GetName()
		repo := fi.File.Commit.Branch.Repo.Name
		branch := fi.File.Commit.Branch.Name
		commit := fi.File.Commit.Id
		mountsForRepo := datumInputsToMounts[client.NewBranch(project, repo, branch).String()]
		name := mountsForRepo[0]

		// Could have multiple files to mount from same repo
		if mi, ok := mountToMi[name]; ok {
			mi.Paths = append(mi.Paths, fi.File.Path)
		} else {
			mountToMi[name] = &MountInfo{
				Name:    name,
				Project: project,
				Repo:    repo,
				Branch:  branch,
				Commit:  commit,
				Paths:   []string{fi.File.Path},
				Mode:    "ro",
			}
		}

		// Cycle to next mount name for that repo
		datumInputsToMounts[client.NewBranch(project, repo, branch).String()] = mountsForRepo[1:]
	}
	mis := []*MountInfo{}
	for _, mi := range mountToMi {
		mis = append(mis, mi)
	}
	return mis
}

// Appends the next page of datums if they exist. Might discover we've already
// received all datums.
func (mm *MountManager) getMoreDatums() error {
	datums, err := mm.GetNextXDatums(NumDatumsPerPage, mm.PaginationMarker)
	if err != nil {
		return err
	}
	func() {
		mm.mu.Lock()
		defer mm.mu.Unlock()
		if len(datums) == 0 {
			// we already had all the datums
			mm.AllDatumsReceived = true
		} else {
			mm.Datums = append(mm.Datums, datums...)
			mm.PaginationMarker = datums[len(datums)-1].Datum.Id
			if len(datums) < NumDatumsPerPage || NumDatumsPerPage == 0 {
				mm.AllDatumsReceived = true
			}
		}
	}()
	return nil
}

func (mm *MountManager) resetDatumState() {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	mm.Datums = nil
	mm.PaginationMarker = ""
	mm.DatumInput = nil
	mm.DatumIdx = -1
	mm.DatumInputsToMounts = nil
	mm.AllDatumsReceived = false
}

func removeOutDir(mm *MountManager) error {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	delete(mm.root.repoOpts, "out")
	cleanPath := mm.root.rootPath + "/out"
	return errors.EnsureStack(os.RemoveAll(cleanPath))
}

func createLocalOutDir(mm *MountManager) {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	// Creates folder "out" in mount directory when datums are mounted to
	// simulate pipeline fs. Apart from name of repo (folder name), no other
	// info is necessary.
	mm.root.repoOpts["out"] = &RepoOptions{
		Name:  "out",
		Files: []*pfs.File{client.NewFile("", "out", "", "", "")},
		Write: true,
	}
}

type MountState struct {
	MountInfo         // mount details. written by client
	State      string `json:"state"`      // "unmounted", "mounting", "mounted", "pushing", "unmounted", "error". written by fsm
	Status     string `json:"status"`     // human readable string with additional info wrt State, e.g. an error message for the error state. written by fsm
	Mountpoint string `json:"mountpoint"` // where on the filesystem it's mounted. written by fsm. can also be derived from {MountDir}/{Name}
	// the following are used by the "refresh" button feature in the jupyter plugin
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
	} else if commit == "" {
		m.Status = "branch does not contain any commits"
		return nil
	}
	m.Commit = commit

	// Get the latest commit on the branch
	branchInfo, err := m.manager.Client.InspectBranch(m.Project, m.Repo, m.Branch)
	if err != nil {
		// If we mounted a specific commit but the branch no longer exists,
		// then how many commits we're behind on the branch is meaningless
		if errutil.IsNotFoundError(err) {
			return nil
		}
		return err
	}
	m.LatestCommit = branchInfo.Head.Id

	commitInfos, err := m.manager.Client.ListCommit(branchInfo.Branch.Repo, branchInfo.Head, nil, 0)
	if err != nil {
		return err
	}
	// reverse slice
	for i, j := 0, len(commitInfos)-1; i < j; i, j = i+1, j-1 {
		commitInfos[i], commitInfos[j] = commitInfos[j], commitInfos[i]
	}
	// iterate over commits in branch, calculating how many commits behind LatestCommit Commit is
	log.Debug(pctx.TODO(), "calculating commits behind for mount", zap.String("name", m.Name))
	indexOfCurrentCommit := -1
	for i, commitInfo := range commitInfos {
		log.Debug(pctx.Child(pctx.TODO(), "", pctx.WithoutRatelimit()), "commitInfo dump", zap.Int("i", i), zap.String("commitID", commitInfo.Commit.Id), zap.String("mountedCommitID", m.Commit))
		if commitInfo.Commit.Id == m.Commit {
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
			first8chars(m.Commit),
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

type ListRepoResponse []RepoResponse

type RepoResponse struct {
	Project       string   `json:"project"`
	Repo          string   `json:"repo"`
	Branches      []string `json:"branches"`
	Authorization string   `json:"authorization"` // "off", "none", "read", "write"
	// TODO: Commits map[string]CommitResponse
}

type ListMountResponse struct {
	Mounted   []MountState   `json:"mounted"`
	Unmounted []RepoResponse `json:"unmounted"`
}

// state machines in the style of Rob Pike's Lexical Scanning
// http://www.timschmelmer.com/2013/08/state-machines-with-state-functions-in.html

// setup

type StateFn func(*MountStateMachine) StateFn

func (m *MountStateMachine) transitionedTo(state, status string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	log.Info(pctx.TODO(), "state transition", zap.Any("state machine", m), zap.String("old state", m.State), zap.String("new state", state))
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
	for {
		req := <-m.requests
		switch req.Action {
		case "mount":
			// check user permissions on repo
			repoInfo, err := m.manager.Client.InspectRepo(req.Project, req.Repo)
			if err != nil {
				m.responses <- Response{
					Project:    req.Project,
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
						Project:    req.Project,
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
						Project:    req.Project,
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
			m.Project = req.Project
			m.Repo = req.Repo
			m.Branch = req.Branch
			m.Commit = req.Commit
			m.Paths = req.Paths
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
		files := []*pfs.File{}
		for _, path := range m.Paths {
			files = append(files, client.NewFile(m.Project, m.Repo, m.Branch, m.Commit, path))
		}
		m.manager.mu.Lock()
		defer m.manager.mu.Unlock()
		m.manager.root.repoOpts[m.Name] = &RepoOptions{
			Name:  m.Name,
			Files: files,
			Write: m.Mode == "rw",
		}
		m.manager.root.branches[m.Name] = m.Branch
		if m.Commit != "" {
			m.manager.root.commits[m.Name] = m.Commit
		}
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
			// being used, the request project, repo, and branch match the
			// project, repo, and branch already associated with mount_name.
			// It's essentially a safety check for when we get to the remounting
			// case below.
			if req.Project != m.Project || req.Repo != m.Repo || req.Branch != m.Branch || req.Commit != m.Commit {
				m.responses <- Response{
					Repo:       req.Repo,
					Project:    req.Project,
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
				Project:    req.Project,
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
		log.Info(pctx.TODO(), "Error while uploading!", zap.Error(err))
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
		log.Info(pctx.TODO(), "Error while uploading!", zap.Error(err))
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
	log.Info(pctx.TODO(), "Removing path", zap.String("path", cleanPath))

	err := os.RemoveAll(cleanPath)
	m.responses <- Response{
		MountState: m.MountState,
		Error:      err,
	}
	if err != nil {
		log.Info(pctx.TODO(), "Error while cleaning!", zap.Error(err))
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
	opts, ok := mm.root.repoOpts[name]
	if !ok {
		return nil, errors.Errorf("could not get fuse repo options for mount %s", name)
	}
	projectName := opts.Files[0].Commit.Branch.Repo.Project.GetName()
	repoName := opts.Files[0].Commit.Branch.Repo.Name
	mfc, err := mm.Client.NewModifyFileClient(client.NewCommit(projectName, repoName, mm.root.branch(name), ""))
	if err != nil {
		return nil, err
	}
	mm.mfcs[name] = mfc
	return mfc, nil
}

func (mm *MountManager) uploadFiles(prefixFilter string) (retErr error) {
	ctx, done := log.SpanContextL(pctx.TODO(), "uploadFiles", log.InfoLevel, zap.String("prefix-filter", prefixFilter))
	defer done(log.Errorp(&retErr))
	log.Info(ctx, "Uploading files to Pachyderm...")

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
	log.Info(ctx, "Done!")
	return nil
}
