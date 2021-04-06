package server

import (
	"context"
	"fmt"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	"github.com/jmoiron/sqlx"

	"github.com/pachyderm/pachyderm/v2/src/client"
	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/transactiondb"
	txnenv "github.com/pachyderm/pachyderm/v2/src/internal/transactionenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
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
	env *serviceenv.ServiceEnv,
	txnEnv *txnenv.TransactionEnv,
) (*driver, error) {
	db := env.GetDBClient()
	transactions, err := transactiondb.Transactions(context.Background(), db, env.GetPostgresListener())
	if err != nil {
		return nil, err
	}

	return &driver{
		txnEnv:       txnEnv,
		db:           env.GetDBClient(),
		transactions: transactions,
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
	err = d.txnEnv.WithWriteContext(ctx, func(txnCtx *txnenv.TransactionContext) error {
		info, err = d.runTransaction(txnCtx, info)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return info, nil
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
	return d.txnEnv.WithWriteContext(ctx, func(txnCtx *txnenv.TransactionContext) error {
		// first try to clean up any aspects of the transaction that live outside the sqlTx
		info := &transaction.TransactionInfo{}
		if err := d.transactions.ReadOnly(ctx).Get(txn.ID, info); err != nil {
			return err
		}
		directTxn := txnenv.NewDirectTransaction(txnCtx)
		for i, req := range info.Requests {
			if req.CreatePipeline == nil || len(info.Responses) <= i {
				continue
			}
			commit := info.Responses[i].Commit
			if commit != nil {
				if err := directTxn.SquashCommit(&pfs.SquashCommitRequest{Commit: commit}); err != nil {
					return err
				}
			}
		}
		// now that all side effects are undone, delete the transaction
		return d.transactions.ReadWrite(txnCtx.SqlTx).Delete(txn.ID)
	})
}

func (d *driver) listTransaction(ctx context.Context) ([]*transaction.TransactionInfo, error) {
	var result []*transaction.TransactionInfo
	transactionInfo := &transaction.TransactionInfo{}
	transactions := d.transactions.ReadOnly(ctx)
	if err := transactions.List(transactionInfo, col.DefaultOptions(), func() error {
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

func (d *driver) runTransaction(txnCtx *txnenv.TransactionContext, info *transaction.TransactionInfo) (*transaction.TransactionInfo, error) {
	responses := []*transaction.TransactionResponse{}
	directTxn := txnenv.NewDirectTransaction(txnCtx)
	for i, request := range info.Requests {
		var err error
		var response *transaction.TransactionResponse

		if request.CreateRepo != nil {
			err = directTxn.CreateRepo(request.CreateRepo)
			response = &transaction.TransactionResponse{}
		} else if request.DeleteRepo != nil {
			err = directTxn.DeleteRepo(request.DeleteRepo)
			response = &transaction.TransactionResponse{}
		} else if request.StartCommit != nil {
			// Do a little extra work here so we can make sure the new commit ID is
			// the same every time.  We store the response the first time and reuse
			// the commit ID on subsequent runs.
			var commit *pfs.Commit
			if len(info.Responses) > i {
				commit = info.Responses[i].Commit
				if commit == nil {
					err = errors.Errorf("unexpected stored response type for StartCommit")
				}
			}
			if err == nil {
				commit, err = directTxn.StartCommit(request.StartCommit, commit)
				response = client.NewCommitResponse(commit)
			}
		} else if request.FinishCommit != nil {
			err = directTxn.FinishCommit(request.FinishCommit)
			response = &transaction.TransactionResponse{}
		} else if request.SquashCommit != nil {
			err = directTxn.SquashCommit(request.SquashCommit)
			response = &transaction.TransactionResponse{}
		} else if request.CreateBranch != nil {
			err = directTxn.CreateBranch(request.CreateBranch)
			response = &transaction.TransactionResponse{}
		} else if request.DeleteBranch != nil {
			err = directTxn.DeleteBranch(request.DeleteBranch)
			response = &transaction.TransactionResponse{}
		} else if request.UpdateJobState != nil {
			err = directTxn.UpdateJobState(request.UpdateJobState)
			response = &transaction.TransactionResponse{}
		} else if request.DeleteAll != nil {
			// TODO: extend this to delete everything through PFS, PPS, Auth and
			// update the client DeleteAll call to use only this, then remove unused
			// RPCs.
			err = d.deleteAll(txnCtx.ClientContext, txnCtx.SqlTx, info.Transaction)
			response = &transaction.TransactionResponse{}
		} else if request.CreatePipeline != nil {
			// find the existing spec commit, if it exists
			var commit *pfs.Commit
			if len(info.Responses) > i {
				commit = info.Responses[i].Commit
				if commit == nil {
					err = errors.Errorf("unexpected stored response type for CreatePipeline")
				}
			}
			if err == nil {
				err = directTxn.CreatePipeline(request.CreatePipeline, &commit)
				response = client.NewCommitResponse(commit)
			}
		} else {
			err = errors.Errorf("unrecognized transaction request type")
		}

		if err != nil {
			return nil, errors.Wrapf(err, "error running request %d of %d", i+1, len(info.Requests))
		}

		responses = append(responses, response)
	}

	info.Responses = responses
	return info, nil
}

func (d *driver) finishTransaction(ctx context.Context, txn *transaction.Transaction) (*transaction.TransactionInfo, error) {
	info := &transaction.TransactionInfo{}
	err := d.txnEnv.WithWriteContext(ctx, func(txnCtx *txnenv.TransactionContext) error {
		err := d.transactions.ReadOnly(ctx).Get(txn.ID, info)
		if err != nil {
			return err
		}
		info, err = d.runTransaction(txnCtx, info)
		if err != nil {
			return err
		}
		return d.transactions.ReadWrite(txnCtx.SqlTx).Delete(txn.ID)
	})
	if err != nil {
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
		var dryrunResponses []*transaction.TransactionResponse

		info := &transaction.TransactionInfo{}
		if err := d.txnEnv.WithReadContext(ctx, func(txnCtx *txnenv.TransactionContext) error {
			// Get the existing transaction and append the new requests
			err := d.transactions.ReadWrite(txnCtx.SqlTx).Get(txn.ID, info)
			if err != nil {
				return err
			}

			// Save the length so that we can check that nothing else modifies the
			// transaction in the meantime
			numRequests = len(info.Requests)
			numResponses = len(info.Responses)
			info.Requests = append(info.Requests, items...)

			info, err = d.runTransaction(txnCtx, info)
			if err != nil {
				return err
			}

			dryrunResponses = info.Responses[numResponses:]
			return nil
		}); err != nil {
			return nil, err
		}

		info = &transaction.TransactionInfo{}
		if err := col.NewSQLTx(ctx, d.db, func(sqlTx *sqlx.Tx) error {
			// Update the existing transaction with the new requests/responses
			return d.transactions.ReadWrite(sqlTx).Update(txn.ID, info, func() error {
				if len(info.Requests) != numRequests || len(info.Responses) != numResponses {
					// Someone else modified the transaction while we did the dry run
					return &transactionConflictError{}
				}

				fmt.Printf("appending requests: %s\n", items)
				fmt.Printf("appending responses: %s\n", dryrunResponses)
				info.Requests = append(info.Requests, items...)
				info.Responses = append(info.Responses, dryrunResponses...)
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
