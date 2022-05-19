# Create a Pipeline

A Pachyderm pipeline is a mechanism that automates a machine learning workflow.
A pipeline reads data from one or more input repositories, runs your code, and
places the results into an output repository within the Pachyderm file system.
To create a pipeline, you need to define a pipeline specification in the JSON
or YAML file format.

This is a simple example of a Pachyderm pipeline specification:

```json
{
  "pipeline": {
    "name": "edges"
  },
  "description": "A pipeline that performs image edge detection by using the OpenCV library.",
  "transform": {
    "cmd": [ "python3", "/edges.py" ],
    "image": "pachyderm/opencv"
  },
  "input": {
    "pfs": {
      "repo": "images",
      "glob": "/*"
    }
  }
}
```

At the very minimum, a standard pipeline needs to have a name, a user code
in the `transform` section, and an input
repository with a glob pattern specified. Special types
of pipelines, such as a service, cron, and spout,
have other requirements.
For more information, see [Pipelines](../../../concepts/pipeline-concepts/pipeline/).

After you have your pipeline spec ready, you need to pass that configuration
to Pachyderm so that it creates a Kubernetes pod or pods that will run your code.

For more information about property fields that you can define in a pipeline,
see [Pipeline Specification](../../../reference/pipeline-spec/).

To create a pipeline, complete the following steps:

1. Create a pipeline specification. For more information, see
[Pipeline Specification](../../../reference/pipeline-spec/).

1. Create a pipeline by passing the pipeline configuration to Pachyderm:

    ```shell
    pachctl create pipeline -f <pipeline_spec>
    ```

1. Verify that the Kubernetes pod has been created for the pipeline:

    ```shell
    pachctl list pipeline
    ```

    **System Response:**

    ```shell
    NAME  VERSION INPUT     CREATED       STATE / LAST JOB   DESCRIPTION
    edges 1       images:/* 5 seconds ago running / starting A pipeline that performs image edge detection by using the OpenCV library.
    ```

    You can also run `kubectl` commands to view the pod that has been created:

    ```shell
    kubectl get pod
    ```

    **System Response:**

    ```shell
    NAME                      READY   STATUS    RESTARTS   AGE
    pachd-5485f6ddd-wx8vw     1/1     Running   1          17d
    pipeline-edges-v1-qhd4f   2/2     Running   0          95s
    ...
    ```

    You should see a pod named after your pipeline in the list of pods.
    In this case, it is `pipeline-edges-v1-qhd4f`.

## Creating a Pipeline When an Output Repository Already Exists

When you create a pipeline, Pachyderm automatically creates an eponymous output
repository. However, if such a repo already exists, your pipeline will take
over the master branch. The files that were stored in the repo before
will still be in the `HEAD` of the branch.

!!! note "See Also:"
    - [Pipelines](../../../concepts/pipeline-concepts/pipeline/)
    - [Pipeline Specification](../../../reference/pipeline-spec/)
    - [Update a Pipeline](../updating-pipelines/)
    - [Delete a Pipeline](../delete-pipeline/)
