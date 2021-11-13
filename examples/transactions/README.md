>![pach_logo](../img/pach_logo.svg) INFO Pachyderm 2.0 introduces profound architectual changes to the product. As a result, our examples pre and post 2.0 are kept in two separate branches:
> - Branch Master: Examples using Pachyderm 2.0 and later versions - https://github.com/pachyderm/pachyderm/tree/master/examples
> - Branch 1.13.x: Examples using Pachyderm 1.13 and older versions - https://github.com/pachyderm/pachyderm/tree/1.13.x/examples

# Use Transactions with Hyperparameter Tuning

!!! note "Summary"
    Transactions can help optimize the use of resources
    by postponing pipeline runs.

Hyperparameter tuning is a machine learning technique
of narrowing down a set of parameters to
an optimal number of parameters to train a learning
algorithm. The Pachyderm documentation includes a
[hyperparameter tuning example](https://github.com/pachyderm/pachyderm/tree/master/examples/ml/hyperparameter)
that describes how this computation works in Pachyderm.

In the hyperparameter example, training data is submitted
to the `data` repository, and the parameters are stored
in the `parameters` repository. In that example, the
data processing takes seconds and, therefore, you can
run this operation for every commit without being worried
about the use of resources. But, if your
data processing takes significant time,
you might want to optimize Pachyderm to run the pipeline against
specific commits in the `data` and `parameters` repositories.
You can do so by using transactions.

## Set up the Hyperparameter Example

To demonstrate the benefits of using transactions, we
will use transactions on the `model` pipeline step from the
[hyperparameter tuning example](https://github.com/pachyderm/pachyderm/tree/master/examples/ml/hyperparameter).
In this transaction example, we omit the splitting step and
have just the `model` pipeline that consumes commits from
the `data` and `parameters` repositories and outputs the
result to the `model` repository. You can adapt this example
to your pipelines as needed.

The following diagram describes the pipeline structure:

![transactions diagram](../../doc/docs/master/assets/images/d_transactions_hyperparameter.svg)

To set up the pipeline, complete the following steps:

1. Create the `data` repository:

   ```shell
   $ pachctl create repo data
   ```

1. Create the `parameters` repository:

   ```shell
   $ pachctl create repo parameters
   ```

1. Verify that the repositories were successfully created:

   ```shell
   $ pachctl list repo
   NAME       CREATED        SIZE (MASTER)
   parameters 44 minutes ago 123B
   raw_data   44 minutes ago 6.858KiB
   ```

1. Clone the Pachyderm repository:

   ```shell
   $ git clone git@github.com:pachyderm/pachyderm.git
   ```

1. Change the directory to `examples/transactions`:

   ```shell
   $ cd examples/transactions/
   ```

1. Create the `model` pipeline:

   ```shell
   $ pachctl create pipeline -f model.json
   ```

1. Verify that the pipeline has been created:

   ```shell
   $ pachctl list pipeline
   NAME       VERSION INPUT                                                                                      CREATED        STATE / LAST JOB
   model      1       (parameters:/c_parameters.txt/* ⨯ parameters:/gamma_parameters.txt/* ⨯ raw_data:/iris.csv) 12 seconds ago running / starting
   ```

## Run the Transaction

To match commits in a pipeline, you need to create
a transaction, open two commits inside of that transaction,
then close the transaction, add your files, and then close both
commits. Pachyderm puts the changes to both repositories simultaneously
only when all commits that you have opened within a transaction are
closed.

To run the transaction, complete the following steps:

1. Start a transaction:

   ```shell
   $ pachctl start transaction
   Started new transaction: 854e8503-6e5d-4542-805c-a73a39200bf8
   ```

1. Open a commit into the `master` branch of the `raw_data` repository:

   ```shell
   $ pachctl start commit raw_data@master
   Added to transaction: 854e8503-6e5d-4542-805c-a73a39200bf8
   42b893e48e7d40f1bb5ed770526a9a07
   ```

1. Open a commit into the `master` branch of the `parameters` repository:

   ```shell
   $ pachctl start commit parameters@master
   Added to transaction: 854e8503-6e5d-4542-805c-a73a39200bf8
   c4dc446b25e54a938a67a5e913b3f9a4
   ```

1. Close the transaction:

   ```shell
   $ pachctl finish transaction
   Completed transaction with 2 requests: 854e8503-6e5d-4542-805c-a73a39200bf8
   ```

1. Add the data to the parameters repository by splitting each line
   into a separate file:

   ```shell
   $ pachctl put file parameters@master -f c_parameters.txt --split line --target-file-datums 1
   $ pachctl put file parameters@master -f gamma_parameters.txt --split line --target-file-datums 1
   ```

   ```shell
   $ pachctl list file parameters@master
   NAME                  TYPE SIZE
   /c_parameters.txt     dir  81B
   /gamma_parameters.txt dir  42B
   ```

   **Note:** Although the files are in the repository, no jobs were
   triggered for the `model` pipeline. You can verify that by running
   the following command:

   ```shell
   $ pachctl list job --pipeline=model --no-pager
   ID                               PIPELINE STARTED      DURATION RESTART PROGRESS  DL UL STATE
   ```

1. Add the data to the `raw_data` repository:

   ```shell
   $ pachctl put file raw_data@master:iris.csv -f noisy_iris.csv
   ```

   If you check whether the pipeline has run or not, you
   can see that it has not yet run:

   ```shell
   $ pachctl list job --pipeline=model --no-pager
   ID PIPELINE STARTED DURATION RESTART PROGRESS DL UL STATE
   ```

1. Close the commit to the `raw_data` repository that you have
   started within the transaction:

   ```shell
   $ pachctl finish commit raw_data@master
   ```

   Still no jobs run for the `model` pipeline. If we had not
   started those commits in a transaction, a job would normally
   be triggered here because Pachyderm normally triggers jobs
   whenever a commit is made on any input. In this case, because
   those commits were added to a transaction, Pachyderm waits
   for both input commits to be finished before a job triggers.

1. Close the commit to the `parameters` repository that you have
   started within the transaction:

   ```shell
   $ pachctl finish commit parameters@master
   ```

   Now, the Pachyderm finishes the transaction by creating one
   job that takes the commits that you have specified within the
   transaction and runs your code against these two commits:

   ```shell
   $ pachctl list job --pipeline=model --no-pager
   ID                               PIPELINE STARTED          DURATION    RESTART  PROGRESS  DL       UL      STATE
   6cdc80ae105f47b4a09f0ab8ce005003 model    37 seconds ago - 0           21 + 0 / 77        115.5KiB 62.2KiB running
   ```

1. View the contents of the model output repo:

   ```shell
   $ pachctl list file model@master
   NAME                      TYPE SIZE
   /model_C0.031_G0.001.pkl  file 5.713KiB
   /model_C0.031_G0.004.pkl  file 5.713KiB
   /model_C0.031_G0.016.pkl  file 5.713KiB
   /model_C0.031_G0.063.pkl  file 5.713KiB
   ...
   ```

In this example, we learned that if a pipeline
takes a lot of time to run, you can optimize it by using
transactions. Transactions enable you to accumulate your
changes in input repositories and postpone pipeline runs
until after the commits are closed.
