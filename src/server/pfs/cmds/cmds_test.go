package cmds_test

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/v2/src/admin"
	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/debug"
	"github.com/pachyderm/pachyderm/v2/src/enterprise"
	"github.com/pachyderm/pachyderm/v2/src/identity"
	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/license"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	pachctl "github.com/pachyderm/pachyderm/v2/src/server/cmd/pachctl/cmd"
	"github.com/pachyderm/pachyderm/v2/src/transaction"
	"github.com/pachyderm/pachyderm/v2/src/version/versionpb"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

/*
func TestCommit(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	require.NoError(t, tu.BashCmd(`
		pachctl create repo {{.repo}}

		# Create a commit and put some data in it
		commit1=$(pachctl start commit {{.repo}}@master)
		echo "file contents" | pachctl put file {{.repo}}@${commit1}:/file -f -
		pachctl finish commit {{.repo}}@${commit1}

		# Check that the commit above now appears in the output
		pachctl list commit {{.repo}} \
		  | match ${commit1}

		# Create a second commit and put some data in it
		commit2=$(pachctl start commit {{.repo}}@master)
		echo "file contents" | pachctl put file {{.repo}}@${commit2}:/file -f -
		pachctl finish commit {{.repo}}@${commit2}

		# Check that the commit above now appears in the output
		pachctl list commit {{.repo}} \
		  | match ${commit1} \
		  | match ${commit2}
		`,
		"repo", tu.UniqueString("TestCommit-repo"),
	).Run())
}

func TestPutFileSplit(t *testing.T) {
	// TODO: Need to implement put file split in V2.
	t.Skip("Put file split not implemented in V2")
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	require.NoError(t, tu.BashCmd(`
		pachctl create repo {{.repo}}

		pachctl put file {{.repo}}@master:/data --split=csv --header-records=1 <<EOF
		name,job
		alice,accountant
		bob,baker
		EOF

		pachctl get file "{{.repo}}@master:/data/*0" \
		  | match "name,job"
		pachctl get file "{{.repo}}@master:/data/*0" \
		  | match "alice,accountant"
		pachctl get file "{{.repo}}@master:/data/*0" \
		  | match -v "bob,baker"

		pachctl get file "{{.repo}}@master:/data/*1" \
		  | match "name,job"
		pachctl get file "{{.repo}}@master:/data/*1" \
		  | match "bob,baker"
		pachctl get file "{{.repo}}@master:/data/*1" \
		  | match -v "alice,accountant"

		pachctl get file "{{.repo}}@master:/data/*" \
		  | match "name,job"
		pachctl get file "{{.repo}}@master:/data/*" \
		  | match "alice,accountant"
		pachctl get file "{{.repo}}@master:/data/*" \
		  | match "bob,baker"
		`,
		"repo", tu.UniqueString("TestPutFileSplit-repo"),
	).Run())
}

func TestMountParsing(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	expected := map[string]*fuse.RepoOptions{
		"repo1": {
			Branch: "branch",
			Write:  true,
		},
		"repo2": {
			Branch: "master",
			Write:  true,
		},
		"repo3": {
			Branch: "master",
		},
	}
	opts, err := parseRepoOpts([]string{"repo1@branch+w", "repo2+w", "repo3"})
	require.NoError(t, err)
	require.Equal(t, 3, len(opts))
	fmt.Printf("%+v\n", opts)
	for repo, ro := range expected {
		require.Equal(t, ro, opts[repo])
	}
}
*/

type TestEnv struct {
	MockClient           *MockClient
	MockEnterpriseClient *MockClient
	Stdin                io.Reader
	Stdout               io.Writer
	Stderr               io.Writer
}

func NewTestEnv(t *testing.T, stdin io.Reader, stdout io.Writer, stderr io.Writer) *TestEnv {
	return &TestEnv{
		MockClient:           NewMockClient(t),
		MockEnterpriseClient: NewMockClient(t),
		Stdin:                stdin,
		Stdout:               stdout,
		Stderr:               stderr,
	}
}

func (env *TestEnv) Client(string, ...client.Option) *client.APIClient {
	return env.MockClient.APIClient
}

func (env *TestEnv) EnterpriseClient(string, ...client.Option) *client.APIClient {
	return env.MockEnterpriseClient.APIClient
}

func (env *TestEnv) In() io.Reader {
	return env.Stdin
}

func (env *TestEnv) Out() io.Writer {
	return env.Stdout
}

func (env *TestEnv) Err() io.Writer {
	return env.Stderr
}

