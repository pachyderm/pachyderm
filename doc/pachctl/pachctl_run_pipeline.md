## pachctl run pipeline

Run an existing Pachyderm pipeline on the specified commits or branches.

### Synopsis


Run a Pachyderm pipeline on the datums from specific commits. Note: pipelines run automatically when data is committed to them. This command is for the case where you want to run the pipeline on a specific set of data, or if you want to rerun the pipeline.

```
pachctl run pipeline <pipeline> [commits...]
```

### Examples

```
```sh

		# Rerun the latest job for the "filter" pipeline
		$ pachctl run pipeline filter

		# Reprocess the pipeline "filter" on the data from commits a23e4 and bf363
		$ pachctl run pipeline filter a23e4 and bf363

		# Run the pipeline "filter" on the data from the "staging" branch
		$ pachctl run pipeline filter staging
```
```

### Options inherited from parent commands

```
      --no-metrics           Don't report user metrics for this command
      --no-port-forwarding   Disable implicit port forwarding
  -v, --verbose              Output verbose logs
```

