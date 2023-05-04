## Testing the collector locally with minikube

* Setup minikube like you would for normal setup
* adjust variables in `etc/testing/circle/workloads/ci-results/collector/test-collect.sh` as needed
* create the repos and pipeline images: `eval $(minikube -p minikube docker-env) && echo $(cd etc/testing/circle/workloads/ci-results && ./build-docker.sh)`
* ensure the results you want to test with are in the `/tmp/test-results` folder 
* run the collector with `etc/testing/circle/workloads/ci-results/collector/test-collect.sh

