package client

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/admin"
	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/debug"
	"github.com/pachyderm/pachyderm/v2/src/enterprise"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"github.com/pachyderm/pachyderm/v2/src/transaction"
	"github.com/pachyderm/pachyderm/v2/src/version/versionpb"

	types "github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const transactionMetadataKey = "pach-transaction"

// WithTransaction (client-side) returns a new APIClient that will run supported
// write operations within the specified transaction.
func (c APIClient) WithTransaction(txn *transaction.Transaction) *APIClient {
	md, _ := metadata.FromOutgoingContext(c.Ctx())
	md = md.Copy()
	if txn != nil {
		md.Set(transactionMetadataKey, txn.ID)
	} else {
		md.Set(transactionMetadataKey)
	}
	ctx := metadata.NewOutgoingContext(c.Ctx(), md)
	return c.WithCtx(ctx)
}

// WithoutTransaction returns a new APIClient which will run all future operations
// outside of any active transaction
// Removing from both incoming and outgoing metadata is necessary because Ctx() merges them
func (c APIClient) WithoutTransaction() *APIClient {
	ctx := c.Ctx()
	incomingMD, _ := metadata.FromIncomingContext(ctx)
	outgoingMD, _ := metadata.FromOutgoingContext(ctx)
	newIn := make(metadata.MD)
	newOut := make(metadata.MD)
	for k, v := range incomingMD {
		if k == transactionMetadataKey {
			continue
		}
		newIn[k] = v
	}
	for k, v := range outgoingMD {
		if k == transactionMetadataKey {
			continue
		}
		newOut[k] = v
	}
	return c.WithCtx(metadata.NewIncomingContext(metadata.NewOutgoingContext(ctx, newOut), newIn))
}

// GetTransaction (should be run from the server-side) loads the active
// transaction from the grpc metadata and returns the associated transaction
// object - or `nil` if no transaction is set.
func GetTransaction(ctx context.Context) (*transaction.Transaction, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errors.Errorf("request metadata could not be parsed from context")
	}

	txns := md.Get(transactionMetadataKey)
	if len(txns) == 0 {
		return nil, nil
	} else if len(txns) > 1 {
		return nil, errors.Errorf("multiple active transactions found in context")
	}
	return &transaction.Transaction{ID: txns[0]}, nil
}

// GetTransaction is a helper function to get the active transaction from the
// client's context metadata.
func (c APIClient) GetTransaction() (*transaction.Transaction, error) {
	return GetTransaction(c.Ctx())
}

// NewCommitResponse is a helper function to instantiate a TransactionResponse
// for a transaction item that returns a Commit ID.
func NewCommitResponse(commit *pfs.Commit) *transaction.TransactionResponse {
	return &transaction.TransactionResponse{
		Commit: commit,
	}
}

// ListTransaction is an RPC that fetches a list of all open transactions in the
// Pachyderm cluster.
func (c APIClient) ListTransaction() ([]*transaction.TransactionInfo, error) {
	response, err := c.TransactionAPIClient.ListTransaction(
		c.Ctx(),
		&transaction.ListTransactionRequest{},
	)
	if err != nil {
		return nil, grpcutil.ScrubGRPC(err)
	}
	return response.TransactionInfo, nil
}

// StartTransaction is an RPC that registers a new transaction with the
// Pachyderm cluster and returns the identifier of the new transaction.
func (c APIClient) StartTransaction() (*transaction.Transaction, error) {
	response, err := c.TransactionAPIClient.StartTransaction(
		c.Ctx(),
		&transaction.StartTransactionRequest{},
	)
	if err != nil {
		return nil, grpcutil.ScrubGRPC(err)
	}
	return response, nil
}

// FinishTransaction is an RPC that closes an existing transaction in the
// Pachyderm cluster and commits its changes to the persisted cluster metadata
// transactionally.
func (c APIClient) FinishTransaction(txn *transaction.Transaction) (*transaction.TransactionInfo, error) {
	response, err := c.TransactionAPIClient.FinishTransaction(
		c.Ctx(),
		&transaction.FinishTransactionRequest{
			Transaction: txn,
		},
	)
	if err != nil {
		return nil, grpcutil.ScrubGRPC(err)
	}
	return response, nil
}

// DeleteTransaction is an RPC that aborts an existing transaction in the
// Pachyderm cluster and removes it from the cluster.
func (c APIClient) DeleteTransaction(txn *transaction.Transaction) error {
	_, err := c.TransactionAPIClient.DeleteTransaction(
		c.Ctx(),
		&transaction.DeleteTransactionRequest{
			Transaction: txn,
		},
	)
	return grpcutil.ScrubGRPC(err)
}

