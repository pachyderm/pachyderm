package server

import (
	"bytes"
	"fmt"
	"io"
	"sync"

	"github.com/golang/snappy"

	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/admin"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/client/pkg/pbutil"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pkg/log"
	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
)

type apiServer struct {
	log.Logger
	address        string
	pachClient     *client.APIClient
	pachClientOnce sync.Once
}

func (a *apiServer) Extract(request *admin.ExtractRequest, extractServer admin.API_ExtractServer) (retErr error) {
	pachClient := a.getPachClient()
	pachClient = pachClient.WithCtx(extractServer.Context())
	handleOp := extractServer.Send
	if request.URL != "" {
		url, err := obj.ParseURL(request.URL)
		if err != nil {
			return fmt.Errorf("error parsing url %v: %v", request.URL, err)
		}
		objClient, err := obj.NewClientFromURLAndSecret(extractServer.Context(), url)
		if err != nil {
			return err
		}
		objW, err := objClient.Writer(url.Object)
		if err != nil {
			return err
		}
		snappyW := snappy.NewBufferedWriter(objW)
		defer func() {
			if err := snappyW.Close(); err != nil && retErr == nil {
				retErr = err
			}
		}()
		w := pbutil.NewWriter(snappyW)
		handleOp = func(op *admin.Op) error { return w.Write(op) }
	}
	v, err := pachClient.VersionAPIClient.GetVersion(pachClient.Ctx(), &types.Empty{})
	if err != nil {
		return err
	}
	if err := handleOp(&admin.Op{
		Version: v,
	}); err != nil {
		return err
	}
	if !request.NoObjects {
		w := &extractObjectWriter{extractServer}
		if err := pachClient.ListObject(func(object *pfs.Object) error {
			if err := pachClient.GetObject(object.Hash, w); err != nil {
				return err
			}
			// empty PutObjectRequest to indicate EOF
			return handleOp(&admin.Op{Object: &pfs.PutObjectRequest{}})
		}); err != nil {
			return err
		}
		if err := pachClient.ListTag(func(resp *pfs.ListTagsResponse) error {
			return handleOp(&admin.Op{
				Tag: &pfs.TagObjectRequest{
					Object: resp.Object,
					Tags:   []*pfs.Tag{resp.Tag},
				},
			})
		}); err != nil {
			return err
		}
	}
	if !request.NoRepos {
		ris, err := pachClient.ListRepo(nil)
		if err != nil {
			return err
		}
		for _, ri := range ris {
			if len(ri.Provenance) > 0 {
				continue
			}
			if err := handleOp(&admin.Op{
				Repo: &pfs.CreateRepoRequest{
					Repo:        ri.Repo,
					Provenance:  ri.Provenance,
					Description: ri.Description,
				},
			}); err != nil {
				return err
			}
			cis, err := pachClient.ListCommit(ri.Repo.Name, "", "", 0)
			if err != nil {
				return err
			}
			for _, ci := range sortCommitInfos(cis) {
				// Even without a parent, ParentCommit is used to indicate which
				// repo to make the commit in.
				if ci.ParentCommit == nil {
					ci.ParentCommit = client.NewCommit(ri.Repo.Name, "")
				}
				if err := handleOp(&admin.Op{
					Commit: &pfs.BuildCommitRequest{
						Parent: ci.ParentCommit,
						Tree:   ci.Tree,
						ID:     ci.Commit.ID,
					},
				}); err != nil {
					return err
				}
			}
			bis, err := pachClient.ListBranch(ri.Repo.Name)
			if err != nil {
				return err
			}
			for _, bi := range bis {
				if err := handleOp(&admin.Op{
					Branch: &pfs.SetBranchRequest{
						Commit: bi.Head,
						Branch: bi.Name,
					},
				},
				); err != nil {
					return err
				}
			}
		}
	}
	if !request.NoPipelines {
		pis, err := pachClient.ListPipeline()
		if err != nil {
			return err
		}
		for _, pi := range pis {
			if err := handleOp(&admin.Op{
				Pipeline: &pps.CreatePipelineRequest{
					Pipeline:           pi.Pipeline,
					Transform:          pi.Transform,
					ParallelismSpec:    pi.ParallelismSpec,
					Egress:             pi.Egress,
					OutputBranch:       pi.OutputBranch,
					ScaleDownThreshold: pi.ScaleDownThreshold,
					ResourceRequests:   pi.ResourceRequests,
					ResourceLimits:     pi.ResourceLimits,
					Input:              pi.Input,
					Description:        pi.Description,
					Incremental:        pi.Incremental,
					CacheSize:          pi.CacheSize,
					EnableStats:        pi.EnableStats,
					Batch:              pi.Batch,
					MaxQueueSize:       pi.MaxQueueSize,
					Service:            pi.Service,
					ChunkSpec:          pi.ChunkSpec,
					DatumTimeout:       pi.DatumTimeout,
					JobTimeout:         pi.JobTimeout,
					Salt:               pi.Salt,
				},
			}); err != nil {
				return err
			}
		}
	}
	return nil
}

