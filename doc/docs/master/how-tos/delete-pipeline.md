# Delete a Pipeline

When you no longer need a pipeline, you can delete it by using the
`pachctl delete pipeline` command or by using the UI wizard. When
you run the delete pipeline, Pachyderm destroys the following components:

* The pipeline Kubernetes pod
* The output repository with all data

Only authorized users can delete pipelines.

When you delete a pipeline, all attributes, such as the output repository
and the job history is deleted as well.
You can use the `--keep repo` flag that preserves the output repo with
all its branches and provenance. Only the information about the pipeline
history itself is erased. Later, you can recreate the pipeline, and it
will resume provenance from the last record.

When Pachyderm cannot delete a pipeline with the standard command,
you might need to enforce deletion by using the `--force` flag. Use
this option with caution.

To delete all pipelines, use the `--all` flag.

To delete a pipeline, run the following command:

```bash
pachctl delete pipeline <pipeine_name>
```

!!! note "See Also"
    - [Update a Pipeline](./updating_pipelines/)
    - [Create a Pipeline](./create-pipeline/)
