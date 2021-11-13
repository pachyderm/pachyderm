## pachctl run pipeline

Run an existing Pachyderm pipeline on the specified commits-branch pairs.

### Synopsis

Run a Pachyderm pipeline on the datums from specific commit-branch pairs. If only the branch is given, the head commit of the branch is used to complete the pair. Note: pipelines run automatically when data is committed to them. This command is for the case where you want to run the pipeline on a specific set of data, or if you want to rerun the pipeline. If a commit or branch is not specified, it will default to using the HEAD of master.

```
pachctl run pipeline <pipeline> [<repo>@<branch>[=<commit>]...] [flags]
```

### Examples

```

		# Rerun the latest job for the "filter" pipeline
		$ pachctl run pipeline filter

		# Process the pipeline "filter" on the data from commit-branch pairs "repo1@A=a23e4" and "repo2@B=bf363"
		$ pachctl run pipeline filter repo1@A=a23e4 repo2@B=bf363

		# Run the pipeline "filter" on the data from commit "167af5" on the "staging" branch on repo "repo1"
		$ pachctl run pipeline filter repo1@staging=167af5
```

### Options

```
  -h, --help         help for pipeline
      --job string   rerun the given job
```

### Options inherited from parent commands

```
      --no-color   Turn off colors.
  -v, --verbose    Output verbose logs
```

