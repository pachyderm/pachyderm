# Ingest Data Via SQL Ingest

!!! Warning
    SQL Ingest is an [experimental feature](../../../reference/supported-releases/#experimental){target=_blank}.

You can inject database content, collected by your data warehouse, by pulling the result of a given query into Pachyderm and saving it as a CSV or JSON file.

## Before You Start 

- You should be familiar with [Jsonnet](https://jsonnet.org/learning/tutorial.html)
- You should be familiar with creating [Jsonnet pipeline specs](https://docs.pachyderm.com/latest/how-tos/pipeline-operations/jsonnet-pipeline-specs/#jsonnet-pipeline-specifications) in Pachydernm.

---

## How to Set Up SQL Ingest

### 1. Create & Upload a Generic Secret

1. Copy the following:
    ```shell
    kubectl create secret generic yourSecretName --from-literal=PACHYDERM_SQL_PASSWORD=yourDatabaseUserPassword --dry-run=client --output=json > yourSecretFile.json
    ```
2. Swap out `yourSecretName`, `yourDatabaseUserPassword`, and `yourSecretFile` with relevant inputs.
3. Open a terminal and run the command.
4. Copy the following: 
    ```shell
    pachctl create secret -f yourSecretFile.json
    ```
5. Swap out `yourSecretfile` with relevant filename. 
6. Run the command. 
   
This creates a secret that looks like the following:

!!! Note
     Not all secret formats are the same. For a full walkthrough on how to create, edit, and view different types of secrets, see [Create and Manage Secrets in Pachyderm](../../advanced-data-operations/secrets/#create-a-secret).

### 2. Create a Database Connection URL 

Pachyderm's SQL Ingest requires a connection string defined as a URL to connect to your database; the URL is structured as follows:

```
<protocol>://<username>@<host>:<port>/<database>?<param1>=<value1>&<param2>=<value2>
```

### 3. Create a Pipeline Spec 

Pachyderm provides a [default jsonnet template](https://raw.githubusercontent.com/pachyderm/pachyderm/{{ config.pach_branch }}/src/templates/sql_ingest_cron.jsonnet) that has key parameters built in.

1. Copy the following:
   ```shell
   pachctl update pipeline --jsonnet https://raw.githubusercontent.com/pachyderm/pachyderm/{{ config.pach_branch }}/src/templates/sql_ingest_cron.jsonnet \
     --arg name=<outputRepoName> \
     --arg url="<connectionStringToDdatabase>" \
     --arg query="<query>" \
     --arg hasHeader=<boolean> \
     --arg cronSpec="<pullInterval>" \
     --arg secretName="<youSecretName>" \
     --arg format=<CsvOrJson> 
   ```
2. Swap out all of the parameter values with relevant inputs. 
3. Open terminal.
4. Run the command.

### 4. View Query & Results 

- **To View Query String**: `pachctl get file outputRepoName@master:/0000`
- **To View Output File**: `pachctl list file outputRepoName@master`
- **To Read Output File**: `pachctl get file myingest@master:/0000` 

---
## How Does This Work?

SQL Ingest's jsonnet pipeline spec [**`sql_ingest_cron.jsonnet`**](https://github.com/pachyderm/pachyderm/blob/{{ config.pach_branch }}/src/templates/sql_ingest_cron.jsonnet) creates all of the following:

- **1 Input Data Repo**: Used to store timestamp files at the cronSpec's set interval rate (`--arg cronSpec="@every 30s" \`) to trigger the pipeline.
- [**1 Cron Pipeline**](../../../concepts/pipeline-concepts/pipeline/cron/#cron-pipeline): Houses the spec details that define the input type and settings and  data transformation.
-  **1 Output Repo**: Used to store the data transformed by the cron pipeline; set by the pipeline spec's `pipeline.name` attribute, which you can define through the jsonnet parameter `--arg name=myingest \`
- **1 Output File**: Used to save the query results (JSON or CSV) and potentially be used as input for a following pipeline.

In the default Jsonnet template, the file generated is obtainable from the output repo, `outputRepoName@master:/0000`. The filename is hardcoded, however you could paramaterize this as well using a custom jsonnet pipeline spec and passing `--arg outputFile='0000'`. The file's contents are the result of the query(`--arg query="..."`) being ran against the database`--arg url="..."` ; both are defined in the `transform.cmd` attribute.

### Other Details

- The name of each pipeline (and their related input/output repos) are derived from the `name` parameter (`--arg name=myingest`).
- The same base image [pachctf](https://hub.docker.com/repository/docker/pachyderm/pachtf){target=_blank} is used in both pipelines.
---

## About SQL Ingest Jsonnet Pipeline Specs

To create an SQL Ingest Jsonnet Pipeline spec, you must have a .jsonnet file and several parameters:

```shell
pachctl update pipeline --jsonnet <https://your-SQL-ingest-pipeline-spec.jsonnet> \
  --arg name=<outputRepoName> \
  --arg url="<connectionStringToDdatabase>" \
  --arg query="<query>" \
  --arg hasHeader=<boolean> \
  --arg cronSpec="<pullInterval>" \
  --arg secretName="<secretName>" \
  --arg format=<CsvOrJson> 
```

### Parameters

| Parameter  | Description | 
| ------------- |-------------| 
| `name`        | The name of output repo where query results will materialize.|
| `url`         | The connection string to the database.|  
| `query`       | The SQL query to be run against the connected database. |
| `hasHeader`   | Adds a header to your CSV file if set to `true`. Ignored if `format="json"` (JSON files always display (header,value) pairs for each returned row). Defaults to `false`. <br><br>Pachyderm creates the header after each element of the comma separated list of your SELECT clause or their aliases (if any). <br>For example `country.country_name_eng` will have `country.country_name_eng` as header while `country.country_name_eng as country_name` will have `country_name`. |
| `cronSpec`    | How often to run the query. For example `"@every 60s"`.|
| `format`      | The type of your output file containing the results of your query (either `json` or `csv`).|
| `secretName`  | The kubernetes secret name that contains the [password to the database](#database-secret).|

#### URL Parameter Details


- Passwords are not included in the URL; they are retrieved from the secret created in step 1.
- The additional parameters after `?` are optional and needed on a case-by-case bases (for example, Snowflake)

| Parameter     | Description | 
| ------------- |-------------| 
|`protocol`  | The name of the database protocol. <br> As of today, we support: <br>- `postgres` and `postgresql` : connect to Postgresql or compatible (for example Redshift).<br>- `mysql` : connect to MySQL or compatible (for example MariaDB). <br>- `snowflake` : connect to Snowflake. |
| `username` | The user used to access the database.|
| `host`     | The hostname of your database instance.|
| `port`     | The port number your instance is listening on.|
| `database` | The name of the database to connect to. | 

---

## Formats & SQL Data Types 

The following comments on formatting reflect the state of this release and are subject to change.

### Formats 

#### Numeric 

All numeric values are converted into strings in your CSV and JSON. 

|Database|CSV|JSON|
|--------|---|----|
| 12345 | 12345 | "12345" |
| 123.45 | 123.45 | "123.45" |

!!! Warning
        - Note that infinite (Inf) and not a number (NaN) values will also be stored as strings in JSON files. 
        - Use this format `#.#` for all decimals that you plan to egress back to a database.

#### Date/Timestamps

|Type|Database|CSV|JSON|
|----|--------|---|----|
|Date|2022-05-09|2022-05-09T00:00:00|"2022-05-09T00:00:00"|
|Timestamp ntz|2022-05-09 16:43:00|2022-05-09T16:43:00|"2022-05-09T16:43:00"|
|Timestamp tz|2022-05-09 16:43:00-05:00|2022-05-09T16:43:00-05:00|"2022-05-09T16:43:00-05:00"|

#### Strings

|Database|CSV|
|--------|---|
|"null"|null|
|\`""\`|""""""|
|""|""|
|nil||
|`"my string"`|"""my string"""|
|"this will be enclosed in quotes because it has a ,"|"this will be enclosed in quotes because it has a ,"|

!!! Tip 
     When parsing your CSVs in your user code, remember to escape `"` with `""`.


### Supported Data Types

Some of the Data Types listed in this section are specific to a particular database.

| Dates/Timestamps | Varchars | Numerics | Booleans |
|------------------|---------|----------|----------|
|`DATE` <br> `TIME`<br> `TIMESTAMP`<br> `TIMESTAMP_LTZ`<br> `TIMESTAMP_NTZ`<br> `TIMESTAMP_TZ`<br> `TIMESTAMPTZ`<br> `TIMESTAMP WITH TIME ZONE`<br> `TIMESTAMP WITHOUT TIME ZONE` |`VARCHAR`<br> `TEXT`<br> `CHARACTER VARYING`|`SMALLINT`<br> `INT2`<br> `INTEGER`<br> `INT`<br> `INT4`<br> `BIGINT`<br> `INT8`<br>`FLOAT`<br> `FLOAT4`<br> `FLOAT8`<br> `REAL`<br> `DOUBLE PRECISION`<br>`NUMERIC`<br> `DECIMAL`<br> `NUMBER`|`BOOL`<br>`BOOLEAN`|