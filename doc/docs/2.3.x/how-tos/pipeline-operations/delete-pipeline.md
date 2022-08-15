# Delete a Pipeline
You can delete a pipeline by running:

```shell
pachctl delete pipeline <pipeline_name>
```

To delete all of your pipelines (be careful with this feature), use the additional  `--all` flag.

When you delete a pipeline: 

* Kubernetes deletes all resources associated with the pipeline - pods (if any), services, and replication controllers.
* Pachyderm deletes the user output repository **with all its data** as well as the system `meta` (stats) and `spec` (historical versions of the pipeline specification file) repositories. Check our [Repositories concept page](../../../concepts/data-concepts/repo/#definition) for more details on repositories types.


!!! Note
     If you are using [Pachyderm authorization features](../../enterprise/auth/authorization/index.md), only authorized users will be able to delete a given pipeline. In particular, they will have to be `repoOwner` of the output repo of the pipeline (i.e., have created the pipeline) or `clusterAdmin`. 

You can **use the `--keep-repo` flag to preserve the output repo** with all its branches. However, important job metadata will still be deleted (including all historical versions of the pipeline specification file).
As a result, **you will not be able to recreate the deleted pipeline** with the same name unless that repo is deleted first. 

!!! Example 
     For example, if a pipeline "xyz" exists, then there is an output repo "xyz". If a user deletes the pipeline with `--keep-repo`, the output repo "xyz" will remain, but the pipeline will be gone. If the user tries to create a new pipeline called "xyz", it will fail (there is already an output repo with that name). For the pipeline creation to be successful, the user would have to delete repo "xyz" first.

!!! Note 
     You can use the output repo of a pipeline deleted with `--keep-repo` as an input repo and add more data.  
     
When Pachyderm cannot delete a pipeline with the standard command, you might need to enforce deletion using the `--force` flag. Because this option can break dependent components in your DAG, **use this option withextreme caution**.


!!! Note  "See Also" 
     - [Update a Pipeline](../updating-pipelines/) 
     - [Create a Pipeline](../create-pipeline/)