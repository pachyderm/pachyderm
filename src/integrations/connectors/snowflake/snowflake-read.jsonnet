function(name, secretName, cronSpec, account, user, database, schema, query, format, hasHeader='false', outputFile='0000')
  {
    pipeline: {
      name: name,
    },
    input: {
      cron: {
        spec: cronSpec,
      },
    },
    transform: {
      cmd: ['sh'],
      stdin: [
        std.format('snowsql -a %s -u %s -d %s -s %s -q %s -o output_format=%s -o header=%s > /pfs/out/%s', [account, user, database, schema, query, format, hasHeader, outputFile]),
      ],
      secrets: [
        {
          name: secretName,
          env_var: 'SNOWSQL_PWD ',
          key: 'SNOWSQL_PWD ',
        },
      ],
      image: '<snowflake-image>',
    },
  }