func (env *TestEnv) Close() error {
	return nil
}

type MockClient struct {
	PFS         *mockPFSClient
	PPS         *mockPPSClient
	Auth        *mockAuthClient
	Identity    *mockIdentityClient
	Version     *mockVersionClient
	Admin       *mockAdminClient
	Transaction *mockTransactionClient
	Debug       *mockDebugClient
	Enterprise  *mockEnterpriseClient
	License     *mockLicenseClient

	APIClient *client.APIClient
}

func NewMockClient(t *testing.T) *MockClient {
	mockClient := &MockClient{
		PFS:         &mockPFSClient{},
		PPS:         &mockPPSClient{},
		Auth:        &mockAuthClient{},
		Identity:    &mockIdentityClient{},
		Version:     &mockVersionClient{},
		Admin:       &mockAdminClient{},
		Transaction: &mockTransactionClient{},
		Debug:       &mockDebugClient{},
		Enterprise:  &mockEnterpriseClient{},
		License:     &mockLicenseClient{},
	}

	mockClient.PFS.Test(t)
	mockClient.PPS.Test(t)
	mockClient.Auth.Test(t)
	mockClient.Identity.Test(t)
	mockClient.Version.Test(t)
	mockClient.Admin.Test(t)
	mockClient.Transaction.Test(t)
	mockClient.Debug.Test(t)
	mockClient.Enterprise.Test(t)
	mockClient.License.Test(t)

	mockClient.APIClient = &client.APIClient{
		PfsAPIClient:         mockClient.PFS,
		PpsAPIClient:         mockClient.PPS,
		AuthAPIClient:        mockClient.Auth,
		IdentityAPIClient:    mockClient.Identity,
		VersionAPIClient:     mockClient.Version,
		AdminAPIClient:       mockClient.Admin,
		TransactionAPIClient: mockClient.Transaction,
		DebugClient:          mockClient.Debug,
		Enterprise:           mockClient.Enterprise,
		License:              mockClient.License,
	}
	return mockClient
}

type mockPFSClient struct {
	mock.Mock
}

