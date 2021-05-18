package server

import (
	"context"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	"github.com/jmoiron/sqlx"

	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/transactiondb"
	txnenv "github.com/pachyderm/pachyderm/v2/src/internal/transactionenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv/context"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"github.com/pachyderm/pachyderm/v2/src/transaction"
)

type driver struct {
	// txnEnv stores references to other pachyderm APIServer instances so we can
	// make calls within the same transaction without serializing through RPCs
	txnEnv       *txnenv.TransactionEnv
	db           *sqlx.DB
	transactions col.PostgresCollection
}

func newDriver(
	env serviceenv.ServiceEnv,
	txnEnv *txnenv.TransactionEnv,
) (*driver, error) {

	return &driver{
		txnEnv:       txnEnv,
		db:           env.GetDBClient(),
		transactions: transactiondb.Transactions(env.GetDBClient(), env.GetPostgresListener()),
	}, nil
}

func now() *types.Timestamp {
	t, err := types.TimestampProto(time.Now())
	if err != nil {
		return &types.Timestamp{}
	}
	return t
}

func (d *driver) batchTransaction(ctx context.Context, req []*transaction.TransactionRequest) (*transaction.TransactionInfo, error) {
	var result *transaction.TransactionInfo
	if err := d.txnEnv.WithWriteContext(ctx, func(txnCtx *txncontext.TransactionContext) error {
		// Because we're building and running the entire transaction atomically here,
		// there is no need to persist the TransactionInfo to the collection
		info := &transaction.TransactionInfo{
			Transaction: &transaction.Transaction{
				ID: uuid.New(),
			},
			Requests: req,
			Started:  now(),
		}

		var err error
		result, err = d.runTransaction(txnCtx, info)
		return err
	}); err != nil {
		return nil, err
	}

	return result, nil
}

func (d *driver) startTransaction(ctx context.Context) (*transaction.Transaction, error) {
	info := &transaction.TransactionInfo{
		Transaction: &transaction.Transaction{
			ID: uuid.New(),
		},
		Requests: []*transaction.TransactionRequest{},
		Started:  now(),
	}

	if err := col.NewSQLTx(ctx, d.db, func(sqlTx *sqlx.Tx) error {
		return d.transactions.ReadWrite(sqlTx).Put(
			info.Transaction.ID,
			info,
		)
	}); err != nil {
		return nil, err
	}
	return info.Transaction, nil
}

func (d *driver) inspectTransaction(ctx context.Context, txn *transaction.Transaction) (*transaction.TransactionInfo, error) {
	info := &transaction.TransactionInfo{}
	if err := d.transactions.ReadOnly(ctx).Get(txn.ID, info); err != nil {
		return nil, err
	}
	return info, nil
}

func (d *driver) deleteTransaction(ctx context.Context, txn *transaction.Transaction) error {
	return d.txnEnv.WithWriteContext(ctx, func(txnCtx *txncontext.TransactionContext) error {
		return d.transactions.ReadWrite(txnCtx.SqlTx).Delete(txn.ID)
	})
}

func (d *driver) listTransaction(ctx context.Context) ([]*transaction.TransactionInfo, error) {
	var result []*transaction.TransactionInfo
	transactionInfo := &transaction.TransactionInfo{}
	transactions := d.transactions.ReadOnly(ctx)
	if err := transactions.List(transactionInfo, col.DefaultOptions(), func(string) error {
		result = append(result, proto.Clone(transactionInfo).(*transaction.TransactionInfo))
		return nil
	}); err != nil {
		return nil, err
	}
	return result, nil
}

