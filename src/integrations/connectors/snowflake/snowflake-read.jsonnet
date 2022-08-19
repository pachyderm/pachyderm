/*
COPY INTO <stage> FROM <query>
GET
*/
function(name, cronSpec, image='pachyderm/snowflake', account, user, warehouse, role, database, schema, query, fileFormat, copyOptions='', hasHeader='false', outputFile='0000')
  local copyInto = 'COPY INTO @~/%(name)s/%(outputFile)s FROM (%(query)s) FILE_FORMAT = %(fileFormat)s %(copyOptions)s' % { name: name, outputFile: outputFile, query: query, fileFormat: fileFormat, copyOptions: copyOptions };
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
      cmd: ['bash'],
      stdin: [
        'set -eo pipefail',
        'outputdir=$(dirname /pfs/out/%s)' % outputFile,
        'mkdir -p $outputdir',
        'snowsql -q %s' % std.escapeStringBash(copyInto),
        'snowsql -q "GET @~/%(name)s/%(outputFile)s file://${outputdir}"' % { name: name, outputFile: outputFile },
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
