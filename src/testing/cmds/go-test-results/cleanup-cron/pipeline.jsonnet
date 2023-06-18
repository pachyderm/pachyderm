function(version, pghost, pguser)
{
    "pipeline": {
        "name": "go-test-results-sql-cleanup"
    },
    "input": {
        "cron": {
            "name": "cleanup-tick",
            "spec": "@daily",
            "overwrite": true
        }
    },
    "transform": {
        "image": "pachyderm/go-test-results:"+version,
        "cmd": [
            "/go/cleanup-cron/cleanup-cron"
        ],
        "env": {
            "LOG_LEVEL": "DEBUG",
            "POSTGRESQL_HOST": pghost, // remote GCP: "cloudsql-auth-proxy.pachyderm.svc.cluster.local." local testing: "postgres"
            "POSTGRESQL_USER": pguser  // remote GCP:"postgres" local testing: "pachyderm"
        },
        "secrets": [{
            "name": "postgres",
            "env_var": "POSTGRESQL_PASSWORD",
            "key": "postgresql-password"
        }]
    }
}

