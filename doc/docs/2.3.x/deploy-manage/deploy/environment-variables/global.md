
# Configure Global Variables

## Variables

| Variable | Required | Type | Default | Description |
|---|---|---|---|---|
| CLUSTER_DEPLOYMENT_ID | No | string |  | The Cluster's deployment ID.  |
| ETCD_PREFIX | No | string |  |  |
| ETCD_SERVICE_HOST | Yes | string |  | The etcd service host name. |
| ETCD_SERVICE_PORT | Yes | string |  | The etcd port number. |
| GC_PERCENT | No | int  | 50 |  |
| GOOGLE_CLOUD_PROFILER_PROJECT | No | string |  | The name of a GCP project; enables GCP-specific continuous profiling and sends profiles to the named project. Requires pachd to have Google application credentials. |
| IDENTITY_SERVER_DATABASE | No | string | dex |  |
| LOG_FORMAT | No | string | json | The output format of logs. |
| LOKI_SERVICE_HOST_VAR | No | string | LOKI_SERVICE_HOST | The Loki service host name. |
| LOKI_SERVICE_PORT_VAR | No | string  | LOKI_SERVICE_PORT | The Loki service port number. |
| METRICS | No | boolean | TRUE |  |
| METRICS_ENDPOINT | No | string |  |  |
| OIDC_PORT | No | uint16  | 1657 | The OIDC port number. |
| PACH_NAMESPACE | No | string | default |  |
| PACH_ROOT | No | string | /pach |  |
| PACHD_SERVICE_HOST | No | string |  | The pachd service host name. |
| PACHD_SERVICE_PORT | No | string |  | The pachd service port number.  |
| PACHYDERM_ENTERPRISE_ETCD_PREFIX | No | string | pachyderm_enterprise |  |
| PEER_PORT | No | uint16 | 1653 | The pachd-to-pachd port number. |
| PG_BOUNCER_HOST | Yes | string |  | The PG Bouncer host name. |
| PG_BOUNCER_MAX_IDLE_CONNS | No | int  | 10 | The maximum number of idle PG Bouncer connections. |
| PG_BOUNCER_MAX_OPEN_CONNS | No | int  | 10 | The maximum number of open PG Bouncer connections. |
| PG_BOUNCER_PORT | Yes | int  |  | The PG Bouncer port number. |
| POSTGRES_CONN_MAX_IDLE_SECONDS | No | int  | 0 | The maximum lifetime of an idle connection, measured in seconds. |
| POSTGRES_CONN_MAX_LIFETIME_SECONDS | No | int  | 0 | The maximum lifetime of a connection, measured in seconds. |
| POSTGRES_DATABASE | Yes | string |  | The Postgres database name. |
| POSTGRES_HOST | No | string |  | The Postres host name. |
| POSTGRES_MAX_IDLE_CONNS | No | int  | 10 | The maximum number of idle Postgres connections. |
| POSTGRES_MAX_OPEN_CONNS | No | int  | 10 | The maximum number of open Postgres connections. |
| POSTGRES_PASSWORD | No | string |  | The Postgres username's password. Pachyderm does not use this account; this password is only required so that administrators can manually perform administrative tasks. |
| POSTGRES_PORT | No | int  |  | The Postgres port number.  |
| POSTGRES_SSL | No | string | disable | The Postgres SSL certificate. |
| POSTGRES_USER | Yes | string |  | The Postgres username. Any sort of install that needs to be secure must specify a secure password here, or provide the `postgresqlExistingSecretName` and `postgresqlExistingSecretKey` secret. |
| PPS_ETCD_PREFIX | No | string | pachyderm_pps |  |
| PPS_MAX_CONCURRENT_K8S_REQUESTS | No | int  | 10 | The number of concurrent requests that the PPS Master can make against kubernetes. |
| PPS_PIPELINE_NAME | No | string |  | The name of the pipeline that this worker belongs to; only set for workers and sidecar pachd instances. |
| PPS_SPEC_COMMIT | No | string |  | The ID of the pipeline that this worker belongs to; only set for workers and sidecar pachd instances. |
| PPS_WORKER_GRPC_PORT | No | uint16 | 1080 | The GRPs port number. |
| PROMETHEUS_PORT | No | uint16 | 1656 | The Prometheus port number. |
| S3GATEWAY_PORT | No | uint16 | 1600 | The s3 gateway port number. |
| SESSION_DURATION_MINUTES | No | int  | 43200 | The duration auth tokens are valid for; defaults to 30 days. |