
# Configure Pipeline Worker Environment Variables

The environment (env) variables in this article define the parameters for your Pachyderm pipeline workers.


## View Current Env Variables

To list your current env variables, you can include the command `env` in the `transform.stdin` attribute inside your pipeline spec. For example, for a repo named `images`, you can view your env variables by using the following:

```json
{
    "pipeline": {
        "name": "env"
    },
    "input": {
        "pfs": {
            "glob": "/",
            "repo": "images"
        }
    },
    "transform": {
        "cmd": ["sh" ],
        "stdin": ["env"],
        "image": "ubuntu:14.04"
    }
}
```

---

## Variables

The following list of variables is not exhaustive; it is, however, a list of the most useful variables.


| Environment Variable       | Description |
| -------------------------- | --------------------------------------------- |
| `PACH_JOB_ID`              | The ID of the current job. For example, <br> `PACH_JOB_ID=8991d6e811554b2a8eccaff10ebfb341`. |
| `PACH_DATUM_ID`             | The ID of the current Datum.|
|`PACH_DATUM_<input.name>_JOIN_ON`|Exposes the `join_on` match to the pipeline's job. |
|`PACH_DATUM_<input.name>_GROUP_BY`|Expose the `group_by` match to the pipeline's job. |
| `PACH_OUTPUT_COMMIT_ID`    | The ID of the commit in the output repo for <br> the current job. For example, <br> `PACH_OUTPUT_COMMIT_ID=a974991ad44d4d37ba5cf33b9ff77394`. |
| `PPS_NAMESPACE`            | The PPS namespace. For example, <br> `PPS_NAMESPACE=default`. |
| `PPS_SPEC_COMMIT`          | The hash of the pipeline specification commit.<br> This value is tied to the pipeline version. Therefore, jobs that use <br> the same version of the same pipeline have the same spec commit. <br> For example, `PPS_SPEC_COMMIT=3596627865b24c4caea9565fcde29e7d`. |
| `PPS_POD_NAME`             | The name of the pipeline pod. For example, <br>`pipeline-env-v1-zbwm2`. |
| `PPS_PIPELINE_NAME`        | The name of the pipeline that this pod runs. <br> For example, `env`. |
| `PIPELINE_SERVICE_PORT_PROMETHEUS_METRICS` | The port that you can use to <br> exposed metrics to Prometheus from within your pipeline. The default value is 9090. |
| `HOME`                     | The path to the home directory. The default value is `/root` |
| `<input-repo>=<path/to/input/repo>` | The path to the filesystem that is <br> defined in the `input` in your pipeline specification. Pachyderm defines <br> such a variable for each input. The path is defined by the `glob` pattern in the <br> spec. For example, if you have an input `images` and a glob pattern of `/`, <br> Pachyderm defines the `images=/pfs/images` variable. If you <br> have a glob pattern of `/*`, Pachyderm matches <br> the files in the `images` repository and, therefore, the path is <br> `images=/pfs/images/liberty.png`. |
| `input_COMMIT`             | The ID of the commit that is used for the input. <br>For example, `images_COMMIT=fa765b5454e3475f902eadebf83eac34`. |
| `S3_ENDPOINT`         | A Pachyderm S3 gateway sidecar container endpoint. <br> If you have an S3 enabled pipeline, this parameter specifies a URL that <br> you can use to access the pipeline's repositories state when a <br> particular job was run. The URL has the following format: <br> `http://<job-ID>-s3:600`. <br> An example of accessing the data by using AWS CLI looks like this: <br>`echo foo_data | aws --endpoint=${S3_ENDPOINT} s3 cp - s3://out/foo_file`. |