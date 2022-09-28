/* Snowflake Read Connector

Runs a user's query in a pipeline at a regular cadence configured via cron. 
Importantly, we expose many useful Snowflake features via fileFormat and copyOptions.

At a high level, this is a 2 step process:
  1: COPY INTO <temp stage> FROM <query> : which unloads data from a user's query into the user's stage under the directory named after this pipeline.
  2: GET <temp stage> file:///pfs/out : downloads the files from the Snoflake user stage to local /pfs/out

Arguments:
  name : pipeline name
  cronSpec : cadence at which to run this pipeline
  image : Docker image, usually changed for testing

  account : Snowflake account identifier (<orgname>-<account_name>)
  user : Snowflake user
  role : Snowflake role
  warehouse : Snowflake warehouse
  database : Snowflake database
  schema : Snowflake schema

  query : the query to run on Snowflake
  fileFormat : documented at https://docs.snowflake.com/en/sql-reference/sql/copy-into-location.html#format-type-options-formattypeoptions
  copyOptions : documented at https://docs.snowflake.com/en/sql-reference/sql/copy-into-location.html#copy-options-copyoptions
  partitionBy: specify an expression used to partition the table rows into separate files; documented at https://docs.snowflake.com/en/sql-reference/sql/copy-into-location.html#copy-options-copyoptions:~:text=Optional%20Parameters-,PARTITION%20BY,-expr
*/
function(name, cronSpec, image='pachyderm/snowflake:latest', account, user, role, warehouse, database, schema, query, fileFormat, copyOptions='', partitionBy='', header=false, debug=false)
  {
    pipeline: {
      name: name,
    },
    input: {
      cron: {
        name: 'cron',
        spec: cronSpec,
        overwrite: true,
      },
    },
    transform: {
      cmd: ['sh'],
      stdin: [
        'snowpach read -query=%(query)s -fileFormat=%(fileFormat)s -partitionBy=%(partitionBy)s -outputDir=/tmp/out -header=%(header)s -debug=%(debug)s' % { query: std.escapeStringBash(query), fileFormat: std.escapeStringBash(fileFormat), partitionBy: std.escapeStringBash(partitionBy), header: header, debug: debug},
        // we are writing to /tmp/out first because for unknown reasons, snowpach writes corrupted data to /pfs but not /tmp
        'mv /tmp/out/* /pfs/out'
      ],
      env: {
        SNOWSQL_ACCOUNT: account,
        SNOWSQL_USER: user,
        SNOWSQL_DATABASE: database,
        SNOWSQL_SCHEMA: schema,
        SNOWSQL_ROLE: role,
        SNOWSQL_WH: warehouse,
      },
      secrets: [
        {
          name: 'snowflake-secret',
          env_var: 'SNOWSQL_PWD',
          key: 'SNOWSQL_PWD',
        },
      ],
      image: image,
    },
  }
