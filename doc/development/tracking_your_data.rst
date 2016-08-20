Tracking Your Data Flow
=======================

Typical workflow:
StartCommit
write foo
FinishCommit

Easy during development, but if you have a long chain of pipelines and have frequent commits, how do you know when a specific commit's results are ready?

Also important for programatically consuming data. Just "waiting some amount of time" is a terrible idea.

Wait Until Your Output Is Ready
-------------------------------


Now I want to GetFile commit ... but need to wait until it exists.

Understanding FlushCommit:



Wait for all commits caused by the specified commits to finish and return them.

Synopsis

Wait for all commits caused by the specified commits to finish and return them.

Examples:

# return commits caused by foo/abc123 and bar/def456
$ pachctl flush-commit foo/abc123 bar/def456

# return commits caused by foo/abc123 leading to repos bar and baz
$ pachctl flush-commit foo/abc123 -r bar -r baz
./pachctl flush-commit commit [commit ...]
Options

  -r, --repos value   Wait only for commits leading to a specific set of repos (default [])

