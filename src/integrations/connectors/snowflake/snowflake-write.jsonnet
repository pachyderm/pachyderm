/*
PUT file:///pfs/in <table_stage>
COPY INTO <table> FROM <table_stage>

Two modes
1. User provides target table name. Copy all files from /pfs/in to table stage, then load into target table.

2. User does not provide target table name. We infer table names from top level directories in /pfs/in. Parallelization is at the table level.
*/
function(name, inputRepo, image='pachyderm/snowflake', account, user='PACHYDERM_USER', role='PACHYDERM_ROLE', warehouse='PACHYDERM_WH', database='PACHYDERM_DB', schema='PACHYDERM_SCHEMA', table='', fileFormat, copyOptions='')
  // local queryTemplate = if table == '' then 'DELETE FROM ${table}; COPY INTO ${table} FROM @%%${table} FILE_FORMAT = %(format)s %(copyOptions)s;' else
  //                                           'DELETE FROM %(table)s; COPY INTO %(table)s FROM @%%%(table)s FILE_FORMAT = %(format)s %(copyOptions)s;';
  // local query = queryTemplate % { table: table, format: fileFormat, copyOptions: copyOptions };
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
