## pachctl run pipeline

Run an existing Pachyderm pipeline on the specified commits-branch pairs.

### Synopsis

Run a Pachyderm pipeline on the datums from specific commit-branch pairs. If you only specify a branch, Pachyderm uses the HEAD commit to complete the pair. Similarly, if you only specify a commit, Pachyderm will try to use the branch the commit originated on. Note: Pipelines run automatically when data is committed to them. This command is for the case where you want to run the pipeline on a specific set of data.

```
pachctl run pipeline <pipeline> [<repo>@[<branch>|<commit>|<branch>=<commit>]...] [flags]
```

### Examples

```

		# Rerun the latest job for the "filter" pipeline
		pachctl run pipeline filter

		# Process the pipeline "filter" on the data from commit-branch pairs "repo1@A=a23e4" and "repo2@B=bf363"
		pachctl run pipeline filter repo1@A=a23e4 repo2@B=bf363

		# Run the pipeline "filter" on the data from commit "167af5" on the "staging" branch on repo "repo1"
		pachctl run pipeline filter repo1@staging=167af5

		# Run the pipeline "filter" on the HEAD commit of the "testing" branch on repo "repo1"
		pachctl run pipeline filter repo1@testing

		# Run the pipeline "filter" on the commit "af159e which originated on the "master" branch on repo "repo1"
		pachctl run pipeline filter repo1@af159
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

