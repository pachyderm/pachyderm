# Batching Pachyderm with Transactions

Transactions were added to Pachyderm as a way to make multiple changes to the state of Pachyderm while only triggering jobs once. This is done by constructing a batch of operations to perform on the cluster state, then running the set of operations in a single ETCD transaction.

The transaction framework provides a method for batching together commit propagation such that changed branches are collected over the course of the transaction and all propagated in one batch at the end. This allows Pachyderm to dedupe changed branches and branches provenant on the changed branches so that the minimum number of new commits are issued.

This is useful in particular for pipelines with multiple inputs. If you need to update two or more input repos, you might not want pipeline jobs for each state change. You can issue a transaction to start commits in each of the input repos, which will create a single downstream commit in the pipeline repo. After the transaction, you can put files and finish the commits at will, and the pipeline job will run once all the input commits have been finished.

## Pachctl
In pachctl, a transaction can be initiated through the start transaction command. This will generate a transaction object in the cluster and save its ID into the local pachyderm config (`~/.pachyderm/config.json` by default).

While there is a transaction object in the config file, all transactionally-supported API requests will append the request to the transaction instead of running directly. These commands include:

!!! example
    ```shell
    create repo
    delete repo
    start commit
    finish commit
    delete commit
    create branch
    delete branch
    ```

Each time a command is added to a transaction, the transaction is dry-run against the current state of the cluster metadata to make sure it is still valid and to obtain any return values (important for commands like `start commit`). If the dry-run fails for any reason, the operation will not be added to the transaction. If the transaction has been invalidated by changing cluster state, the transaction will need to be deleted and started over, taking into account the new state of the cluster.

From a command-line perspective, these commands should work identically within a transaction as without with the exception that the changes will not be committed until `finish transaction` is run, and a message will be logged to `stderr` to indicate that the command was placed in a transaction rather than run directly.

There are several other supporting commands for transactions:

- `list transaction` - list all unfinished transactions available in the pachyderm cluster
- `stop transaction` - remove the currently active transaction from the local pachyderm config file - it remains in the pachyderm cluster and may be resumed later
- `resume transaction` - set an already-existing transaction as the active transaction in the local pachyderm config file
- `delete transaction` - deletes a transaction from the pachyderm cluster
- `inspect transaction` - provide detailed information about an existing transaction, including which operations it will perform
