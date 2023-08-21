function(version, pghost, pguser)
{
    "pipeline": {
        "name": "go-test-results-sql-egress"
    },
    "input": {
        "pfs": {
            "repo": "go-test-results-raw",
            "project": "ci-metrics",
            "glob": "/*/*/*"
        }
    },
    "transform": {
        "image": "pachyderm/go-test-results:"+version,
        "cmd": [
            "/go/egress/egress",
            "/pfs/go-test-results-raw"
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
    },
    "parallelism_spec": {
        "constant": 2
    },
}

