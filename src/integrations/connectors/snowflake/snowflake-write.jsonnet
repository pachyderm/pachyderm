function(name, inputRepo, image='pachyderm/snowflake', account, user='PACHYDERM_USER', role='PACHYDERM_ROLE', warehouse='PACHYDERM_WH', database='PACHYDERM_DB', schema='PACHYDERM_SCHEMA', table='', fileFormat, copyOptions='')
  local options = copyOptions + ' PURGE=true'; // removes files from staging after COPY is done
  local query = 'DELETE FROM %(table)s; COPY INTO %(table)s FROM @%%%(table)s FILE_FORMAT = %(format)s %(copyOptions)s;' % { table: table, format: fileFormat, copyOptions: options};
  {
    pipeline: {
      name: name,
    },
    input: {
      pfs: {
        repo: inputRepo,
        glob: '/*',
        name: 'in',
      },
    },
    transform: {
      cmd: ['bash'],
      stdin: [
        'for f in $(find /pfs/in -type f -follow -print); do',
        'snowsql -q "put file://${f} @%%%s OVERWRITE = TRUE"' % table,
        'done',
        'snowsql --single-transaction -q %s' % std.escapeStringBash(query),
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