func (m *mockPFSClient) CreateRepo(ctx context.Context, in *pfs.CreateRepoRequest, opts ...grpc.CallOption) (*types.Empty, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*types.Empty), args.Error(1)
}
func (m *mockPFSClient) InspectRepo(ctx context.Context, in *pfs.InspectRepoRequest, opts ...grpc.CallOption) (*pfs.RepoInfo, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*pfs.RepoInfo), args.Error(1)
}
func (m *mockPFSClient) ListRepo(ctx context.Context, in *pfs.ListRepoRequest, opts ...grpc.CallOption) (*pfs.ListRepoResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*pfs.ListRepoResponse), args.Error(1)
}
func (m *mockPFSClient) DeleteRepo(ctx context.Context, in *pfs.DeleteRepoRequest, opts ...grpc.CallOption) (*types.Empty, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*types.Empty), args.Error(1)
}
func (m *mockPFSClient) StartCommit(ctx context.Context, in *pfs.StartCommitRequest, opts ...grpc.CallOption) (*pfs.Commit, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*pfs.Commit), args.Error(1)
}
func (m *mockPFSClient) FinishCommit(ctx context.Context, in *pfs.FinishCommitRequest, opts ...grpc.CallOption) (*types.Empty, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*types.Empty), args.Error(1)
}
func (m *mockPFSClient) ClearCommit(ctx context.Context, in *pfs.ClearCommitRequest, opts ...grpc.CallOption) (*types.Empty, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*types.Empty), args.Error(1)
}
func (m *mockPFSClient) InspectCommit(ctx context.Context, in *pfs.InspectCommitRequest, opts ...grpc.CallOption) (*pfs.CommitInfo, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*pfs.CommitInfo), args.Error(1)
}
func (m *mockPFSClient) ListCommit(ctx context.Context, in *pfs.ListCommitRequest, opts ...grpc.CallOption) (pfs.API_ListCommitClient, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(pfs.API_ListCommitClient), args.Error(1)
}
func (m *mockPFSClient) SubscribeCommit(ctx context.Context, in *pfs.SubscribeCommitRequest, opts ...grpc.CallOption) (pfs.API_SubscribeCommitClient, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(pfs.API_SubscribeCommitClient), args.Error(1)
}
func (m *mockPFSClient) InspectCommitSet(ctx context.Context, in *pfs.InspectCommitSetRequest, opts ...grpc.CallOption) (pfs.API_InspectCommitSetClient, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(pfs.API_InspectCommitSetClient), args.Error(1)
}
func (m *mockPFSClient) SquashCommitSet(ctx context.Context, in *pfs.SquashCommitSetRequest, opts ...grpc.CallOption) (*types.Empty, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*types.Empty), args.Error(1)
}
func (m *mockPFSClient) CreateBranch(ctx context.Context, in *pfs.CreateBranchRequest, opts ...grpc.CallOption) (*types.Empty, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*types.Empty), args.Error(1)
}
func (m *mockPFSClient) InspectBranch(ctx context.Context, in *pfs.InspectBranchRequest, opts ...grpc.CallOption) (*pfs.BranchInfo, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*pfs.BranchInfo), args.Error(1)
}
func (m *mockPFSClient) ListBranch(ctx context.Context, in *pfs.ListBranchRequest, opts ...grpc.CallOption) (*pfs.BranchInfos, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*pfs.BranchInfos), args.Error(1)
}
func (m *mockPFSClient) DeleteBranch(ctx context.Context, in *pfs.DeleteBranchRequest, opts ...grpc.CallOption) (*types.Empty, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*types.Empty), args.Error(1)
}
func (m *mockPFSClient) ModifyFile(ctx context.Context, opts ...grpc.CallOption) (pfs.API_ModifyFileClient, error) {
	args := m.Called(ctx, opts)
	return args.Get(0).(pfs.API_ModifyFileClient), args.Error(1)
}
func (m *mockPFSClient) GetFileTAR(ctx context.Context, in *pfs.GetFileRequest, opts ...grpc.CallOption) (pfs.API_GetFileTARClient, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(pfs.API_GetFileTARClient), args.Error(1)
}
func (m *mockPFSClient) InspectFile(ctx context.Context, in *pfs.InspectFileRequest, opts ...grpc.CallOption) (*pfs.FileInfo, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*pfs.FileInfo), args.Error(1)
}
func (m *mockPFSClient) ListFile(ctx context.Context, in *pfs.ListFileRequest, opts ...grpc.CallOption) (pfs.API_ListFileClient, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(pfs.API_ListFileClient), args.Error(1)
}
func (m *mockPFSClient) WalkFile(ctx context.Context, in *pfs.WalkFileRequest, opts ...grpc.CallOption) (pfs.API_WalkFileClient, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(pfs.API_WalkFileClient), args.Error(1)
}
func (m *mockPFSClient) GlobFile(ctx context.Context, in *pfs.GlobFileRequest, opts ...grpc.CallOption) (pfs.API_GlobFileClient, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(pfs.API_GlobFileClient), args.Error(1)
}
func (m *mockPFSClient) DiffFile(ctx context.Context, in *pfs.DiffFileRequest, opts ...grpc.CallOption) (pfs.API_DiffFileClient, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(pfs.API_DiffFileClient), args.Error(1)
}
func (m *mockPFSClient) ActivateAuth(ctx context.Context, in *pfs.ActivateAuthRequest, opts ...grpc.CallOption) (*pfs.ActivateAuthResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*pfs.ActivateAuthResponse), args.Error(1)
}
func (m *mockPFSClient) DeleteAll(ctx context.Context, in *types.Empty, opts ...grpc.CallOption) (*types.Empty, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*types.Empty), args.Error(1)
}
func (m *mockPFSClient) Fsck(ctx context.Context, in *pfs.FsckRequest, opts ...grpc.CallOption) (pfs.API_FsckClient, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(pfs.API_FsckClient), args.Error(1)
}
func (m *mockPFSClient) CreateFileSet(ctx context.Context, opts ...grpc.CallOption) (pfs.API_CreateFileSetClient, error) {
	args := m.Called(ctx, opts)
	return args.Get(0).(pfs.API_CreateFileSetClient), args.Error(1)
}
func (m *mockPFSClient) GetFileSet(ctx context.Context, in *pfs.GetFileSetRequest, opts ...grpc.CallOption) (*pfs.CreateFileSetResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*pfs.CreateFileSetResponse), args.Error(1)
}
func (m *mockPFSClient) AddFileSet(ctx context.Context, in *pfs.AddFileSetRequest, opts ...grpc.CallOption) (*types.Empty, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*types.Empty), args.Error(1)
}
func (m *mockPFSClient) RenewFileSet(ctx context.Context, in *pfs.RenewFileSetRequest, opts ...grpc.CallOption) (*types.Empty, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*types.Empty), args.Error(1)
}
func (m *mockPFSClient) RunLoadTest(ctx context.Context, in *pfs.RunLoadTestRequest, opts ...grpc.CallOption) (*pfs.RunLoadTestResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*pfs.RunLoadTestResponse), args.Error(1)
}

