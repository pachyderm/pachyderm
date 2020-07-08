# Skip Failed Datums in Your Pipeline

This example describes how you can use the `err_cmd` and `err_stdin` fields in
your pipeline to fail a datum without failing the job and the whole pipeline.
This feature is useful when you have large datasets and multiple datums, and you
do not need to have them all processed successfully to move to the next step in
your DAG.

For more information about the `err_cmd` command, see [](../../docs/err_cmd.md)

## Prerequisites

Before you begin, verify that you have the following configured in your
environment:

-   Pachyderm version 1.9.x or later
-   A clone of the Pachyderm repository

To clone the Pachyderm repository, run the following command:

```bash
$ git clone git@github.com:pachyderm/pachyderm.git
```

## Create a Repository

The first step is to create a repository called `input` by running the following
command:

```bash
$ pachctl create repo input
```

## Create a Pipeline

Next, you need to create a pipeline that uses the `input` repository as input
and has the `err_cmd` and `err_stdin` fields specified.

In this example, we use the [error_test.json](error_test.json) pipeline:

```json
{
    "pipeline": {
        "name": "error_test"
    "description": "A pipeline that checks if the `file1` is present in the datum.",
    },
    "input": {
        "pfs": {
            "glob": "/*",
            "repo": "input"
        }
    },
    "transform": {
        "cmd": [ "bash" ] ,
        "stdin": [ "if", "[ -a /pfs/input/file1 ]", "then cp /pfs/input/* /pfs/out/", "exit 0",  "fi", "exit 1" ] ,
        "err_cmd": [ "bash" ] ,
        "err_stdin": [ "if", "[ -a /pfs/input/file2 ]",  "then", "exit 0", "fi", " exit 1" ]
    }
}
```

In the pipeline above, the code checks if the datum contains `file1`. If it
does, then the code copies everything in `/pfs/input/` to the `/pfs/out`
directory. If the datum does not include `file1`, the datum is checked against
the code in `err_stdin`. That code checks if the datum has `file2`. If it does,
the code marks the datum as recovered, and the job succeeds. If it does not, the
job fails.

Create a pipeline by running the following command from the `examples/err_cmd/`
directory:

```bash
$ pachctl create pipeline -f error_test.json
```

Verify that the pipeline was successfully created:

```bash
$ pachctl list pipeline
NAME       VERSION INPUT    CREATED       STATE / LAST JOB
error_test 1       input:/* 5 seconds ago running / starting
```

## Add Files to the Input Repository

Now, let's add some files to the input repository to watch how your pipeline
code and error code work.

You will add three files, `file1`, `file2`, and `file3`, that each contains one
line in them.

1. Add `file1`:

    ```bash
    $ echo "foo" | pachctl put file input@master:file1
    ```

    When you add `file1`, your pipeline should succeed:

    ```bash
    $ pachctl list job --no-pager
    ID                               PIPELINE   STARTED        DURATION           RESTART PROGRESS  DL UL STATE
    c8860dae5a054ec38a33068f75fe9690 error_test 13 seconds ago Less than a second 0       1 + 0 / 1 4B 4B success
    ```

    As you can see in the `PROGRESS` column – `1 + 0 / 1`, you have one
    successfully processed datum.

1. Add `file2`:

    ```bash
    $ echo "bar" | pachctl put file input@master:file2
    ```

    Processing of this datum fails, but because the `err_cmd` code ran
    successfully, the datum is marked as _recovered_, and the job finishes
    without errors. Only `file1` is available in the output commit.

    ```bash
    $ pachctl list job --no-pager
    ID                               PIPELINE   STARTED       DURATION           RESTART PROGRESS      DL UL STATE
    bc3da288ff884d5a9bcb312dd6cf07cb error_test 3 seconds ago Less than a second 0       0 + 1 + 1 / 2 0B 0B success
    c8860dae5a054ec38a33068f75fe9690 error_test 3 minutes ago Less than a second 0       1 + 0 / 1     4B 4B success
    ```

    In the `PROGRESS` column, you can see that the last job did not have any
    successfully processed datums, but it had a skipped datum and a recovered
    datum – `0 + 1 + 1 / 2`.

1. Add `file3`:

```bash
$ echo "baz" | pachctl put file input@master:file3
```

Because the processed datum does not have neither `file1`, nor `file2`, this job
results in failure. Therefore, both `cmd` and `err_cmd` codes result in non-zero
status:

```bash
ID                               PIPELINE   STARTED        DURATION           RESTART PROGRESS      DL UL STATE
272370ec03c24cc1be660bf97403712f error_test 26 seconds ago Less than a second 0       0 + 2 / 3     0B 0B failure: failed to process datum:...
bc3da288ff884d5a9bcb312dd6cf07cb error_test 6 minutes ago  Less than a second 0       0 + 1 + 1 / 2 0B 0B success
c8860dae5a054ec38a33068f75fe9690 error_test 10 minutes ago Less than a second 0       1 + 0 / 1     4B 4B success
```
