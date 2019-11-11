package client

import (
	"context"
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

// GetTransaction (should be run from the server-side) loads the active
// transaction from the grpc metadata and returns the associated transaction
// object - or `nil` if no transaction is set.
func GetTransaction(ctx context.Context) (*transaction.Transaction, error) {
	md, ok := metadata.FromIncomingContext(ctx)
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
