# Transaction

!!! note "Summary"
    Use transactions to run multiple Pachyderm commands
    simultaneously in one job run.

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

## Use Cases

Pachyderm users implement transactions to their own workflows finding
unique ways to benefit from this feature, whether it is a small
research or an enterprise-grade machine learning workflow.

Here are the most commonly employed ways of using transactions:

###Commit to Separate Repositories Simultaneously

For example, you have a Pachyderm pipeline with two input
repositories. One repository includes training data and the
other `parameters` for your machine learning pipeline. If you need
to run specific data against specific parameters, you need to
run your pipeline against specific commits in both repositories.
To achieve this, you need to commit to these repositories
simultaneously.

If you use a regular Pachyderm workflow, the data is uploaded consequently
by running two separate Pachyderm jobs. One job commits changes to
the data repository and the other updates the parameters.
The following animation shows the standard Pachyderm workflow:

![Standard workflow](../../assets/images/transactions_wrong.gif)

If you need to update both the parameters and the data, this approach
does not work.

With transactions, you can ensure that `data` and `parameters` have
the latest commits in the same pipeline run. The following animation
demonstrates how a transaction works:

![Transactions workflow](../../assets/images/transactions_right.gif)

The transaction ensures that a job runs for the two latest commits.
While you could probably achieve the same without transactions by
storing all your data in one repository, often you prefer to separate
it in individual repositories for organizational and logistics reasons.

### Switching from Staging to Master Simultaneously

If you use the [deferred processing](../../how-tos/deferred_processing/)
model when you want to commit your changes often but do not want your
pipeline to be triggered as often as you commit. When you want to postpone
pipeline execution, you can create a staging and master branch in
the same repository. You commit your changes to the staging branch and
when needed, switch the HEAD of you master branch to a commit in the
staging branch. To do this simultaneously, you can use transactions.

For example, you have two repositories `data` and `parameters`, both
of which have a `master` and `staging` branch. You commit your
changes to the staging branch while your pipeline is subscribed to the
master branch. To switch to these branches simultaneously, you can
use transactions like this:

```bash
$ pachctl start transaction
Started new transaction: 0d6f0bc3-37a0-4936-96e3-82034a2a2055
$ pachctl pachctl create branch data@master --head staging
Added to transaction: 0d6f0bc3-37a0-4936-96e3-82034a2a2055
$ pachctl create branch parameters@master --head staging
Added to transaction: 0d6f0bc3-37a0-4936-96e3-82034a2a2055
$ pachctl finish transaction
Completed transaction with 2 requests: 0d6f0bc3-37a0-4936-96e3-82034a2a2055
```

When you finish the transaction, both repositories switch to
to the master at the same time which crates new commits in
the corresponding `master` branches and the pipeline triggers
one job to process the commits together.

## Start and Finish Transactions

To start a transaction, run the following command:

```bash
$ pachctl start transaction
Started new transaction: 7a81eab5-e6c6-430a-a5c0-1deb06852ca5
```

This command generates a transaction object in the cluster and saves
its ID in the local Pachyderm configuration file. By default this file
is stored at `~/.pachyderm/config.json`.

!!! example

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

After you start a transaction, you can add supported commands to it, such
as create a repo, create a branch, and so on. You can also start a commit
within a transaction. If you close a transaction before you close a commit,
the commands that you have specified inside the transaction and inside the
commit will not be applied until you close that commit.

To finish a transaction, run:

```bash
$ pachctl finsh transaction
Completed transaction with 1 requests: 7a81eab5-e6c6-430a-a5c0-1deb06852ca5
```

If you have opened a commit within the transaction, you can apply
all the operations that are not supported by transactions, such as
`pachctl put file` within that commit. Then, the operations are applied
in one job after you close the commit.

## Other Transaction Commands

Other supporting commands for transactions include the following commands:

| Command      | Description |
| ------------ | ----------- |
| `pachctl list transaction` | List all unfinished transactions available in the Pachyderm cluster. |
| `pachctl stop transaction` | Remove the currently active transaction from the local Pachyderm config file. The transaction remains in the Pachyderm cluster and can be resumed later. |
| `pachctl resume transaction` | Set an already-existing transaction as the active transaction in the local Pachyderm config file. |
| `pachctl delete transaction` | Deletes a transaction from the Pachyderm cluster. |
| `pachctl inspect transaction` | Provides detailed information about an existing transaction, including which operations it will perform. By default, displays information about the current transaction. If you specify a transaction ID, displays information
about the corresponding transaction. |

## Supported Operations

While there is a transaction object in the Pachyderm configuration
file, all supported API requests append the request to the
transaction instead of running directly. These supported commands include:

```bash
create repo
delete repo
start commit
finish commit
delete commit
create branch
delete branch
```

Each time you ad a command to a transaction, Pachyderm validates the
transaction against the current state of the cluster metadata and obtains
any return values, which is important for such commands as
`start commit`. If validation fails for any reason, Pachyderm does
not add the operation to the transaction. If the transaction has been
invalidated by changing the cluster state, you must delete the transaction
and start over, taking into account the new state of the cluster.

From a command-line perspective, these commands work identically within
a transaction as without. The only difference is that you do not apply
your changes until you run `finish transaction`, and a message that
Pachyderm logs to `stderr` to indicate that the command was placed
in a transaction rather than run directly.

## Multiple Opened Transactions

Some systems has a notion of *nested* transactions. That is when you
open transactions within an opened transaction. In such systems, the
operations added to the subsequent transactions are not executed
until you close all the nested transaction and the main transaction.

Pachyderm does not support such behaviour. Instead, when you open a
transaction, the transaction ID is written to the Pachyderm configuration
file. If you open another transaction while the first one is open, Pachyderm
suspends the first transaction and overwrites the transaction ID in the
configuration file. All operations that you add to the consequent
transactions will be executed as soon as you close those
transactions. To resume the initial transaction, you need to run
`pachctl resume transaction`.

Because nested transactions are not supported, transactions cannot
conflict with each other. Every time you add a command to a transaction,
Pachyderm creates a blueprint of the commit and verifies that the
command is valid.



