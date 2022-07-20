function(name, url, query, format, cronSpec, secretName, outputFile='0000', hasHeader='false')
  {
    pipeline: {
      name: name,
    },
    transform: {
      cmd: ['/pach-bin/pachtf', 'sql-run', url, format, query, outputFile, hasHeader],
      secrets: [
        {
          name: secretName,
          env_var: 'PACHYDERM_SQL_PASSWORD',
          key: 'PACHYDERM_SQL_PASSWORD',
        },
      ],
    },
    input: {
      cron: {
        name: 'cron',
        spec: cronSpec,
        overwrite: true,
      },
    },
  }
