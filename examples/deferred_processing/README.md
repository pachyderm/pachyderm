# Deferred Processing and Transactions

This example, which uses a simple DAG based on our
[OpenCV example](https://github.com/pachyderm/pachyderm/tree/master/examples/opencv),
illustrates two Pachyderm usage patterns for fine-grain control over when
pipelines trigger jobs.

[Deferred processing](https://docs.pachyderm.com/latest/how-tos/deferred_processing/)
is a Pachyderm technique for controlling when data gets processed. Deferred
processing uses branches to prevent pipelines from triggering on every input
commit.

[Transactions](https://docs.pachyderm.com/latest/how-tos/use-transactions-to-run-multiple-commands/)
are a Pachyderm feature that allows you to batch match multiple operations at
once, such as committing data to two different repos, but only trigger a single
job, so data from both repos gets processed together.

## Prerequisites

Before you begin, you need to have Pachyderm v1.9.8 or later installed on your
computer or cloud platform. See
[Deploy Pachyderm](https://docs.pachyderm.com/latest/deploy-manage/deploy/).

## Pipelines

The following diagram demonstrates the DAG that is used in this example.

![Example DAG](example_dag.png)

The DAG shown is a simple elaboration on the OpenCV example, with pipeline and
repo names chosen to avoid collisions with that example if installed in the same
cluster.

The functionality is slightly different. The `edges_dp` pipeline performs edge
detection on images committed to `images_dp_1`. `montage_dp` pipeline creates a
montage out of images committed to `images_dp_2` and images in the master branch
of `edges_dp`. This configuration enables you to verify the files being
processed by visually inspecting the montage.

The most significant change from the OpenCV example is the pipeline spec for
`edges_dp`, which has the `output_branch` attribute set to `dev`. It also has
added a `name` field to the `input` repo, to avoid having to change the code in
the example.

```json hl_lines="9,12"
{
    "pipeline": {
        "name": "edges_dp"
    },
    "input": {
        "pfs": {
            "glob": "/*",
            "repo": "images_dp_1",
            "name": "images"
        }
    },
    "output_branch": "dev",
    "transform": {
        "cmd": ["python3", "/edges.py"],
        "image": "pachyderm/opencv"
    }
}
```

Therefore, this pipeline puts the output in the `dev` branch instead of putting
it in the `master` branch.

Since `montage_dp` is subscribed to the master branch of `edges_dp`, jobs will
not trigger when the edges pipeline outputs files to the `dev` branch. Instead,
to trigger a montage job, we can simply create a `master` branch attached to any
commit in `edges_dp` to trigger the pipeline.

## Example run-through

This section provides steps that you can run to test this example.

### Deferred Processing

You should have a Pachyderm cluster set up and access to it configured from your
local computer before you run this example.

1. Run the script `setup.sh` included in this repo. The script executes the
   following commands:

    ```sh
    pachctl create repo images_dp_1
    pachctl create repo images_dp_2
    pachctl create pipeline -f ./edges_dp.json
    pachctl create pipeline -f ./montage_dp.json
    pachctl put file images_dp_1@master -i ./images.txt
    pachctl put file images_dp_1@master -i ./images2.txt
    pachctl put file images_dp_2@master -i ./images3.txt
    ```

2. Once the demo is loaded, check the commits in `edges_dp`. You should see an
   output similar to this. Note that there are two commits in `edges_dp` both in
   the `dev` branch that is the output for the pipeline:

    ```sh
    $  pachctl list commit edges_dp
    REPO     BRANCH COMMIT                           FINISHED
    edges_dp dev    364f49663dd848098b60c1ac97a332af 36 seconds ago
    edges_dp dev    a07c857b91a14add9f8309a81d86dbe8 44 seconds ago
    ```

    Remember that the `edges_dp` pipeline outputs to the `dev` branch. Since the
    `montage_dp` pipeline subscribes to the `master` branch, it will not be
    triggered when `edges_dp` jobs complete, since that output goes into the
    `dev` branch.

3. List the branches in `edges_dp`.

    ```sh
    $ pachctl list branch edges_dp
    BRANCH HEAD
    master -
    dev    364f49663dd848098b60c1ac97a332af
    ```

    Note that the `dev` branch has the most recent commit. Take note of the
    commit id and match it to the id above. The `master` branch does not have
    any commits.

4. View the list of jobs:

    ```sh
    $ pachctl list job
    ID                               PIPELINE   STARTED            DURATION  RESTART PROGRESS  DL       UL       STATE
    2288709b4d8044409c2232d673ec8f23 montage_dp 55 seconds ago     1 second  0       0 + 0 / 0 0B       0B       success
    6d9d4cf0f6524b0ca126fa97141303ea edges_dp   About a minute ago 4 seconds 0       2 + 1 / 3 181.1KiB 111.4KiB success
    fcaf537975554935b0f15d184d7a0984 edges_dp   About a minute ago 3 seconds 0       1 + 0 / 1 57.27KiB 22.22KiB success

    ```

    You should see that there are three jobs:

    - one 0-datum job for `montage_dp` and
    - two jobs for `edges_dp` with the appropriate number of datums in each.

    This is what you should expect. There is no data in the master branch of
    `edges_dp`, so an empty job was created in `montage_dp` when data was
    commited to `images_dp_2` because of its `cross` input.

5. Commit a file to `images_dp_1`.

    ```sh
    $ pachctl put file images_dp_1@master:1VqcWw9.jpg -f http://imgur.com/1VqcWw9.jpg
    ```

6. View the list of jobs, again.

    ```sh
    $ pachctl list job
    ID                               PIPELINE   STARTED            DURATION  RESTART PROGRESS  DL       UL       STATE
    c7e69e46e9954611ad8efc8aeac47f2a edges_dp   12 seconds ago     3 seconds 0       1 + 3 / 4 175.1KiB 92.18KiB success
    2288709b4d8044409c2232d673ec8f23 montage_dp About a minute ago 1 second  0       0 + 0 / 0 0B       0B       success
    6d9d4cf0f6524b0ca126fa97141303ea edges_dp   About a minute ago 4 seconds 0       2 + 1 / 3 181.1KiB 111.4KiB success
    fcaf537975554935b0f15d184d7a0984 edges_dp   About a minute ago 3 seconds 0       1 + 0 / 1 57.27KiB 22.22KiB success
    ```

    You see that one job was triggered in `edges_dp` with the one datum we
    committed, above, processed and the three existing datums skipped. You may
    also confirm that no job was triggered for `montage_dp`.

7. To trigger `montage_dp` on the set of data in our `dev` branch, you create a
   `master` branch with `dev` as its head.

    ```sh
    $ pachctl create branch edges_dp@master --head dev
    ```

8. Listing jobs will show that a job got triggered on `montage_dp`:

    ```sh
    $ pachctl list job
    ID                               PIPELINE   STARTED        DURATION  RESTART PROGRESS  DL       UL       STATE
    e5a116fd9c2e4678a0f49fcb2f8c8331 montage_dp 10 seconds ago 4 seconds 0       1 + 0 / 1 919.6KiB 1.055MiB success
    c7e69e46e9954611ad8efc8aeac47f2a edges_dp   42 seconds ago 3 seconds 0       1 + 3 / 4 175.1KiB 92.18KiB success
    2288709b4d8044409c2232d673ec8f23 montage_dp 2 minutes ago  1 second  0       0 + 0 / 0 0B       0B       success
    6d9d4cf0f6524b0ca126fa97141303ea edges_dp   2 minutes ago  4 seconds 0       2 + 1 / 3 181.1KiB 111.4KiB success
    fcaf537975554935b0f15d184d7a0984 edges_dp   2 minutes ago  3 seconds 0       1 + 0 / 1 57.27KiB 22.22KiB success
    ```

9. Commit more data to `images_dp_1`. It will only trigger a job in `edges_dp`:

    ```sh
    $ pachctl put file images_dp_1@master:2GI70mb.jpg -f http://imgur.com/2GI70mb.jpg
    $ pachctl list job
    ID                               PIPELINE   STARTED        DURATION  RESTART PROGRESS  DL       UL       STATE
    65eacaae2e63461bbfc1ed609e8b6f5e edges_dp   8 seconds ago  3 seconds 0       1 + 4 / 5 204KiB   18.89KiB success
    e5a116fd9c2e4678a0f49fcb2f8c8331 montage_dp 13 minutes ago 4 seconds 0       1 + 0 / 1 919.6KiB 1.055MiB success
    c7e69e46e9954611ad8efc8aeac47f2a edges_dp   13 minutes ago 3 seconds 0       1 + 3 / 4 175.1KiB 92.18KiB success
    2288709b4d8044409c2232d673ec8f23 montage_dp 15 minutes ago 1 second  0       0 + 0 / 0 0B       0B       success
    6d9d4cf0f6524b0ca126fa97141303ea edges_dp   15 minutes ago 4 seconds 0       2 + 1 / 3 181.1KiB 111.4KiB success
    fcaf537975554935b0f15d184d7a0984 edges_dp   15 minutes ago 3 seconds 0       1 + 0 / 1 57.27KiB 22.22KiB success
    ```

10. Move the `master` branch in `edges_dp` to point dev, again. It will trigger
    a job against the data currently in dev.

    ```sh
    $ pachctl create branch edges_dp@master --head dev
    $ pachctl list job
    ID                               PIPELINE   STARTED        DURATION  RESTART PROGRESS  DL       UL       STATE
    65eddcb60ae1475aa6d59b2baa69c78e montage_dp 8 seconds ago  5 seconds 0       1 + 0 / 1 938.5KiB 1.066MiB success
    65eacaae2e63461bbfc1ed609e8b6f5e edges_dp   3 minutes ago  3 seconds 0       1 + 4 / 5 204KiB   18.89KiB success
    e5a116fd9c2e4678a0f49fcb2f8c8331 montage_dp 16 minutes ago 4 seconds 0       1 + 0 / 1 919.6KiB 1.055MiB success
    c7e69e46e9954611ad8efc8aeac47f2a edges_dp   16 minutes ago 3 seconds 0       1 + 3 / 4 175.1KiB 92.18KiB success
    2288709b4d8044409c2232d673ec8f23 montage_dp 18 minutes ago 1 second  0       0 + 0 / 0 0B       0B       success
    6d9d4cf0f6524b0ca126fa97141303ea edges_dp   18 minutes ago 4 seconds 0       2 + 1 / 3 181.1KiB 111.4KiB success
    fcaf537975554935b0f15d184d7a0984 edges_dp   18 minutes ago 3 seconds 0       1 + 0 / 1 57.27KiB 22.22KiB success
    ```

### Transactions

After you test deferred processing, you can explore how transactions work in
combination with deferred processing.

1. If you want to run a particular set of data in `images_dp_2` against a
   particular branch of `edges_dp`, you need to perform two operations

    - commit data to `images_dp_2` and
    - point `edges_dp@master` to the specific commit of interest.

    If you do not use a transaction, this will result in two jobs being
    triggered, one for the new commit and a second when we move
    `edges_dp@master` branch.

    - `images_dp_2@master` running against whatever is currently in
      `edges_dp@master`
    - `images_dp_2@master` running against whatever you set `edges_dp@master` to

    Remember that in step 10 above, we performed the `create branch` operation
    against `edges_dp`. Now we perform the commit to `images_dp_2`. and see that
    another job got triggered.

    ```sh
    $ pachctl put file images_dp_2@master:3Kr6Mr6.jpg  -f http://imgur.com/3Kr6Mr6.jpg
    $ pachctl list job
    $ pachctl list job
    ID                               PIPELINE   STARTED        DURATION  RESTART PROGRESS  DL       UL       STATE
    9c97578031544cab9cc5fb64e9d77153 montage_dp 9 seconds ago  5 seconds 0       1 + 0 / 1 1015KiB  1.292MiB success
    65eddcb60ae1475aa6d59b2baa69c78e montage_dp 28 seconds ago 5 seconds 0       1 + 0 / 1 938.5KiB 1.066MiB success
    65eacaae2e63461bbfc1ed609e8b6f5e edges_dp   3 minutes ago  3 seconds 0       1 + 4 / 5 204KiB   18.89KiB success
    e5a116fd9c2e4678a0f49fcb2f8c8331 montage_dp 16 minutes ago 4 seconds 0       1 + 0 / 1 919.6KiB 1.055MiB success
    c7e69e46e9954611ad8efc8aeac47f2a edges_dp   16 minutes ago 3 seconds 0       1 + 3 / 4 175.1KiB 92.18KiB success
    2288709b4d8044409c2232d673ec8f23 montage_dp 18 minutes ago 1 second  0       0 + 0 / 0 0B       0B       success
    6d9d4cf0f6524b0ca126fa97141303ea edges_dp   18 minutes ago 4 seconds 0       2 + 1 / 3 181.1KiB 111.4KiB success
    fcaf537975554935b0f15d184d7a0984 edges_dp   18 minutes ago 3 seconds 0       1 + 0 / 1 57.27KiB 22.22KiB success
    ```

2. If you want to just have one job where `images_dp_2@master` runs against
   whatever you set `edges_dp@master` to, you can use Pachyderm transactions.
   First step is to start a transaction.

    ```sh
    $ pachctl start transaction
    Started new transaction: 11fbbcbd-6cda-42fa-b1fe-cd63b292582e
    ```

3. Once the transaction is started, you start all commits and branch creations
   within the scope of the transaction.

    ```sh
    $ pachctl start commit  images_dp_2@master
    Added to transaction: 11fbbcbd-6cda-42fa-b1fe-cd63b292582e
    de55d4856e814c41a65836321fe672fa
    $ pachctl create branch edges_dp@master --head dev
    Added to transaction: 11fbbcbd-6cda-42fa-b1fe-cd63b292582e
    ```

4)  Before you put any files in a repo, you need to finish the transaction. When
    you run `pachctl finish transaction`, Pachyderm groups all the commits and
    branches together, triggering when the last commit in the transaction is
    finished.

    ```sh
    $ pachctl finish transaction
    Completed transaction with 2 requests: 11fbbcbd-6cda-42fa-b1fe-cd63b292582e
    ```

5)  Commit a file, and job list will show no new jobs.

    ```sh
    $ pachctl put file images_dp_2@master:9iIlokw.jpg -f http://imgur.com/9iIlokw.jpg
    $ pachctl list job
    ID                               PIPELINE   STARTED        DURATION  RESTART PROGRESS  DL       UL       STATE
    9c97578031544cab9cc5fb64e9d77153 montage_dp 18 minutes ago 5 seconds 0       1 + 0 / 1 1015KiB  1.292MiB success
    65eddcb60ae1475aa6d59b2baa69c78e montage_dp 19 minutes ago 5 seconds 0       1 + 0 / 1 938.5KiB 1.066MiB success
    65eacaae2e63461bbfc1ed609e8b6f5e edges_dp   22 minutes ago 3 seconds 0       1 + 4 / 5 204KiB   18.89KiB success
    e5a116fd9c2e4678a0f49fcb2f8c8331 montage_dp 35 minutes ago 4 seconds 0       1 + 0 / 1 919.6KiB 1.055MiB success
    c7e69e46e9954611ad8efc8aeac47f2a edges_dp   36 minutes ago 3 seconds 0       1 + 3 / 4 175.1KiB 92.18KiB success
    2288709b4d8044409c2232d673ec8f23 montage_dp 37 minutes ago 1 second  0       0 + 0 / 0 0B       0B       success
    6d9d4cf0f6524b0ca126fa97141303ea edges_dp   38 minutes ago 4 seconds 0       2 + 1 / 3 181.1KiB 111.4KiB success
    fcaf537975554935b0f15d184d7a0984 edges_dp   38 minutes ago 3 seconds 0       1 + 0 / 1 57.27KiB 22.22KiB success
    ```

6)  Finish the commit that you started during the transaction. That will start
    the job.

    ```
    $ pachctl finish commit images_dp_2@master
    $ pachctl list job
    ID                               PIPELINE   STARTED        DURATION  RESTART PROGRESS  DL       UL       STATE
    76f1e7c311fd4529938653787b1d283a montage_dp 14 seconds ago 6 seconds 0       1 + 0 / 1 1.175MiB 1.587MiB success
    9c97578031544cab9cc5fb64e9d77153 montage_dp 19 minutes ago 5 seconds 0       1 + 0 / 1 1015KiB  1.292MiB success
    65eddcb60ae1475aa6d59b2baa69c78e montage_dp 20 minutes ago 5 seconds 0       1 + 0 / 1 938.5KiB 1.066MiB success
    65eacaae2e63461bbfc1ed609e8b6f5e edges_dp   23 minutes ago 3 seconds 0       1 + 4 / 5 204KiB   18.89KiB success
    e5a116fd9c2e4678a0f49fcb2f8c8331 montage_dp 37 minutes ago 4 seconds 0       1 + 0 / 1 919.6KiB 1.055MiB success
    c7e69e46e9954611ad8efc8aeac47f2a edges_dp   37 minutes ago 3 seconds 0       1 + 3 / 4 175.1KiB 92.18KiB success
    2288709b4d8044409c2232d673ec8f23 montage_dp 38 minutes ago 1 second  0       0 + 0 / 0 0B       0B       success
    6d9d4cf0f6524b0ca126fa97141303ea edges_dp   39 minutes ago 4 seconds 0       2 + 1 / 3 181.1KiB 111.4KiB success
    fcaf537975554935b0f15d184d7a0984 edges_dp   39 minutes ago 3 seconds 0       1 + 0 / 1 57.27KiB 22.22KiB success
    ```

## Summary

Deferred processing with transactions in Pachyderm will give you fine-grained
control of jobs and datums while preserving Pachyderm's advantages of data
lineage and incremental processing.
