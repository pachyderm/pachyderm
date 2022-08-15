function(name, secretName, account, user, stage="pachyderm", database, schema='pachyderm', file_format, inputRepo)
  {
    pipeline: {
      name: name,
    },
    load_external: {},
    input: {
        pfs: {
            repo: inputRepo,
            glob: "/*",
            name: "in"
        }
    },
    transform: {
      cmd: ['sh'],
      stdin: [
        std.format("snowsql -a %s -u %s -q 'put file:///pfs/in/* @~/%s'", [account, user, stage]),
        "table=$(basename -- /pfs/in/*)",
        std.format("snowsql -a %s -u %s -d %s -s %s -q \"DELETE FROM $table\"", [account, user, database, schema]),
        std.format("snowsql -a %s -u %s -d %s -s %s -q \"COPY INTO $table FROM @~/%s/$table file_format = %s\"", [account, user, database, schema, file_format]),
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
