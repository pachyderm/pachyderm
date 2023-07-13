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

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)

const transactionMetadataKey = "pach-transaction"

// WithTransaction (client-side) returns a new APIClient that will run supported
// write operations within the specified transaction.
func (c APIClient) WithTransaction(txn *transaction.Transaction) *APIClient {
	md, _ := metadata.FromOutgoingContext(c.Ctx())
	md = md.Copy()
	if txn != nil {
		md.Set(transactionMetadataKey, txn.Id)
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
	return &transaction.Transaction{Id: txns[0]}, nil
}

// GetTransaction is a helper function to get the active transaction from the
// client's context metadata.
func (c APIClient) GetTransaction() (*transaction.Transaction, error) {
	return GetTransaction(c.Ctx())
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
		// TODO multierr
		_ = c.DeleteTransaction(txn)
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
	unsupportedPfsBuilderClient
	tb *TransactionBuilder
}

type ppsBuilderClient struct {
	unsupportedPpsBuilderClient
	tb *TransactionBuilder
}

type authBuilderClient struct {
	unsupportedAuthBuilderClient
	tb *TransactionBuilder
}

type versionpbBuilderClient struct {
	unsupportedVersionpbBuilderClient
	tb *TransactionBuilder
}

type adminBuilderClient struct {
	unsupportedAdminBuilderClient
	tb *TransactionBuilder
}

type transactionBuilderClient struct {
	unsupportedTransactionBuilderClient
	tb *TransactionBuilder
}

type debugBuilderClient struct {
	unsupportedDebugBuilderClient
	tb *TransactionBuilder
}

type enterpriseBuilderClient struct {
	unsupportedEnterpriseBuilderClient
	tb *TransactionBuilder
}

func newPfsBuilderClient(tb *TransactionBuilder) pfs.APIClient {
	return &pfsBuilderClient{tb: tb}
}

func newPpsBuilderClient(tb *TransactionBuilder) pps.APIClient {
	return &ppsBuilderClient{tb: tb}
}

func newAuthBuilderClient(tb *TransactionBuilder) auth.APIClient {
	return &authBuilderClient{tb: tb}
}

func newEnterpriseBuilderClient(tb *TransactionBuilder) enterprise.APIClient {
	return &enterpriseBuilderClient{tb: tb}
}

func newVersionpbBuilderClient(tb *TransactionBuilder) versionpb.APIClient {
	return &versionpbBuilderClient{tb: tb}
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
	tb.AuthAPIClient = newAuthBuilderClient(tb)
	tb.Enterprise = newEnterpriseBuilderClient(tb)
	tb.VersionAPIClient = newVersionpbBuilderClient(tb)
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
func (tb *TransactionBuilder) GetAddress() *grpcutil.PachdAddress {
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

func (c *pfsBuilderClient) CreateRepo(ctx context.Context, req *pfs.CreateRepoRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	c.tb.requests = append(c.tb.requests, &transaction.TransactionRequest{CreateRepo: req})
	return nil, nil
}
func (c *pfsBuilderClient) DeleteRepo(ctx context.Context, req *pfs.DeleteRepoRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
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
func (c *pfsBuilderClient) FinishCommit(ctx context.Context, req *pfs.FinishCommitRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	c.tb.requests = append(c.tb.requests, &transaction.TransactionRequest{FinishCommit: req})
	return nil, nil
}
func (c *pfsBuilderClient) SquashCommitSet(ctx context.Context, req *pfs.SquashCommitSetRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	c.tb.requests = append(c.tb.requests, &transaction.TransactionRequest{SquashCommitSet: req})
	return nil, nil
}
func (c *pfsBuilderClient) CreateBranch(ctx context.Context, req *pfs.CreateBranchRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	c.tb.requests = append(c.tb.requests, &transaction.TransactionRequest{CreateBranch: req})
	return nil, nil
}
func (c *pfsBuilderClient) DeleteBranch(ctx context.Context, req *pfs.DeleteBranchRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	c.tb.requests = append(c.tb.requests, &transaction.TransactionRequest{DeleteBranch: req})
	return nil, nil
}
func (c *ppsBuilderClient) StopJob(ctx context.Context, req *pps.StopJobRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	c.tb.requests = append(c.tb.requests, &transaction.TransactionRequest{StopJob: req})
	return nil, nil
}
func (c *ppsBuilderClient) UpdateJobState(ctx context.Context, req *pps.UpdateJobStateRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	c.tb.requests = append(c.tb.requests, &transaction.TransactionRequest{UpdateJobState: req})
	return nil, nil
}
func (c *ppsBuilderClient) CreatePipeline(ctx context.Context, req *pps.CreatePipelineRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	c.tb.requests = append(c.tb.requests, &transaction.TransactionRequest{CreatePipeline: req})
	return nil, nil
}
