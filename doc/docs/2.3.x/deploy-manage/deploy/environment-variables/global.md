
# Configure Global Variables 

## Variables  

| Variable | Type | Default Value | Description | Required |
|---|---|---|---|---|
| `ETCD_SERVICE_HOST` | string |  |  | Yes |
| `ETCD_SERVICE_PORT` | string |  |  | Yes |
| `PPS_WORKER_GRPC_PORT` | uint16 | 1080 |  | No |
| `PORT` | uint16 | 1650 |  | No |
| `PROMETHEUS_PORT` | uint16 | 1656 |  | No |
| `PEER_PORT` | uint16 | 1653 |  | No |
| `S3GATEWAY_PORT` | uint16 | 1600 |  | No |
| `PPS_ETCD_PREFIX` | string | pachyderm_pps |  | No |
| `PACH_NAMESPACE` | string | default |  | No |
| `PACH_ROOT` | string | /pach |  | No |
| `GC_PERCENT` | int  | 50 |  | No |
| `LOKI_SERVICE_HOST_VAR` | string | LOKI_SERVICE_HOST |  | No |
| `LOKI_SERVICE_PORT_VAR` | string  | LOKI_SERVICE_PORT |  | No |
| `OIDC_PORT` | uint16  | 1657 |  | No |
| `PG_BOUNCER_HOST` | string |  |  | Yes |
| `PG_BOUNCER_PORT` | int  |  |  | Yes |
| `POSTGRES_SSL` | string | disable |  | No |
| `POSTGRES_HOST` | string |  |  | No |
| `POSTGRES_PORT` | int  |  |  | No |
| `POSTGRES_DATABASE` | string |  |  | Yes |
| `POSTGRES_USER` | string |  |  | Yes |
| `POSTGRES_PASSWORD` | string |  |  | No |
| `POSTGRES_MAX_OPEN_CONNS` | int  | 10 |  | No |
| `POSTGRES_MAX_IDLE_CONNS` | int  | 10 |  | No |
| `PG_BOUNCER_MAX_OPEN_CONNS` | int  | 10 |  | No |
| `PG_BOUNCER_MAX_IDLE_CONNS` | int  | 10 |  | No |
| `POSTGRES_CONN_MAX_LIFETIME_SECONDS` | int  | 0 |  | No |
| `POSTGRES_CONN_MAX_IDLE_SECONDS` | int  | 0 |  | No |
| `PACHD_SERVICE_HOST` | string |  |  | No |
| `PACHD_SERVICE_PORT` | string |  |  | No |
| `ETCD_PREFIX` | string |  |  | No |
| `CLUSTER_DEPLOYMENT_ID` | string |  |  | No |
| `LOG_FORMAT` | string | json |  | No |
| `LOG_LEVEL` | string | info |  | No |
| `PACHYDERM_ENTERPRISE_ETCD_PREFIX` | string | pachyderm_enterprise |  | No |
| `METRICS` | boolean | TRUE |  | No |
| `METRICS_ENDPOINT` | string |  |  | No |
| `SESSION_DURATION_MINUTES` | int  | 43200 |How long auth tokens are valid for; defaults to 30 days.  | No |
| `IDENTITY_SERVER_DATABASE` | string | dex |  | No |
| `PPS_SPEC_COMMIT` | string |  |  | No |
| `PPS_PIPELINE_NAME` | string |  |  | No |
| `GOOGLE_CLOUD_PROFILER_PROJECT` | string |  |  | No |
| `PPS_MAX_CONCURRENT_K8S_REQUESTS` | int  | 10 |  | No |