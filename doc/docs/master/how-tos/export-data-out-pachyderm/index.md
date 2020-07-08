# Export Your Data From Pachyderm

After you build a pipeline, you probably want to see the results that the
pipeline has produced. Every commit into an input repository results in a
corresponding commit into an output repository.

To access the results of a pipeline, you can use one of the following methods:

-   By running the `pachctl get file` command. This command returns the contents
    of the specified file.<br> To get the list of files in a repo, you should
    first run the `pachctl list file` command. See
    [Export Your Data with `pachctl`](export-data-pachctl/).<br>

-   By configuring the pipeline. A pipeline can push or expose output data to
    external sources. You can configure the following data exporting methods in
    a Pachyderm pipeline:

    -   An `egress` property enables you to export your data to an external
        datastore, such as Amazon S3, Google Cloud Storage, and others.<br> See
        [Export data by using `egress`](export-data-egress/).<br>

    -   A service. A Pachyderm service exposes the results of the pipeline
        processing on a specific port in the form of a dashboard or similar
        endpoint.<br> See
        [Service](../../../concepts/pipeline-concepts/pipeline/service/).<br>

    -   Configure your code to connect to an external data source. Because a
        pipeline is a Docker container that runs your code, you can egress your
        data to any data source, even to those that the `egress` field does not
        support, by connecting to that source from within your code.

-   By using the S3 gateway. Pachyderm Enterprise users can reuse their existing
    tools and libraries that work with object store to export their data with
    the S3 gateway.<br> See
    [Using the S3 Gateway](../../../deploy-manage/manage/s3gateway/).

-   By mounting your data to a local filesystem with `pachctl mount`. See
    [Mount a Repo to a Local Computer](mount-repo-to-local-computer/)