// InspectTransaction is an RPC that fetches the detailed information for an
// existing transaction in the Pachyderm cluster.
func (c APIClient) InspectTransaction(txn *transaction.Transaction) (*transaction.TransactionInfo, error) {
	response, err := c.TransactionAPIClient.InspectTransaction(
		c.Ctx(),
		&transaction.InspectTransactionRequest{
			Transaction: txn,
		},
	)
	if err != nil {
		return nil, grpcutil.ScrubGRPC(err)
	}
	return response, nil
}

// ExecuteInTransaction executes a callback within a transaction.
// The callback should use the passed in APIClient.
// If the callback returns a nil error, then the transaction will be finished.
// If the callback returns a non-nil error, then the transaction will be deleted.
func (c APIClient) ExecuteInTransaction(f func(c *APIClient) error) (*transaction.TransactionInfo, error) {
	txn, err := c.StartTransaction()
	if err != nil {
		return nil, err
	}
	if err := f(c.WithTransaction(txn)); err != nil {
		// We ignore the delete error, because we are more interested in the error from the callback.
		c.DeleteTransaction(txn)
		return nil, err
	}
	return c.FinishTransaction(txn)
}

// TransactionBuilder presents the same interface as a pachyderm APIClient, but
// captures requests rather than sending to the server. If a request is not
// supported by the transaction system, it immediately errors.
type TransactionBuilder struct {
	APIClient

	parent *APIClient

	requests []*transaction.TransactionRequest
}

type pfsBuilderClient struct {
	tb *TransactionBuilder
}

type ppsBuilderClient struct {
	tb *TransactionBuilder
}

type objectBuilderClient struct {
	tb *TransactionBuilder
}

type authBuilderClient struct {
	tb *TransactionBuilder
}

type versionBuilderClient struct {
	tb *TransactionBuilder
}

type adminBuilderClient struct {
	tb *TransactionBuilder
}

type transactionBuilderClient struct {
	tb *TransactionBuilder
}

type debugBuilderClient struct {
	tb *TransactionBuilder
}

type enterpriseBuilderClient struct {
	tb *TransactionBuilder
}

func newPfsBuilderClient(tb *TransactionBuilder) pfs.APIClient {
	return &pfsBuilderClient{tb: tb}
}

func newPpsBuilderClient(tb *TransactionBuilder) pps.APIClient {
	return &ppsBuilderClient{tb: tb}
}

func newObjectBuilderClient(tb *TransactionBuilder) pfs.ObjectAPIClient {
	return &objectBuilderClient{tb: tb}
}

func newAuthBuilderClient(tb *TransactionBuilder) auth.APIClient {
	return &authBuilderClient{tb: tb}
}

func newEnterpriseBuilderClient(tb *TransactionBuilder) enterprise.APIClient {
	return &enterpriseBuilderClient{tb: tb}
}

func newVersionBuilderClient(tb *TransactionBuilder) versionpb.APIClient {
	return &versionBuilderClient{tb: tb}
}

func newAdminBuilderClient(tb *TransactionBuilder) admin.APIClient {
	return &adminBuilderClient{tb: tb}
}

func newTransactionBuilderClient(tb *TransactionBuilder) transaction.APIClient {
	return &transactionBuilderClient{tb: tb}
}

func newDebugBuilderClient(tb *TransactionBuilder) debug.DebugClient {
	return &debugBuilderClient{tb: tb}
}

func newTransactionBuilder(parent *APIClient) *TransactionBuilder {
	tb := &TransactionBuilder{parent: parent}
	tb.PfsAPIClient = newPfsBuilderClient(tb)
	tb.PpsAPIClient = newPpsBuilderClient(tb)
	tb.ObjectAPIClient = newObjectBuilderClient(tb)
	tb.AuthAPIClient = newAuthBuilderClient(tb)
	tb.Enterprise = newEnterpriseBuilderClient(tb)
	tb.VersionAPIClient = newVersionBuilderClient(tb)
	tb.AdminAPIClient = newAdminBuilderClient(tb)
	tb.TransactionAPIClient = newTransactionBuilderClient(tb)
	tb.DebugClient = newDebugBuilderClient(tb)
	return tb
}

// Close does not exist on a TransactionBuilder because it doesn't represent
// ownership of a connection to the API server. We need this to shadow the
// inherited Close, though.
func (tb *TransactionBuilder) Close() error {
	return errors.Errorf("Close is not implemented on a TransactionBuilder instance")
}

// GetAddress should not exist on a TransactionBuilder because it doesn't represent
// ownership of a connection to the API server, but it also doesn't return an error,
// so we just passthrough to the parent client's implementation.
func (tb *TransactionBuilder) GetAddress() string {
	return tb.parent.GetAddress()
}

