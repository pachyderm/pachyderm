## ./pachctl



### Synopsis


Access the Pachyderm API.

Environment variables:
  ADDRESS=<host>:<port>, the pachd server to connect to (e.g. 127.0.0.1:30650).


### Options

```
      --no-metrics   Don't report user metrics for this command
  -v, --verbose      Output verbose logs
```

### SEE ALSO
* [./pachctl auth](./pachctl_auth.md)	 - Auth commands manage access to data in a Pachyderm cluster
* [./pachctl commit](./pachctl_commit.md)	 - Docs for commits.
* [./pachctl completion](./pachctl_completion.md)	 - Install bash completion code.
* [./pachctl copy-file](./pachctl_copy-file.md)	 - Copy files between pfs paths.
* [./pachctl create-pipeline](./pachctl_create-pipeline.md)	 - Create a new pipeline.
* [./pachctl create-repo](./pachctl_create-repo.md)	 - Create a new repo.
* [./pachctl delete-all](./pachctl_delete-all.md)	 - Delete everything.
* [./pachctl delete-branch](./pachctl_delete-branch.md)	 - Delete a branch
* [./pachctl delete-commit](./pachctl_delete-commit.md)	 - Delete an unfinished commit.
* [./pachctl delete-file](./pachctl_delete-file.md)	 - Delete a file.
* [./pachctl delete-job](./pachctl_delete-job.md)	 - Delete a job.
* [./pachctl delete-pipeline](./pachctl_delete-pipeline.md)	 - Delete a pipeline.
* [./pachctl delete-repo](./pachctl_delete-repo.md)	 - Delete a repo.
* [./pachctl deploy](./pachctl_deploy.md)	 - Deploy a Pachyderm cluster.
* [./pachctl diff-file](./pachctl_diff-file.md)	 - Return a diff of two file trees.
* [./pachctl enterprise](./pachctl_enterprise.md)	 - Enterprise commands enable Pachyderm Enterprise features
* [./pachctl file](./pachctl_file.md)	 - Docs for files.
* [./pachctl finish-commit](./pachctl_finish-commit.md)	 - Finish a started commit.
* [./pachctl flush-commit](./pachctl_flush-commit.md)	 - Wait for all commits caused by the specified commits to finish and return them.
* [./pachctl garbage-collect](./pachctl_garbage-collect.md)	 - Garbage collect unused data.
* [./pachctl get-file](./pachctl_get-file.md)	 - Return the contents of a file.
* [./pachctl get-logs](./pachctl_get-logs.md)	 - Return logs from a job.
* [./pachctl get-object](./pachctl_get-object.md)	 - Return the contents of an object
* [./pachctl get-tag](./pachctl_get-tag.md)	 - Return the contents of a tag
* [./pachctl glob-file](./pachctl_glob-file.md)	 - Return files that match a glob pattern in a commit.
* [./pachctl inspect-commit](./pachctl_inspect-commit.md)	 - Return info about a commit.
* [./pachctl inspect-datum](./pachctl_inspect-datum.md)	 - Display detailed info about a single datum.
* [./pachctl inspect-file](./pachctl_inspect-file.md)	 - Return info about a file.
* [./pachctl inspect-job](./pachctl_inspect-job.md)	 - Return info about a job.
* [./pachctl inspect-pipeline](./pachctl_inspect-pipeline.md)	 - Return info about a pipeline.
* [./pachctl inspect-repo](./pachctl_inspect-repo.md)	 - Return info about a repo.
* [./pachctl job](./pachctl_job.md)	 - Docs for jobs.
* [./pachctl list-branch](./pachctl_list-branch.md)	 - Return all branches on a repo.
* [./pachctl list-commit](./pachctl_list-commit.md)	 - Return all commits on a set of repos.
* [./pachctl list-datum](./pachctl_list-datum.md)	 - Return the datums in a job.
* [./pachctl list-file](./pachctl_list-file.md)	 - Return the files in a directory.
* [./pachctl list-job](./pachctl_list-job.md)	 - Return info about jobs.
* [./pachctl list-pipeline](./pachctl_list-pipeline.md)	 - Return info about all pipelines.
* [./pachctl list-repo](./pachctl_list-repo.md)	 - Return all repos.
* [./pachctl migrate](./pachctl_migrate.md)	 - Migrate the internal state of Pachyderm from one version to another.
* [./pachctl mount](./pachctl_mount.md)	 - Mount pfs locally. This command blocks.
* [./pachctl pipeline](./pachctl_pipeline.md)	 - Docs for pipelines.
* [./pachctl port-forward](./pachctl_port-forward.md)	 - Forward a port on the local machine to pachd. This command blocks.
* [./pachctl put-file](./pachctl_put-file.md)	 - Put a file into the filesystem.
* [./pachctl repo](./pachctl_repo.md)	 - Docs for repos.
* [./pachctl restart-datum](./pachctl_restart-datum.md)	 - Restart a datum.
* [./pachctl run-pipeline](./pachctl_run-pipeline.md)	 - Run a pipeline once.
* [./pachctl set-branch](./pachctl_set-branch.md)	 - Set a commit and its ancestors to a branch
* [./pachctl start-commit](./pachctl_start-commit.md)	 - Start a new commit.
* [./pachctl start-pipeline](./pachctl_start-pipeline.md)	 - Restart a stopped pipeline.
* [./pachctl stop-job](./pachctl_stop-job.md)	 - Stop a job.
* [./pachctl stop-pipeline](./pachctl_stop-pipeline.md)	 - Stop a running pipeline.
* [./pachctl subscribe-commit](./pachctl_subscribe-commit.md)	 - Print commits as they are created (finished).
* [./pachctl undeploy](./pachctl_undeploy.md)	 - Tear down a deployed Pachyderm cluster.
* [./pachctl unmount](./pachctl_unmount.md)	 - Unmount pfs.
* [./pachctl update-dash](./pachctl_update-dash.md)	 - Update and redeploy the Pachyderm Dashboard at the latest compatible version.
* [./pachctl update-pipeline](./pachctl_update-pipeline.md)	 - Update an existing Pachyderm pipeline.
* [./pachctl update-repo](./pachctl_update-repo.md)	 - Update a repo.
* [./pachctl version](./pachctl_version.md)	 - Return version information.

###### Auto generated by spf13/cobra on 9-Oct-2017
