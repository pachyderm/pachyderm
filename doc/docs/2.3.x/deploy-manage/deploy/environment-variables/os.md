# Configure OS Variables

## Variables

| Variable | Required | Type | Default | Description |
|---|---|---|---|---|
| CIRCLE_REPOSITORY_URL |  | string |  |  |
| DATABASE_HOST |  | string |  | Defines the database host (e.g., `localhost`) |
| DATABASE_PASSWORD |  | string |  | Contains the database user's password. |
| DATABASE_PORT |  | int |  | Defines the database port (e.g., `3306`) |
| DATABASE_USERNAME |  | string |  | Contains the database user's username. |
| ECS_CLIENT_BUCKET |  | string  |  |  |
| ECS_CLIENT_CUSTOM_ENDPOINT |   | string  |  |  |
| ECS_CLIENT_ID |  |string  |  |  |
| ECS_CLIENT_SECRET |  | string  |  |  |
| EDITOR |  |  |  |  |
| GOBIN |  | string |  |  |
| GOOGLE_CLIENT_BUCKET |  | string |  |  |
| GOOGLE_CLIENT_CREDS |  | string |  |  |
| GOOGLE_CLIENT_HMAC_BUCKET |  | string  |  |  |
| GOOGLE_CLIENT_HMAC_ID |  | string  |  |  |
| GOOGLE_CLIENT_HMAC_SECRET |  | string |  |  |
| GOOGLE_CLIENT_REGION |  | string |  |  |
| GOPATH |  | string |  |  |
| HOME |  | string  |  | Defines the path to the home directory. The default value is `/root`  |
| HOST |  | string  |  |  |
| JAEGER_COLLECTOR_SERVICE_PORT_JAEGER_COLLECTOR_HTTP |  | string  |  |  |
| KUBERNETES_SERVICE_HOST |  | string  |  |  |
| LOCAL_TEST |  | string |  |  |
| LOG_LEVEL |  |  | `info` | Defines the verbosity of the log output; options include `debug`, `info`, `error`, and `0` to disable. |
| MICROSOFT_CLIENT_CONTAINER |  | string |  |  |
| MICROSOFT_CLIENT_ID |  | string  |  |  |
| MICROSOFT_CLIENT_SECRET |  | string  |  |  |
| PACH_DATUM_ID |  | string  |  | Contains the ID of the current Datum. |
| PACHD_PEER_SERVICE_HOST |  | string  |  |  |
| PACHD_PEER_SERVICE_PORT |  | int  |  |  |
| PAGER |  | string |  |  |
| PIPELINE_HOME |  |string  |  | Defines the PFS home folder.  |
| PIPELINE_INPUT |  | string |  | Defines the PFS input folder. |
| PIPELINE_OUTPUT |  |string |  | Defines the PFS output folder. |
| PORT |  | int  |  | Defines the `pachd` port number. |
| SNOWFLAKE_ACCOUNT |  | string  |  | Contains the Snowflake account name. |
| SNOWFLAKE_PASSWORD |  | string  |  | Contains the Snowflake user's password. |
| SNOWFLAKE_USER |  | string  |  | Contains the Snowflake user's username. |
| SNOWFLAKE_USER_ROLE |  | string |  | Contains the Snowflake user's role.|
| TEST_IMAGE_SHA |  | string  |  |  |
| TOPIC |  | string  |  |  |
| UNPAUSED_MODE |  | string  |  | Options include `full`, and empty. |
| VM_IP |  | string |  | Defines the virtual machine IP address (e.g., `localhost`) |