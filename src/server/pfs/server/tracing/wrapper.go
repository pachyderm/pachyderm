package main

import (
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"google.golang.org/grpc/metadata"
)

// tracingServer wraps the given PFS APIServer with another implementation of
// pfs.APIServer that
type tracingServer struct {
	pfs.APIServer
}

// Tracing wraps 'server' in a tracing pfs.APIServer that delegates all logic to
// the inner server, but wraps all calls in tracing logic that propagates any
// incoming traces and tags them with request info
func Tracing(server pfs.APIServer) pfs.APIServer {
	return &tracingServer{
		APIServer: server,
	}
}

func intercept(ctx, tags ...string) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	}
	spanContext, err := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, metadataReaderWriter{md})
	if err != nil && err != opentracing.ErrSpanContextNotFound {
		// TODO: establish some sort of error reporting mechanism here. We
		// don't know where to put such an error and must rely on Tracer
		// implementations to do something appropriate for the time being.
	}
	serverSpan := tracer.StartSpan(
		info.FullMethod,
		ext.RPCServerOption(spanContext),
		gRPCComponentTag,
	)
	defer serverSpan.Finish()

	ctx = opentracing.ContextWithSpan(ctx, serverSpan)
	if otgrpcOpts.logPayloads {
		serverSpan.LogFields(log.Object("gRPC request", req))
	}
	resp, err = handler(ctx, req)
	if err == nil {
		if otgrpcOpts.logPayloads {
			serverSpan.LogFields(log.Object("gRPC response", resp))
		}
	} else {
		SetSpanTags(serverSpan, err, false)
		serverSpan.LogFields(log.String("event", "error"), log.String("message", err.Error()))
	}
	if otgrpcOpts.decorator != nil {
		otgrpcOpts.decorator(serverSpan, info.FullMethod, req, resp, err)
	}
	return resp, err
}

// CreateRepo implements the corresponding method of the pfs.APIServer interface
func (s *tracingServer) CreateRepo(context.Context, *CreateRepoRequest) (*types.Empty, error) {
	opentracing.start_TK
	opentracing.endTK()
	return s.APIClient.CreateRepo(context.Context, *CreateRepoRequest)
}

// InspectRepo implements the corresponding method of the pfs.APIServer interface
func (s *tracingServer) InspectRepo(context.Context, *InspectRepoRequest) (*RepoInfo, error) {
	opentracing.start_TK
	opentracing.endTK()
	return s.APIClient.InspectRepo(context.Context, *InspectRepoRequest)
}

// ListRepo implements the corresponding method of the pfs.APIServer interface
func (s *tracingServer) ListRepo(context.Context, *ListRepoRequest) (*ListRepoResponse, error) {
	opentracing.start_TK
	opentracing.endTK()
	return s.APIClient.ListRepo(context.Context, *ListRepoRequest)
}

// DeleteRepo implements the corresponding method of the pfs.APIServer interface
func (s *tracingServer) DeleteRepo(context.Context, *DeleteRepoRequest) (*types.Empty, error) {
	opentracing.start_TK
	opentracing.endTK()
	return s.APIClient.DeleteRepo(context.Context, *DeleteRepoRequest)
}

// StartCommit implements the corresponding method of the pfs.APIServer interface
func (s *tracingServer) StartCommit(context.Context, *StartCommitRequest) (*Commit, error) {
	opentracing.start_TK
	opentracing.endTK()
	return s.APIClient.StartCommit(context.Context, *StartCommitRequest)
}

// FinishCommit implements the corresponding method of the pfs.APIServer interface
func (s *tracingServer) FinishCommit(context.Context, *FinishCommitRequest) (*types.Empty, error) {
	opentracing.start_TK
	opentracing.endTK()
	return s.APIClient.FinishCommit(context.Context, *FinishCommitRequest)
}

// InspectCommit implements the corresponding method of the pfs.APIServer interface
func (s *tracingServer) InspectCommit(context.Context, *InspectCommitRequest) (*CommitInfo, error) {
	opentracing.start_TK
	opentracing.endTK()
	return s.APIClient.InspectCommit(context.Context, *InspectCommitRequest)
}

// ListCommit implements the corresponding method of the pfs.APIServer interface
func (s *tracingServer) ListCommit(context.Context, *ListCommitRequest) (*CommitInfos, error) {
	opentracing.start_TK
	opentracing.endTK()
	return s.APIClient.ListCommit(context.Context, *ListCommitRequest)
}

// ListCommitStream implements the corresponding method of the pfs.APIServer interface
func (s *tracingServer) ListCommitStream(*ListCommitRequest, API_ListCommitStreamServer) error {
	opentracing.start_TK
	opentracing.endTK()
	return s.APIClient.ListCommitStream(*ListCommitRequest, API_ListCommitStreamServer)
}

// DeleteCommit implements the corresponding method of the pfs.APIServer interface
func (s *tracingServer) DeleteCommit(context.Context, *DeleteCommitRequest) (*types.Empty, error) {
	opentracing.start_TK
	opentracing.endTK()
	return s.APIClient.DeleteCommit(context.Context, *DeleteCommitRequest)
}

// FlushCommit implements the corresponding method of the pfs.APIServer interface
func (s *tracingServer) FlushCommit(*FlushCommitRequest, API_FlushCommitServer) error {
	opentracing.start_TK
	opentracing.endTK()
	return s.APIClient.FlushCommit(*FlushCommitRequest, API_FlushCommitServer)
}

