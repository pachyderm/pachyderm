package server

import (
	"context"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/errutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/transactiondb"
	txnenv "github.com/pachyderm/pachyderm/v2/src/internal/transactionenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv/txncontext"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"github.com/pachyderm/pachyderm/v2/src/transaction"
)

type driver struct {
	// txnEnv stores references to other pachyderm APIServer instances so we can
	// make calls within the same transaction without serializing through RPCs
	txnEnv       *txnenv.TransactionEnv
	db           *pachsql.DB
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

func (d *driver) batchTransaction(ctx context.Context, req []*transaction.TransactionRequest) (*transaction.TransactionInfo, error) {
	if err := d.txnEnv.PreTxOps(ctx, req); err != nil {
		return nil, errors.Wrap(err, "run pre-transaction operations")
	}
	var result *transaction.TransactionInfo
	if err := d.txnEnv.WithWriteContext(ctx, func(txnCtx *txncontext.TransactionContext) error {
		// Because we're building and running the entire transaction atomically here,
		// there is no need to persist the TransactionInfo to the collection
		info := &transaction.TransactionInfo{
			Transaction: &transaction.Transaction{
				Id: uuid.NewWithoutDashes(),
			},
			Requests: req,
			Started:  timestamppb.Now(),
		}

		var err error
		result, err = d.runTransaction(ctx, txnCtx, info)
		return err
	}); err != nil {
		return nil, err
	}

	return result, nil
}

func (d *driver) startTransaction(ctx context.Context) (*transaction.Transaction, error) {
	info := &transaction.TransactionInfo{
		Transaction: &transaction.Transaction{
			Id: uuid.NewWithoutDashes(),
		},
		Requests: []*transaction.TransactionRequest{},
		Started:  timestamppb.Now(),
	}

	if err := dbutil.WithTx(ctx, d.db, func(ctx context.Context, sqlTx *pachsql.Tx) error {
		return errors.EnsureStack(d.transactions.ReadWrite(sqlTx).Put(
			info.Transaction.Id,
			info,
		))
	}); err != nil {
		return nil, err
	}
	return info.Transaction, nil
}

func (d *driver) inspectTransaction(ctx context.Context, txn *transaction.Transaction) (*transaction.TransactionInfo, error) {
	info := &transaction.TransactionInfo{}
	if err := d.transactions.ReadOnly(ctx).Get(txn.Id, info); err != nil {
		return nil, errors.EnsureStack(err)
	}
	return info, nil
}

func (d *driver) deleteTransaction(ctx context.Context, txn *transaction.Transaction) error {
	return d.txnEnv.WithWriteContext(ctx, func(txnCtx *txncontext.TransactionContext) error {
		return errors.EnsureStack(d.transactions.ReadWrite(txnCtx.SqlTx).Delete(txn.Id))
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
		return nil, errors.EnsureStack(err)
	}
	return result, nil
}

// deleteAll deletes all transactions from etcd except the currently running
// transaction (if any).
func (d *driver) deleteAll(ctx context.Context, sqlTx *pachsql.Tx, running *transaction.Transaction) error {
	txns, err := d.listTransaction(ctx)
	if err != nil {
		return err
	}

	transactions := d.transactions.ReadWrite(sqlTx)
	for _, info := range txns {
		if running == nil || info.Transaction.Id != running.Id {
			err := transactions.Delete(info.Transaction.Id)
			if err != nil {
				return errors.EnsureStack(err)
			}
		}
	}
	return nil
}

func (d *driver) runTransaction(ctx context.Context, txnCtx *txncontext.TransactionContext, info *transaction.TransactionInfo) (*transaction.TransactionInfo, error) {
	result := proto.Clone(info).(*transaction.TransactionInfo)
	for len(result.Responses) < len(result.Requests) {
		result.Responses = append(result.Responses, &transaction.TransactionResponse{})
	}

	// Set the transaction's CommitSetID to be the same as the transaction ID, which
	// will be used for any newly made commits.
	txnCtx.CommitSetID = info.Transaction.Id

	directTxn := txnenv.NewDirectTransaction(ctx, d.txnEnv, txnCtx)
	for i, request := range info.Requests {
		var err error
		response := result.Responses[i]

		if request.CreateRepo != nil {
			err = directTxn.CreateRepo(request.CreateRepo)
		} else if request.DeleteRepo != nil {
			err = directTxn.DeleteRepo(request.DeleteRepo)
		} else if request.StartCommit != nil {
			response.Commit, err = directTxn.StartCommit(request.StartCommit)
		} else if request.FinishCommit != nil {
			err = directTxn.FinishCommit(request.FinishCommit)
		} else if request.SquashCommitSet != nil {
			err = directTxn.SquashCommitSet(request.SquashCommitSet)
		} else if request.CreateBranch != nil {
			err = directTxn.CreateBranch(request.CreateBranch)
		} else if request.DeleteBranch != nil {
			err = directTxn.DeleteBranch(request.DeleteBranch)
		} else if request.UpdateJobState != nil {
			err = directTxn.UpdateJobState(request.UpdateJobState)
		} else if request.StopJob != nil {
			err = directTxn.StopJob(request.StopJob)
		} else if request.CreatePipelineV2 != nil {
			err = directTxn.CreatePipeline(request.CreatePipelineV2)
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
	return d.updateTransaction(ctx, true, txn, func(txnCtx *txncontext.TransactionContext, info *transaction.TransactionInfo, restarted bool) (*transaction.TransactionInfo, error) {
		info, err := d.runTransaction(ctx, txnCtx, info)
		if err != nil {
			return info, err
		}
		if err := d.transactions.ReadWrite(txnCtx.SqlTx).Delete(txn.Id); err != nil {
			return info, errors.EnsureStack(err)
		}
		// no need to update the transaction, since it's gone
		// because the transaction info was read in the same sql transaction as the delete,
		// we don't have to worry about checking for additional transaction changes
		return info, errutil.ErrBreak
	})
}

// Error to be returned when the transaction has been modified between our two sqlTx calls
type transactionModifiedError struct{}

func (e *transactionModifiedError) Error() string {
	return "transaction could not be modified due to concurrent modifications"
}

func (d *driver) appendTransaction(
	ctx context.Context,
	txn *transaction.Transaction,
	items []*transaction.TransactionRequest,
) (*transaction.TransactionInfo, error) {
	if err := d.txnEnv.PreTxOps(ctx, items); err != nil {
		return nil, errors.Wrap(err, "run pre-transaction operations")
	}
	// We do a dryrun of the transaction to
	// 1. make sure the appended request is valid
	// 2. Capture the result of the request to be returned
	return d.updateTransaction(ctx, false, txn, func(txnCtx *txncontext.TransactionContext, info *transaction.TransactionInfo, restarted bool) (*transaction.TransactionInfo, error) {
		if restarted {
			info.Requests = append(info.Requests, items...)
		}
		return d.runTransaction(ctx, txnCtx, info)
	})
}

// updateTransaction accepts a function that uses and (possibly) updates a transaction and runs it
// in either a Read- or WriteContext, saving an updated version of the transaction on a successful run
func (d *driver) updateTransaction(
	ctx context.Context,
	writeTxn bool,
	txn *transaction.Transaction,
	f func(txnCtx *txncontext.TransactionContext, info *transaction.TransactionInfo, restarted bool) (*transaction.TransactionInfo, error),
) (*transaction.TransactionInfo, error) {
	// we want to respect ErrBreak, so we have to smuggle some information out of the transaction
	var gotBreak bool
	// local info will hold a version of the transaction with any modifications needed to get the current version to run
	// it will only be written to the collection after the update operation is successful
	var localInfo *transaction.TransactionInfo
	attempt := func(txnCtx *txncontext.TransactionContext) error {
		storedInfo := new(transaction.TransactionInfo)
		var err error
		if err := d.transactions.ReadWrite(txnCtx.SqlTx).Get(txn.Id, storedInfo); err != nil {
			return errors.EnsureStack(err)
		}
		restarted := localInfo == nil || storedInfo.Version != localInfo.Version
		if restarted {
			// something has changed, reset our saved info
			localInfo = storedInfo
		}
		// update local info even if there was an error
		localInfo, err = f(txnCtx, localInfo, restarted)
		if err != nil && errors.Is(err, errutil.ErrBreak) {
			gotBreak = true
			return nil
		}
		gotBreak = false
		return err
	}

	// prefetch transaction info and add data to refresher ahead of time
	var prefetch transaction.TransactionInfo
	if err := d.txnEnv.WithReadContext(ctx, func(txnCtx *txncontext.TransactionContext) error {
		return errors.EnsureStack(d.transactions.ReadWrite(txnCtx.SqlTx).Get(txn.Id, &prefetch))
	}); err != nil {
		return nil, err
	}

	// Run this thing in a loop in case we get a conflict, time out after some tries
	for i := 0; i < 10; i++ {
		var err error
		if writeTxn {
			err = d.txnEnv.WithWriteContext(ctx, attempt)
		} else {
			err = d.txnEnv.WithReadContext(ctx, attempt)
		}
		if err == nil && gotBreak {
			return localInfo, nil // no need to update
		}

		if err == nil {
			// only persist the transaction if we succeeded, otherwise just update localInfo
			var storedInfo transaction.TransactionInfo
			if err = dbutil.WithTx(ctx, d.db, func(ctx context.Context, sqlTx *pachsql.Tx) error {
				// Update the existing transaction with the new requests/responses
				err := d.transactions.ReadWrite(sqlTx).Update(txn.Id, &storedInfo, func() error {
					if storedInfo.Version != localInfo.Version {
						return &transactionModifiedError{}
					}
					storedInfo.Requests = localInfo.Requests
					storedInfo.Responses = localInfo.Responses
					storedInfo.Version += 1
					return nil
				})
				return errors.EnsureStack(err)
			}); err == nil {
				// update succeeded, put the incremented version in the returned info
				localInfo.Version = storedInfo.Version
			}
		}
		// if we got a transactionModifiedError trying to update the stored TransactionInfo, just try again
		if err == nil {
			return localInfo, nil
		} else if !errors.Is(err, &transactionModifiedError{}) {
			return nil, err
		}
	}
	return nil, &transactionModifiedError{}
}