// RunBatchInTransaction will execute a batch of API calls in a single round-trip
// transactionally. The callback is used to build the request, which is executed
// when the callback returns.
func (c APIClient) RunBatchInTransaction(cb func(builder *TransactionBuilder) error) (*transaction.TransactionInfo, error) {
	tb := newTransactionBuilder(&c)
	if err := cb(tb); err != nil {
		return nil, err
	}

	return c.BatchTransaction(c.Ctx(), &transaction.BatchTransactionRequest{Requests: tb.requests})
}

func (c *pfsBuilderClient) CreateRepo(ctx context.Context, req *pfs.CreateRepoRequest, opts ...grpc.CallOption) (*types.Empty, error) {
	c.tb.requests = append(c.tb.requests, &transaction.TransactionRequest{CreateRepo: req})
	return nil, nil
}
func (c *pfsBuilderClient) DeleteRepo(ctx context.Context, req *pfs.DeleteRepoRequest, opts ...grpc.CallOption) (*types.Empty, error) {
	c.tb.requests = append(c.tb.requests, &transaction.TransactionRequest{DeleteRepo: req})
	return nil, nil
}
func (c *pfsBuilderClient) StartCommit(ctx context.Context, req *pfs.StartCommitRequest, opts ...grpc.CallOption) (*pfs.Commit, error) {
	// Note that since we are batching requests (no extra round-trips), we do not
	// have the commit id to return here. If you need an operation that relies
	// upon the commit id, use ExecuteInTransaction instead.
	c.tb.requests = append(c.tb.requests, &transaction.TransactionRequest{StartCommit: req})
	return nil, nil
}
func (c *pfsBuilderClient) FinishCommit(ctx context.Context, req *pfs.FinishCommitRequest, opts ...grpc.CallOption) (*types.Empty, error) {
	c.tb.requests = append(c.tb.requests, &transaction.TransactionRequest{FinishCommit: req})
	return nil, nil
}
func (c *pfsBuilderClient) DeleteCommit(ctx context.Context, req *pfs.DeleteCommitRequest, opts ...grpc.CallOption) (*types.Empty, error) {
	c.tb.requests = append(c.tb.requests, &transaction.TransactionRequest{DeleteCommit: req})
	return nil, nil
}
func (c *pfsBuilderClient) CreateBranch(ctx context.Context, req *pfs.CreateBranchRequest, opts ...grpc.CallOption) (*types.Empty, error) {
	c.tb.requests = append(c.tb.requests, &transaction.TransactionRequest{CreateBranch: req})
	return nil, nil
}
func (c *pfsBuilderClient) DeleteBranch(ctx context.Context, req *pfs.DeleteBranchRequest, opts ...grpc.CallOption) (*types.Empty, error) {
	c.tb.requests = append(c.tb.requests, &transaction.TransactionRequest{DeleteBranch: req})
	return nil, nil
}
func (c *ppsBuilderClient) UpdateJobState(ctx context.Context, req *pps.UpdateJobStateRequest, opts ...grpc.CallOption) (*types.Empty, error) {
	c.tb.requests = append(c.tb.requests, &transaction.TransactionRequest{UpdateJobState: req})
	return nil, nil
}

// Boilerplate for making unsupported API requests error when used on a TransactionBuilder
func unsupportedError(name string) error {
	return errors.Errorf("the '%s' API call is not supported in transactions", name)
}

