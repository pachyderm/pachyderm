package client

import (
	"fmt"

	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/client/transaction"

	"github.com/gogo/protobuf/types"
)

func NewEmptyResponse() *transaction.TransactionResponse {
	return &transaction.TransactionResponse{
		Response: &transaction.TransactionResponse_None{
			None: &types.Empty{},
		},
	}
}

func NewCommitResponse(commit *pfs.Commit) *transaction.TransactionResponse {
	return &transaction.TransactionResponse{
		Response: &transaction.TransactionResponse_Commit{
			Commit: commit,
		},
	}
}

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

func (c APIClient) DeleteTransaction(txn *transaction.Transaction) error {
	_, err := c.TransactionAPIClient.DeleteTransaction(
		c.Ctx(),
		&transaction.DeleteTransactionRequest{
			Transaction: txn,
		},
	)
	return grpcutil.ScrubGRPC(err)
}

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

// helper function for the most common case,
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

// This is a rare write operation that actually has a return value
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
