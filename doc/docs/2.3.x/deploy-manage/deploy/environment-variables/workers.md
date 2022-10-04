# Configure Worker Variables

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


## Variables 

| Variable | Required | Type | Default | Description |
|---|---|---|---|---|
| PPS_WORKER_IP | Yes | string |  | Contains the IP address of the PPS Worker  |
| PPS_POD_NAME | Yes | string |  | Contains the pod name of the PPS worker.  |