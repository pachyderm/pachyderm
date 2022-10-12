# Spark writing to `s3_out` in `raw_s3_out` mode

Spark likes to write to `s3_out` in ways that the normal Pachyderm S3 gateway doesn't like.

We have a special alpha `s3_out` feature mode called `raw_s3_out`.
You can trigger it by adding an annotation to your pipeline json:

```
  "metadata": {
    "annotations": {
      "raw_s3_out": "true"
    }
  },
```

Writes to `s3_out` will then work with Spark, especially when Spark is writing a large amount of data. (With the normal S3 gateway, you see slow-downs and errors relating to "copyFile" failing.)

This directory contains a worked example. We've built and pushed the Docker image for you already, so all you need to do is run:

```
pachctl create repo poke_s3
pachctl create branch poke_s3@master
```
This gives us a pipeline to poke to make spark start.

```
pachctl create pipeline -f s3-spark.json
```

And observe that the result (about 190MB of the phrase "INFINITEIMPROBABILITY" repeated over and over again, a homage to [The Guide](https://sites.google.com/site/h2g2theguide/Index/i/149246)) is written to the output repo:

```
pachctl put file -f /etc/passwd poke_s3@master:/test01
```
Pokes the pipeline to start it.
```
pachctl list file spark_s3_demo@master
```
And observe the final files are written therein!


## How it works (advanced details 🤓)

You don't need to know how this mode works to use it, but in case you're interested.

We implement this by passing those S3 requests directly through to the backing S3 store (with some light protocol hacking to rewrite bucket names, paths and authentication credentials).

This means that real S3 (or minio) is processing the complex things that Spark does with the S3 protocol, and when it's finished, we just copy the result back into the output repo in PFS.