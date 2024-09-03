package cmds

import (
	"fmt"
	"os"

	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachctl"
	"github.com/pachyderm/pachyderm/v2/src/internal/tabwriter"
	"github.com/pachyderm/pachyderm/v2/src/server/transaction/pretty"
	"github.com/pachyderm/pachyderm/v2/src/transaction"
	"github.com/spf13/cobra"
)

// Cmds returns the set of commands used for managing transactions with the
// Pachyderm CLI tool pachctl.
func Cmds(pachctlCfg *pachctl.Config) []*cobra.Command {
	var commands []*cobra.Command

	var raw bool
	var output string
	outputFlags := cmdutil.OutputFlags(&raw, &output)

	var fullTimestamps bool
	timestampFlags := cmdutil.TimestampFlags(&fullTimestamps)

	transactionDocs := &cobra.Command{
		Short: "Docs for transactions.",
		Long: "Transactions modify several Pachyderm objects in a single operation. " +
			"The following pachctl commands are supported in transactions: \n\n" +
			"\t- create repo\n" +
			"\t- delete repo\n" +
			"\t- start commit\n" +
			"\t- finish commit\n" +
			"\t- delete commit\n" +
			"\t- create branch\n" +
			"\t- delete branch\n" +
			"\t- create pipeline\n" +
			"\t- update pipeline\n\n" +
			"A transaction can be started with `pachctl start transaction`, after which the above commands will be stored in the transaction rather than immediately executed. \n\n" +
			"The stored commands can be executed as a single operation with `pachctl finish transaction` or cancelled with `pachctl delete transaction`.",
	}
	commands = append(commands, cmdutil.CreateDocsAlias(transactionDocs, "transaction", " transaction$"))

	listTransaction := &cobra.Command{
		Short: "List transactions.",
		Long:  "This command lists transactions.",
		Run: cmdutil.RunFixedArgs(0, func(cmd *cobra.Command, args []string) (retErr error) {
			c, err := pachctlCfg.NewOnUserMachine(cmd.Context(), false)
			if err != nil {
				return err
			}
			defer errors.Close(&retErr, c, "close client")
			transactions, err := c.ListTransaction()
			if err != nil {
				return err
			}
			if raw {
				encoder := cmdutil.Encoder(output, os.Stdout)
				for _, transaction := range transactions {
					if err := encoder.EncodeProto(transaction); err != nil {
						return errors.EnsureStack(err)
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
		Long:  "This command starts a new transaction.",
		Run: cmdutil.RunFixedArgs(0, func(cmd *cobra.Command, args []string) (retErr error) {
			c, err := pachctlCfg.NewOnUserMachine(cmd.Context(), false)
			if err != nil {
				return err
			}
			defer errors.Close(&retErr, c, "close client")
			txn, err := getActiveTransaction()
			if err != nil {
				return err
			}
			if txn != nil {
				return errors.Errorf("cannot start a new transaction, since transaction with ID %q already exists", txn.Id)
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
			fmt.Printf("started new transaction: %q\n", transaction.Id)
			return nil
		}),
	}
	commands = append(commands, cmdutil.CreateAlias(startTransaction, "start transaction"))

	stopTransaction := &cobra.Command{
		Short: "Stop modifying the current transaction.",
		Long:  "This command stops modifying the current transaction; to be used with `pachctl resume transaction`. This command is ideal for drafting a transaction in multiple steps.",
		Run: cmdutil.RunFixedArgs(0, func(cmd *cobra.Command, args []string) error {
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

			fmt.Printf("Cleared active transaction: %s\n", txn.Id)
			return nil
		}),
	}
	commands = append(commands, cmdutil.CreateAlias(stopTransaction, "stop transaction"))

	finishTransaction := &cobra.Command{
		Use:   "{{alias}} [<transaction>]",
		Short: "Execute and clear the currently active transaction.",
		Long:  "This command executes and clears the currently active transaction.",
		Run: cmdutil.RunBoundedArgs(0, 1, func(cmd *cobra.Command, args []string) (retErr error) {
			c, err := pachctlCfg.NewOnUserMachine(cmd.Context(), false)
			if err != nil {
				return err
			}
			defer errors.Close(&retErr, c, "close client")

			// TODO: use advisory locks on config so we don't have a race condition if
			// two commands are run simultaneously
			var txn *transaction.Transaction
			if len(args) > 0 {
				txn = &transaction.Transaction{Id: args[0]}
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

			fmt.Printf("Completed transaction with %d requests: %s\n", len(info.Responses), info.Transaction.Id)
			return nil
		}),
	}
	commands = append(commands, cmdutil.CreateAlias(finishTransaction, "finish transaction"))

	deleteTransaction := &cobra.Command{
		Use:   "{{alias}} [<transaction>]",
		Short: "Cancel and delete an existing transaction.",
		Long:  "This command cancels and deletes an existing transaction.",
		Run: cmdutil.RunBoundedArgs(0, 1, func(cmd *cobra.Command, args []string) (retErr error) {
			c, err := pachctlCfg.NewOnUserMachine(cmd.Context(), false)
			if err != nil {
				return err
			}
			defer errors.Close(&retErr, c, "close client")

			// TODO: use advisory locks on config so we don't have a race condition if
			// two commands are run simultaneously
			var txn *transaction.Transaction
			isActive := false
			if len(args) > 0 {
				txn = &transaction.Transaction{Id: args[0]}

				// Don't check err here, this is just a quality-of-life check to clean
				// up the config after a successful delete
				activeTxn, _ := requireActiveTransaction()
				if activeTxn != nil {
					isActive = txn.Id == activeTxn.Id
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
				if err := ClearActiveTransaction(); err != nil {
					return err
				}
			}
			return nil
		}),
	}
	commands = append(commands, cmdutil.CreateAlias(deleteTransaction, "delete transaction"))

	inspectTransaction := &cobra.Command{
		Use:   "{{alias}} [<transaction>]",
		Short: "Print information about an open transaction.",
		Long:  "This command prints information about an open transaction.",
		Run: cmdutil.RunBoundedArgs(0, 1, func(cmd *cobra.Command, args []string) (retErr error) {
			c, err := pachctlCfg.NewOnUserMachine(cmd.Context(), false)
			if err != nil {
				return err
			}
			defer errors.Close(&retErr, c, "close client")

			var txn *transaction.Transaction
			if len(args) > 0 {
				txn = &transaction.Transaction{Id: args[0]}
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
				return errors.Errorf("transaction %s not found", txn.Id)
			}
			if raw {
				return errors.EnsureStack(cmdutil.Encoder(output, os.Stdout).EncodeProto(info))
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
		Long:  "This command sets an existing transaction as active; to be used with `pachctl stop transaction`.",
		Run: cmdutil.RunFixedArgs(1, func(cmd *cobra.Command, args []string) (retErr error) {
			c, err := pachctlCfg.NewOnUserMachine(cmd.Context(), false)
			if err != nil {
				return err
			}
			defer errors.Close(&retErr, c, "close client")
			info, err := c.InspectTransaction(&transaction.Transaction{Id: args[0]})
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

			fmt.Printf("Resuming existing transaction with %d requests: %s\n", len(info.Requests), info.Transaction.Id)
			return nil
		}),
	}
	commands = append(commands, cmdutil.CreateAlias(resumeTransaction, "resume transaction"))

	return commands
}