type mockPPSClient struct {
	mock.Mock
}

func (m *mockPPSClient) InspectJob(ctx context.Context, in *pps.InspectJobRequest, opts ...grpc.CallOption) (*pps.JobInfo, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*pps.JobInfo), args.Error(1)
}
func (m *mockPPSClient) InspectJobSet(ctx context.Context, in *pps.InspectJobSetRequest, opts ...grpc.CallOption) (pps.API_InspectJobSetClient, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(pps.API_InspectJobSetClient), args.Error(1)
}
func (m *mockPPSClient) ListJob(ctx context.Context, in *pps.ListJobRequest, opts ...grpc.CallOption) (pps.API_ListJobClient, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(pps.API_ListJobClient), args.Error(1)
}
func (m *mockPPSClient) SubscribeJob(ctx context.Context, in *pps.SubscribeJobRequest, opts ...grpc.CallOption) (pps.API_SubscribeJobClient, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(pps.API_SubscribeJobClient), args.Error(1)
}
func (m *mockPPSClient) DeleteJob(ctx context.Context, in *pps.DeleteJobRequest, opts ...grpc.CallOption) (*types.Empty, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*types.Empty), args.Error(1)
}
func (m *mockPPSClient) StopJob(ctx context.Context, in *pps.StopJobRequest, opts ...grpc.CallOption) (*types.Empty, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*types.Empty), args.Error(1)
}
func (m *mockPPSClient) InspectDatum(ctx context.Context, in *pps.InspectDatumRequest, opts ...grpc.CallOption) (*pps.DatumInfo, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*pps.DatumInfo), args.Error(1)
}
func (m *mockPPSClient) ListDatum(ctx context.Context, in *pps.ListDatumRequest, opts ...grpc.CallOption) (pps.API_ListDatumClient, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(pps.API_ListDatumClient), args.Error(1)
}
func (m *mockPPSClient) RestartDatum(ctx context.Context, in *pps.RestartDatumRequest, opts ...grpc.CallOption) (*types.Empty, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*types.Empty), args.Error(1)
}
func (m *mockPPSClient) CreatePipeline(ctx context.Context, in *pps.CreatePipelineRequest, opts ...grpc.CallOption) (*types.Empty, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*types.Empty), args.Error(1)
}
func (m *mockPPSClient) InspectPipeline(ctx context.Context, in *pps.InspectPipelineRequest, opts ...grpc.CallOption) (*pps.PipelineInfo, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*pps.PipelineInfo), args.Error(1)
}
func (m *mockPPSClient) ListPipeline(ctx context.Context, in *pps.ListPipelineRequest, opts ...grpc.CallOption) (*pps.PipelineInfos, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*pps.PipelineInfos), args.Error(1)
}
func (m *mockPPSClient) DeletePipeline(ctx context.Context, in *pps.DeletePipelineRequest, opts ...grpc.CallOption) (*types.Empty, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*types.Empty), args.Error(1)
}
func (m *mockPPSClient) StartPipeline(ctx context.Context, in *pps.StartPipelineRequest, opts ...grpc.CallOption) (*types.Empty, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*types.Empty), args.Error(1)
}
func (m *mockPPSClient) StopPipeline(ctx context.Context, in *pps.StopPipelineRequest, opts ...grpc.CallOption) (*types.Empty, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*types.Empty), args.Error(1)
}
func (m *mockPPSClient) RunPipeline(ctx context.Context, in *pps.RunPipelineRequest, opts ...grpc.CallOption) (*types.Empty, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*types.Empty), args.Error(1)
}
func (m *mockPPSClient) RunCron(ctx context.Context, in *pps.RunCronRequest, opts ...grpc.CallOption) (*types.Empty, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*types.Empty), args.Error(1)
}
func (m *mockPPSClient) CreateSecret(ctx context.Context, in *pps.CreateSecretRequest, opts ...grpc.CallOption) (*types.Empty, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*types.Empty), args.Error(1)
}
func (m *mockPPSClient) DeleteSecret(ctx context.Context, in *pps.DeleteSecretRequest, opts ...grpc.CallOption) (*types.Empty, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*types.Empty), args.Error(1)
}
func (m *mockPPSClient) ListSecret(ctx context.Context, in *types.Empty, opts ...grpc.CallOption) (*pps.SecretInfos, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*pps.SecretInfos), args.Error(1)
}
func (m *mockPPSClient) InspectSecret(ctx context.Context, in *pps.InspectSecretRequest, opts ...grpc.CallOption) (*pps.SecretInfo, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*pps.SecretInfo), args.Error(1)
}
func (m *mockPPSClient) DeleteAll(ctx context.Context, in *types.Empty, opts ...grpc.CallOption) (*types.Empty, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*types.Empty), args.Error(1)
}
func (m *mockPPSClient) GetLogs(ctx context.Context, in *pps.GetLogsRequest, opts ...grpc.CallOption) (pps.API_GetLogsClient, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(pps.API_GetLogsClient), args.Error(1)
}
func (m *mockPPSClient) ActivateAuth(ctx context.Context, in *pps.ActivateAuthRequest, opts ...grpc.CallOption) (*pps.ActivateAuthResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*pps.ActivateAuthResponse), args.Error(1)
}
func (m *mockPPSClient) UpdateJobState(ctx context.Context, in *pps.UpdateJobStateRequest, opts ...grpc.CallOption) (*types.Empty, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*types.Empty), args.Error(1)
}

