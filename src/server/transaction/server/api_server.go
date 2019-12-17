package server

import (
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/src/client/transaction"
	col "github.com/pachyderm/pachyderm/src/server/pkg/collection"
	"github.com/pachyderm/pachyderm/src/server/pkg/log"
	"github.com/pachyderm/pachyderm/src/server/pkg/serviceenv"
	txnenv "github.com/pachyderm/pachyderm/src/server/pkg/transactionenv"

	"golang.org/x/net/context"
)

type apiServer struct {
	log.Logger
	driver *driver

	// env generates clients for pachyderm's downstream services
	env *serviceenv.ServiceEnv
}

func newAPIServer(
	env *serviceenv.ServiceEnv,
	txnEnv *txnenv.TransactionEnv,
	etcdPrefix string,
) (*apiServer, error) {
	d, err := newDriver(env, txnEnv, etcdPrefix)
	if err != nil {
		return nil, err
	}
	s := &apiServer{
		Logger: log.NewLogger("transaction.API"),
		driver: d,
		env:    env,
	}
	go func() { s.env.GetPachClient(context.Background()) }() // Begin dialing connection on startup
	return s, nil
}

func (a *apiServer) StartTransaction(ctx context.Context, request *transaction.StartTransactionRequest) (response *transaction.Transaction, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())

	return a.driver.startTransaction(ctx)
}

func (a *apiServer) InspectTransaction(ctx context.Context, request *transaction.InspectTransactionRequest) (response *transaction.TransactionInfo, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())

	return a.driver.inspectTransaction(ctx, request.Transaction)
}

func (a *apiServer) DeleteTransaction(ctx context.Context, request *transaction.DeleteTransactionRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())

	err := a.driver.deleteTransaction(ctx, request.Transaction)
	if err != nil {
		return nil, err
	}

	return &types.Empty{}, nil
}

func (a *apiServer) ListTransaction(ctx context.Context, request *transaction.ListTransactionRequest) (response *transaction.TransactionInfos, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())

	transactions, err := a.driver.listTransaction(ctx)
	if err != nil {
		return nil, err
	}
	return &transaction.TransactionInfos{TransactionInfo: transactions}, nil
}

func (a *apiServer) FinishTransaction(ctx context.Context, request *transaction.FinishTransactionRequest) (response *transaction.TransactionInfo, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())

	return a.driver.finishTransaction(ctx, request.Transaction)
}

func (a *apiServer) DeleteAll(ctx context.Context, request *transaction.DeleteAllRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())

	_, err := col.NewSTM(ctx, a.driver.etcdClient, func(stm col.STM) error {
		return a.driver.deleteAll(ctx, stm, nil)
	})
	if err != nil {
		return nil, err
	}

	return &types.Empty{}, nil
}

// AppendRequest is not an RPC, but is called from other systems in pachd to
// add an operation to an existing transaction.
func (a *apiServer) AppendRequest(ctx context.Context, txn *transaction.Transaction, request *transaction.TransactionRequest) (response *transaction.TransactionResponse, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())

	items := []*transaction.TransactionRequest{request}
	info, err := a.driver.appendTransaction(ctx, txn, items)
	if err != nil {
		return nil, err
	}

	return info.Responses[len(info.Responses)-1], nil
}
