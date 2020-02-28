package server

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"sync"
	"time"

	"github.com/golang/snappy"
	"golang.org/x/net/context"

	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/admin"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/client/pkg/pbutil"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pkg/ancestry"
	"github.com/pachyderm/pachyderm/src/server/pkg/errutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/log"
	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
	"github.com/pachyderm/pachyderm/src/server/pkg/ppsconsts"
	"github.com/pachyderm/pachyderm/src/server/pkg/ppsutil"
)

var objHashRE = regexp.MustCompile("[0-9a-f]{128}")

type apiServer struct {
	log.Logger
	address        string
	storageRoot    string // for downloading/converting hashtrees
	pachClient     *client.APIClient
	pachClientOnce sync.Once
	clusterInfo    *admin.ClusterInfo
}

func (a *apiServer) InspectCluster(ctx context.Context, request *types.Empty) (*admin.ClusterInfo, error) {
	return a.clusterInfo, nil
}

type opVersion int8

const (
	undefined opVersion = iota
	v1_7
	v1_8
	v1_9
)

func (v opVersion) String() string {
	switch v {
	case v1_7:
		return "1.7"
	case v1_8:
		return "1.8"
	case v1_9:
		return "1.9"
	}
	return "undefined"
}

func version(op *admin.Op) opVersion {
	switch {
	case op.Op1_7 != nil:
		return v1_7
	case op.Op1_8 != nil:
		return v1_8
	case op.Op1_9 != nil:
		return v1_9
	default:
		return undefined
	}
}

func (a *apiServer) Extract(request *admin.ExtractRequest, extractServer admin.API_ExtractServer) (retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, nil, retErr, time.Since(start)) }(time.Now())
	ctx := extractServer.Context()
	pachClient := a.getPachClient().WithCtx(ctx)
	writeOp := extractServer.Send
	if request.URL != "" {
		url, err := obj.ParseURL(request.URL)
		if err != nil {
			return fmt.Errorf("error parsing url %v: %v", request.URL, err)
		}
		if url.Object == "" {
			return fmt.Errorf("URL must be <svc>://<bucket>/<object> (no object in %s)", request.URL)
		}
		objClient, err := obj.NewClientFromURLAndSecret(url, false)
		if err != nil {
			return err
		}
		objW, err := objClient.Writer(extractServer.Context(), url.Object)
		if err != nil {
			return err
		}
		defer func() {
			if err := objW.Close(); err != nil && retErr == nil {
				retErr = err
			}
		}()
		snappyW := snappy.NewBufferedWriter(objW)
		defer func() {
			if err := snappyW.Close(); err != nil && retErr == nil {
				retErr = err
			}
		}()
		w := pbutil.NewWriter(snappyW)
		writeOp = func(op *admin.Op) error {
			_, err := w.Write(op)
			return err
		}
	}
	if !request.NoObjects {
		if err := pachClient.ListBlock(func(block *pfs.Block) error {
			w := &extractBlockWriter{f: writeOp, block: block}
			if err := pachClient.GetBlock(block.Hash, w); err != nil {
				return err
			}
			return w.Close()
		}); err != nil {
			return err
		}
		if err := pachClient.ListObject(func(oi *pfs.ObjectInfo) error {
			return writeOp(&admin.Op{Op1_9: &admin.Op1_9{CreateObject: &pfs.CreateObjectRequest{
				Object:   oi.Object,
				BlockRef: oi.BlockRef,
			}}})
		}); err != nil {
			return err
		}
		if err := pachClient.ListTag(func(resp *pfs.ListTagsResponse) error {
			return writeOp(&admin.Op{Op1_9: &admin.Op1_9{
				Tag: &pfs.TagObjectRequest{
					Object: resp.Object,
					Tags:   []*pfs.Tag{resp.Tag},
				},
			}})
		}); err != nil {
			return err
		}
	}
	if !request.NoRepos {
		ris, err := pachClient.ListRepo()
		if err != nil {
			return err
		}
		ris = append(ris, &pfs.RepoInfo{Repo: &pfs.Repo{Name: ppsconsts.SpecRepo}})
		for i := range ris {
			ri := ris[len(ris)-1-i]
			if err := writeOp(&admin.Op{Op1_9: &admin.Op1_9{
				Repo: &pfs.CreateRepoRequest{
					Repo:        ri.Repo,
					Description: ri.Description,
				}},
			}); err != nil {
				return err
			}
		}
		if err := pachClient.ListCommitF("", "", "", 0, true, func(ci *pfs.CommitInfo) error {
			if ci.ParentCommit == nil {
				ci.ParentCommit = client.NewCommit(ci.Commit.Repo.Name, "")
			}
			return writeOp(&admin.Op{Op1_9: &admin.Op1_9{Commit: &pfs.BuildCommitRequest{
				Parent:     ci.ParentCommit,
				Tree:       ci.Tree,
				ID:         ci.Commit.ID,
				Trees:      ci.Trees,
				Datums:     ci.Datums,
				SizeBytes:  ci.SizeBytes,
				Provenance: ci.Provenance,
			}}})
		}); err != nil {
			return err
		}
		bis, err := pachClient.PfsAPIClient.ListBranch(pachClient.Ctx(),
			&pfs.ListBranchRequest{
				Repo:    client.NewRepo(""),
				Reverse: true,
			},
		)
		if err != nil {
			return err
		}
		for _, bi := range bis.BranchInfo {
			if err := writeOp(&admin.Op{Op1_9: &admin.Op1_9{
				Branch: &pfs.CreateBranchRequest{
					Head:       bi.Head,
					Branch:     bi.Branch,
					Provenance: bi.DirectProvenance,
				},
			}}); err != nil {
				return err
			}
		}
	}
	if !request.NoPipelines {
		pis, err := pachClient.ListPipeline()
		if err != nil {
			return err
		}
		pis = sortPipelineInfos(pis)
		for _, pi := range pis {
			cPR := ppsutil.PipelineReqFromInfo(pi)
			cPR.SpecCommit = pi.SpecCommit
			if err := writeOp(&admin.Op{Op1_9: &admin.Op1_9{Pipeline: cPR}}); err != nil {
				return err
			}
			if err := pachClient.ListJobF(pi.Pipeline.Name, nil, nil, -1, false, func(ji *pps.JobInfo) error {
				return writeOp(&admin.Op{Op1_9: &admin.Op1_9{Job: &pps.CreateJobRequest{
					Pipeline:      pi.Pipeline,
					OutputCommit:  ji.OutputCommit,
					Restart:       ji.Restart,
					DataProcessed: ji.DataProcessed,
					DataSkipped:   ji.DataSkipped,
					DataTotal:     ji.DataTotal,
					DataFailed:    ji.DataFailed,
					DataRecovered: ji.DataRecovered,
					Stats:         ji.Stats,
					StatsCommit:   ji.StatsCommit,
					State:         ji.State,
					Reason:        ji.Reason,
					Started:       ji.Started,
					Finished:      ji.Finished,
				}}})
			}); err != nil {
				return err
			}
		}
	}
	return nil
}