func (c *pfsBuilderClient) InspectRepo(ctx context.Context, req *pfs.InspectRepoRequest, opts ...grpc.CallOption) (*pfs.RepoInfo, error) {
	return nil, unsupportedError("InspectRepo")
}
func (c *pfsBuilderClient) ListRepo(ctx context.Context, req *pfs.ListRepoRequest, opts ...grpc.CallOption) (*pfs.ListRepoResponse, error) {
	return nil, unsupportedError("ListRepo")
}
func (c *pfsBuilderClient) InspectCommit(ctx context.Context, req *pfs.InspectCommitRequest, opts ...grpc.CallOption) (*pfs.CommitInfo, error) {
	return nil, unsupportedError("InspectCommit")
}
func (c *pfsBuilderClient) ListCommit(ctx context.Context, req *pfs.ListCommitRequest, opts ...grpc.CallOption) (pfs.API_ListCommitClient, error) {
	return nil, unsupportedError("ListCommit")
}
func (c *pfsBuilderClient) FlushCommit(ctx context.Context, req *pfs.FlushCommitRequest, opts ...grpc.CallOption) (pfs.API_FlushCommitClient, error) {
	return nil, unsupportedError("FlushCommit")
}
func (c *pfsBuilderClient) SubscribeCommit(ctx context.Context, req *pfs.SubscribeCommitRequest, opts ...grpc.CallOption) (pfs.API_SubscribeCommitClient, error) {
	return nil, unsupportedError("SubscribeCommit")
}
func (c *pfsBuilderClient) ClearCommit(ctx context.Context, req *pfs.ClearCommitRequest, opts ...grpc.CallOption) (*types.Empty, error) {
	return nil, unsupportedError("ClearCommit")
}
func (c *pfsBuilderClient) InspectBranch(ctx context.Context, req *pfs.InspectBranchRequest, opts ...grpc.CallOption) (*pfs.BranchInfo, error) {
	return nil, unsupportedError("InspectBranch")
}
func (c *pfsBuilderClient) ListBranch(ctx context.Context, req *pfs.ListBranchRequest, opts ...grpc.CallOption) (*pfs.BranchInfos, error) {
	return nil, unsupportedError("ListBranch")
}
func (c *pfsBuilderClient) ModifyFile(ctx context.Context, opts ...grpc.CallOption) (pfs.API_ModifyFileClient, error) {
	return nil, unsupportedError("ModifyFile")
}
func (c *pfsBuilderClient) CopyFile(ctx context.Context, req *pfs.CopyFileRequest, opts ...grpc.CallOption) (*types.Empty, error) {
	return nil, unsupportedError("CopyFile")
}
func (c *pfsBuilderClient) GetFile(ctx context.Context, req *pfs.GetFileRequest, opts ...grpc.CallOption) (pfs.API_GetFileClient, error) {
	return nil, unsupportedError("GetFile")
}
func (c *pfsBuilderClient) InspectFile(ctx context.Context, req *pfs.InspectFileRequest, opts ...grpc.CallOption) (*pfs.FileInfo, error) {
	return nil, unsupportedError("InspectFile")
}
func (c *pfsBuilderClient) ListFile(ctx context.Context, req *pfs.ListFileRequest, opts ...grpc.CallOption) (pfs.API_ListFileClient, error) {
	return nil, unsupportedError("ListFile")
}
func (c *pfsBuilderClient) WalkFile(ctx context.Context, req *pfs.WalkFileRequest, opts ...grpc.CallOption) (pfs.API_WalkFileClient, error) {
	return nil, unsupportedError("WalkFile")
}
func (c *pfsBuilderClient) GlobFile(ctx context.Context, req *pfs.GlobFileRequest, opts ...grpc.CallOption) (pfs.API_GlobFileClient, error) {
	return nil, unsupportedError("GlobFile")
}
func (c *pfsBuilderClient) DiffFile(ctx context.Context, req *pfs.DiffFileRequest, opts ...grpc.CallOption) (pfs.API_DiffFileClient, error) {
	return nil, unsupportedError("DiffFile")
}
func (c *pfsBuilderClient) DeleteAll(ctx context.Context, req *types.Empty, opts ...grpc.CallOption) (*types.Empty, error) {
	return nil, unsupportedError("DeleteAll")
}
func (c *pfsBuilderClient) Fsck(ctx context.Context, req *pfs.FsckRequest, opts ...grpc.CallOption) (pfs.API_FsckClient, error) {
	return nil, unsupportedError("Fsck")
}
func (c *pfsBuilderClient) CreateFileset(ctx context.Context, opts ...grpc.CallOption) (pfs.API_CreateFilesetClient, error) {
	return nil, unsupportedError("CreateFileset")
}
func (c *pfsBuilderClient) RenewFileset(ctx context.Context, req *pfs.RenewFilesetRequest, opts ...grpc.CallOption) (*types.Empty, error) {
	return nil, unsupportedError("RenewFileset")
}

