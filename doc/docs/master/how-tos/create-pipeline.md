# Create a Pipeline

Pipelines are computational entities that run your code. A standard
pipeline injects data from an input repository into your code, runs
your code against that data, and places the results into an output
repository within the Pachyderm file system. To create a pipeline,
you need to define a pipeline specification in the JSON or YAML file
format.

This is a simple example of a Pachyderm pipeline specification:

```# edges.json
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

At the very minimum, a standard pipeline needs to have a name, a
transformation code, a Docker image with your code, and an input
repository with a branch and glob pattern specified. Special types
of pipelines, such as service, cron, and spout,
have other requirements.
For more information, see [Pipelines](../concepts/pipeline-concepts/pipeline/).

When you have a pipeline spec ready, you need to use it to create a pipeline
pod or pods that will run your code.

For more information about property fields that you can define in a pipeline,
see [Pipeline Specification](../reference/pipeline_spec/).

!!! note
    To create a pipeline, you can use either the Pachyderm UI or the CLI.
    This section provides the CLI instructions only. In the UI, follow the
    wizard to create a pipeline.

To create a pipeline, complete the following steps:

1. Create a pipeline specification. For more information, see
[Pipeline Specification](../reference/pipeline_spec/).

1. Create a pipeline pod:

   ```bash
   pachctl create pipeline -f <pipeline_spec>
   ```

1. Verify that the Kubernetes pod has been created for the pipeline:

   ```bash
   pachctl list pipeline
   ```

   **System Response:**

   ```bash
   NAME  VERSION INPUT     CREATED       STATE / LAST JOB   DESCRIPTION
   edges 1       images:/* 5 seconds ago running / starting A pipeline that performs image edge detection by using the OpenCV library.
   ```

   You can also run `kubectl` commands to view the pod that has been created:

   ```bash
   kubectl get pod
   ```

   **System Response:**

   ```bash hl_lines="5"
   NAME                      READY   STATUS    RESTARTS   AGE
   dash-676d6cdf6f-lmfc5     2/2     Running   2          17d
   etcd-79ffc76f58-ppf28     1/1     Running   1          17d
   pachd-5485f6ddd-wx8vw     1/1     Running   1          17d
   pipeline-edges-v1-qhd4f   2/2     Running   0          95s
   ```

   You should see a pod named after your pipeline in the list of pods.

!!! note "See Also:"
    - [Pipelines](../concepts/pipeline-concepts/pipeline/)
    - [Pipeline Specification](../reference/pipeline_spec/)
    - [Update a Pipeline](./updating_pipelines/)
    - [Delete a Pipelie](./delete-pipeline)