type mockAuthClient struct {
	mock.Mock
}

func (m *mockAuthClient) Activate(ctx context.Context, in *auth.ActivateRequest, opts ...grpc.CallOption) (*auth.ActivateResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*auth.ActivateResponse), args.Error(1)
}
func (m *mockAuthClient) Deactivate(ctx context.Context, in *auth.DeactivateRequest, opts ...grpc.CallOption) (*auth.DeactivateResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*auth.DeactivateResponse), args.Error(1)
}
func (m *mockAuthClient) GetConfiguration(ctx context.Context, in *auth.GetConfigurationRequest, opts ...grpc.CallOption) (*auth.GetConfigurationResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*auth.GetConfigurationResponse), args.Error(1)
}
func (m *mockAuthClient) SetConfiguration(ctx context.Context, in *auth.SetConfigurationRequest, opts ...grpc.CallOption) (*auth.SetConfigurationResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*auth.SetConfigurationResponse), args.Error(1)
}
func (m *mockAuthClient) Authenticate(ctx context.Context, in *auth.AuthenticateRequest, opts ...grpc.CallOption) (*auth.AuthenticateResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*auth.AuthenticateResponse), args.Error(1)
}
func (m *mockAuthClient) Authorize(ctx context.Context, in *auth.AuthorizeRequest, opts ...grpc.CallOption) (*auth.AuthorizeResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*auth.AuthorizeResponse), args.Error(1)
}
func (m *mockAuthClient) GetPermissions(ctx context.Context, in *auth.GetPermissionsRequest, opts ...grpc.CallOption) (*auth.GetPermissionsResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*auth.GetPermissionsResponse), args.Error(1)
}
func (m *mockAuthClient) GetPermissionsForPrincipal(ctx context.Context, in *auth.GetPermissionsForPrincipalRequest, opts ...grpc.CallOption) (*auth.GetPermissionsResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*auth.GetPermissionsResponse), args.Error(1)
}
func (m *mockAuthClient) WhoAmI(ctx context.Context, in *auth.WhoAmIRequest, opts ...grpc.CallOption) (*auth.WhoAmIResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*auth.WhoAmIResponse), args.Error(1)
}
func (m *mockAuthClient) ModifyRoleBinding(ctx context.Context, in *auth.ModifyRoleBindingRequest, opts ...grpc.CallOption) (*auth.ModifyRoleBindingResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*auth.ModifyRoleBindingResponse), args.Error(1)
}
func (m *mockAuthClient) GetRoleBinding(ctx context.Context, in *auth.GetRoleBindingRequest, opts ...grpc.CallOption) (*auth.GetRoleBindingResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*auth.GetRoleBindingResponse), args.Error(1)
}
func (m *mockAuthClient) GetOIDCLogin(ctx context.Context, in *auth.GetOIDCLoginRequest, opts ...grpc.CallOption) (*auth.GetOIDCLoginResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*auth.GetOIDCLoginResponse), args.Error(1)
}
func (m *mockAuthClient) GetRobotToken(ctx context.Context, in *auth.GetRobotTokenRequest, opts ...grpc.CallOption) (*auth.GetRobotTokenResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*auth.GetRobotTokenResponse), args.Error(1)
}
func (m *mockAuthClient) RevokeAuthToken(ctx context.Context, in *auth.RevokeAuthTokenRequest, opts ...grpc.CallOption) (*auth.RevokeAuthTokenResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*auth.RevokeAuthTokenResponse), args.Error(1)
}
func (m *mockAuthClient) RevokeAuthTokensForUser(ctx context.Context, in *auth.RevokeAuthTokensForUserRequest, opts ...grpc.CallOption) (*auth.RevokeAuthTokensForUserResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*auth.RevokeAuthTokensForUserResponse), args.Error(1)
}
func (m *mockAuthClient) SetGroupsForUser(ctx context.Context, in *auth.SetGroupsForUserRequest, opts ...grpc.CallOption) (*auth.SetGroupsForUserResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*auth.SetGroupsForUserResponse), args.Error(1)
}
func (m *mockAuthClient) ModifyMembers(ctx context.Context, in *auth.ModifyMembersRequest, opts ...grpc.CallOption) (*auth.ModifyMembersResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*auth.ModifyMembersResponse), args.Error(1)
}
func (m *mockAuthClient) GetGroups(ctx context.Context, in *auth.GetGroupsRequest, opts ...grpc.CallOption) (*auth.GetGroupsResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*auth.GetGroupsResponse), args.Error(1)
}
func (m *mockAuthClient) GetGroupsForPrincipal(ctx context.Context, in *auth.GetGroupsForPrincipalRequest, opts ...grpc.CallOption) (*auth.GetGroupsResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*auth.GetGroupsResponse), args.Error(1)
}
func (m *mockAuthClient) GetUsers(ctx context.Context, in *auth.GetUsersRequest, opts ...grpc.CallOption) (*auth.GetUsersResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*auth.GetUsersResponse), args.Error(1)
}
func (m *mockAuthClient) ExtractAuthTokens(ctx context.Context, in *auth.ExtractAuthTokensRequest, opts ...grpc.CallOption) (*auth.ExtractAuthTokensResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*auth.ExtractAuthTokensResponse), args.Error(1)
}
func (m *mockAuthClient) RestoreAuthToken(ctx context.Context, in *auth.RestoreAuthTokenRequest, opts ...grpc.CallOption) (*auth.RestoreAuthTokenResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*auth.RestoreAuthTokenResponse), args.Error(1)
}
func (m *mockAuthClient) DeleteExpiredAuthTokens(ctx context.Context, in *auth.DeleteExpiredAuthTokensRequest, opts ...grpc.CallOption) (*auth.DeleteExpiredAuthTokensResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*auth.DeleteExpiredAuthTokensResponse), args.Error(1)
}
func (m *mockAuthClient) RotateRootToken(ctx context.Context, in *auth.RotateRootTokenRequest, opts ...grpc.CallOption) (*auth.RotateRootTokenResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*auth.RotateRootTokenResponse), args.Error(1)
}

