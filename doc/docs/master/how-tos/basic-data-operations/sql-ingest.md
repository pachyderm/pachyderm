
# Ingest Structured Data - Integration with SQL

Part of your data might live in databases requiring some level of integration to your warehouse to retrieve and inject them into Pachyderm.

Our **SQL ingest** tool provides a seamless connection between databases and Pachyderm,  allowing you to import data from a SQL database into Pachyderm-powered pipelines. By bringing data-driven pipelines, versioning & lineage to structured data, we are allowing Data Science teams to easily combine structured and unstructured data, thus improving the performance of prediction models.

Specifically, we help you connect to a remote Database of your choice and pull the result of a given query at regular intervals in the form of a CSV or JSON file.  

## Use SQL Ingest
Pachyderm's SQL Ingest uses [pipeline templates](../pipeline-operations/pipeline-templates) populated with the following parameters to automatically create the pipelines that access, query, and materialize the results in the form of CSV or JSON files.

Pass in the following parameters and get your results commited to an output commit, ready for the next downstream pipeline:
```shell
pachctl update pipeline --jsonnet ./src/templates/sql_ingest_cron.jsonnet \
  --arg name=myingest \
  --arg url="mysql://root@mysql:3306/test_db" \
  --arg query="SELECT * FROM test_data" \
  --arg cronSpec="@every 30s" \
  --arg secretName="mysql-creds" \
  --arg format=json
```

Where the parameters passed to the template are:

| Parameter     | Description | 
| ------------- |-------------| 
| `name`        | The name of output repo in which your query results will materialize.|
| `url`         | The [connection string to the database](#database-connection).|  
| `query`       | The SQL query that will be run against your database. |
| `cronSpec`    | How often to run the query. For example `"@every 60s"`.|
|`format`       | The type of your output file containing the results of your query (either `json` or `yaml`).|
| `secretName`  | The kubernetes secret name that contains the [password to the database](#database-secret).|

When the pipeline template command is run the database will be queried on a schedule defined in your `cronSpec` parameter.

### Database Secret
Before you create your SQL Ingest pipelines, make sure to create a [generic secret](../advanced-data-operations/secrets/#create-a-secret) containing your database password in the field `PACHYDERM_SQL_PASSWORD`.

!!! Example
    ```yaml
    apiVersion: v1
    kind: Secret
    metadata:
      name: mysql-creds
    data:
      "PACHYDERM_SQL_PASSWORD": "cm9vdA==" # base64 encoded
    ```

### Database Connection URL
Pachyderm's SQL Ingest will take one URL as its connection string to the database of your choice.

The URL is structured as follows:
```shell
<protocol>://<username>@<host>:<port>/<database>?<param1>=<value1>&<param2>=<value2>
```

Where:

| Parameter     | Description | 
| ------------- |-------------| 
| **protocol**   | The name of the database protocol. As of today, you have the ability to connect to  `postgres`or `mysql`. Note that `mysql` allows you to connect to mariadb. |
| **username**  | The user used to access the database.|
| **host**      | The hostname of your database instance.|
| **port**      | The port number your instance is listening on.|
| **database**  | The name of the database to connect to. | 
| The following parameters are optional and specific to the driver.|

!!! Note 
    The password is not included in the URL.  It is retrieved from a [kubernetes secret](#database-secret) or file on disk at the time of the query.


## How does this work?

SQL Ingest's pipeline template <link to gihub template> creates two pipelines:

- A **[Cron Pipeline](../../../concepts/pipeline-concepts/pipeline/cron/#cron-pipeline)** trigering at an interval set by `cronSpec` and outputting a file `/0000` containing a timestamp and the SQL statement.
- A following pipeline taking the `/0000` as input and running the query againts the database.The result of the query is then materialized in a file (JSON or CSV) commited to its output repo.

The same base image [pachctf]() is used in both pipelines.


List pipelines:
![List pipelines](../images/sqlingest-list-pipeline.png)

List repos:
![List repo](../images/sqlingest-list-repo.png)

```shell
pachctl list file myingest_queries_in@master
```
```
NAME                  TYPE SIZE
/2022-01-26T22:16:55Z file 0B
```
```shell
pachctl list file myingest_queries@master
```
```
NAME  TYPE SIZE
/0000 file 37B
```
```shell
pachctl list file myingest@master
```
```
NAME  TYPE SIZE
/0000 file 52B
```

```shell
pachctl get file myingest@master:/0000
```
```yaml
{"mycolumn":"hello world","id":1}
{"mycolumn":"hello you","id":2}
```

```shell
pachctl get file myingest_queries@master:/0000
```
```
-- 1643235475
SELECT * FROM test_data
```


