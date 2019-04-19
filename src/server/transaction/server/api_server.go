package server

import (
	"sync"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/transaction"
	"github.com/pachyderm/pachyderm/src/server/pkg/log"
	"github.com/pachyderm/pachyderm/src/server/pkg/serviceenv"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var (
	grpcErrorf = grpc.Errorf // needed to get passed govet
)

type apiServer struct {
	log.Logger
	driver *driver

	// env generates clients for pachyderm's downstream services
	env *serviceenv.ServiceEnv
	// pachClientOnce ensures that _pachClient is only initialized once
	pachClientOnce sync.Once
	// pachClient is a cached Pachd client that connects to Pachyderm's object
	// store API and auth API. Instead of accessing it directly, functions should
	// call a.env.GetPachClient()
	_pachClient *client.APIClient
}

func newAPIServer(
	env *serviceenv.ServiceEnv,
	txnEnv *TransactionEnv,
	etcdPrefix string,
	memoryRequest int64,
) (*apiServer, error) {
	d, err := newDriver(env, txnEnv, etcdPrefix, memoryRequest)
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

	return a.driver.startTransaction(a.env.GetPachClient(ctx))
}

func (a *apiServer) InspectTransaction(ctx context.Context, request *transaction.InspectTransactionRequest) (response *transaction.TransactionInfo, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())

	return a.driver.inspectTransaction(a.env.GetPachClient(ctx), request.Transaction)
}

func (a *apiServer) DeleteTransaction(ctx context.Context, request *transaction.DeleteTransactionRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())

	err := a.driver.deleteTransaction(a.env.GetPachClient(ctx), request.Transaction)
	if err != nil {
		return nil, err
	}

	return &types.Empty{}, nil
}

func (a *apiServer) ListTransaction(ctx context.Context, request *transaction.ListTransactionRequest) (response *transaction.TransactionInfos, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())

	transactions, err := a.driver.listTransaction(a.env.GetPachClient(ctx))
	if err != nil {
		return nil, err
	}
	return &transaction.TransactionInfos{TransactionInfo: transactions}, nil
}

func (a *apiServer) FinishTransaction(ctx context.Context, request *transaction.FinishTransactionRequest) (response *transaction.TransactionInfo, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())

	return a.driver.finishTransaction(a.env.GetPachClient(ctx), request.Transaction)
}

func (a *apiServer) AppendTransaction(ctx context.Context, request *transaction.AppendTransactionRequest) (response *transaction.TransactionInfo, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())

	info, err := a.driver.appendTransaction(a.env.GetPachClient(ctx), request.Transaction, request.Items)
	if err != nil {
		return nil, err
	}

	return info, nil
}