type mockIdentityClient struct {
	mock.Mock
}

func (m *mockIdentityClient) SetIdentityServerConfig(ctx context.Context, in *identity.SetIdentityServerConfigRequest, opts ...grpc.CallOption) (*identity.SetIdentityServerConfigResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*identity.SetIdentityServerConfigResponse), args.Error(1)
}
func (m *mockIdentityClient) GetIdentityServerConfig(ctx context.Context, in *identity.GetIdentityServerConfigRequest, opts ...grpc.CallOption) (*identity.GetIdentityServerConfigResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*identity.GetIdentityServerConfigResponse), args.Error(1)
}
func (m *mockIdentityClient) CreateIDPConnector(ctx context.Context, in *identity.CreateIDPConnectorRequest, opts ...grpc.CallOption) (*identity.CreateIDPConnectorResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*identity.CreateIDPConnectorResponse), args.Error(1)
}
func (m *mockIdentityClient) UpdateIDPConnector(ctx context.Context, in *identity.UpdateIDPConnectorRequest, opts ...grpc.CallOption) (*identity.UpdateIDPConnectorResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*identity.UpdateIDPConnectorResponse), args.Error(1)
}
func (m *mockIdentityClient) ListIDPConnectors(ctx context.Context, in *identity.ListIDPConnectorsRequest, opts ...grpc.CallOption) (*identity.ListIDPConnectorsResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*identity.ListIDPConnectorsResponse), args.Error(1)
}
func (m *mockIdentityClient) GetIDPConnector(ctx context.Context, in *identity.GetIDPConnectorRequest, opts ...grpc.CallOption) (*identity.GetIDPConnectorResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*identity.GetIDPConnectorResponse), args.Error(1)
}
func (m *mockIdentityClient) DeleteIDPConnector(ctx context.Context, in *identity.DeleteIDPConnectorRequest, opts ...grpc.CallOption) (*identity.DeleteIDPConnectorResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*identity.DeleteIDPConnectorResponse), args.Error(1)
}
func (m *mockIdentityClient) CreateOIDCClient(ctx context.Context, in *identity.CreateOIDCClientRequest, opts ...grpc.CallOption) (*identity.CreateOIDCClientResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*identity.CreateOIDCClientResponse), args.Error(1)
}
func (m *mockIdentityClient) UpdateOIDCClient(ctx context.Context, in *identity.UpdateOIDCClientRequest, opts ...grpc.CallOption) (*identity.UpdateOIDCClientResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*identity.UpdateOIDCClientResponse), args.Error(1)
}
func (m *mockIdentityClient) GetOIDCClient(ctx context.Context, in *identity.GetOIDCClientRequest, opts ...grpc.CallOption) (*identity.GetOIDCClientResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*identity.GetOIDCClientResponse), args.Error(1)
}
func (m *mockIdentityClient) ListOIDCClients(ctx context.Context, in *identity.ListOIDCClientsRequest, opts ...grpc.CallOption) (*identity.ListOIDCClientsResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*identity.ListOIDCClientsResponse), args.Error(1)
}
func (m *mockIdentityClient) DeleteOIDCClient(ctx context.Context, in *identity.DeleteOIDCClientRequest, opts ...grpc.CallOption) (*identity.DeleteOIDCClientResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*identity.DeleteOIDCClientResponse), args.Error(1)
}
func (m *mockIdentityClient) DeleteAll(ctx context.Context, in *identity.DeleteAllRequest, opts ...grpc.CallOption) (*identity.DeleteAllResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*identity.DeleteAllResponse), args.Error(1)
}

