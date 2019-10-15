## pachctl run pipeline

Run an existing Pachyderm pipeline on the specified commits or branches.

### Synopsis


Run a Pachyderm pipeline on the datums from specific commits. Note: pipelines run automatically when data is committed to them. This command is for the case where you want to run the pipeline on a specific set of data, or if you want to rerun the pipeline. If a commit or branch is not specified, it will default to using the HEAD of master.

```
pachctl run pipeline <pipeline> [<repo>@<commit or branch>...]
```

### Examples

```

		# Rerun the latest job for the "filter" pipeline
		$ pachctl run pipeline filter

		# Process the pipeline "filter" on the data from commits repo1@a23e4 and repo2@bf363
		$ pachctl run pipeline filter repo1@a23e4 repo2@bf363

		# Run the pipeline "filter" on the data from the "staging" branch on repo repo1
		$ pachctl run pipeline filter repo1@staging
```

### Options inherited from parent commands

```
      --no-color   Turn off colors.
  -v, --verbose    Output verbose logs
```

