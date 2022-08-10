
export ETCD_SERVICE_HOST=localhost 
export ETCD_SERVICE_PORT=2379
export STORAGE_BACKEND=LOCAL
export PG_BOUNCER_HOST=localhost
export PG_BOUNCER_PORT=5432
export POSTGRES_DATABASE=postgres
export POSTGRES_USER=alon
export POSTGRES_HOST=localhost
export POSTGRES_PORT=5432
export PACHD_POD_NAME=pachd
export KUBE=false

export CGO_ENABLED=0


export PACH_ROOT="/Users/alon/.pachyderm/data/pachd"

go run src/server/cmd/pachd/main.go 