func (c *objectBuilderClient) PutObject(ctx context.Context, opts ...grpc.CallOption) (pfs.ObjectAPI_PutObjectClient, error) {
	return nil, unsupportedError("PutObject")
}
func (c *objectBuilderClient) PutObjectSplit(ctx context.Context, opts ...grpc.CallOption) (pfs.ObjectAPI_PutObjectSplitClient, error) {
	return nil, unsupportedError("PutObjectSplit")
}
func (c *objectBuilderClient) PutObjects(ctx context.Context, opts ...grpc.CallOption) (pfs.ObjectAPI_PutObjectsClient, error) {
	return nil, unsupportedError("PutObjects")
}
func (c *objectBuilderClient) CreateObject(ctx context.Context, req *pfs.CreateObjectRequest, opts ...grpc.CallOption) (*types.Empty, error) {
	return nil, unsupportedError("CreateObject")
}
func (c *objectBuilderClient) GetObject(ctx context.Context, req *pfs.Object, opts ...grpc.CallOption) (pfs.ObjectAPI_GetObjectClient, error) {
	return nil, unsupportedError("GetObject")
}
func (c *objectBuilderClient) GetObjects(ctx context.Context, req *pfs.GetObjectsRequest, opts ...grpc.CallOption) (pfs.ObjectAPI_GetObjectsClient, error) {
	return nil, unsupportedError("GetObjects")
}
func (c *objectBuilderClient) PutBlock(ctx context.Context, opts ...grpc.CallOption) (pfs.ObjectAPI_PutBlockClient, error) {
	return nil, unsupportedError("PutBlock")
}
func (c *objectBuilderClient) GetBlock(ctx context.Context, req *pfs.GetBlockRequest, opts ...grpc.CallOption) (pfs.ObjectAPI_GetBlockClient, error) {
	return nil, unsupportedError("GetBlock")
}
func (c *objectBuilderClient) GetBlocks(ctx context.Context, req *pfs.GetBlocksRequest, opts ...grpc.CallOption) (pfs.ObjectAPI_GetBlocksClient, error) {
	return nil, unsupportedError("GetBlocks")
}
func (c *objectBuilderClient) ListBlock(ctx context.Context, req *pfs.ListBlockRequest, opts ...grpc.CallOption) (pfs.ObjectAPI_ListBlockClient, error) {
	return nil, unsupportedError("ListBlock")
}
func (c *objectBuilderClient) TagObject(ctx context.Context, req *pfs.TagObjectRequest, opts ...grpc.CallOption) (*types.Empty, error) {
	return nil, unsupportedError("TagObject")
}
func (c *objectBuilderClient) InspectObject(ctx context.Context, req *pfs.Object, opts ...grpc.CallOption) (*pfs.ObjectInfo, error) {
	return nil, unsupportedError("InspectObject")
}
func (c *objectBuilderClient) CheckObject(ctx context.Context, req *pfs.CheckObjectRequest, opts ...grpc.CallOption) (*pfs.CheckObjectResponse, error) {
	return nil, unsupportedError("CheckObject")
}
func (c *objectBuilderClient) ListObjects(ctx context.Context, req *pfs.ListObjectsRequest, opts ...grpc.CallOption) (pfs.ObjectAPI_ListObjectsClient, error) {
	return nil, unsupportedError("ListObjects")
}
func (c *objectBuilderClient) DeleteObjects(ctx context.Context, req *pfs.DeleteObjectsRequest, opts ...grpc.CallOption) (*pfs.DeleteObjectsResponse, error) {
	return nil, unsupportedError("DeleteObjects")
}
func (c *objectBuilderClient) GetTag(ctx context.Context, req *pfs.Tag, opts ...grpc.CallOption) (pfs.ObjectAPI_GetTagClient, error) {
	return nil, unsupportedError("GetTag")
}
func (c *objectBuilderClient) InspectTag(ctx context.Context, req *pfs.Tag, opts ...grpc.CallOption) (*pfs.ObjectInfo, error) {
	return nil, unsupportedError("InspectTag")
}
func (c *objectBuilderClient) ListTags(ctx context.Context, req *pfs.ListTagsRequest, opts ...grpc.CallOption) (pfs.ObjectAPI_ListTagsClient, error) {
	return nil, unsupportedError("ListTags")
}
func (c *objectBuilderClient) DeleteTags(ctx context.Context, req *pfs.DeleteTagsRequest, opts ...grpc.CallOption) (*pfs.DeleteTagsResponse, error) {
	return nil, unsupportedError("DeleteTags")
}
func (c *objectBuilderClient) Compact(ctx context.Context, req *types.Empty, opts ...grpc.CallOption) (*types.Empty, error) {
	return nil, unsupportedError("Compact")
}
func (c *objectBuilderClient) PutObjDirect(ctx context.Context, opts ...grpc.CallOption) (pfs.ObjectAPI_PutObjDirectClient, error) {
	return nil, unsupportedError("PutObjDirect")
}
func (c *objectBuilderClient) GetObjDirect(ctx context.Context, req *pfs.GetObjDirectRequest, opts ...grpc.CallOption) (pfs.ObjectAPI_GetObjDirectClient, error) {
	return nil, unsupportedError("GetObjDirect")
}
func (c *objectBuilderClient) DeleteObjDirect(ctx context.Context, req *pfs.DeleteObjDirectRequest, opts ...grpc.CallOption) (*types.Empty, error) {
	return nil, unsupportedError("DeleteObjDirect")
}