type mockVersionClient struct {
	mock.Mock
}

func (m *mockVersionClient) GetVersion(ctx context.Context, in *types.Empty, opts ...grpc.CallOption) (*versionpb.Version, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*versionpb.Version), args.Error(1)
}

type mockAdminClient struct {
	mock.Mock
}

func (m *mockAdminClient) InspectCluster(ctx context.Context, in *types.Empty, opts ...grpc.CallOption) (*admin.ClusterInfo, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*admin.ClusterInfo), args.Error(1)
}

type mockTransactionClient struct {
	mock.Mock
}

func (m *mockTransactionClient) BatchTransaction(ctx context.Context, in *transaction.BatchTransactionRequest, opts ...grpc.CallOption) (*transaction.TransactionInfo, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*transaction.TransactionInfo), args.Error(1)
}
func (m *mockTransactionClient) StartTransaction(ctx context.Context, in *transaction.StartTransactionRequest, opts ...grpc.CallOption) (*transaction.Transaction, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*transaction.Transaction), args.Error(1)
}
func (m *mockTransactionClient) InspectTransaction(ctx context.Context, in *transaction.InspectTransactionRequest, opts ...grpc.CallOption) (*transaction.TransactionInfo, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*transaction.TransactionInfo), args.Error(1)
}
func (m *mockTransactionClient) DeleteTransaction(ctx context.Context, in *transaction.DeleteTransactionRequest, opts ...grpc.CallOption) (*types.Empty, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*types.Empty), args.Error(1)
}
func (m *mockTransactionClient) ListTransaction(ctx context.Context, in *transaction.ListTransactionRequest, opts ...grpc.CallOption) (*transaction.TransactionInfos, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*transaction.TransactionInfos), args.Error(1)
}
func (m *mockTransactionClient) FinishTransaction(ctx context.Context, in *transaction.FinishTransactionRequest, opts ...grpc.CallOption) (*transaction.TransactionInfo, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*transaction.TransactionInfo), args.Error(1)
}
func (m *mockTransactionClient) DeleteAll(ctx context.Context, in *transaction.DeleteAllRequest, opts ...grpc.CallOption) (*types.Empty, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*types.Empty), args.Error(1)
}

type mockDebugClient struct {
	mock.Mock
}

