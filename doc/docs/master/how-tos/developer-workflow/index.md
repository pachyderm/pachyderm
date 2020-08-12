# Developer Workflow

In general, the developer workflow for Pachyderm involves adding 
data to a versioned data repository and adding a pipeline that 
reads from that data repository to execute your code against that data. 
Both the data and pipeline can be iterated on independenly with Pachyderm
handling the code execution according to the pipeline specfication.
The workflow steps are shown below. 

![Developer workflow](../../assets/images/d_steps_analysis_pipeline.svg)

## Data Workflow - Load Your Data into Pachyderm

You need to add your data to Pachyderm so that your pipeline runs your code
against it. You can do so by using one of the following methods:

* By using the `pachctl put file` command
* By using a special type of pipeline, such as a [spout](../../concepts/pipeline-concepts/pipeline/spout/) or [cron](../../concepts/pipeline-concepts/pipeline/cron/) 
* By using one of the Pachyderm's [language clients](../../../reference/clients/)
* By using a compatible S3 client
* By using the Pachyderm UI (Enterprise version or free trial)

For more information, see [Load Your Data Into Pachyderm](../load-data-into-pachyderm/).

## Pipeline Workflow - Processes Data in Pachyderm

The fundamental concepts of Pachyderm are very powerful,  but the manual build steps mentioned in the [pipeline workflow](working-with-pipelines.md) can become cumbersome during development cycles. We've created a few helpful developer workflows and tools to automate steps that are error-prone or repetitive:

* [Build Pipelines](build-pipelines.md) map code changes into a the pipeline using a default base Docker image without rebuilding it. They are most useful when iterating on the code, with few changes to the Docker image.
* The [build flag](build-flag.md) or `--build` is a optional flag that can be passed to the `create` or `update` pipeline command. This option is most useful when iterating on the Docker image, since it rebuilds and pushes the image before updating the pipeline. 
* [CI/CD Integration](ci-cd-integration.md) provides a way to incorporate Pachyderm functions into the CI process. This is most useful when working with a complex project or for code collaboration. 
* [Create Python_Pachyderm](https://pachyderm.github.io/python-pachyderm/python_pachyderm.m.html#python_pachyderm.create_python_pipeline) is Python-specific way to quickly update pipelines and was the predecessor to Build Pipelines. They are only available for Python via the [Python Pachyderm](https://github.com/pachyderm/python-pachyderm) package. This tool can be useful when using the [Pachyderm IDE](../use-pachyderm-ide).
