package cmds

import (
	"fmt"
	"os"

	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/tabwriter"
	"github.com/pachyderm/pachyderm/v2/src/server/transaction/pretty"
	"github.com/pachyderm/pachyderm/v2/src/transaction"
	"github.com/spf13/cobra"
)

// Cmds returns the set of commands used for managing transactions with the
// Pachyderm CLI tool pachctl.
func Cmds() []*cobra.Command {
	var commands []*cobra.Command

	var raw bool
	var output string
	outputFlags := cmdutil.OutputFlags(&raw, &output)

	var fullTimestamps bool
	timestampFlags := cmdutil.TimestampFlags(&fullTimestamps)

	transactionDocs := &cobra.Command{
		Short: "Docs for transactions.",
		Long: `Transactions modify several Pachyderm objects in a single operation.

The following pachctl commands are supported in transactions:
  create repo
  delete repo
  start commit
  finish commit
  delete commit
  create branch
  delete branch
  create pipeline
  update pipeline

A transaction can be started with 'start transaction', after which the above
commands will be stored in the transaction rather than immediately executed.
The stored commands can be executed as a single operation with 'finish
transaction' or cancelled with 'delete transaction'.`,
	}
	commands = append(commands, cmdutil.CreateDocsAlias(transactionDocs, "transaction", " transaction$"))

	listTransaction := &cobra.Command{
		Short: "List transactions.",
		Long:  "List transactions.",
		Run: cmdutil.RunFixedArgs(0, func([]string) error {
			c, err := client.NewOnUserMachine("user")
			if err != nil {
				return err
			}
			defer c.Close()
			transactions, err := c.ListTransaction()
			if err != nil {
				return err
			}
			if raw {
				encoder := cmdutil.Encoder(output, os.Stdout)
				for _, transaction := range transactions {
					if err := encoder.EncodeProto(transaction); err != nil {
						return err
					}
				}
				return nil
			} else if output != "" {
				return errors.New("cannot set --output (-o) without --raw")
			}
			writer := tabwriter.NewWriter(os.Stdout, pretty.TransactionHeader)
			for _, transaction := range transactions {
				pretty.PrintTransactionInfo(writer, transaction, fullTimestamps)
			}
			return writer.Flush()
		}),
	}
	listTransaction.Flags().AddFlagSet(outputFlags)
	listTransaction.Flags().AddFlagSet(timestampFlags)
	commands = append(commands, cmdutil.CreateAlias(listTransaction, "list transaction"))

	startTransaction := &cobra.Command{
		Short: "Start a new transaction.",
		Long:  "Start a new transaction.",
		Run: cmdutil.RunFixedArgs(0, func([]string) error {
			c, err := client.NewOnUserMachine("user")
			if err != nil {
				return err
			}
			defer c.Close()
			txn, err := getActiveTransaction()
			if err != nil {
				return err
			}
			if txn != nil {
				return errors.Errorf("cannot start a new transaction, since transaction with ID %q already exists", txn.ID)
			}

			transaction, err := c.StartTransaction()
			if err != nil {
				return grpcutil.ScrubGRPC(err)
			}
			// TODO: use advisory locks on config so we don't have a race condition if
			// two commands are run simultaneously
			err = setActiveTransaction(transaction)
			if err != nil {
				return err
			}
			fmt.Printf("started new transaction: %q\n", transaction.ID)
			return nil
		}),
	}
	commands = append(commands, cmdutil.CreateAlias(startTransaction, "start transaction"))

	stopTransaction := &cobra.Command{
		Short: "Stop modifying the current transaction.",
		Long:  "Stop modifying the current transaction.",
		Run: cmdutil.RunFixedArgs(0, func([]string) error {
			// TODO: use advisory locks on config so we don't have a race condition if
			// two commands are run simultaneously
			txn, err := requireActiveTransaction()
			if err != nil {
				return err
			}

			err = ClearActiveTransaction()
			if err != nil {
				return err
			}

			fmt.Printf("Cleared active transaction: %s\n", txn.ID)
			return nil
		}),
	}
	commands = append(commands, cmdutil.CreateAlias(stopTransaction, "stop transaction"))

	finishTransaction := &cobra.Command{
		Use:   "{{alias}} [<transaction>]",
		Short: "Execute and clear the currently active transaction.",
		Long:  "Execute and clear the currently active transaction.",
		Run: cmdutil.RunBoundedArgs(0, 1, func(args []string) error {
			c, err := client.NewOnUserMachine("user")
			if err != nil {
				return err
			}
			defer c.Close()

			// TODO: use advisory locks on config so we don't have a race condition if
			// two commands are run simultaneously
			var txn *transaction.Transaction
			if len(args) > 0 {
				txn = &transaction.Transaction{ID: args[0]}
			} else {
				txn, err = requireActiveTransaction()
				if err != nil {
					return err
				}
			}

			info, err := c.FinishTransaction(txn)
			if err != nil {
				return grpcutil.ScrubGRPC(err)
			}

			err = ClearActiveTransaction()
			if err != nil {
				return err
			}

			fmt.Printf("Completed transaction with %d requests: %s\n", len(info.Responses), info.Transaction.ID)
			return nil
		}),
	}
	commands = append(commands, cmdutil.CreateAlias(finishTransaction, "finish transaction"))

	deleteTransaction := &cobra.Command{
		Use:   "{{alias}} [<transaction>]",
		Short: "Cancel and delete an existing transaction.",
		Long:  "Cancel and delete an existing transaction.",
		Run: cmdutil.RunBoundedArgs(0, 1, func(args []string) error {
			c, err := client.NewOnUserMachine("user")
			if err != nil {
				return err
			}
			defer c.Close()

			// TODO: use advisory locks on config so we don't have a race condition if
			// two commands are run simultaneously
			var txn *transaction.Transaction
			isActive := false
			if len(args) > 0 {
				txn = &transaction.Transaction{ID: args[0]}

				// Don't check err here, this is just a quality-of-life check to clean
				// up the config after a successful delete
				activeTxn, _ := requireActiveTransaction()
				if activeTxn != nil {
					isActive = txn.ID == activeTxn.ID
				}
			} else {
				txn, err = requireActiveTransaction()
				if err != nil {
					return err
				}
				isActive = true
			}

			err = c.DeleteTransaction(txn)
			if err != nil {
				return grpcutil.ScrubGRPC(err)
			}
			if isActive {
				// The active transaction was successfully deleted, clean it up so the
				// user doesn't need to manually 'stop transaction' it.
				ClearActiveTransaction()
			}
			return nil
		}),
	}
	commands = append(commands, cmdutil.CreateAlias(deleteTransaction, "delete transaction"))

	inspectTransaction := &cobra.Command{
		Use:   "{{alias}} [<transaction>]",
		Short: "Print information about an open transaction.",
		Long:  "Print information about an open transaction.",
		Run: cmdutil.RunBoundedArgs(0, 1, func(args []string) error {
			c, err := client.NewOnUserMachine("user")
			if err != nil {
				return err
			}
			defer c.Close()

			var txn *transaction.Transaction
			if len(args) > 0 {
				txn = &transaction.Transaction{ID: args[0]}
			} else {
				txn, err = requireActiveTransaction()
				if err != nil {
					return err
				}
			}

			info, err := c.InspectTransaction(txn)
			if err != nil {
				return grpcutil.ScrubGRPC(err)
			}
			if info == nil {
				return errors.Errorf("transaction %s not found", txn.ID)
			}
			if raw {
				return cmdutil.Encoder(output, os.Stdout).EncodeProto(info)
			} else if output != "" {
				return errors.New("cannot set --output (-o) without --raw")
			}
			return pretty.PrintDetailedTransactionInfo(&pretty.PrintableTransactionInfo{
				TransactionInfo: info,
				FullTimestamps:  fullTimestamps,
			})
		}),
	}
	inspectTransaction.Flags().AddFlagSet(outputFlags)
	inspectTransaction.Flags().AddFlagSet(timestampFlags)
	commands = append(commands, cmdutil.CreateAlias(inspectTransaction, "inspect transaction"))

	resumeTransaction := &cobra.Command{
		Use:   "{{alias}} <transaction>",
		Short: "Set an existing transaction as active.",
		Long:  "Set an existing transaction as active.",
		Run: cmdutil.RunFixedArgs(1, func(args []string) error {
			c, err := client.NewOnUserMachine("user")
			if err != nil {
				return err
			}
			defer c.Close()
			info, err := c.InspectTransaction(&transaction.Transaction{ID: args[0]})
			if err != nil {
				return grpcutil.ScrubGRPC(err)
			}
			if info == nil {
				return errors.Errorf("transaction %s not found", args[0])
			}

			err = setActiveTransaction(info.Transaction)
			if err != nil {
				return err
			}

			fmt.Printf("Resuming existing transaction with %d requests: %s\n", len(info.Requests), info.Transaction.ID)
			return nil
		}),
	}
	commands = append(commands, cmdutil.CreateAlias(resumeTransaction, "resume transaction"))

	return commands
}
