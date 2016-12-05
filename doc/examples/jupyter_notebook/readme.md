# Jupyter Notebook using versioned Pachyderm Data

1) Create a repo w data you would like to access

```
pachctl create-repo foo
pachctl put-file foo master iris.csv -c -f iris.csv
```

2) Deploy a Jupyter Notebook using a Pachyderm Service

Note the `jupyter.json` file in this directory. We use a standard jupyter docker image, and run the Jupyter notebook webserver as part of the Pachyderm Transform. 

Just like a normal Pachyderm Job, a container is created with a specific version of any data sets loaded into the container's filesystem.

In this case we'll see our data under `/pfs/foo`

To deploy the service:

```
pachctl create-job -f jupyter.json
```
