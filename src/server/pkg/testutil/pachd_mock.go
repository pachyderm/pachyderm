package testutil

import (
	"context"
	"fmt"

	"github.com/gogo/protobuf/types"

	"github.com/pachyderm/pachyderm/src/client/admin"
	"github.com/pachyderm/pachyderm/src/client/auth"
	"github.com/pachyderm/pachyderm/src/client/enterprise"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/client/transaction"
	version "github.com/pachyderm/pachyderm/src/client/version/versionpb"

	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

type mockAdminServer struct{}

func (mock *mockAdminServer) Extract(*admin.ExtractRequest, admin.API_ExtractServer) error {
	return fmt.Errorf("Mock")
}
func (mock *mockAdminServer) ExtractPipeline(context.Context, *admin.ExtractPipelineRequest) (*admin.Op, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockAdminServer) Restore(admin.API_RestoreServer) error {
	return fmt.Errorf("Mock")
}
func (mock *mockAdminServer) InspectCluster(context.Context, *types.Empty) (*admin.ClusterInfo, error) {
	return nil, fmt.Errorf("Mock")
}

type mockAuthServer struct{}

func (mock *mockAuthServer) Activate(context.Context, *auth.ActivateRequest) (*auth.ActivateResponse, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockAuthServer) Deactivate(context.Context, *auth.DeactivateRequest) (*auth.DeactivateResponse, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockAuthServer) GetConfiguration(context.Context, *auth.GetConfigurationRequest) (*auth.GetConfigurationResponse, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockAuthServer) SetConfiguration(context.Context, *auth.SetConfigurationRequest) (*auth.SetConfigurationResponse, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockAuthServer) GetAdmins(context.Context, *auth.GetAdminsRequest) (*auth.GetAdminsResponse, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockAuthServer) ModifyAdmins(context.Context, *auth.ModifyAdminsRequest) (*auth.ModifyAdminsResponse, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockAuthServer) Authenticate(context.Context, *auth.AuthenticateRequest) (*auth.AuthenticateResponse, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockAuthServer) Authorize(context.Context, *auth.AuthorizeRequest) (*auth.AuthorizeResponse, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockAuthServer) WhoAmI(context.Context, *auth.WhoAmIRequest) (*auth.WhoAmIResponse, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockAuthServer) GetScope(context.Context, *auth.GetScopeRequest) (*auth.GetScopeResponse, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockAuthServer) SetScope(context.Context, *auth.SetScopeRequest) (*auth.SetScopeResponse, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockAuthServer) GetACL(context.Context, *auth.GetACLRequest) (*auth.GetACLResponse, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockAuthServer) SetACL(context.Context, *auth.SetACLRequest) (*auth.SetACLResponse, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockAuthServer) GetAuthToken(context.Context, *auth.GetAuthTokenRequest) (*auth.GetAuthTokenResponse, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockAuthServer) ExtendAuthToken(context.Context, *auth.ExtendAuthTokenRequest) (*auth.ExtendAuthTokenResponse, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockAuthServer) RevokeAuthToken(context.Context, *auth.RevokeAuthTokenRequest) (*auth.RevokeAuthTokenResponse, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockAuthServer) SetGroupsForUser(context.Context, *auth.SetGroupsForUserRequest) (*auth.SetGroupsForUserResponse, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockAuthServer) ModifyMembers(context.Context, *auth.ModifyMembersRequest) (*auth.ModifyMembersResponse, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockAuthServer) GetGroups(context.Context, *auth.GetGroupsRequest) (*auth.GetGroupsResponse, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockAuthServer) GetUsers(context.Context, *auth.GetUsersRequest) (*auth.GetUsersResponse, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockAuthServer) GetOneTimePassword(context.Context, *auth.GetOneTimePasswordRequest) (*auth.GetOneTimePasswordResponse, error) {
	return nil, fmt.Errorf("Mock")
}

type mockEnterpriseServer struct{}

func (mock *mockEnterpriseServer) Activate(context.Context, *enterprise.ActivateRequest) (*enterprise.ActivateResponse, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockEnterpriseServer) GetState(context.Context, *enterprise.GetStateRequest) (*enterprise.GetStateResponse, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockEnterpriseServer) Deactivate(context.Context, *enterprise.DeactivateRequest) (*enterprise.DeactivateResponse, error) {
	return nil, fmt.Errorf("Mock")
}

type mockPfsServer struct{}

func (mock *mockPfsServer) CreateRepo(context.Context, *pfs.CreateRepoRequest) (*types.Empty, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockPfsServer) InspectRepo(context.Context, *pfs.InspectRepoRequest) (*pfs.RepoInfo, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockPfsServer) ListRepo(context.Context, *pfs.ListRepoRequest) (*pfs.ListRepoResponse, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockPfsServer) DeleteRepo(context.Context, *pfs.DeleteRepoRequest) (*types.Empty, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockPfsServer) StartCommit(context.Context, *pfs.StartCommitRequest) (*pfs.Commit, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockPfsServer) FinishCommit(context.Context, *pfs.FinishCommitRequest) (*types.Empty, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockPfsServer) InspectCommit(context.Context, *pfs.InspectCommitRequest) (*pfs.CommitInfo, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockPfsServer) ListCommit(context.Context, *pfs.ListCommitRequest) (*pfs.CommitInfos, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockPfsServer) ListCommitStream(*pfs.ListCommitRequest, pfs.API_ListCommitStreamServer) error {
	return fmt.Errorf("Mock")
}
func (mock *mockPfsServer) DeleteCommit(context.Context, *pfs.DeleteCommitRequest) (*types.Empty, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockPfsServer) FlushCommit(*pfs.FlushCommitRequest, pfs.API_FlushCommitServer) error {
	return fmt.Errorf("Mock")
}
func (mock *mockPfsServer) SubscribeCommit(*pfs.SubscribeCommitRequest, pfs.API_SubscribeCommitServer) error {
	return fmt.Errorf("Mock")
}
func (mock *mockPfsServer) BuildCommit(context.Context, *pfs.BuildCommitRequest) (*pfs.Commit, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockPfsServer) CreateBranch(context.Context, *pfs.CreateBranchRequest) (*types.Empty, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockPfsServer) InspectBranch(context.Context, *pfs.InspectBranchRequest) (*pfs.BranchInfo, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockPfsServer) ListBranch(context.Context, *pfs.ListBranchRequest) (*pfs.BranchInfos, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockPfsServer) DeleteBranch(context.Context, *pfs.DeleteBranchRequest) (*types.Empty, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockPfsServer) PutFile(pfs.API_PutFileServer) error {
	return fmt.Errorf("Mock")
}
func (mock *mockPfsServer) CopyFile(context.Context, *pfs.CopyFileRequest) (*types.Empty, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockPfsServer) GetFile(*pfs.GetFileRequest, pfs.API_GetFileServer) error {
	return fmt.Errorf("Mock")
}
func (mock *mockPfsServer) InspectFile(context.Context, *pfs.InspectFileRequest) (*pfs.FileInfo, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockPfsServer) ListFile(context.Context, *pfs.ListFileRequest) (*pfs.FileInfos, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockPfsServer) ListFileStream(*pfs.ListFileRequest, pfs.API_ListFileStreamServer) error {
	return fmt.Errorf("Mock")
}
func (mock *mockPfsServer) WalkFile(*pfs.WalkFileRequest, pfs.API_WalkFileServer) error {
	return fmt.Errorf("Mock")
}
func (mock *mockPfsServer) GlobFile(context.Context, *pfs.GlobFileRequest) (*pfs.FileInfos, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockPfsServer) GlobFileStream(*pfs.GlobFileRequest, pfs.API_GlobFileStreamServer) error {
	return fmt.Errorf("Mock")
}
func (mock *mockPfsServer) DiffFile(context.Context, *pfs.DiffFileRequest) (*pfs.DiffFileResponse, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockPfsServer) DeleteFile(context.Context, *pfs.DeleteFileRequest) (*types.Empty, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockPfsServer) DeleteAll(context.Context, *types.Empty) (*types.Empty, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockPfsServer) Fsck(*pfs.FsckRequest, pfs.API_FsckServer) error {
	return fmt.Errorf("Mock")
}

type mockPpsServer struct{}

func (mock *mockPpsServer) CreateJob(context.Context, *pps.CreateJobRequest) (*pps.Job, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockPpsServer) InspectJob(context.Context, *pps.InspectJobRequest) (*pps.JobInfo, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockPpsServer) ListJob(context.Context, *pps.ListJobRequest) (*pps.JobInfos, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockPpsServer) ListJobStream(*pps.ListJobRequest, pps.API_ListJobStreamServer) error {
	return fmt.Errorf("Mock")
}
func (mock *mockPpsServer) FlushJob(*pps.FlushJobRequest, pps.API_FlushJobServer) error {
	return fmt.Errorf("Mock")
}
func (mock *mockPpsServer) DeleteJob(context.Context, *pps.DeleteJobRequest) (*types.Empty, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockPpsServer) StopJob(context.Context, *pps.StopJobRequest) (*types.Empty, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockPpsServer) InspectDatum(context.Context, *pps.InspectDatumRequest) (*pps.DatumInfo, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockPpsServer) ListDatum(context.Context, *pps.ListDatumRequest) (*pps.ListDatumResponse, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockPpsServer) ListDatumStream(*pps.ListDatumRequest, pps.API_ListDatumStreamServer) error {
	return fmt.Errorf("Mock")
}
func (mock *mockPpsServer) RestartDatum(context.Context, *pps.RestartDatumRequest) (*types.Empty, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockPpsServer) CreatePipeline(context.Context, *pps.CreatePipelineRequest) (*types.Empty, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockPpsServer) InspectPipeline(context.Context, *pps.InspectPipelineRequest) (*pps.PipelineInfo, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockPpsServer) ListPipeline(context.Context, *pps.ListPipelineRequest) (*pps.PipelineInfos, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockPpsServer) DeletePipeline(context.Context, *pps.DeletePipelineRequest) (*types.Empty, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockPpsServer) StartPipeline(context.Context, *pps.StartPipelineRequest) (*types.Empty, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockPpsServer) StopPipeline(context.Context, *pps.StopPipelineRequest) (*types.Empty, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockPpsServer) RunPipeline(context.Context, *pps.RunPipelineRequest) (*types.Empty, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockPpsServer) DeleteAll(context.Context, *types.Empty) (*types.Empty, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockPpsServer) GetLogs(*pps.GetLogsRequest, pps.API_GetLogsServer) error {
	return fmt.Errorf("Mock")
}
func (mock *mockPpsServer) GarbageCollect(context.Context, *pps.GarbageCollectRequest) (*pps.GarbageCollectResponse, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockPpsServer) ActivateAuth(context.Context, *pps.ActivateAuthRequest) (*pps.ActivateAuthResponse, error) {
	return nil, fmt.Errorf("Mock")
}

type mockTransactionServer struct{}

func (mock *mockTransactionServer) StartTransaction(context.Context, *transaction.StartTransactionRequest) (*transaction.Transaction, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockTransactionServer) InspectTransaction(context.Context, *transaction.InspectTransactionRequest) (*transaction.TransactionInfo, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockTransactionServer) DeleteTransaction(context.Context, *transaction.DeleteTransactionRequest) (*types.Empty, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockTransactionServer) ListTransaction(context.Context, *transaction.ListTransactionRequest) (*transaction.TransactionInfos, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockTransactionServer) FinishTransaction(context.Context, *transaction.FinishTransactionRequest) (*transaction.TransactionInfo, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockTransactionServer) DeleteAll(context.Context, *transaction.DeleteAllRequest) (*types.Empty, error) {
	return nil, fmt.Errorf("Mock")
}

type mockVersionServer struct{}

func (mock *mockVersionServer) GetVersion(context.Context, *types.Empty) (*version.Version, error) {
	return nil, fmt.Errorf("Mock")
}

type mockObjectServer struct{}

func (mock *mockObjectServer) PutObject(pfs.ObjectAPI_PutObjectServer) error {
	return fmt.Errorf("Mock")
}
func (mock *mockObjectServer) PutObjectSplit(pfs.ObjectAPI_PutObjectSplitServer) error {
	return fmt.Errorf("Mock")
}
func (mock *mockObjectServer) PutObjects(pfs.ObjectAPI_PutObjectsServer) error {
	return fmt.Errorf("Mock")
}
func (mock *mockObjectServer) CreateObject(context.Context, *pfs.CreateObjectRequest) (*types.Empty, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockObjectServer) GetObject(*pfs.Object, pfs.ObjectAPI_GetObjectServer) error {
	return fmt.Errorf("Mock")
}
func (mock *mockObjectServer) GetObjects(*pfs.GetObjectsRequest, pfs.ObjectAPI_GetObjectsServer) error {
	return fmt.Errorf("Mock")
}
func (mock *mockObjectServer) PutBlock(pfs.ObjectAPI_PutBlockServer) error {
	return fmt.Errorf("Mock")
}
func (mock *mockObjectServer) GetBlock(*pfs.GetBlockRequest, pfs.ObjectAPI_GetBlockServer) error {
	return fmt.Errorf("Mock")
}
func (mock *mockObjectServer) GetBlocks(*pfs.GetBlocksRequest, pfs.ObjectAPI_GetBlocksServer) error {
	return fmt.Errorf("Mock")
}
func (mock *mockObjectServer) ListBlock(*pfs.ListBlockRequest, pfs.ObjectAPI_ListBlockServer) error {
	return fmt.Errorf("Mock")
}
func (mock *mockObjectServer) TagObject(context.Context, *pfs.TagObjectRequest) (*types.Empty, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockObjectServer) InspectObject(context.Context, *pfs.Object) (*pfs.ObjectInfo, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockObjectServer) CheckObject(context.Context, *pfs.CheckObjectRequest) (*pfs.CheckObjectResponse, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockObjectServer) ListObjects(*pfs.ListObjectsRequest, pfs.ObjectAPI_ListObjectsServer) error {
	return fmt.Errorf("Mock")
}
func (mock *mockObjectServer) DeleteObjects(context.Context, *pfs.DeleteObjectsRequest) (*pfs.DeleteObjectsResponse, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockObjectServer) GetTag(*pfs.Tag, pfs.ObjectAPI_GetTagServer) error {
	return fmt.Errorf("Mock")
}
func (mock *mockObjectServer) InspectTag(context.Context, *pfs.Tag) (*pfs.ObjectInfo, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockObjectServer) ListTags(*pfs.ListTagsRequest, pfs.ObjectAPI_ListTagsServer) error {
	return fmt.Errorf("Mock")
}
func (mock *mockObjectServer) DeleteTags(context.Context, *pfs.DeleteTagsRequest) (*pfs.DeleteTagsResponse, error) {
	return nil, fmt.Errorf("Mock")
}
func (mock *mockObjectServer) Compact(context.Context, *types.Empty) (*types.Empty, error) {
	return nil, fmt.Errorf("Mock")
}

// PachdMock provides an interface for running the interface for a Pachd API
// server locally without any of its dependencies. Tests may mock out specific
// API calls by providing a handler function, and later check information about
// the mocked calls.
type PachdMock struct {
	cancel context.CancelFunc
	eg     *errgroup.Group

	ObjectAPI      mockObjectServer
	PfsAPI         mockPfsServer
	PpsAPI         mockPpsServer
	AuthAPI        mockAuthServer
	TransactionAPI mockTransactionServer
	EnterpriseAPI  mockEnterpriseServer
	VersionAPI     mockVersionServer
	AdminAPI       mockAdminServer
}

// NewPachdMock constructs a mock Pachd API server whose behavior can be
// controlled through the PachdMock instance. By default, all API calls will
// error, unless a handler is specified.
func NewPachdMock(port uint16) *PachdMock {
	mock := &PachdMock{}

	ctx := context.Background()
	ctx, mock.cancel = context.WithCancel(ctx)
	mock.eg, ctx = errgroup.WithContext(ctx)

	mock.eg.Go(func() error {
		err := grpcutil.Serve(
			grpcutil.ServerOptions{
				Port:       port,
				MaxMsgSize: grpcutil.MaxMsgSize,
				RegisterFunc: func(s *grpc.Server) error {
					admin.RegisterAPIServer(s, &mock.AdminAPI)
					auth.RegisterAPIServer(s, &mock.AuthAPI)
					enterprise.RegisterAPIServer(s, &mock.EnterpriseAPI)
					pfs.RegisterObjectAPIServer(s, &mock.ObjectAPI)
					pfs.RegisterAPIServer(s, &mock.PfsAPI)
					pps.RegisterAPIServer(s, &mock.PpsAPI)
					transaction.RegisterAPIServer(s, &mock.TransactionAPI)
					version.RegisterAPIServer(s, &mock.VersionAPI)
					return nil
				},
			},
		)
		if err != nil {
			log.Printf("error starting grpc server %v\n", err)
		}
		return err
	})

	return mock
}

// Close will cancel the mock Pachd API server goroutine and return its result
func (mock *PachdMock) Close() error {
	mock.cancel()
	return mock.eg.Wait()
}
