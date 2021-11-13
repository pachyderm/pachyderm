!!! note 
    To deploy and configure a Pachyderm cluster
    to ship logs to Loki, 
    a ***Pachyderm Enterprise License*** is required. 

# Enabling Loki

Enabling Loki logging will require to set the following [environment variables](https://docs.pachyderm.com/latest/deploy-manage/deploy/environment-variables/) on the pachd container:

- `LOKI_LOGGING` to `true`  to ship the logs to Loki
- `LOKI_SERVICE_HOST` and `LOKI_SERVICE_PORT` to fetch the logs from Loki

## Shipping logs to Loki

Loki retrieves logs from pods in Kubernetes through
an agent service called **Promtail**. 
**Promtail** runs on each node and
sends logs from Kubernetes pods to the Loki API Server,
tagging each log entry with information
about the pod that produced it. 

You need to [configure Promtail](https://grafana.com/docs/loki/latest/clients/promtail/configuration/) for your environment
to ship logs to your Loki instance. 
If you are running **multiple nodes**, 
then you will need to install and configure Promtail
**for each node** shipping logs to Loki.


## Fetching logs

While enabling Loki will enable the collection of logs, commands such as `pachctl logs` will not fetch logs directly from Loki until the `LOKI_SERVICE_HOST` and `LOKI_SERVICE_PORT` environment variables are set on the `pachd` container.

For example, a `deployment.json` generated with 
```shell
    pachctl deploy local --dry-run > deployment.json`     
```
can be modified to make logs available for Loki as follows:

```json
{
    "containers": [{
        "name": "pachd",
        "image": "pachyderm/pachd:local",
        "command": ["/pachd"],
        "ports": [],
        "env": [
            {
                "name": "LOKI_LOGGING",
                "value": "true"
            },
            {
                "name": "LOKI_SERVICE_HOST",
                "value": "10.107.254.102"
            },
            {
                "name": "LOKI_SERVICE_PORT",
                "value": "3100"
            }
        ]
    }]
}
```

Pachyderm reads logs from the Loki API Server with a particular set of tags. The URI at which Pachyderm reads from the Loki API Server is set by the `LOKI_SERVICE_HOST` and `LOKI_SERVICE_PORT` values.

!!! note 
    If you are not running Promtail on the node 
    where your Pachyderm pods are located, you
    will be unable to get logs for pipelines running
    on that node via `pachctl logs -p pipelineName`.

## References

* Loki Documentation - https://grafana.com/docs/loki/latest/
* Promtail Documentation - https://grafana.com/docs/loki/latest/clients/promtail/
* Operating Loki - https://grafana.com/docs/loki/latest/operations/