func sortCommitInfos(cis []*pfs.CommitInfo) []*pfs.CommitInfo {
	commitMap := make(map[string]*pfs.CommitInfo)
	for _, ci := range cis {
		commitMap[ci.Commit.ID] = ci
	}
	var result []*pfs.CommitInfo
	for _, ci := range cis {
		if commitMap[ci.Commit.ID] == nil {
			continue
		}
		var localResult []*pfs.CommitInfo
		for ci != nil {
			localResult = append(localResult, ci)
			delete(commitMap, ci.Commit.ID)
			if ci.ParentCommit != nil {
				ci = commitMap[ci.ParentCommit.ID]
			} else {
				ci = nil
			}
		}
		for i := range localResult {
			result = append(result, localResult[len(localResult)-i-1])
		}
	}
	return result
}

func (a *apiServer) Restore(restoreServer admin.API_RestoreServer) (retErr error) {
	ctx := restoreServer.Context()
	pachClient := a.getPachClient()
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
				objClient, err := obj.NewClientFromURLAndSecret(restoreServer.Context(), url)
				if err != nil {
					return err
				}
				objR, err := objClient.Reader(url.Object, 0, 0)
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
		switch {
		case op.Version != nil:
		case op.Object != nil:
			r := &extractObjectReader{adminAPIRestoreServer: restoreServer}
			r.buf.Write(op.Object.Value)
			if _, _, err := pachClient.PutObject(r); err != nil {
				return fmt.Errorf("error putting object: %v", err)
			}
		case op.Tag != nil:
			if _, err := pachClient.ObjectAPIClient.TagObject(ctx, op.Tag); err != nil {
				return fmt.Errorf("error tagging object: %v", grpcutil.ScrubGRPC(err))
			}
		case op.Repo != nil:
			if _, err := pachClient.PfsAPIClient.CreateRepo(ctx, op.Repo); err != nil {
				return fmt.Errorf("error creating repo: %v", grpcutil.ScrubGRPC(err))
			}
		case op.Commit != nil:
			if _, err := pachClient.PfsAPIClient.BuildCommit(ctx, op.Commit); err != nil {
				return fmt.Errorf("error creating commit: %v", grpcutil.ScrubGRPC(err))
			}
		case op.Branch != nil:
			if _, err := pachClient.PfsAPIClient.SetBranch(ctx, op.Branch); err != nil {
				return fmt.Errorf("error creating branch: %v", grpcutil.ScrubGRPC(err))
			}
		case op.Pipeline != nil:
			if _, err := pachClient.PpsAPIClient.CreatePipeline(ctx, op.Pipeline); err != nil {
				return fmt.Errorf("error creating pipeline: %v", grpcutil.ScrubGRPC(err))
			}
		}
	}
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

type extractObjectWriter struct {
	admin.API_ExtractServer
}

func (w *extractObjectWriter) Write(p []byte) (int, error) {
	chunkSize := grpcutil.MaxMsgSize / 2
	var n int
	for i := 0; i*(chunkSize) < len(p); i++ {
		value := p[i*chunkSize:]
		if len(value) > chunkSize {
			value = value[:chunkSize]
		}
		if err := w.Send(&admin.Op{Object: &pfs.PutObjectRequest{Value: value}}); err != nil {
			return n, err
		}
		n += len(value)
	}
	return n, nil
}

type adminAPIRestoreServer admin.API_RestoreServer

type extractObjectReader struct {
	adminAPIRestoreServer
	buf bytes.Buffer
	eof bool
}

func (r *extractObjectReader) Read(p []byte) (int, error) {
	for len(p) > r.buf.Len() && !r.eof {
		request, err := r.Recv()
		if err != nil {
			return 0, grpcutil.ScrubGRPC(err)
		}
		op := request.Op
		if op.Object == nil {
			return 0, fmt.Errorf("expected an object, but got: %v", op)
		}
		r.buf.Write(op.Object.Value)
		if len(op.Object.Value) == 0 {
			r.eof = true
		}
	}
	return r.buf.Read(p)
}
