package client

import (
	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/client/transaction"
	"github.com/pachyderm/pachyderm/src/server/pkg/errutil"
)

func (c APIClient) ListTransaction() ([]*pfs.TransactionInfo, error) {
	response, err := c.PfsAPIClient.ListTransaction(
		c.Ctx(),
		&pfs.ListTransactionRequest{},
	)
	if err != nil {
		return nil, grpcutil.ScrubGRPC(err)
	}
	return response.TransactionInfo, nil
}

func (c APIClient) StartTransaction() (*pfs.Transaction, error) {
	response, err := c.PfsAPIClient.StartTransaction(
		c.Ctx(),
		&pfs.StartTransactionRequest{},
	)
	if err != nil {
		return nil, grpcutil.ScrubGRPC(err)
	}
	return response, nil
}

func (c APIClient) FinishTransaction(transaction *pfs.Transaction) (*pfs.TransactionInfo, error) {
	response, err := c.PfsAPIClient.FinishTransaction(
		c.Ctx(),
		&pfs.FinishTransactionRequest{
			Transaction: transaction,
		},
	)
	if err != nil {
		return nil, grpcutil.ScrubGRPC(err)
	}
	return response, nil
}

func (c APIClient) DeleteTransaction(transaction *pfs.Transaction) error {
	_, err := c.PfsAPIClient.DeleteTransaction(
		c.Ctx(),
		&pfs.DeleteTransactionRequest{
			Transaction: transaction,
		},
	)
	return grpcutil.ScrubGRPC(err)
}

func (c APIClient) InspectTransaction(transaction *pfs.Transaction) (*pfs.TransactionInfo, error) {
	response, err := c.PfsAPIClient.InspectTransaction(
		c.Ctx(),
		&pfs.InspectTransactionRequest{
			Transaction: transaction,
		},
	)
	if err != nil {
		return nil, grpcutil.ScrubGRPC(err)
	}
	return response, nil
}