// deleteAll deletes all transactions from etcd except the currently running
// transaction (if any).
func (d *driver) deleteAll(ctx context.Context, sqlTx *sqlx.Tx, running *transaction.Transaction) error {
	txns, err := d.listTransaction(ctx)
	if err != nil {
		return err
	}

	transactions := d.transactions.ReadWrite(sqlTx)
	for _, info := range txns {
		if running == nil || info.Transaction.ID != running.ID {
			err := transactions.Delete(info.Transaction.ID)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (d *driver) runTransaction(txnCtx *txncontext.TransactionContext, info *transaction.TransactionInfo) (*transaction.TransactionInfo, error) {
	result := proto.Clone(info).(*transaction.TransactionInfo)
	for len(result.Responses) < len(result.Requests) {
		result.Responses = append(result.Responses, &transaction.TransactionResponse{})
	}

	directTxn := txnenv.NewDirectTransaction(d.txnEnv, txnCtx)
	for i, request := range info.Requests {
		var err error
		response := result.Responses[i]

		if request.CreateRepo != nil {
			err = directTxn.CreateRepo(request.CreateRepo)
		} else if request.DeleteRepo != nil {
			err = directTxn.DeleteRepo(request.DeleteRepo)
		} else if request.StartCommit != nil {
			// Do a little extra work here so we can make sure the new commit ID is
			// the same every time.  We store the response the first time and reuse
			// the commit ID on subsequent runs.
			response.Commit, err = directTxn.StartCommit(request.StartCommit, response.Commit)
		} else if request.FinishCommit != nil {
			err = directTxn.FinishCommit(request.FinishCommit)
		} else if request.SquashCommit != nil {
			err = directTxn.SquashCommit(request.SquashCommit)
		} else if request.CreateBranch != nil {
			err = directTxn.CreateBranch(request.CreateBranch)
		} else if request.DeleteBranch != nil {
			err = directTxn.DeleteBranch(request.DeleteBranch)
		} else if request.UpdatePipelineJobState != nil {
			err = directTxn.UpdatePipelineJobState(request.UpdatePipelineJobState)
		} else if request.DeleteAll != nil {
			// TODO: extend this to delete everything through PFS, PPS, Auth and
			// update the client DeleteAll call to use only this, then remove unused
			// RPCs.  This is not currently feasible because it does an orderly
			// deletion that generates a very large transaction.
			err = d.deleteAll(txnCtx.ClientContext, txnCtx.SqlTx, info.Transaction)
		} else if request.CreatePipeline != nil {
			if response.CreatePipelineResponse == nil {
				response.CreatePipelineResponse = &transaction.CreatePipelineTransactionResponse{}
			}
			filesetID := &response.CreatePipelineResponse.FilesetId
			prevSpecCommit := &response.CreatePipelineResponse.PrevSpecCommit

			// CreatePipeline may update the fileset and prevSpecCommit even if it
			// fails (because these refer to things outside of the transaction) - we
			// need to save them into the response so they can be seen the next time
			// the transaction is attempted.
			err = directTxn.CreatePipeline(request.CreatePipeline, filesetID, prevSpecCommit)
		} else {
			err = errors.New("unrecognized transaction request type")
		}

		if err != nil {
			return result, errors.Wrapf(err, "error running request %d of %d", i+1, len(info.Requests))
		}
	}
	return result, nil
}

func (d *driver) finishTransaction(ctx context.Context, txn *transaction.Transaction) (*transaction.TransactionInfo, error) {
	info := &transaction.TransactionInfo{}
	if err := d.txnEnv.WithWriteContext(ctx, func(txnCtx *txncontext.TransactionContext) error {
		err := d.transactions.ReadOnly(ctx).Get(txn.ID, info)
		if err != nil {
			return err
		}
		info, err = d.runTransaction(txnCtx, info)
		if err != nil {
			return err
		}
		return d.transactions.ReadWrite(txnCtx.SqlTx).Delete(txn.ID)
	}); err != nil {
		return nil, err
	}
	return info, nil
}

// Error to be returned when the transaction has been modified between our two sqlTx calls
type transactionConflictError struct{}

func (e *transactionConflictError) Error() string {
	return "transaction could not be modified due to concurrent modifications"
}

func (d *driver) appendTransaction(
	ctx context.Context,
	txn *transaction.Transaction,
	items []*transaction.TransactionRequest,
) (*transaction.TransactionInfo, error) {
	// Run this thing in a loop in case we get a conflict, time out after some tries
	for i := 0; i < 10; i++ {
		// We first do a dryrun of the transaction to
		// 1. make sure the appended request is valid
		// 2. Capture the result of the request to be returned
		var numRequests, numResponses int
		var newResponses []*transaction.TransactionResponse

		if err := d.txnEnv.WithReadContext(ctx, func(txnCtx *txncontext.TransactionContext) error {
			// Get the existing transaction and append the new requests
			info := &transaction.TransactionInfo{}
			err := d.transactions.ReadWrite(txnCtx.SqlTx).Get(txn.ID, info)
			if err != nil {
				return err
			}

			// Save the length so that we can check that nothing else modifies the
			// transaction in the meantime
			numRequests = len(info.Requests)
			numResponses = len(info.Responses)
			info.Requests = append(info.Requests, items...)
			if newResponses != nil {
				info.Responses = newResponses
			}

			// This may error with a postgres-level transaction conflict, in which
			// case we want to save the generated responses and reattempt.  The saved
			// responses will contain any references to objects generated outside of
			// the transaction and need to be reused on the next attempt.
			info, err = d.runTransaction(txnCtx, info)
			newResponses = info.Responses
			return err
		}); err != nil {
			return nil, err
		}

		info := &transaction.TransactionInfo{}
		if err := col.NewSQLTx(ctx, d.db, func(sqlTx *sqlx.Tx) error {
			// Update the existing transaction with the new requests/responses
			return d.transactions.ReadWrite(sqlTx).Update(txn.ID, info, func() error {
				if len(info.Requests) != numRequests || len(info.Responses) != numResponses {
					// Someone else modified the transaction while we did the dry run
					return &transactionConflictError{}
				}

				info.Requests = append(info.Requests, items...)
				info.Responses = newResponses
				return nil
			})
		}); err == nil {
			return info, nil
		} else if !errors.As(err, &transactionConflictError{}) {
			return nil, err
		}
	}
	return nil, &transactionConflictError{}
}