func (c *ppsBuilderClient) CreateJob(ctx context.Context, req *pps.CreateJobRequest, opts ...grpc.CallOption) (*pps.Job, error) {
	return nil, unsupportedError("CreateJob")
}
func (c *ppsBuilderClient) InspectJob(ctx context.Context, req *pps.InspectJobRequest, opts ...grpc.CallOption) (*pps.JobInfo, error) {
	return nil, unsupportedError("InspectJob")
}
func (c *ppsBuilderClient) ListJob(ctx context.Context, req *pps.ListJobRequest, opts ...grpc.CallOption) (pps.API_ListJobClient, error) {
	return nil, unsupportedError("ListJob")
}
func (c *ppsBuilderClient) FlushJob(ctx context.Context, req *pps.FlushJobRequest, opts ...grpc.CallOption) (pps.API_FlushJobClient, error) {
	return nil, unsupportedError("FlushJob")
}
func (c *ppsBuilderClient) DeleteJob(ctx context.Context, req *pps.DeleteJobRequest, opts ...grpc.CallOption) (*types.Empty, error) {
	return nil, unsupportedError("DeleteJob")
}
func (c *ppsBuilderClient) StopJob(ctx context.Context, req *pps.StopJobRequest, opts ...grpc.CallOption) (*types.Empty, error) {
	return nil, unsupportedError("StopJob")
}
func (c *ppsBuilderClient) InspectDatum(ctx context.Context, req *pps.InspectDatumRequest, opts ...grpc.CallOption) (*pps.DatumInfo, error) {
	return nil, unsupportedError("InspectDatum")
}
func (c *ppsBuilderClient) ListDatum(ctx context.Context, req *pps.ListDatumRequest, opts ...grpc.CallOption) (pps.API_ListDatumClient, error) {
	return nil, unsupportedError("ListDatum")
}
func (c *ppsBuilderClient) RestartDatum(ctx context.Context, req *pps.RestartDatumRequest, opts ...grpc.CallOption) (*types.Empty, error) {
	return nil, unsupportedError("RestartDatum")
}
func (c *ppsBuilderClient) CreatePipeline(ctx context.Context, req *pps.CreatePipelineRequest, opts ...grpc.CallOption) (*types.Empty, error) {
	return nil, unsupportedError("CreatePipeline")
}
func (c *ppsBuilderClient) InspectPipeline(ctx context.Context, req *pps.InspectPipelineRequest, opts ...grpc.CallOption) (*pps.PipelineInfo, error) {
	return nil, unsupportedError("InspectPipeline")
}
func (c *ppsBuilderClient) ListPipeline(ctx context.Context, req *pps.ListPipelineRequest, opts ...grpc.CallOption) (*pps.PipelineInfos, error) {
	return nil, unsupportedError("ListPipeline")
}
func (c *ppsBuilderClient) DeletePipeline(ctx context.Context, req *pps.DeletePipelineRequest, opts ...grpc.CallOption) (*types.Empty, error) {
	return nil, unsupportedError("DeletePipeline")
}
func (c *ppsBuilderClient) StartPipeline(ctx context.Context, req *pps.StartPipelineRequest, opts ...grpc.CallOption) (*types.Empty, error) {
	return nil, unsupportedError("StartPipeline")
}
func (c *ppsBuilderClient) StopPipeline(ctx context.Context, req *pps.StopPipelineRequest, opts ...grpc.CallOption) (*types.Empty, error) {
	return nil, unsupportedError("StopPipeline")
}
func (c *ppsBuilderClient) RunPipeline(ctx context.Context, req *pps.RunPipelineRequest, opts ...grpc.CallOption) (*types.Empty, error) {
	return nil, unsupportedError("RunPipeline")
}
func (c *ppsBuilderClient) DeleteAll(ctx context.Context, req *types.Empty, opts ...grpc.CallOption) (*types.Empty, error) {
	return nil, unsupportedError("DeleteAll")
}
func (c *ppsBuilderClient) GetLogs(ctx context.Context, req *pps.GetLogsRequest, opts ...grpc.CallOption) (pps.API_GetLogsClient, error) {
	return nil, unsupportedError("GetLogs")
}
func (c *ppsBuilderClient) ActivateAuth(ctx context.Context, req *pps.ActivateAuthRequest, opts ...grpc.CallOption) (*pps.ActivateAuthResponse, error) {
	return nil, unsupportedError("ActivateAuth")
}
func (c *ppsBuilderClient) RunCron(ctx context.Context, req *pps.RunCronRequest, opts ...grpc.CallOption) (*types.Empty, error) {
	return nil, unsupportedError("RunCron")
}
func (c *ppsBuilderClient) CreateSecret(ctx context.Context, req *pps.CreateSecretRequest, opts ...grpc.CallOption) (*types.Empty, error) {
	return nil, unsupportedError("CreateSecret")
}
func (c *ppsBuilderClient) DeleteSecret(ctx context.Context, req *pps.DeleteSecretRequest, opts ...grpc.CallOption) (*types.Empty, error) {
	return nil, unsupportedError("DeleteSecret")
}
func (c *ppsBuilderClient) InspectSecret(ctx context.Context, req *pps.InspectSecretRequest, opt ...grpc.CallOption) (*pps.SecretInfo, error) {
	return nil, unsupportedError("InspectSecret")
}
func (c *ppsBuilderClient) ListSecret(ctx context.Context, in *types.Empty, opt ...grpc.CallOption) (*pps.SecretInfos, error) {
	return nil, unsupportedError("ListSecret")
}

