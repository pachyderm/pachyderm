# Service

Service is a special type of pipeline that does not process data but provides
a capability to expose it to the outside world. For example, you can use
a service to serve a machine learning model as an API that has the most
up-to-date version of your data.

The following pipeline spec extract is an example of how you can expose your
Jupyter notebook as a service by adding a `service` field:

```json
{
    "pipeline": {
      "name": "notebook"
    },
    "input": {
        "pfs": {
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

The service section specifies the following parameters:

| Parameter         | Description   |
| ----------------- | ------------- |
| `"internal_port"` | The port that the code running inside the container binds to. |
| `"external_port"` | The port that is exposed outside of the container. You must set this value in the range of `30000 — 32767`. You can access the service from any Kubernetes node through the following address: `http://<kubernetes-host>:<external_port>`. |

!!! Info
        The Service starts running *at the first commit* in the input repo.

!!! Note "See Also:"    
        - [Service](../../../../reference/pipeline-spec/#service-optional)
