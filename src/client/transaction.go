package client

import (
	"fmt"

	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/client/transaction"

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

// GetTransaction (run server-side) loads the active transaction from the grpc
// metadata and returns the associated transaction object - or `nil` if no
// transaction is set.
func (c APIClient) GetTransaction() (*transaction.Transaction, error) {
	md, ok := metadata.FromIncomingContext(c.Ctx())
	if !ok {
		return nil, fmt.Errorf("request metadata could not be parsed from context")
	}

	txns := md.Get(transactionMetadataKey)
	if txns == nil || len(txns) == 0 {
		return nil, nil
	} else if len(txns) > 1 {
		return nil, fmt.Errorf("multiple active transactions found in context")
	}
	return &transaction.Transaction{ID: txns[0]}, nil
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

/*
// appendTransactionHelper is a helper function for the most common case of
// appending a transaction, because there's quite a bit of boilerplate here.
// This is suitable when appending a single request to a transaction when that
// request type should return an empty response.
func (c APIClient) appendTransactionHelper(
	txn *transaction.Transaction,
	request *transaction.TransactionRequest,
) (*types.Empty, error) {
	info, err := c.TransactionAPIClient.AppendTransaction(
		c.Ctx(),
		&transaction.AppendTransactionRequest{
			Transaction: txn,
			Items:       []*transaction.TransactionRequest{request},
		},
	)
	if err != nil {
		return nil, grpcutil.ScrubGRPC(err)
	}

	switch response := info.Responses[len(info.Responses)-1].Response.(type) {
	case *transaction.TransactionResponse_None:
		return response.None, nil
	}
	return nil, fmt.Errorf("added to transaction but received an unexpected response type")
}

// AppendCreateRepo is a wrapper function for the AppendTransaction RPC which
// will append a CreateRepoRequest to an existing transaction in the Pachyderm
// cluster and return the expected response.
func (c APIClient) AppendCreateRepo(txn *transaction.Transaction, request *pfs.CreateRepoRequest) (*types.Empty, error) {
	return c.appendTransactionHelper(
		txn,
		&transaction.TransactionRequest{
			Request: &transaction.TransactionRequest_CreateRepo{
				CreateRepo: request,
			},
		},
	)
}

// AppendDeleteRepo is a wrapper function for the AppendTransaction RPC which
// will append a DeleteRepoRequest to an existing transaction in the Pachyderm
// cluster and return the expected response.
func (c APIClient) AppendDeleteRepo(txn *transaction.Transaction, request *pfs.DeleteRepoRequest) (*types.Empty, error) {
	return c.appendTransactionHelper(
		txn,
		&transaction.TransactionRequest{
			Request: &transaction.TransactionRequest_DeleteRepo{
				DeleteRepo: request,
			},
		},
	)
}

// AppendFinishCommit is a wrapper function for the AppendTransaction RPC which
// will append a FinishCommitRequest to an existing transaction in the Pachyderm
// cluster and return the expected response.
func (c APIClient) AppendFinishCommit(txn *transaction.Transaction, request *pfs.FinishCommitRequest) (*types.Empty, error) {
	return c.appendTransactionHelper(
		txn,
		&transaction.TransactionRequest{
			Request: &transaction.TransactionRequest_FinishCommit{
				FinishCommit: request,
			},
		},
	)
}

// AppendDeleteCommit is a wrapper function for the AppendTransaction RPC which
// will append a DeleteCommitRequest to an existing transaction in the Pachyderm
// cluster and return the expected response.
func (c APIClient) AppendDeleteCommit(txn *transaction.Transaction, request *pfs.DeleteCommitRequest) (*types.Empty, error) {
	return c.appendTransactionHelper(
		txn,
		&transaction.TransactionRequest{
			Request: &transaction.TransactionRequest_DeleteCommit{
				DeleteCommit: request,
			},
		},
	)
}

// AppendCreateBranch is a wrapper function for the AppendTransaction RPC which
// will append a CreateBranchRequest to an existing transaction in the Pachyderm
// cluster and return the expected response.
func (c APIClient) AppendCreateBranch(txn *transaction.Transaction, request *pfs.CreateBranchRequest) (*types.Empty, error) {
	return c.appendTransactionHelper(
		txn,
		&transaction.TransactionRequest{
			Request: &transaction.TransactionRequest_CreateBranch{
				CreateBranch: request,
			},
		},
	)
}

// AppendDeleteBranch is a wrapper function for the AppendTransaction RPC which
// will append a DeleteBranchRequest to an existing transaction in the Pachyderm
// cluster and return the expected response.
func (c APIClient) AppendDeleteBranch(txn *transaction.Transaction, request *pfs.DeleteBranchRequest) (*types.Empty, error) {
	return c.appendTransactionHelper(
		txn,
		&transaction.TransactionRequest{
			Request: &transaction.TransactionRequest_DeleteBranch{
				DeleteBranch: request,
			},
		},
	)
}

// AppendCopyFile is a wrapper function for the AppendTransaction RPC which
// will append a CopyFileRequest to an existing transaction in the Pachyderm
// cluster and return the expected response.
func (c APIClient) AppendCopyFile(txn *transaction.Transaction, request *pfs.CopyFileRequest) (*types.Empty, error) {
	return c.appendTransactionHelper(
		txn,
		&transaction.TransactionRequest{
			Request: &transaction.TransactionRequest_CopyFile{
				CopyFile: request,
			},
		},
	)
}

// AppendDeleteFile is a wrapper function for the AppendTransaction RPC which
// will append a DeleteFileRequest to an existing transaction in the Pachyderm
// cluster and return the expected response.
func (c APIClient) AppendDeleteFile(txn *transaction.Transaction, request *pfs.DeleteFileRequest) (*types.Empty, error) {
	return c.appendTransactionHelper(
		txn,
		&transaction.TransactionRequest{
			Request: &transaction.TransactionRequest_DeleteFile{
				DeleteFile: request,
			},
		},
	)
}

// AppendStartCommit is a wrapper function for the AppendTransaction RPC which
// will append a StartCommitRequest to an existing transaction in the Pachyderm
// cluster and return the expected response. This is a rare write operation that
// actually has a return value - the ID of the new Commit.
func (c APIClient) AppendStartCommit(txn *transaction.Transaction, request *pfs.StartCommitRequest) (*pfs.Commit, error) {
	info, err := c.TransactionAPIClient.AppendTransaction(
		c.Ctx(),
		&transaction.AppendTransactionRequest{
			Transaction: txn,
			Items: []*transaction.TransactionRequest{
				&transaction.TransactionRequest{
					Request: &transaction.TransactionRequest_StartCommit{
						StartCommit: request,
					},
				},
			},
		},
	)
	if err != nil {
		return nil, grpcutil.ScrubGRPC(err)
	}

	switch response := info.Responses[len(info.Responses)-1].Response.(type) {
	case *transaction.TransactionResponse_Commit:
		return response.Commit, nil
	}
	return nil, fmt.Errorf("added to transaction but received an unexpected response type")
}
*/
