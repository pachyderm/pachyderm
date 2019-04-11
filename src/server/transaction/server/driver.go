package server

import (
	"fmt"
	"time"

	"golang.org/x/sync/semaphore"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/transaction"
	authserver "github.com/pachyderm/pachyderm/src/server/auth/server"
	pfsserver "github.com/pachyderm/pachyderm/src/server/pfs/server"
	col "github.com/pachyderm/pachyderm/src/server/pkg/collection"
	"github.com/pachyderm/pachyderm/src/server/pkg/serviceenv"
	"github.com/pachyderm/pachyderm/src/server/pkg/transactiondb"
	"github.com/pachyderm/pachyderm/src/server/pkg/uuid"
	ppsserver "github.com/pachyderm/pachyderm/src/server/pps/server"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
)

type driver struct {
	// etcdClient and prefix write repo and other metadata to etcd
	etcdClient *etcd.Client
	prefix     string

	pfsServer  *pfsserver.APIServer
	ppsServer  *ppsserver.APIServer
	authServer *authserver.APIServer

	// collections
	transactions col.Collection

	// memory limiter (useful for limiting operations that could use a lot of memory)
	memoryLimiter *semaphore.Weighted
}

// newDriver
func newDriver(
	env *serviceenv.ServiceEnv,
	etcdPrefix string,
	pfsServer *pfsserver.APIServer,
	ppsServer *ppsserver.APIServer,
	authServer *authserver.APIServer,
	memoryRequest int64,
) (*driver, error) {
	// Initialize driver
	etcdClient := env.GetEtcdClient()
	d := &driver{
		etcdClient:   etcdClient,
		prefix:       etcdPrefix,
		pfsServer:    pfsServer,
		ppsServer:    ppsServer,
		authServer:   authServer,
		transactions: transactiondb.Transactions(etcdClient, etcdPrefix),
		// Allow up to a third of the requested memory to be used for memory intensive operations
		memoryLimiter: semaphore.NewWeighted(memoryRequest / 3),
	}
	return d, nil
}

func now() *types.Timestamp {
	t, err := types.TimestampProto(time.Now())
	if err != nil {
		panic(err)
	}
	return t
}

func (d *driver) startTransaction(pachClient *client.APIClient) (*transaction.Transaction, error) {
	ctx := pachClient.Ctx()
	info := &transaction.TransactionInfo{
		Transaction: &transaction.Transaction{
			ID: uuid.New(),
		},
		Requests: []*transaction.TransactionRequest{},
		Started:  now(),
	}

	_, err := col.NewSTM(ctx, d.etcdClient, func(stm col.STM) error {
		return d.transactions.ReadWrite(stm).Put(
			info.Transaction.ID,
			info,
		)
	})
	if err != nil {
		return nil, err
	}
	return info.Transaction, nil
}

func (d *driver) inspectTransaction(pachClient *client.APIClient, txn *transaction.Transaction) (*transaction.TransactionInfo, error) {
	ctx := pachClient.Ctx()
	info := &transaction.TransactionInfo{}
	err := d.transactions.ReadOnly(ctx).Get(txn.ID, info)
	if err != nil {
		return nil, err
	}
	return info, nil
}

func (d *driver) deleteTransaction(pachClient *client.APIClient, txn *transaction.Transaction) error {
	ctx := pachClient.Ctx()
	_, err := col.NewSTM(ctx, d.etcdClient, func(stm col.STM) error {
		return d.transactions.ReadWrite(stm).Delete(txn.ID)
	})
	return err
}

func (d *driver) listTransaction(pachClient *client.APIClient) ([]*transaction.TransactionInfo, error) {
	var result []*transaction.TransactionInfo
	transactionInfo := &transaction.TransactionInfo{}
	transactions := d.transactions.ReadOnly(pachClient.Ctx())
	if err := transactions.List(transactionInfo, col.DefaultOptions, func(string) error {
		result = append(result, proto.Clone(transactionInfo).(*transaction.TransactionInfo))
		return nil
	}); err != nil {
		return nil, err
	}
	return result, nil
}

