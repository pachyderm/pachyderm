# Transactions in Pachyderm

Transactions were added to Pachyderm as a way to make multiple changes to the
state of Pachyderm while only triggering jobs once. This is done by constructing
a batch of operations to perform on the cluster state, then running the set of
operations in a single ETCD transaction.

The transaction framework provides a method for batching together commit
propagation such that changed branches are collected over the course of the
transaction and all propagated in one batch at the end. This allows Pachyderm to
dedupe changed branches and branches provenant on the changed branches so that
the minimum number of new commits are issued.

This is useful in particular for pipelines with multiple inputs. If you need to
update two or more input repos, you might not want pipeline jobs for each state
change. You can issue a transaction to start commits in each of the input repos,
which will create a single downstream commit in the pipeline repo. After the
transaction, you can put files and finish the commits at will, and the pipeline
job will run once all the input commits have been finished.

## Pachctl

In `pachctl`, a transaction can be initiated through the `start transaction`
command. This will generate a transaction object in the cluster and save its ID
into the local pachyderm config (`~/.pachyderm/config.json` by default).

While there is a transaction object in the config file, all
transactionally-supported API requests will append the request to the
transaction instead of running directly. These commands (as of v1.9.0) are:

-   `create repo`
-   `delete repo`
-   `start commit`
-   `finish commit`
-   `delete commit`
-   `create branch`
-   `delete branch`

Each time a command is added to a transaction, the transaction is dry-run
against the current state of the cluster metadata to make sure it is still valid
and to obtain any return values (important for commands like `start commit`). If
the dry-run fails for any reason, the operation will not be added to the
transaction. If the transaction has been invalidated by changing cluster state,
the transaction will need to be deleted and started over, taking into account
the new state of the cluster.

From a command-line perspective, these commands should work identically within a
transaction as without with the exception that the changes will not be committed
until `finish transaction` is run, and a message will be logged to `stderr` to
indicate that the command was placed in a transaction rather than run directly.

There are several other supporting commands for transactions:

-   `list transaction` - list all unfinished transactions available in the
    pachyderm cluster
-   `stop transaction` - remove the currently active transaction from the local
    pachyderm config file - it remains in the pachyderm cluster and may be
    resumed later
-   `resume transaction` - set an already-existing transaction as the active
    transaction in the local pachyderm config file
-   `delete transaction` - deletes a transaction from the pachyderm cluster
-   `inspect transaction` - provide detailed information about an existing
    transaction, including which operations it will perform

## Implementation Details

Files and Packages:

-   `src/client/transaction.go` - client helper functions for transactions
-   `src/client/transaction/transaction.proto` - protobuf definitions for the
    API
-   `src/server/transaction/cmds` - implementation of the `pachctl` transaction
    commands
-   `src/server/transaction/pretty` - pretty-printing code used by `pachctl`
    commands
-   `src/server/transaction/server` - implementation of the GRPC API defined in
    the protobuf file
-   `src/server/pkg/transactiondb` - definition of transaction metadata in etcd
-   `src/server/pkg/transactionenv` - an environment object passed to each API
    server in `pachd` that coordinates calls across package boundaries, and
    provides common abstractions needed for transaction support

### TransactionEnv

The `transactionenv.TransactionEnv` object allows us to coordinate calls across
the API-server objects in `pachd` without going through an RPC. This means we
can include the state of an open STM transaction and guarantee consistent reads
and writes efficiently. This interface introduces the `TransactionContext`
object, which is a simple container with getters the full suite of objects
needed:

-   `ClientContext()` - the client context from the API client which initiated
    the current request
-   `Client()` - a pachyderm API client for making RPC calls to other
    subsystems. Using this does _not_ result in consistent reads and writes.
-   `Stm()` - the `col.STM` object associated with the current request
-   `PfsDefer()` - the `pfs.TransactionDefer` object associated with the current
    request. Primarily for its `PropagateCommit` call.

Two main functions are provided for starting an etcd transaction with a
`TransactionContext`. Each one takes a callback that will be provided with the
`TransactionContext` that is valid for the duration of the callback. When the
callback finishes, deferred tasks (i.e. the `PropagateCommit` calls) will be run
in the STM before committing the changes. `WithReadContext` uses a dry-run STM
so that all changes are discarded, and `WithWriteContext` uses a normal STM.

### Auth considerations

The transaction API does not use auth at all. This is fine because each
operation in a transaction checks auth as usual. When adding an operation to a
transaction, the entire transaction is dry-run. If a user attempts to add an
operation they cannot perform, it will be rejected when the dry-run fails.
Similarly, if a user attempts to finish a transaction that contains operations
they cannot perform, it will use their auth credentials and fail.

The most an adversarial user can do is delete or invalidate transactions that
other users are building. Any other commands they might add to a transaction,
they could run directly on the cluster.

### Limitations

The main limitation in transactions are the etcd operation limit, which means
that a transaction may grow large enough that it will be rejected. There is no
easy way to predict when this will happen at the moment, and the only workaround
is to break up the operations into multiple transactions.

In addition, transactions (especially large transactions) may be very slow,
because a dry-run is executed every time an operation is added to a transaction.
This means that we end up performing `O(n)` dry-runs and `O(n^2)` operations for
a transaction with `n` operations in it.

At the moment, there is no way to modify the files of a commit within a
transaction. Primarily, this is because we do not have a good way to provide an
STM interface that supports list operations. Modifying files on open commits
involves merging the tree from the committed state with the tree stored in the
open commits collection in etcd, which uses list operations. As such, this will
likely not be available without major changes to the architecture. Luckily, due
to how pipeline triggering works, it is not important to change files
transactionally - starting commits transactionally is the important part.

### Future Work

-   Support transactions in PPS calls
-   Have `delete all` be performed transactionally
-   Provide an API method for issuing a batch of operations as a transaction
    without round-trips
