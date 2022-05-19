!!! note 
    To deploy and configure a Pachyderm cluster
    to read logs from Loki,
    a ***Pachyderm Enterprise License*** is required. 

## Shipping logs to Loki

Loki retrieves logs from pods in Kubernetes through
an agent service called **Promtail**. 
**Promtail** runs on each node and
sends logs from Kubernetes pods to the Loki API Server,
tagging each log entry with information
about the pod that produced it. 

You need to [configure Promtail](https://grafana.com/docs/loki/latest/clients/promtail/configuration/){target=_blank} for your environment
to ship logs to your Loki instance. 
If you are running **multiple nodes**, 
then you will need to install and configure Promtail
**for each node** shipping logs to Loki.


## Fetching logs

While installing Loki will enable the collection of logs, commands such as `pachctl logs` will not fetch logs directly
from Loki until the `LOKI_LOGGING` environment variable on the `pachd` container is **true**.

This is controlled by the helm value `pachd.lokiLogging`, which can be set by adding the following to your [values.yaml](../../../reference/helm-values/) file:

```yaml
    pachd:
        lokiLogging: true
```

Pachyderm reads logs from the Loki API Server with a particular set of tags. 
The URI at which Pachyderm reads from the Loki API Server is determined by the `LOKI_SERVICE_HOST` and `LOKI_SERVICE_PORT` environment values **automatically added by Loki Kubernetes service**. 

If Loki is deployed after the `pachd` container,
the `pachd` container will need to be redeployed to receive these connection parameters.

!!! note 
    If you are not running Promtail on the node 
    where your Pachyderm pods are located, you
    will be unable to get logs for pipelines running
    on that node via `pachctl logs -p pipelineName`.

## References

* Loki Documentation - https://grafana.com/docs/loki/latest/
* Promtail Documentation - https://grafana.com/docs/loki/latest/clients/promtail/
* Operating Loki - https://grafana.com/docs/loki/latest/operations/
