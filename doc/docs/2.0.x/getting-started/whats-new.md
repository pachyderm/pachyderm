# New Features and Major Changes

`Pachyderm 2.0`, also referred to as `Pachyderm 2`, comes with significant architectural refactoring and functional enhancements. 

In this page, we highlight Pachyderm 2 notable changes 
and point to their relevant documentation. 

!!! Note
    Not all changes have a visible impact; therefore some might not be mentioned.

    For a complete overview of what Pachyderm 2 entails, [read this blog post](https://www.pachyderm.com/blog/getting-ready-for-pachyderm-2/){target=_blank}. 

 So, what is new?

## Major Changes

### Architecture Overhaul

The most important change in Pachyderm 2 is its **storage architecture**.

- Most of Pachyderm's metadata storage has been moved from etcd to [**Postgresql**](https://www.postgresql.org/docs/){target=_blank}. 

    This change ultimately alters the deployment of Pachyderm in production. Along with the required object store for storing your data, you now need to create a managed instance of PostgreSQL on your favorite Cloud (RDS on AWS, CloudSQL on GCP, PostgreSQL Server on Azure). These deployment changes are addressed in the [Helm](./#helm) section below.

    To understand how this new PostgreSQL component plays in the overall architecture of the product, take a look at the high-level [architectural overview](../../deploy-manage/) of Pachyderm. Additionally, this [infrastructure diagram](../../deploy-manage/deploy/ingress/#deliver-external-traffic-to-pachyderm) will give you a high-level picture of the physical architecture of Pachyderm deployed in the enterprise.


!!! Note
      By default, Pachyderm runs with a bundled version of PostgreSQL for quick and easy deployment. This is a suitable environment to design your pipelines or simply test the product. 
      
      However, for production settings, Pachyderm **strongly recommends that you disable the bundle and use a managed PostgreSQL instance** instead. 
 

- Although this is not part of this documentation, it might be interesting to note that we have changed the inner working of how data is broken down into "chunks". 

    We transitioned from a `content addressing` strategy (which would sometimes lead to storage and performance issues depending on the edit pattern of an existing file) to a **`content-defined chunking`** strategy. As a result, storing old versions of files can be done efficiently regardless of deletes, updates, or insertions, thus optimizing storage and simplifying the deduplication. 


### Helm

**Helm** is now the authoritative deployment method for Pachyderm. 

Read about the general [principles of a deployment with Helm](../../deploy-manage/deploy/helm-install/) in Pachyderm in our Deployment section.

All existing `pachctl deploy` commands are [EOL](../../reference/supported-releases/#end-of-life-eol). You can now configure the Helm values passed to [Pachyderm's chart](https://artifacthub.io/packages/helm/pachyderm/pachyderm){target=_blank} depending on your targeted setup.

- Look at how to configure your Helm values:

    - For a quick deployment on a [specific Cloud target](../../deploy-manage/deploy/quickstart/), or [Locally](../local-installation/), on your machine's Docker Desktop or Minikube.
    - For [Production](../../deploy-manage/deploy/) settings.

- As a reference, check the complete list of all configurable fields in our [Reference](../../reference/helm-values/) section or on [GitHub](https://github.com/pachyderm/pachyderm/blob/master/etc/helm/pachyderm/values.yaml){target=_blank}.

### Elimination Of Automatic Merge
 
In Pachyderm 1, multiple [datums](../../concepts/pipeline-concepts/datum/) from the same input repo, in a same job, could write to the same output file. The results would be automatically merged, with an indeterminate ordering of results in the merged file. 

In Pachyderm 2, **if two datums from the same repo write to the same output file, it will raise an error**.   All pipelines relying on a merge behavior now  need to **add a following "Reduce" pipeline** that groups the files into single datums (using filename metadata) and merges them by using their own code. As a result, you now  **control your own [merge behavior](../../concepts/pipeline-concepts/datum/relationship-between-datums/#5-next-add-a-reduce-merge-pipeline)**.


Check our illustration of this new [`Single Datum Provenance Rule`](../../concepts/pipeline-concepts/datum/relationship-between-datums/#example-two-steps-mapreduce-pattern-and-single-datum-provenance-rule) in our documentation. 
 
Alternatively, you can take a look at an implementation of the addition of a new "Reduce/Merge" pipeline in our [examples](https://github.com/pachyderm/pachyderm/tree/2.0.x/examples/joins){target=_blank}.
 
### Default Overwrite Behavior

The default upload behavior changes from append to [overwrite](../../concepts/data-concepts/file/#overwriting-files): `pachctl put file` now **overwrites files by default**. 

For example, if you
have a file `foo` in the repository `A`
and add the same file `foo` to that repository again by
using the `pachctl put file` command, Pachyderm will
overwrite that file in the repo. 

For more information, and learn how to change this behavior, see [File](../../concepts/data-concepts/file/).

### More Changes 

In Pachyderm 2, directories are implied from the paths of the files. Directories will not be created in input or output repos [unless they contain at least one file](../../concepts/data-concepts/file/#file).

## New Common Core Feature

### Global ID 

Global ID can be seen as **a shared TAG or identifier for all provenance-dependent commits and jobs**.
In other words, an initial change to your data (For example, a `put file` in a repository, an `update pipeline`...) triggers a set of related commits and jobs in your data-driven DAG. The set of those commits and jobs will share the same identifier.

Visit the [Global ID](../../concepts/advanced-concepts/globalID/) page to learn more about Global ID or check this didactical example to understand how [one single ID lets you track all provenance-dependent commits and jobs](https://github.com/pachyderm/pachyderm/tree/2.0.x/examples/globalID){target=_blank} at once.  

!!! Note
    Pachyderm [transactions](../../how-tos/advanced-data-operations/use-transactions-to-run-multiple-commands/#use-transactions) also use Global ID. 


## New Enterprise Features And Tooling

### New Authentication, Role-Based Access Control, And Enterprise Server
User Access Management is an Enterprise feature.

- **Authentication**: Bring your Identity Provider.

    Pachyderm allows for [authentication against any OIDC provider](../../enterprise/auth/authentication/idp-dex/). As a result, users can now authenticate to Pachyderm using their existing credentials from various back-ends.

- **Authorization**: 

    We have added a [Role Based Access Control model (RBAC)](../../enterprise/auth/authorization) to Pachyderm's resources. You can assign Roles to Users, granting them a set of permissions on Pachyderm's Ressources (Repo and Cluster).

- **Enterprise Server**: Single-point management of multiple clusters.

    Pachyderm now includes a new [Enterprise Management](../../enterprise/auth/enterprise-server/setup/) capability which allows for site-wide configuration of licensing, authentication, and access control.


### New Console 

We have entirely re-worked our Web UI (`Console`): Console replaces our Dashboard in Pachyderm 1.x.

This is the first iteration of the product in which we have focused on pipeline visualization and data exploration, allowing for easy access to commits, files display, jobs, and logs information.

Console is an Enterprise feature. By adding the relevant fields in your Helm values, you can [deploy Console with Pachyderm](../../deploy-manage/deploy/console/). The deployment of Console in production requires the setup of an [Ingress Controller and a DNS](../../deploy-manage/deploy/ingress/).

!!! Note
    [Deploy Console Locally](../../deploy-manage/deploy/console/#deploy-locally) and browse through your DAGs' pipelines, check the content of your commits in a repo, look at the files they contain, check your DAG's jobs, or zoom in on their logs.


!!! Info "See Also"
    Check our [Changelog](https://github.com/pachyderm/pachyderm/blob/master/CHANGELOG.md){target=_blank}.





