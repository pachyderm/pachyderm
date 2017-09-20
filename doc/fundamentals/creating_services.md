# Creating Services

Services are special kinds of pipelines, rather then processing data, they
serve it to the outside world. For example you might use a service to expose a
Jupyter notebook that's always got the most up-to-date version of your data
exposed to it. Creating a service is much like creating a pipeline, the only
difference is that your pipeline spec should contain a `"Service"` field. This
is an example of Jupyter service:

```json
{
    "input": {
        "atom": {
            "glob": "/",
            "repo": "input"
        }
    },
    "service": {
        "external_port": 30888,
        "internal_port": 8888
    },
    "transform": {
        "cmd": [
            "start-notebook.sh"
        ],
        "image": "jupyter/datascience-notebook"
    }
}
```

# Accessing Services

The service section specifies 2 ports, `"internal_port"` and `"external_port"`.
`"internal_port"` is the port that the code running inside the container (in
this case Jupyter) binds to. `"external_port"` is the port that will be
exposed outside the container, this value must be in the range `30000-32767`.
Once the service is created you should be able to access it by going to
`http://<kubernetes-host>:<external_port>` on any of the kubernetes nodes.
