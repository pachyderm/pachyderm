/* Snowflake Write Connector

Loads files from PFS to Snowflake table. Importantly, it is the responsibility of the user to ensure matching schema between files in Pachyderm and target tables.
To ensure proper syncing of deletes, we first wipe the target table before inserting new data.

At a high level, this is a three step process:
1. PUT file:///pfs/in/... <table_stage>
2. DELETE FROM <table>
3. COPY INTO <table> FROM <table_stage>

We support two modes:
Mode 1: The target table name is provided via Jsonnet argument `table`. As result, we copy all files from /pfs/in/* to this target table.

Mode 2. We infer table names from top level directories in /pfs/in. Datums are sharded at the top level directory level.

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
  image : Docker image containing snowsql
  --------------
  account : Snowflake account identifier (<orgname>-<account_name>)
  user : Snowflake user
  role : Snowflake role
  warehouse : Snowflake warehouse
  database : Snowflake database
  schema : Snowflake schema
  fileFormat : documented at https://docs.snowflake.com/en/sql-reference/sql/copy-into-table.html#format-type-options-formattypeoptions
  copyOptions : documented at https://docs.snowflake.com/en/sql-reference/sql/copy-into-table.html#copy-options-copyoptions 
  --------------
  table : target table name (leave blank if you want Pachyderm to infer table names from input repo directories)
*/
function(name, inputRepo, image='pachyderm/snowflake', account, user, role, warehouse, database, schema, fileFormat, copyOptions='PURGE = TRUE', table='')
  local stdin = if table == '' then [
    'table=$(basename /pfs/in/*)',
    'snowsql -q "put file:///pfs/in/${table}/* @%${table} OVERWRITE = TRUE"',
    'snowsql --single-transaction -q "DELETE FROM ${table}; COPY INTO ${table} FROM @%%${table} FILE_FORMAT = %(format)s %(copyOptions)s"' % { format: fileFormat, copyOptions: copyOptions },
  ] else [
    'for f in $(find /pfs/in -type f -follow -print); do',
    'snowsql -q "put file://${f} @%%%s OVERWRITE = TRUE"' % table,
    'done',
    'snowsql --single-transaction -q "DELETE FROM %(table)s; COPY INTO %(table)s FROM @%%%(table)s FILE_FORMAT = %(format)s %(copyOptions)s;"' % { table: table, format: fileFormat, copyOptions: copyOptions },
  ];
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
      cmd: ['bash'],
      stdin: stdin,
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
