## Testing the collector locally with minikube

* Setup minikube like you would for normal setup
* adjust variables in `etc/testing/circle/workloads/ci-results/collector/test-collect.sh` as needed
* create the repos and pipeline images: `eval $(minikube -p minikube docker-env) && echo $(cd etc/testing/circle/workloads/ci-results && ./build-docker.sh)`
* ensure the results you want to test with are in the `/tmp/test-results` folder 
* Apply the schema `etc/testing/circle/workloads/go-test-results/schema.sql` to the DB if the tables do not exist.
* run the collector with `etc/testing/circle/workloads/ci-results/collector/test-collect.sh

If you want to test with local grafana check this page out: https://grafana.com/docs/grafana/latest/setup-grafana/installation/kubernetes/


## Making changes
To make a change the production docker image needs to be updated then pulled by pachyderm@pachops.com. To do this:
* login as a docker user with push permissions to the pachyderm repo
* Run ./build-docker.sh in this folder with the updated `<version number>`
* Run `docker push pachyderm/go-test-results:<version number>`

To update the pipeline you can run the command 
`pachctl update pipeline --jsonnet src/testing/cmds/go-test-results/egress/pipeline.jsonnet --arg version=<version> --arg pghost=cloudsql-auth-proxy.pachyderm.svc.cluster.local. --arg pguser=postgres --project ci-metrics`