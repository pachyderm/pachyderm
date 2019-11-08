# Transaction

A transaction is a Pachyderm operation that enables you to create
a collection of Pachyderm commands and execute them concurrently in a
single job run. Regular Pachyderm operations are executed consequently,
one after another. However, when you need to run multiple commands
at the same time, you can use transactions.

The transaction framework provides a method for batching together
commit propagation such that changed branches are collected over
the course of the transaction and all propagated in one batch at
the end. This method allows Pachyderm to dedupe changed branches and
branches provenant on the changed branches so that the minimum
number of new commits are issued.

This functionality is useful in particular for pipelines with multiple
inputs. If you need to update two or more input repos, you might not want
pipeline jobs for each state change. You can issue a transaction
to start commits in each of the input repos, which creates a single
downstream commit in the pipeline repo. After the transaction, you
can put files and finish the commits at will, and the pipeline job
will run once all the input commits have been finished.

!!! tip
    While you cannot use `pachctl put file` in a transaction, you can
    start a commit within a transaction, finish the transation,
    then put as many files as you need, and then finish your commit.
    Your changes will only be applied in one batch when you close
    the commit.

## Starting and Finishing Transactions

To start a transaction, run the following command:

```bash
$ pachctl start transaction
Started new transaction: 7a81eab5-e6c6-430a-a5c0-1deb06852ca5
```

This command generates a transaction object in the cluster and saves
its ID in the local Pachyderm configuration file. By default this file
is stored at `~/.pachyderm/config.json`.

**Example:**

```json hl_lines="9"
{
   "user_id": "b4fe4317-be21-4836-824f-6661c68b8fba",
   "v2": {
     "active_context": "local-2",
     "contexts": {
       "default": {},
       "local-2": {
         "source": 3,
         "active_transaction": "7a81eab5-e6c6-430a-a5c0-1deb06852ca5",
         "cluster_name": "minikube",
         "auth_info": "minikube",
         "namespace": "default"
       },
```

## Other Transaction Commands

Other supporting commands for transactions include the following commands:

| Command      | Description |
| ------------ | ----------- |
| `pachctl list transaction` | List all unfinished transactions available in the Pachyderm cluster. |
| `pachctl stop transaction` | Remove the currently active transaction from the local Pachyderm config file. The transaction remains in the Pachyderm cluster and can be resumed later. |
| `pachctl resume transaction` | Set an already-existing transaction as the active transaction in the local Pachyderm config file. |
| `pachctl delete transaction` | Deletes a transaction from the Pachyderm cluster. |
| `inspect transaction` | Provides detailed information about an existing transaction, including which operations it will perform. |

## Supported Operations

While there is a transaction object in the configuration file, all
supported API requests append the request to the transaction instead
of running directly. These commands include:

```bash
create repo
delete repo
start commit
finish commit
delete commit
create branch
delete branch
```

Each time a command is added to a transaction, the transaction is dry-run
against the current state of the cluster metadata to make sure it is still
valid and to obtain any return values which is important for such commands
as `start commit`. If the dry-run fails for any reason, the operation is
not be added to the transaction. If the transaction has been invalidated
by changing cluster state, the transaction must be deleted and started over,
taking into account the new state of the cluster.

From a command-line perspective, these commands should work identically
within a transaction as without with the exception that the changes will
not be committed until you run `finish transaction`, and a message will
be logged to `stderr` to indicate that the command was placed in a
transaction rather than run directly.

There are several other supporting commands for transactions:

- `list transaction` - list all unfinished transactions available in the pachyderm cluster
- `stop transaction` - remove the currently active transaction from the local pachyderm config file - it remains in the pachyderm cluster and may be resumed later
- `resume transaction` - set an already-existing transaction as the active transaction in the local pachyderm config file
- `delete transaction` - deletes a transaction from the pachyderm cluster
- `inspect transaction` - provide detailed information about an existing transaction, including which operations it will perform
