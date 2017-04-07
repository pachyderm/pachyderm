**Note**: This is a Pachyderm pre version 1.4 tutorial.  It needs to be updated for the latest versions of Pachyderm.

# Exporting Pachyderm Data with SQL

## This tutorial is incomplete

I've started committing the files used by this tuturial before the full
tutorial is ready so that users can see an outline of how to use command-line
utilities to interact with external systems from pachyderm (this tutorial uses
the `psql` tool to write data to Amazon Redshift). However, the tutorial isn't
finished, many of the pieces haven't been tested, and this particular example
will soon be obsolete, as pachyderm will soon provide native support for
writing output to SQL databases from pipelines.

That said, some very basic notes on writing data to Redshift from Pachyderm:
- Amazon Redshift speaks the PostgreSQL wire protocol, so any postgres client
  can be used to get data into Redshift. This example uses `psql`

- Since figuring out how to get `psql` into a container seemed hard, I used
  postgres:9.6.1-alpine as the base container for my pipeline. In addition to
  the entire implementation of PostgreSQL, this container has a copy of the
  `psql` client

- Also, since `psql` can only execute SQL queries, I wrote a little go script
  (in `json_to_sql`) that consumes arbitrary json records and outputs SQL
  commands. Since go binaries are statically linked, it's possible to just add
  the compiled binary to the pipeline container image (see `Dockerfile`) and
  run it in the pipeline command (see `transform.stdin` in `pipeline.json`)

  - If you actually want to do this, you'll need to build the docker container
    described by `Dockerfile`. That will look something like:
    ```
    $ DOCKER_IMAGE_NAME=msteffenpachyderm/to_sql
    $ cd json_to_sql && go build to_sql.go && cd .. && docker build ./ -t "${DOCKER_IMAGE_NAME}"
    $ docker push "${DOCKER_IMAGE_NAME}"
    ```

  - Then, set the `transform.image` field in `pipeline.json` to the docker
    image you just pushed

- For `psql` to connect to Redshift, you need to give it your Redshift
  credentials.  Fortunately, Pachyderm makes it easy to access [Kubernetes
  secrets](https://kubernetes.io/docs/user-guide/secrets/) from inside pipeline
  containers.  You can use this to authenticate with Redshift by:
    - creating a [pgpass
      file](https://www.postgresql.org/docs/9.4/static/libpq-pgpass.html) with
      your Redshift credentials
    - creating a Kubernetes secret containing that file, and then
    - setting the PGPASSFILE environment variable in the pipeline to point to
      the Kubernetes secret (see `pipeline.json` for an outline of how that
      looks. The `chmod` command at the beginning is necessary because `psql`
      won't use a pgpass file that's too accessible).

- The redshift pipeline also needs information about your Redshift cluster to
  find it. See the `REDSHIFT_*` environment variables defined in
  `pipeline.json`

- Finally, make sure you set up your network ingress/egress rules
  appropriately. EC2 nodes and Redshift clusters can't talk to each other by
  default