// SubscribeCommit implements the corresponding method of the pfs.APIServer interface
func (s *tracingServer) SubscribeCommit(*SubscribeCommitRequest, API_SubscribeCommitServer) error {
	opentracing.start_TK
	opentracing.endTK()
	return s.APIClient.SubscribeCommit(*SubscribeCommitRequest, API_SubscribeCommitServer)
}

// BuildCommit implements the corresponding method of the pfs.APIServer interface
func (s *tracingServer) BuildCommit(context.Context, *BuildCommitRequest) (*Commit, error) {
	opentracing.start_TK
	opentracing.endTK()
	return s.APIClient.BuildCommit(context.Context, *BuildCommitRequest)
}

// CreateBranch implements the corresponding method of the pfs.APIServer interface
func (s *tracingServer) CreateBranch(context.Context, *CreateBranchRequest) (*types.Empty, error) {
	opentracing.start_TK
	opentracing.endTK()
	return s.APIClient.CreateBranch(context.Context, *CreateBranchRequest)
}

// InspectBranch implements the corresponding method of the pfs.APIServer interface
func (s *tracingServer) InspectBranch(context.Context, *InspectBranchRequest) (*BranchInfo, error) {
	opentracing.start_TK
	opentracing.endTK()
	return s.APIClient.InspectBranch(context.Context, *InspectBranchRequest)
}

// ListBranch implements the corresponding method of the pfs.APIServer interface
func (s *tracingServer) ListBranch(context.Context, *ListBranchRequest) (*BranchInfos, error) {
	opentracing.start_TK
	opentracing.endTK()
	return s.APIClient.ListBranch(context.Context, *ListBranchRequest)
}

// DeleteBranch implements the corresponding method of the pfs.APIServer interface
func (s *tracingServer) DeleteBranch(context.Context, *DeleteBranchRequest) (*types.Empty, error) {
	opentracing.start_TK
	opentracing.endTK()
	return s.APIClient.DeleteBranch(context.Context, *DeleteBranchRequest)
}

// PutFile implements the corresponding method of the pfs.APIServer interface
func (s *tracingServer) PutFile(API_PutFileServer) error {
	opentracing.start_TK
	opentracing.endTK()
	return s.APIClient.PutFile(API_PutFileServer)
}

// CopyFile implements the corresponding method of the pfs.APIServer interface
func (s *tracingServer) CopyFile(context.Context, *CopyFileRequest) (*types.Empty, error) {
	opentracing.start_TK
	opentracing.endTK()
	return s.APIClient.CopyFile(context.Context, *CopyFileRequest)
}

// GetFile implements the corresponding method of the pfs.APIServer interface
func (s *tracingServer) GetFile(*GetFileRequest, API_GetFileServer) error {
	opentracing.start_TK
	opentracing.endTK()
	return s.APIClient.GetFile(*GetFileRequest, API_GetFileServer)
}

// InspectFile implements the corresponding method of the pfs.APIServer interface
func (s *tracingServer) InspectFile(context.Context, *InspectFileRequest) (*FileInfo, error) {
	opentracing.start_TK
	opentracing.endTK()
	return s.APIClient.InspectFile(context.Context, *InspectFileRequest)
}

// ListFile implements the corresponding method of the pfs.APIServer interface
func (s *tracingServer) ListFile(context.Context, *ListFileRequest) (*FileInfos, error) {
	opentracing.start_TK
	opentracing.endTK()
	return s.APIClient.ListFile(context.Context, *ListFileRequest)
}

// ListFileStream implements the corresponding method of the pfs.APIServer interface
func (s *tracingServer) ListFileStream(*ListFileRequest, API_ListFileStreamServer) error {
	opentracing.start_TK
	opentracing.endTK()
	return s.APIClient.ListFileStream(*ListFileRequest, API_ListFileStreamServer)
}

// WalkFile implements the corresponding method of the pfs.APIServer interface
func (s *tracingServer) WalkFile(*WalkFileRequest, API_WalkFileServer) error {
	opentracing.start_TK
	opentracing.endTK()
	return s.APIClient.WalkFile(*WalkFileRequest, API_WalkFileServer)
}

// GlobFile implements the corresponding method of the pfs.APIServer interface
func (s *tracingServer) GlobFile(context.Context, *GlobFileRequest) (*FileInfos, error) {
	opentracing.start_TK
	opentracing.endTK()
	return s.APIClient.GlobFile(context.Context, *GlobFileRequest)
}

// GlobFileStream implements the corresponding method of the pfs.APIServer interface
func (s *tracingServer) GlobFileStream(*GlobFileRequest, API_GlobFileStreamServer) error {
	opentracing.start_TK
	opentracing.endTK()
	return s.APIClient.GlobFileStream(*GlobFileRequest, API_GlobFileStreamServer)
}

// DiffFile implements the corresponding method of the pfs.APIServer interface
func (s *tracingServer) DiffFile(context.Context, *DiffFileRequest) (*DiffFileResponse, error) {
	opentracing.start_TK
	opentracing.endTK()
	return s.APIClient.DiffFile(context.Context, *DiffFileRequest)
}

// DeleteFile implements the corresponding method of the pfs.APIServer interface
func (s *tracingServer) DeleteFile(context.Context, *DeleteFileRequest) (*types.Empty, error) {
	opentracing.start_TK
	opentracing.endTK()
	return s.APIClient.DeleteFile(context.Context, *DeleteFileRequest)
}

// DeleteAll implements the corresponding method of the pfs.APIServer interface
func (s *tracingServer) DeleteAll(context.Context, *types.Empty) (*types.Empty, error) {
	opentracing.start_TK
	opentracing.endTK()
	return s.APIClient.DeleteAll(context.Context, *types.Empty)
}