func (m *mockDebugClient) Profile(ctx context.Context, in *debug.ProfileRequest, opts ...grpc.CallOption) (debug.Debug_ProfileClient, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(debug.Debug_ProfileClient), args.Error(1)
}
func (m *mockDebugClient) Binary(ctx context.Context, in *debug.BinaryRequest, opts ...grpc.CallOption) (debug.Debug_BinaryClient, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(debug.Debug_BinaryClient), args.Error(1)
}
func (m *mockDebugClient) Dump(ctx context.Context, in *debug.DumpRequest, opts ...grpc.CallOption) (debug.Debug_DumpClient, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(debug.Debug_DumpClient), args.Error(1)
}

type mockEnterpriseClient struct {
	mock.Mock
}

func (m *mockEnterpriseClient) Activate(ctx context.Context, in *enterprise.ActivateRequest, opts ...grpc.CallOption) (*enterprise.ActivateResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*enterprise.ActivateResponse), args.Error(1)
}
func (m *mockEnterpriseClient) GetState(ctx context.Context, in *enterprise.GetStateRequest, opts ...grpc.CallOption) (*enterprise.GetStateResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*enterprise.GetStateResponse), args.Error(1)
}
func (m *mockEnterpriseClient) GetActivationCode(ctx context.Context, in *enterprise.GetActivationCodeRequest, opts ...grpc.CallOption) (*enterprise.GetActivationCodeResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*enterprise.GetActivationCodeResponse), args.Error(1)
}
func (m *mockEnterpriseClient) Heartbeat(ctx context.Context, in *enterprise.HeartbeatRequest, opts ...grpc.CallOption) (*enterprise.HeartbeatResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*enterprise.HeartbeatResponse), args.Error(1)
}
func (m *mockEnterpriseClient) Deactivate(ctx context.Context, in *enterprise.DeactivateRequest, opts ...grpc.CallOption) (*enterprise.DeactivateResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*enterprise.DeactivateResponse), args.Error(1)
}

type mockLicenseClient struct {
	mock.Mock
}

func (m *mockLicenseClient) Activate(ctx context.Context, in *license.ActivateRequest, opts ...grpc.CallOption) (*license.ActivateResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*license.ActivateResponse), args.Error(1)
}
func (m *mockLicenseClient) GetActivationCode(ctx context.Context, in *license.GetActivationCodeRequest, opts ...grpc.CallOption) (*license.GetActivationCodeResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*license.GetActivationCodeResponse), args.Error(1)
}
func (m *mockLicenseClient) DeleteAll(ctx context.Context, in *license.DeleteAllRequest, opts ...grpc.CallOption) (*license.DeleteAllResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*license.DeleteAllResponse), args.Error(1)
}
func (m *mockLicenseClient) AddCluster(ctx context.Context, in *license.AddClusterRequest, opts ...grpc.CallOption) (*license.AddClusterResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*license.AddClusterResponse), args.Error(1)
}
func (m *mockLicenseClient) DeleteCluster(ctx context.Context, in *license.DeleteClusterRequest, opts ...grpc.CallOption) (*license.DeleteClusterResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*license.DeleteClusterResponse), args.Error(1)
}
func (m *mockLicenseClient) ListClusters(ctx context.Context, in *license.ListClustersRequest, opts ...grpc.CallOption) (*license.ListClustersResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*license.ListClustersResponse), args.Error(1)
}
func (m *mockLicenseClient) UpdateCluster(ctx context.Context, in *license.UpdateClusterRequest, opts ...grpc.CallOption) (*license.UpdateClusterResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*license.UpdateClusterResponse), args.Error(1)
}
func (m *mockLicenseClient) Heartbeat(ctx context.Context, in *license.HeartbeatRequest, opts ...grpc.CallOption) (*license.HeartbeatResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*license.HeartbeatResponse), args.Error(1)
}
func (m *mockLicenseClient) ListUserClusters(ctx context.Context, in *license.ListUserClustersRequest, opts ...grpc.CallOption) (*license.ListUserClustersResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*license.ListUserClustersResponse), args.Error(1)
}

func RunPachctlSync(t *testing.T, env cmdutil.Env, args ...string) error {
	ctx := context.WithValue(context.Background(), "env", env)
	root := pachctl.PachctlCmd()
	root.SetArgs(args)
	return root.ExecuteContext(ctx)
}

func TestInspectRepo(t *testing.T) {
	env := NewTestEnv(t, &bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{})

	env.MockClient.PFS.On("InspectRepo", mock.Anything, mock.Anything, mock.Anything).Return(&pfs.RepoInfo{}, nil)

	err := RunPachctlSync(t, env, "inspect", "repo", "foo")
	require.NoError(t, err)
}