func (c *authBuilderClient) Activate(ctx context.Context, req *auth.ActivateRequest, opts ...grpc.CallOption) (*auth.ActivateResponse, error) {
	return nil, unsupportedError("Activate")
}
func (c *authBuilderClient) Deactivate(ctx context.Context, req *auth.DeactivateRequest, opts ...grpc.CallOption) (*auth.DeactivateResponse, error) {
	return nil, unsupportedError("Deactivate")
}
func (c *authBuilderClient) GetConfiguration(ctx context.Context, req *auth.GetConfigurationRequest, opts ...grpc.CallOption) (*auth.GetConfigurationResponse, error) {
	return nil, unsupportedError("GetConfiguration")
}
func (c *authBuilderClient) SetConfiguration(ctx context.Context, req *auth.SetConfigurationRequest, opts ...grpc.CallOption) (*auth.SetConfigurationResponse, error) {
	return nil, unsupportedError("SetConfiguration")
}
func (c *authBuilderClient) GetRoleBinding(ctx context.Context, req *auth.GetRoleBindingRequest, opts ...grpc.CallOption) (*auth.GetRoleBindingResponse, error) {
	return nil, unsupportedError("GetRoleBinding")
}
func (c *authBuilderClient) ModifyRoleBinding(ctx context.Context, req *auth.ModifyRoleBindingRequest, opts ...grpc.CallOption) (*auth.ModifyRoleBindingResponse, error) {
	return nil, unsupportedError("ModifyRoleBinding")
}
func (c *authBuilderClient) DeleteRoleBinding(ctx context.Context, req *auth.DeleteRoleBindingRequest, opts ...grpc.CallOption) (*auth.DeleteRoleBindingResponse, error) {
	return nil, unsupportedError("DeleteRoleBinding")
}
func (c *authBuilderClient) Authenticate(ctx context.Context, req *auth.AuthenticateRequest, opts ...grpc.CallOption) (*auth.AuthenticateResponse, error) {
	return nil, unsupportedError("Authenticate")
}
func (c *authBuilderClient) Authorize(ctx context.Context, req *auth.AuthorizeRequest, opts ...grpc.CallOption) (*auth.AuthorizeResponse, error) {
	return nil, unsupportedError("Authorize")
}
func (c *authBuilderClient) WhoAmI(ctx context.Context, req *auth.WhoAmIRequest, opts ...grpc.CallOption) (*auth.WhoAmIResponse, error) {
	return nil, unsupportedError("WhoAmI")
}
func (c *authBuilderClient) GetAuthToken(ctx context.Context, req *auth.GetAuthTokenRequest, opts ...grpc.CallOption) (*auth.GetAuthTokenResponse, error) {
	return nil, unsupportedError("GetAuthToken")
}
func (c *authBuilderClient) GetOIDCLogin(ctx context.Context, req *auth.GetOIDCLoginRequest, opts ...grpc.CallOption) (*auth.GetOIDCLoginResponse, error) {
	return nil, unsupportedError("GetOIDCLogin")
}
func (c *authBuilderClient) ExtendAuthToken(ctx context.Context, req *auth.ExtendAuthTokenRequest, opts ...grpc.CallOption) (*auth.ExtendAuthTokenResponse, error) {
	return nil, unsupportedError("ExtendAuthToken")
}
func (c *authBuilderClient) RevokeAuthToken(ctx context.Context, req *auth.RevokeAuthTokenRequest, opts ...grpc.CallOption) (*auth.RevokeAuthTokenResponse, error) {
	return nil, unsupportedError("RevokeAuthToken")
}
func (c *authBuilderClient) SetGroupsForUser(ctx context.Context, req *auth.SetGroupsForUserRequest, opts ...grpc.CallOption) (*auth.SetGroupsForUserResponse, error) {
	return nil, unsupportedError("SetGroupsForUser")
}
func (c *authBuilderClient) ModifyMembers(ctx context.Context, req *auth.ModifyMembersRequest, opts ...grpc.CallOption) (*auth.ModifyMembersResponse, error) {
	return nil, unsupportedError("ModifyMembers")
}
func (c *authBuilderClient) GetGroups(ctx context.Context, req *auth.GetGroupsRequest, opts ...grpc.CallOption) (*auth.GetGroupsResponse, error) {
	return nil, unsupportedError("GetGroups")
}
func (c *authBuilderClient) GetUsers(ctx context.Context, req *auth.GetUsersRequest, opts ...grpc.CallOption) (*auth.GetUsersResponse, error) {
	return nil, unsupportedError("GetUsers")
}
func (c *authBuilderClient) ExtractAuthTokens(ctx context.Context, req *auth.ExtractAuthTokensRequest, opts ...grpc.CallOption) (*auth.ExtractAuthTokensResponse, error) {
	return nil, unsupportedError("ExtractAuthTokens")
}
func (c *authBuilderClient) RestoreAuthToken(ctx context.Context, req *auth.RestoreAuthTokenRequest, opts ...grpc.CallOption) (*auth.RestoreAuthTokenResponse, error) {
	return nil, unsupportedError("RestoreAuthToken")
}

