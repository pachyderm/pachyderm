# Create an S3-enabled Pipeline

If you want to use Pachyderm with such platforms like Kubeflow or
Apache™ Spark, you need to create an S3-enabled Pachyderm pipeline.
Such a pipeline ensures that data provenance of the pipelines that
run in those external systems is properly preserved and is tied to
corresponding Pachyderm jobs.

Pachyderm can deploys the S3 gateway in the `pachd` pod. Also,
you can deploy a separate S3 gateway instance as a sidecar container
in your pipeline worker pod. The former is
typically used when you need to configure an ingress or egress with
object storage tooling, such as MinIO, boto3, and others. The latter
is needed when you use Pachyderm with external data processing
platforms, such as Kubeflow or Apache Spark, that interact with
object stores but do not work with local file systems.

The master S3 gateway exists independently and outside of the
pipeline lifecycle. Therefore, if a
Kubeflow pod connects through the master S3 gateway, the Pachyderm pipelines
created in Kubeflow do not properly maintain data provenance. When the
S3 functionality is exposed through a sidecar instance in the
pipeline worker pod, Kubeflow can access the files stored in S3 buckets
in the pipeline pod, which ensures the provenance is maintained
correctly. The S3 gateway sidecar instance is created together with the
pipeline and shut down when the pipeline is destroyed.

The following diagram shows communication between the S3 gateway
deployed in a sidecar and the Kubeflow pod.

![Kubeflow S3 gateway](../../../assets/images/d_kubeflow_sidecar.png)

## Limitations

Pipelines exposed through a sidecar S3 gateway have the following limitations:

* As with a standard Pachyderm pipeline, in which the input repo is read-only
and output is write-only, the same applies to using the S3 gateway within
pipelines. The input bucket(s) are read-only and the output bucket that
you define using the `s3_out` parameter is write-only. This limitation
guarantees that pipeline provenance is preserved just as it is with normal
Pachyderm pipelines.

* The `glob` field in the pipeline must be set to `"glob": "/"`. All files
are processed as a single datum. In this configuration, already processed
datums are not skipped which
could be an important performance consideration for some processing steps.

* Join and union inputs are not supported, but you can create a cross or
a single input.

* You can create a cross of an S3-enabled input with a non-S3 input.
For a non-S3 input in such a cross you can still specify a glob pattern.

* Statistics collection for S3-enabled pipelines is not supported. If you
set `"s3_out": true`, you need to disable the `enable_stats`
parameter in your pipeline. 

## Expose a Pipeline through an S3 Gateway in a Sidecar

When you work with platforms like Kubeflow or Apache Spark, you need
to spin up an S3 gateway instance that runs alongside the pipeline worker
pod as a sidecar container. To do so, set the `s3` parameter in the `input`
part of your pipeline specification to `true`. When enabled, this parameter
mounts S3 buckets for input repositories in the S3 gateway sidecar instance
instead of in `/pfs/`. You can set this property for each PFS input in
a pipeline. The address of the input repository will be `s3://<input_repo>`.

You can also expose the output repository through the same S3 gateway
instance by setting the `s3_out` property to `true` in the root of
the pipeline spec.  If set to `true`, Pachyderm creates another S3 bucket
on the sidecar, and the output files will be written there instead of
`/pfs/out`. By default, `s3_out` is set to `false`. The address of the
output repository will be `s3://<output_repo>`, which is always the name
of the pipeline.

You can connect to the S3 gateway sidecar instance through its Kubernetes
service. To access the sidecar instance and the buckets on it, you need
to know the address of the buckets. Because PPS handles all permissions,
no special authentication configuration is needed.

The following text is an example of a pipeline exposed through a sidecar
S3 gateway instance:

```json
{
  "pipeline": {
    "name": "test"
  },
  "input": {
    "pfs": {
      "glob": "/",
      "repo": "s3://images",
      "s3": "true"
    }
  },
  "transform": {
    "cmd": [ "python3", "/edges.py" ],
    "image": "pachyderm/opencv"
  },
  "s3_out": true
}
```

!!! note "See Also:"
    - [Pipeline Specification](../../../../reference/pipeline_spec/#input-required)
    - [Configure Environment Variables](../../../deploy/environment-variables/)
