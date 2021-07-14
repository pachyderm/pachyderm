package cmds

import (
	"fmt"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/config"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/transaction"
)

// getActiveTransaction will read the active transaction from the config file
// (if it exists) and return it.  If the config file is uninitialized or the
// active transaction is unset, `nil` will be returned.
func getActiveTransaction() (*transaction.Transaction, error) {
	cfg, err := config.Read(false, false)
	if err != nil {
		return nil, errors.Wrapf(err, "error reading Pachyderm config")
	}
	_, context, err := cfg.ActiveContext(true)
	if err != nil {
		return nil, errors.Wrapf(err, "error getting the active context")
	}
	if context.ActiveTransaction == "" {
		return nil, nil
	}
	return &transaction.Transaction{ID: context.ActiveTransaction}, nil
}

func requireActiveTransaction() (*transaction.Transaction, error) {
	txn, err := getActiveTransaction()
	if err != nil {
		return nil, err
	} else if txn == nil {
		return nil, errors.Errorf("no active transaction")
	}
	return txn, nil
}

func setActiveTransaction(txn *transaction.Transaction) error {
	cfg, err := config.Read(false, false)
	if err != nil {
		return errors.Wrapf(err, "error reading Pachyderm config")
	}
	_, context, err := cfg.ActiveContext(true)
	if err != nil {
		return errors.Wrapf(err, "error getting the active context")
	}
	if txn == nil {
		context.ActiveTransaction = ""
	} else {
		context.ActiveTransaction = txn.ID
	}
	if err := cfg.Write(); err != nil {
		return errors.Wrapf(err, "error writing Pachyderm config")
	}
	return nil
}

// ClearActiveTransaction will remove the active transaction from the pachctl
// config file - used by the 'delete all' command.
func ClearActiveTransaction() error {
	return setActiveTransaction(nil)
}

// WithActiveTransaction is a helper function that will attach the given
// transaction to the env's client (resulting in a new client), then run the
// given callback with the new client.  This is for executing RPCs that can be
// run inside a transaction - if this isn't supported, it will have no effect.
func WithActiveTransaction(env cmdutil.Env, callback func(*client.APIClient) error) error {
	txn, err := getActiveTransaction()
	if err != nil {
		return err
	}
	c := env.Client("user")
	if txn != nil {
		c = c.WithTransaction(txn)
	}
	err = callback(c)
	if err == nil && txn != nil {
		fmt.Fprintf(env.Stdout(), "Added to transaction: %s\n", txn.ID)
	}
	return err
}
