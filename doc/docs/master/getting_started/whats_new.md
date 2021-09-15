# So What's New?

This new 2.0.0 release comes with major architectural changes, 
and functional enhancements.

## Why? 

Primarily, this is a long overdue refactoring meant to increase Pachyderm performance and reduce resource consumption.
We have also added many changes to our Enterprise Authorization and Authentication module, a compelling way to track data provenance and lineage, and you now deploy Pachyderm with Helm.

Last but not least, we have a brand new Web UI (`Console`) replacing our old Dashboard and a new integrated development environment (`Notebooks`) - namely, JupyterLab on Pachyderm - both of which are readily accessible on [Hub](https://hub.pachyderm.com/) or as an enterprise feature of our core product.

 
## Major Changes And New Features:


### Architectural Changes
Our storage is now optimized for reproducing any state of the system at any point in time.
How? In brief:

- most of Pachyderm's metadata are stored in Postgresql
- data is deduplicated below the level of files, in 64-byte chunks,
improving performance and decreasing storage costs. 

Check our new, high level, [architecture overview](../../deploy-manage/) featuring [Postgresql](https://www.postgresql.org/docs/).


### Helm
As of this release, [Helm is the authoritative deployment method for Pachyderm](../../deploy-manage/deploy/helm_install/)

Check our [reference values.yaml](../../reference/helm_values/) for all possible configuration options or choose the page that fits your particular target deployment in the `Production Deployment` menu.

!!! Warning
    `pachctl deploy` is now [EOL](../../contributing/supported-releases/#end-of-life-eol) and replaced by the package manager Helm.

### Global ID And New Scope For Commits And Jobs
One identifier (`Global ID`) is all that is needed to determine the provenance of data across a complex DAG.

!!! Note "Underlying Key Concepts"
      We recommend to read:

      - the [Repository](../../concepts/data-concepts/repo/) page as we have introduced new types of repositories
      - the [Commit](../../concepts/data-concepts/commit/#definition) page in which we have explained the `origin` of a commit 


Check this simple, didactical, example to grasp how [one single ID lets you track all provenance-dependent commits and jobs](https://github.com/pachyderm/pachyderm/tree/master/examples/globalID) in one command.  

Or visit the [Global ID](../../concepts/advanced-concepts/globalID/) page.

!!! Info
    [Transactions](../../how-tos/advanced-data-operations/use-transactions-to-run-multiple-commands/#use-transactions) also use a single identifier.

### Single Datum Provenance And New Map/Reduce Pattern     
 You can now control and implement a [merge behavior](../../concepts/pipeline-concepts/datum/relationship-between-datums/#5-next-add-a-reduce-pipeline) that fits your use case:
 In Pachyderm 1, multiple datums from the same input repo in the same job in a pipeline could write to the same output file and the results would be merged, with indeterminate ordering of results in the merged file. In Pachyderm 2, **if two datums from the same repo write to the same output file, it will raise an error** (Check our illustration of this [`Single Datum Provenance Rule`](concepts/pipeline-concepts/datum/relationship-between-datums/#example-two-steps-mapreduce-pattern-and-single-datum-provenance-rule). All pipelines relying on a merge behavior can add a following "Reduce" pipeline that groups the files into single datums using filename metadata and merges them according to the appropriate use case by using their own code.
### Default Overwrite Behavior
- The default upload behavior changes from append to [overwrite](../../concepts/data-concepts/file/#overwriting-files): `pachctl put file` now **overwrite files by default**.


### New Authentification, Authorization, And Enterprise Server

Pachyderm now includes new [Enterprise Management](../../enterprise/auth/enterprise-server/setup/) options which allow for site-wide configuration of licensing, authentication and [access control](../../enterprise/auth/authorization/), as well as single-point Pachyderm configuration synchronization. With one command, your users can now gain access to every cluster in your enterprise, with the appropriate level of access control in each cluster. It also allows for [authentication against any OIDC provider](../../enterprise/auth/authentication/idp-dex/).


### New Console and Notebooks
We are introducing a new web UI, the Pachyderm `Console`, that replaces the Dashboard in Pachyderm 1. 
This is a first iteration of the product in which we have focused on DAG visualization, allowing for easy access to job and log information in your pipelines.

We have also kicked off a first iteration of our Notebooks product in which you will be able to run your pipelines and data experiments from your favorite Jupiter notebook.

### More
- No more empty directories in repos
In Pachyderm 2, directories are implied from the paths of the files. They are no longer explicit objects in the file system. Directories will not be created in input or output repos unless they contain at least ome file.


!!! See Also "More"
      - [Read our blog](https://www.pachyderm.com/blog/getting-ready-for-pachyderm-2/) 
      - [Changelog](https://github.com/pachyderm/pachyderm/blob/master/CHANGELOG.md)



