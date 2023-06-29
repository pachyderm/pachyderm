package server

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	txnenv "github.com/pachyderm/pachyderm/v2/src/internal/transactionenv"
	"github.com/pachyderm/pachyderm/v2/src/transaction"
	"google.golang.org/protobuf/types/known/emptypb"
)

type apiServer struct {
	transaction.UnimplementedAPIServer

	driver *driver
}

func newAPIServer(
	env serviceenv.ServiceEnv,
	txnEnv *txnenv.TransactionEnv,
) (*apiServer, error) {
	d, err := newDriver(env, txnEnv)
	if err != nil {
		return nil, err
	}
	s := &apiServer{
		driver: d,
	}
	return s, nil
}

func (a *apiServer) BatchTransaction(ctx context.Context, request *transaction.BatchTransactionRequest) (response *transaction.TransactionInfo, retErr error) {
	return a.driver.batchTransaction(ctx, request.Requests)
}

func (a *apiServer) StartTransaction(ctx context.Context, request *transaction.StartTransactionRequest) (response *transaction.Transaction, retErr error) {
	return a.driver.startTransaction(ctx)
}

func (a *apiServer) InspectTransaction(ctx context.Context, request *transaction.InspectTransactionRequest) (response *transaction.TransactionInfo, retErr error) {
	return a.driver.inspectTransaction(ctx, request.Transaction)
}

func (a *apiServer) DeleteTransaction(ctx context.Context, request *transaction.DeleteTransactionRequest) (response *emptypb.Empty, retErr error) {
	err := a.driver.deleteTransaction(ctx, request.Transaction)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (a *apiServer) ListTransaction(ctx context.Context, request *transaction.ListTransactionRequest) (response *transaction.TransactionInfos, retErr error) {
	transactions, err := a.driver.listTransaction(ctx)
	if err != nil {
		return nil, err
	}
	return &transaction.TransactionInfos{TransactionInfo: transactions}, nil
}

func (a *apiServer) FinishTransaction(ctx context.Context, request *transaction.FinishTransactionRequest) (response *transaction.TransactionInfo, retErr error) {
	return a.driver.finishTransaction(ctx, request.Transaction)
}

func (a *apiServer) DeleteAll(ctx context.Context, request *transaction.DeleteAllRequest) (response *emptypb.Empty, retErr error) {
	if err := dbutil.WithTx(ctx, a.driver.db, func(ctx context.Context, sqlTx *pachsql.Tx) error {
		return a.driver.deleteAll(ctx, sqlTx, nil)
	}); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// AppendRequest is not an RPC, but is called from other systems in pachd to
// add an operation to an existing transaction.
func (a *apiServer) AppendRequest(ctx context.Context, txn *transaction.Transaction, request *transaction.TransactionRequest) (response *transaction.TransactionResponse, retErr error) {
	items := []*transaction.TransactionRequest{request}
	info, err := a.driver.appendTransaction(ctx, txn, items)
	if err != nil {
		return nil, err
	}

	return info.Responses[len(info.Responses)-1], nil
}
