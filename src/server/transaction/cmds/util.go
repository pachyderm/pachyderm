package cmds

import (
	"fmt"

	"github.com/pachyderm/pachyderm/src/client/pkg/config"
	"github.com/pachyderm/pachyderm/src/client/transaction"
)

func GetActiveTransaction() (*transaction.Transaction, error) {
	cfg, err := config.Read()
	if err != nil {
		return nil, fmt.Errorf("error reading Pachyderm config: %v", err)
	}
	if cfg.V1 == nil || cfg.V1.ActiveTransaction == "" {
		return nil, nil
	}
	return &transaction.Transaction{ID: cfg.V1.ActiveTransaction}, nil
}

func requireActiveTransaction() (*transaction.Transaction, error) {
	txn, err := GetActiveTransaction()
	if err != nil {
		return nil, err
	} else if txn == nil {
		return nil, fmt.Errorf("no active transaction")
	}
	return txn, nil
}

func setActiveTransaction(txn *transaction.Transaction) error {
	cfg, err := config.Read()
	if err != nil {
		return fmt.Errorf("error reading Pachyderm config: %v", err)
	}
	if cfg.V1 == nil {
		cfg.V1 = &config.ConfigV1{}
	}
	if txn == nil {
		cfg.V1.ActiveTransaction = ""
	} else {
		cfg.V1.ActiveTransaction = txn.ID
	}
	if err := cfg.Write(); err != nil {
		return fmt.Errorf("error writing Pachyderm config: %v", err)
	}
	return nil
}