func (d *driver) runTransaction(pachClient *client.APIClient, stm col.STM, txnInfo *transaction.TransactionInfo) (*transaction.TransactionInfo, error) {
	responses := []*transaction.TransactionResponse{}
	for i, request := range txnInfo.Requests {
		var err error
		var response *transaction.TransactionResponse
		switch request.Request.(type) {
		case *transaction.TransactionRequest_CreateRepo:
			err = d.pfsServer.CreateRepoInTransaction(pachClient, stm, request.Request)
			response = client.NewEmptyResponse()
		case *transaction.TransactionRequest_DeleteRepo:
			err = d.pfsServer.DeleteRepoInTransaction(pachClient, stm, request.Request)
			response = client.NewEmptyResponse()
		case *transaction.TransactionRequest_StartCommit:
			var commit *pfs.Commit
			commit, err = d.pfsServer.StartCommitInTransaction(pachClient, stm, request.Request)
			response = client.NewCommitResponse(commit)
		case *transaction.TransactionRequest_FinishCommit:
			result, err = d.pfsServer.FinishCommitInTransaction(pachClient, stm, request.Request)
			response = client.NewEmptyResponse()
		case *transaction.TransactionRequest_DeleteCommit:
			result, err = d.pfsServer.DeleteCommitInTransaction(pachClient, stm, request.Request)
			response = client.NewEmptyResponse()
		case *transaction.TransactionRequest_CreateBranch:
			result, err = d.pfsServer.CreateBranchInTransaction(pachClient, stm, request.Request)
			response = client.NewEmptyResponse()
		case *transaction.TransactionRequest_DeleteBranch:
			result, err = d.pfsServer.DeleteBranchInTransaction(pachClient, stm, request.Request)
			response = client.NewEmptyResponse()
		case *transaction.TransactionRequest_PutFile:
			result, err = d.pfsServer.PutFileInTransaction(pachClient, stm, request.Request)
			response = client.NewEmptyResponse()
		case *transaction.TransactionRequest_CopyFile:
			result, err = d.pfsServer.CopyFileInTransaction(pachClient, stm, request.Request)
			response = client.NewEmptyResponse()
		case *transaction.TransactionRequest_DeleteFile:
			result, err = d.pfsServer.DeleteFileInTransaction(pachClient, stm, request.Request)
			response = client.NewEmptyResponse()
		case *transaction.TransactionRequest_DeleteAll:
			err = fmt.Errorf("not yet implemented")
			/*
				_, err = d.authServer.RunInTransaction(pachClient, stm, &auth.DeactivateRequest{})
				if err == nil {
					_, err = d.pfsServer.RunInTransaction(pachClient, stm, &pfs.DeleteAllRequest{})
				}
				if err == nil {
					_, err = d.ppsServer.RunInTransaction(pachClient, stm, &pps.DeleteAllRequest{})
				}
			*/
		case *transaction.TransactionRequest_CreatePipeline:
			result, err = ppsClient.RunInTransaction(pachClient, stm, request.Request)
			response = client.NewEmptyResponse()
		default:
			err = fmt.Errorf("unrecognized transaction request type")
		}

		if err != nil {
			return nil, fmt.Errorf("error running request %d of %d: %v", err)
		}

		responses = append(responses, response)
	}

	// TODO: check that responses match the cached responses in the transaction info
}

func (d *driver) finishTransaction(pachClient *client.APIClient, txn *transaction.Transaction) (*transaction.TransactionInfo, error) {
	ctx := pachClient.Ctx()
	info := &transaction.TransactionInfo{}
	_, err := col.NewSTM(ctx, d.etcdClient, func(stm col.STM) error {
		err := d.transactions.ReadOnly(ctx).Get(txn.ID, info)
		if err != nil {
			return err
		}
		info, err = d.runTransaction(pachClient, stm, info)
		return err
	})
	if err != nil {
		return nil, err
	}
	return info, nil
}

// Error to be returned when aborting a dryrun transaction (following success)
type TransactionDryrunError struct{}

func (e *TransactionDryrunError) Error() string {
	return fmt.Sprintf("transaction dry run successful")
}

// Error to be returned when the transaction has been modified between our two STM calls
type TransactionConflictError struct{}

func (e *TransactionConflictError) Error() string {
	return fmt.Sprintf("transaction could not be modified due to concurrent modifications")
}

func (d *driver) appendTransaction(
	pachClient *client.APIClient,
	txn *transaction.Transaction,
	items []*transaction.TransactionRequest,
) (*transaction.TransactionInfo, error) {
	ctx := pachClient.Ctx()

	// Run this thing in a loop in case we get a conflict, time out after some tries
	for i := 0; i < 10; i++ {
		// We first do a dryrun of the transaction to
		// 1. make sure the appended request is valid
		// 2. Capture the result of the request to be returned
		var transactionLength int
		var dryrunResponses []*transaction.TransactionResponse

		info := &transaction.TransactionInfo{}
		_, err := col.NewSTM(ctx, d.etcdClient, func(stm col.STM) error {
			transactions := d.transactions.ReadWrite(stm)

			// Get the existing transaction and append the new requests
			err := transactions.Get(txn.ID, info)
			if err != nil {
				return err
			}

			// Save the length so that we can check that nothing else modifies the
			// transaction in the meantime
			transactionLength = len(info.Requests)
			info.Requests = append(info.Requests, items...)

			info, err = d.runTransaction(pachClient, stm, info)
			if err != nil {
				return err
			}

			dryrunResponses = info.Responses[transactionLength:]
			return &TransactionDryrunError{}
		})

		info = &transaction.TransactionInfo{}
		_, err = col.NewSTM(ctx, d.etcdClient, func(stm col.STM) error {
			transactions := d.transactions.ReadWrite(stm)

			// Get the existing transaction, append the new requests, save it back
			err := transactions.Get(txn.ID, info)
			if err != nil {
				return err
			}

			if len(info.Requests) != transactionLength {
				return &TransactionConflictError{}
			}

			info.Requests = append(info.Requests, items...)
			info.Responses = append(info.Responses, dryrunResponses...)

			return transactions.Put(txn.ID, info)
		})

		if err == nil {
			return info, nil
		}
		switch err.(type) {
		case *TransactionConflictError: // pass
		default:
			return nil, err
		}
	}
	return nil, &TransactionConflictError{}
}