func (c *enterpriseBuilderClient) Activate(ctx context.Context, req *enterprise.ActivateRequest, opts ...grpc.CallOption) (*enterprise.ActivateResponse, error) {
	return nil, unsupportedError("Activate")
}
func (c *enterpriseBuilderClient) GetState(ctx context.Context, req *enterprise.GetStateRequest, opts ...grpc.CallOption) (*enterprise.GetStateResponse, error) {
	return nil, unsupportedError("GetState")
}
func (c *enterpriseBuilderClient) GetActivationCode(ctx context.Context, req *enterprise.GetActivationCodeRequest, opts ...grpc.CallOption) (*enterprise.GetActivationCodeResponse, error) {
	return nil, unsupportedError("GetActivationCode")
}
func (c *enterpriseBuilderClient) Deactivate(ctx context.Context, req *enterprise.DeactivateRequest, opts ...grpc.CallOption) (*enterprise.DeactivateResponse, error) {
	return nil, unsupportedError("Deactivate")
}
func (c *enterpriseBuilderClient) Heartbeat(ctx context.Context, req *enterprise.HeartbeatRequest, opts ...grpc.CallOption) (*enterprise.HeartbeatResponse, error) {
	return nil, unsupportedError("Heartbeat")
}

func (c *versionBuilderClient) GetVersion(ctx context.Context, req *types.Empty, opts ...grpc.CallOption) (*versionpb.Version, error) {
	return nil, unsupportedError("GetVersion")
}

func (c *adminBuilderClient) InspectCluster(ctx context.Context, req *types.Empty, opts ...grpc.CallOption) (*admin.ClusterInfo, error) {
	return nil, unsupportedError("InspectCluster")
}

func (c *transactionBuilderClient) BatchTransaction(ctx context.Context, req *transaction.BatchTransactionRequest, opts ...grpc.CallOption) (*transaction.TransactionInfo, error) {
	return nil, unsupportedError("BatchTransaction")
}
func (c *transactionBuilderClient) StartTransaction(ctx context.Context, req *transaction.StartTransactionRequest, opts ...grpc.CallOption) (*transaction.Transaction, error) {
	return nil, unsupportedError("StartTransaction")
}
func (c *transactionBuilderClient) InspectTransaction(ctx context.Context, req *transaction.InspectTransactionRequest, opts ...grpc.CallOption) (*transaction.TransactionInfo, error) {
	return nil, unsupportedError("InspectTransaction")
}
func (c *transactionBuilderClient) DeleteTransaction(ctx context.Context, req *transaction.DeleteTransactionRequest, opts ...grpc.CallOption) (*types.Empty, error) {
	return nil, unsupportedError("DeleteTransaction")
}
func (c *transactionBuilderClient) ListTransaction(ctx context.Context, req *transaction.ListTransactionRequest, opts ...grpc.CallOption) (*transaction.TransactionInfos, error) {
	return nil, unsupportedError("ListTransaction")
}
func (c *transactionBuilderClient) FinishTransaction(ctx context.Context, req *transaction.FinishTransactionRequest, opts ...grpc.CallOption) (*transaction.TransactionInfo, error) {
	return nil, unsupportedError("FinishTransaction")
}
func (c *transactionBuilderClient) DeleteAll(ctx context.Context, req *transaction.DeleteAllRequest, opts ...grpc.CallOption) (*types.Empty, error) {
	return nil, unsupportedError("DeleteAll")
}

func (c *debugBuilderClient) Profile(ctx context.Context, req *debug.ProfileRequest, opts ...grpc.CallOption) (debug.Debug_ProfileClient, error) {
	return nil, unsupportedError("Profile")
}
func (c *debugBuilderClient) Binary(ctx context.Context, req *debug.BinaryRequest, opts ...grpc.CallOption) (debug.Debug_BinaryClient, error) {
	return nil, unsupportedError("Binary")
}
func (c *debugBuilderClient) Dump(ctx context.Context, req *debug.DumpRequest, opts ...grpc.CallOption) (debug.Debug_DumpClient, error) {
	return nil, unsupportedError("Dump")
}