func (a *apiServer) ExtractPipeline(ctx context.Context, request *admin.ExtractPipelineRequest) (response *admin.Op, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	pachClient := a.getPachClient().WithCtx(ctx)
	pi, err := pachClient.InspectPipeline(request.Pipeline.Name)
	if err != nil {
		return nil, err
	}
	return &admin.Op{Op1_9: &admin.Op1_9{Pipeline: ppsutil.PipelineReqFromInfo(pi)}}, nil
}

func sortPipelineInfos(pis []*pps.PipelineInfo) []*pps.PipelineInfo {
	piMap := make(map[string]*pps.PipelineInfo)
	for _, pi := range pis {
		piMap[pi.Pipeline.Name] = pi
	}
	var result []*pps.PipelineInfo
	var add func(string)
	add = func(name string) {
		if pi, ok := piMap[name]; ok {
			pps.VisitInput(pi.Input, func(input *pps.Input) {
				if input.Pfs != nil {
					add(input.Pfs.Repo)
				}
			})
			result = append(result, pi)
			delete(piMap, name)
		}
	}
	for _, pi := range pis {
		add(pi.Pipeline.Name)
	}
	return result
}

func (a *apiServer) Restore(restoreServer admin.API_RestoreServer) (retErr error) {
	func() { a.Log(nil, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(nil, nil, retErr, time.Since(start)) }(time.Now())
	pachClient := a.getPachClient().WithCtx(restoreServer.Context())
	defer func() {
		for {
			_, err := restoreServer.Recv()
			if err != nil {
				break
			}
		}
		if err := restoreServer.SendAndClose(&types.Empty{}); err != nil && retErr == nil {
			retErr = err
		}
	}()
	var r pbutil.Reader
	var streamVersion opVersion
	for {
		var op *admin.Op
		if r == nil {
			req, err := restoreServer.Recv()
			if err != nil {
				if err == io.EOF {
					return nil
				}
				return err
			}
			if req.URL != "" {
				url, err := obj.ParseURL(req.URL)
				if err != nil {
					return fmt.Errorf("error parsing url %v: %v", req.URL, err)
				}
				if url.Object == "" {
					return fmt.Errorf("URL must be <svc>://<bucket>/<object> (no object in %s)", req.URL)
				}
				objClient, err := obj.NewClientFromURLAndSecret(url, false)
				if err != nil {
					return err
				}
				objR, err := objClient.Reader(restoreServer.Context(), url.Object, 0, 0)
				if err != nil {
					return err
				}
				snappyR := snappy.NewReader(objR)
				r = pbutil.NewReader(snappyR)
				continue
			} else {
				op = req.Op
			}
		} else {
			op = &admin.Op{}
			if err := r.Read(op); err != nil {
				if err == io.EOF {
					return nil
				}
				return err
			}
		}

		// validate op version
		if streamVersion == undefined {
			streamVersion = version(op)
		} else if streamVersion != version(op) {
			return fmt.Errorf("cannot mix different versions of pachd operation "+
				"within a metadata dumps (found both %s and %s)", version(op), streamVersion)
		}

		// apply op
		if op.Op1_7 != nil {
			if op.Op1_7.Object != nil {
				extractReader := &extractObjectReader{
					adminAPIRestoreServer: restoreServer,
					restoreURLReader:      r,
					version:               v1_7,
				}
				extractReader.buf.Write(op.Op1_7.Object.Value)
				if _, _, err := pachClient.PutObject(extractReader); err != nil {
					return fmt.Errorf("error putting object: %v", err)
				}
			} else {
				newOp1_8, err := convert1_7Op(pachClient, a.storageRoot, op.Op1_7)
				if err != nil {
					return err
				}
				newOp, err := convert1_8Op(newOp1_8)
				if err != nil {
					return err
				}
				if err := a.applyOp(pachClient, newOp); err != nil {
					return err
				}
			}
		} else if op.Op1_8 != nil {
			if op.Op1_8.Object != nil {
				extractReader := &extractObjectReader{
					adminAPIRestoreServer: restoreServer,
					restoreURLReader:      r,
					version:               v1_8,
				}
				extractReader.buf.Write(op.Op1_8.Object.Value)
				if _, _, err := pachClient.PutObject(extractReader); err != nil {
					return fmt.Errorf("error putting object: %v", err)
				}
			} else {
				newOp, err := convert1_8Op(op.Op1_8)
				if err != nil {
					return err
				}
				if err := a.applyOp(pachClient, newOp); err != nil {
					return err
				}
			}
		} else if op.Op1_9 != nil {
			if op.Op1_9.Object != nil {
				extractReader := &extractObjectReader{
					adminAPIRestoreServer: restoreServer,
					restoreURLReader:      r,
					version:               v1_9,
				}
				extractReader.buf.Write(op.Op1_9.Object.Value)
				if _, _, err := pachClient.PutObject(extractReader); err != nil {
					return fmt.Errorf("error putting object: %v", err)
				}
			} else if op.Op1_9.Block != nil {
				if len(op.Op1_9.Block.Value) == 0 {
					// Empty block
					if _, err := pachClient.PutBlock(op.Op1_9.Block.Block.Hash, bytes.NewReader(nil)); err != nil {
						return fmt.Errorf("error putting block: %v", err)
					}
				} else {
					extractReader := &extractBlockReader{
						adminAPIRestoreServer: restoreServer,
						restoreURLReader:      r,
						version:               v1_9,
					}
					extractReader.buf.Write(op.Op1_9.Block.Value)
					if _, err := pachClient.PutBlock(op.Op1_9.Block.Block.Hash, extractReader); err != nil {
						return fmt.Errorf("error putting block: %v", err)
					}
				}
			} else {
				if err := a.applyOp(pachClient, op.Op1_9); err != nil {
					return err
				}
			}
		}
	}
}

func (a *apiServer) applyOp(pachClient *client.APIClient, op *admin.Op1_9) error {
	switch {
	case op.CreateObject != nil:
		if _, err := pachClient.ObjectAPIClient.CreateObject(pachClient.Ctx(), op.CreateObject); err != nil {
			return fmt.Errorf("error creating object: %v", grpcutil.ScrubGRPC(err))
		}
	case op.Tag != nil:
		if _, err := pachClient.ObjectAPIClient.TagObject(pachClient.Ctx(), op.Tag); err != nil {
			return fmt.Errorf("error tagging object: %v", grpcutil.ScrubGRPC(err))
		}
	case op.Repo != nil:
		op.Repo.Repo.Name = ancestry.SanitizeName(op.Repo.Repo.Name)
		if _, err := pachClient.PfsAPIClient.CreateRepo(pachClient.Ctx(), op.Repo); err != nil && !errutil.IsAlreadyExistError(err) {
			return fmt.Errorf("error creating repo: %v", grpcutil.ScrubGRPC(err))
		}
	case op.Commit != nil:
		if _, err := pachClient.PfsAPIClient.BuildCommit(pachClient.Ctx(), op.Commit); err != nil && !errutil.IsAlreadyExistError(err) {
			return fmt.Errorf("error creating commit: %v", grpcutil.ScrubGRPC(err))
		}
	case op.Branch != nil:
		if op.Branch.Branch == nil {
			op.Branch.Branch = client.NewBranch(op.Branch.Head.Repo.Name, ancestry.SanitizeName(op.Branch.SBranch))
		}
		if _, err := pachClient.PfsAPIClient.CreateBranch(pachClient.Ctx(), op.Branch); err != nil && !errutil.IsAlreadyExistError(err) {
			return fmt.Errorf("error creating branch: %v", grpcutil.ScrubGRPC(err))
		}
	case op.Pipeline != nil:
		sanitizePipeline(op.Pipeline)
		if _, err := pachClient.PpsAPIClient.CreatePipeline(pachClient.Ctx(), op.Pipeline); err != nil && !errutil.IsAlreadyExistError(err) {
			return fmt.Errorf("error creating pipeline: %v", grpcutil.ScrubGRPC(err))
		}
	case op.Job != nil:
		if _, err := pachClient.PpsAPIClient.CreateJob(pachClient.Ctx(), op.Job); err != nil && !errutil.IsAlreadyExistError(err) {
			return fmt.Errorf("error creating job: %v", grpcutil.ScrubGRPC(err))
		}
	}
	return nil
}

func sanitizePipeline(req *pps.CreatePipelineRequest) {
	req.Pipeline.Name = ancestry.SanitizeName(req.Pipeline.Name)
	pps.VisitInput(req.Input, func(input *pps.Input) {
		if input.Pfs != nil {
			if input.Pfs.Branch != "" {
				input.Pfs.Branch = ancestry.SanitizeName(input.Pfs.Branch)
			}
			input.Pfs.Repo = ancestry.SanitizeName(input.Pfs.Repo)
		}
	})
}

func (a *apiServer) getPachClient() *client.APIClient {
	a.pachClientOnce.Do(func() {
		var err error
		a.pachClient, err = client.NewFromAddress(a.address)
		if err != nil {
			panic(fmt.Sprintf("pps failed to initialize pach client: %v", err))
		}
	})
	return a.pachClient
}

type adminAPIRestoreServer admin.API_RestoreServer

type extractObjectReader struct {
	// One of these two must be set (whether user is restoring over the wire or
	// via URL)
	adminAPIRestoreServer
	restoreURLReader pbutil.Reader

	version opVersion
	buf     bytes.Buffer
	eof     bool
}

func (r *extractObjectReader) Read(p []byte) (int, error) {
	// Shortcut -- if object is done just return EOF
	if r.eof {
		return 0, io.EOF
	}

	// Read leftover bytes in buffer (from prior Read() call) into 'p'
	n, err := r.buf.Read(p)
	if n == len(p) || err != nil && err != io.EOF {
		return n, err // quit early if done; ignore EOF--just means buf is now empty
	}
	r.buf.Reset() // discard data now in 'p'; ready to refill 'r.buf'
	p = p[n:]     // only want to fill remainder of p

	// refill 'r.buf'
	for len(p) > r.buf.Len() && !r.eof {
		var op *admin.Op
		if r.restoreURLReader == nil {
			request, err := r.Recv()
			if err != nil {
				return 0, grpcutil.ScrubGRPC(err)
			}
			op = request.Op
		} else {
			if op == nil {
				op = &admin.Op{}
			} else {
				*op = admin.Op{} // clear 'op' without making old contents into garbage
			}
			if err := r.restoreURLReader.Read(op); err != nil {
				return 0, fmt.Errorf("unexpected error while restoring object: %v", err)
			}
		}

		// Validate op version
		if r.version != version(op) {
			return 0, fmt.Errorf("cannot mix different versions of pachd operation "+
				"within a metadata dumps (found both %s and %s)", version(op), r.version)
		}

		// extract object bytes
		var value []byte
		if r.version == v1_7 {
			if op.Op1_7.Object == nil {
				return 0, fmt.Errorf("expected an object, but got: %v", op)
			}
			value = op.Op1_7.Object.Value
		} else if r.version == v1_8 {
			if op.Op1_8.Object == nil {
				return 0, fmt.Errorf("expected an object, but got: %v", op)
			}
			value = op.Op1_8.Object.Value
		} else {
			if op.Op1_9.Object == nil {
				return 0, fmt.Errorf("expected an object, but got: %v", op)
			}
			value = op.Op1_9.Object.Value
		}

		if len(value) == 0 {
			r.eof = true
		} else {
			r.buf.Write(value)
		}
	}
	dn, err := r.buf.Read(p)
	return n + dn, err
}

type extractBlockWriter struct {
	f     func(*admin.Op) error
	block *pfs.Block
}

func (w extractBlockWriter) Write(p []byte) (int, error) {
	chunkSize := grpcutil.MaxMsgSize / 2
	var n int
	for i := 0; i*(chunkSize) < len(p); i++ {
		value := p[i*chunkSize:]
		if len(value) > chunkSize {
			value = value[:chunkSize]
		}
		if err := w.f(&admin.Op{Op1_9: &admin.Op1_9{Block: &pfs.PutBlockRequest{Block: w.block, Value: value}}}); err != nil {
			return n, err
		}
		w.block = nil // only need to send block on the first request
		n += len(value)
	}
	return n, nil
}

func (w extractBlockWriter) Close() error {
	return w.f(&admin.Op{Op1_9: &admin.Op1_9{Block: &pfs.PutBlockRequest{Block: w.block}}})
}

type extractBlockReader struct {
	// One of these two must be set (whether user is restoring over the wire or
	// via URL)
	adminAPIRestoreServer
	restoreURLReader pbutil.Reader

	version opVersion
	buf     bytes.Buffer
	eof     bool
}

func (r *extractBlockReader) Read(p []byte) (int, error) {
	// Shortcut -- if object is done just return EOF
	if r.eof {
		return 0, io.EOF
	}

	// Read leftover bytes in buffer (from prior Read() call) into 'p'
	n, err := r.buf.Read(p)
	if n == len(p) || err != nil && err != io.EOF {
		return n, err // quit early if done; ignore EOF--just means buf is now empty
	}
	r.buf.Reset() // discard data now in 'p'; ready to refill 'r.buf'
	p = p[n:]     // only want to fill remainder of p

	// refill 'r.buf'
	for len(p) > r.buf.Len() && !r.eof {
		var op *admin.Op
		if r.restoreURLReader == nil {
			request, err := r.Recv()
			if err != nil {
				return 0, grpcutil.ScrubGRPC(err)
			}
			op = request.Op
		} else {
			if op == nil {
				op = &admin.Op{}
			} else {
				*op = admin.Op{} // clear 'op' without making old contents into garbage
			}
			if err := r.restoreURLReader.Read(op); err != nil {
				return 0, fmt.Errorf("unexpected error while restoring object: %v", err)
			}
		}

		// Validate op version
		if r.version != version(op) {
			return 0, fmt.Errorf("cannot mix different versions of pachd operation "+
				"within a metadata dumps (found both %s and %s)", version(op), r.version)
		}

		// extract object bytes
		var value []byte
		if r.version == v1_7 {
			return 0, fmt.Errorf("invalid version 1.7 doesn't have extracted blocks")
		} else if r.version == v1_8 {
			return 0, fmt.Errorf("invalid version 1.8 doesn't have extracted blocks")
		} else {
			if op.Op1_9.Block == nil {
				return 0, fmt.Errorf("expected a block, but got: %v", op)
			}
			value = op.Op1_9.Block.Value
		}

		if len(value) == 0 {
			r.eof = true
		} else {
			r.buf.Write(value)
		}
	}
	dn, err := r.buf.Read(p)
	return n + dn, err
}
