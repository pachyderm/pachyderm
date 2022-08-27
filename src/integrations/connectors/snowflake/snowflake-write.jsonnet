/* Snowflake Write Connector

Loads files from PFS to Snowflake table. Importantly, it is the responsibility of the user to ensure matching schema between files in Pachyderm and target tables.
To ensure proper syncing of deletes, we first wipe the target table before inserting new data.

At a high level, this is a 2 step process:
1. PUT file:///pfs/in/... <temp stage>
2. Within a transaction
  2.1. DELETE FROM <table>
  2.2. COPY INTO <table> FROM <temp stage>
Note that we use a transaction to protect against losing data in the event of an error.

Mode 1: The target table name is provided via Jsonnet argument `table`. As result, we copy all files from /pfs/in/* to this target table.

Mode 2: We infer table names from top level directories in /pfs/in. Datums are sharded at the top level directory level.

Example:

Given the following files:
  /pfs/in/table1/file1
  /pfs/in/table1/file2
  /pfs/in/table2/file1

There would be two datums: ['/pfs/in/table1', '/pfs/in/table2']
For each datum, we infer target table name from directory name, and copy all of its files to the target table.

Arguments:
  name : pipeline name
  inputRepo: input repo name
  image : Docker image, usually used for testing

  account : Snowflake account identifier (<orgname>-<account_name>)
  user : Snowflake user
  role : Snowflake role
  warehouse : Snowflake warehouse
  database : Snowflake database
  schema : Snowflake schema

  fileFormat : documented at https://docs.snowflake.com/en/sql-reference/sql/copy-into-table.html#format-type-options-formattypeoptions
  copyOptions : documented at https://docs.snowflake.com/en/sql-reference/sql/copy-into-table.html#copy-options-copyoptions
  table : target table name (leave blank if you want Pachyderm to infer table names from input repo directories)
*/
function(name, inputRepo, image='pachyderm/snowflake:latest', account, user, role, warehouse, database, schema, fileFormat, copyOptions='PURGE = TRUE', table='', debug=false)
  {
    pipeline: {
      name: name,
    },
    input: {
      pfs: {
        repo: inputRepo,
        glob: if table == '' then '/*' else '/',
        name: 'in',
      },
    },
    transform: {
      cmd: ['sh'],
      stdin: [
        'snowpach write -fileFormat=%(fileFormat)s -copyOptions=%(copyOptions)s -table=%(table)s -inputDir=/pfs/in -debug=%(debug)s' % { fileFormat: std.escapeStringBash(fileFormat), copyOptions: std.escapeStringBash(copyOptions), table: table, debug: debug },
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
