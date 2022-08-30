/*
Package transaction implements user-level transactions consisting of one or more API calls.

Transactions were added to Pachyderm as a way to make multiple changes to the state of Pachyderm while only triggering jobs once.
This is done by constructing a batch of operations to perform on the cluster state, then running the set of operations in a single SQL transaction.
This allows users to coordinate changes across multiple branches with a minimum of new commits.

This is useful in particular for pipelines with multiple inputs.
If you need to update two or more input repos, you might not want jobs for each state change.
You can issue a transaction to start commits in each of the input repos, which will create a single downstream commit in the pipeline repo.
After the transaction, you can put files and finish the commits at will, and the job will run once all the input commits have been finished.

# Pachctl

In pachctl, a transaction can be initiated via

	start transaction

This will generate a transaction object in the cluster and save its ID into the local pachyderm config (~/.pachyderm/config.json by default).

While there is a transaction object in the config file, all transaction-supporting API requests will append the request to the transaction instead of running directly.
Transaction behavior is triggered by the pach-transaction field in the gRPC context metadata.
Both the client (to attach the metadata) and server (to interpret it) need to support transactions for this to work.
Non-transactional commands issued with an active transaction will be executed normally, which can lead to unexpected results.
The supported commands are:
  - create repo
  - delete repo
  - start commit
  - finish commit
  - squash commit
  - delete commit
  - create branch
  - delete branch
  - create/update pipeline

Each time a command is added to a transaction, the transaction is dry-run against the current state of the cluster metadata to make sure it is still valid.
If the dry-run fails for any reason, the operation will not be added to the transaction.
If the transaction has been invalidated by changing cluster state, the transaction will need to be deleted and started over, taking into account the new state of the cluster.

From a command-line perspective, these commands should work identically within a transaction as without with the exception that the changes will not be committed until

	finish transaction

is run, and a message will be logged to stderr to indicate that the command was placed in a transaction rather than run directly.

There are several other supporting commands for transactions:

  - list transaction - list all unfinished transactions available in the pachyderm cluster
  - stop transaction - remove the currently active transaction from the local pachyderm config file - it remains in the pachyderm cluster and may be resumed later
  - resume transaction - set an already-existing transaction as the active transaction in the local pachyderm config file
  - delete transaction - deletes a transaction from the pachyderm cluster
  - inspect transaction - provide detailed information about an existing transaction, including which operations it will perform

# Auth considerations

The transaction API does not use auth at all.
This is fine because each operation in a transaction checks auth as usual.
When adding an operation to a transaction, the entire transaction is dry-run.
If a user attempts to add an operation they cannot perform, it will be rejected when the dry-run fails.
Similarly, if a user attempts to finish a transaction that contains operations they cannot perform, it will use their auth credentials and fail.

The most an adversarial user can do is delete or invalidate transactions that other users are building.
Any other commands they might add to a transaction, they could run directly on the cluster.

# Limitations

Transactions (especially large transactions) may be very slow, because a dry-run is executed every time an operation is added.
This means that we end up performing O(n) dry-runs and O(n^2) operations for a transaction with n operations in it.

None of the user-exposed transaction commands interact with the file system.
Luckily, due to how pipeline triggering works, it is not important for users to change files transactionally.
As long as commits are started in the same transaction, their contents will be processed together without creating undesired intermediate jobs.
*/
package transaction
