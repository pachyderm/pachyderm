function(name, cronSpec, account, user, warehouse, role, database, schema, query, format, hasHeader='false', outputFile='0000', image='pachyderm/snowflake')
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
        'mkdir -p $(dirname /pfs/out/%s)' % outputFile,
        'snowsql -o friendly=false -o header=%s -o timing=false -o output_format=%s -q %s > /pfs/out/%s' % [hasHeader, format, std.escapeStringBash(query), outputFile],
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
