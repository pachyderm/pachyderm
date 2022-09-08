# New Features and Major Changes

`Pachyderm 2.1`, comes with new features and performance improvements. 

On this page, we highlight Pachyderm 2.1 notable changes and point to their relevant documentation. 

<!!! Note 
     Not all changes have a visible impact; therefore, some might not be mentioned.
     Refer to the Changelog at the bottom of this page for an exhaustive list of what 2.1 entails.
        <!--For a complete overview of what Pachyderm 2.1 entails, [read this blog post](https://www.pachyderm.com/blog/getting-ready-for-pachyderm-2/){target=_blank}. />
 
 So, what is new?
## New Common Core Feature
### Jsonnet Pipeline Specifications 

A jsonnet pipeline specification file is a wrapping function atop a JSON file, allowing you to parameterize pipeline specifications, thus adding a dynamic component to the creation and update of pipelines.

Visit the [Jsonnet pipeline specifications](../../how-tos/pipeline-operations/jsonnet-pipeline-specs/) page to learn more about our parameterized pipeline specs. 

Additionally, once you are familiar with the [opencv](https://github.com/pachyderm/pachyderm/tree/{{ config.pach_branch }}/examples/opencv){target=_blank}, [group](https://github.com/pachyderm/pachyderm/tree/{{ config.pach_branch }}/examples/group){target=_blank} or the [join](https://github.com/pachyderm/pachyderm/tree/{{ config.pach_branch }}/examples/joins){target=_blank} examples, we have added a Jsonnet version of their pipeline specs ([opencv jsonnet pipeline specs](https://github.com/pachyderm/pachyderm/tree/{{ config.pach_branch }}/examples/opencv/jsonnet){target=_blank}, [retail group jsonnet pipeline spec](https://github.com/pachyderm/pachyderm/blob/{{ config.pach_branch }}/examples/group/pipelines/retail/retail_group.jsonnet){target=_blank}, [outer join jsonnet pipeline spec](https://github.com/pachyderm/pachyderm/blob/{{ config.pach_branch }}/examples/joins/pipelines/outer/outer_join.jsonnet){target=_blank}) as an illustration of how they can be used. 
### More Change

### More useful changes:

- We are now exposing the capture groups of the `JOIN_ON` and `GROUP_BY` to the pipelines jobs in environment variables:

    - The content of the capture group defined in the [`group_by`](../../concepts/pipeline-concepts/datum/group/){target=_blank} parameter is available to your pipeline's code in the environment variable: `PACH_DATUM_<input.name>_GROUP_BY`. 
    - Similarly, the content of the capture group defined in the [`join_on`](../../concepts/pipeline-concepts/datum/join/){target=_blank} parameter is available to your pipeline's code in the environment variable: `PACH_DATUM_<input.name>_JOIN_ON`.
    This new addition will tremendously simplify your pipeline code. For illustrative purposes, we have used it in our examples.

- Pachyderm now ships with a [embedded Loki](../../deploy-manage/deploy/loki#default-loki-bundle) for a more persistent log collection.
## New Enterprise Features 
### Console Improvements

- We are introducing a new **Getting Started Tutorial** built right into Console to teach users about Pachyderm's key concepts, such as Data Versioning and Pipelines.
- **Drag-and-Drop File Upload**: Console now supports file ingress from your local computer via an intuitive UI.
- **List View**: Along with your traditional DAG view, our new List View provides an alternative to the DAG view that makes it easier to search and filter through all your repos and pipelines. 

!!! Info "See Also" 
     Check our [Changelog](https://github.com/pachyderm/pachyderm/blob/master/CHANGELOG.md){target=_blank}.




