# Split Load Test

The "split" load test constists of running several large jobs in which each
datum writes to a large number of output files, stressing Pachyderm's merge
operation. This is roughly analogous to building an inverted index of several
database records.

## Building the Load Test
`Dockerfile.buildenv` creates a container that builds both the `supervisor` and
`split` components of the load test. Run it with `make docker-build`, which
builds both components, puts them in docker containers
(`pachyderm/split-loadtest-supervisor` and `pachyderm/split-loadtest-pipeline`,
respectively), and pushes them to dockerhub.

## Running the Load Test
1. Create a pachyderm cluster and start a pod using the
   `pachyderm/split-loadtest-supervisor` docker image:
   ```
   make docker-build
   pachctl deploy <args>
   kubectl apply -f kube/supervisor.yaml
   ```

## To Do
- Deployment could be more ergonomic, especially if we want to run multiple
  load tests, or the same load test with multiple parameter configurations in
  multiple cloud providers
  - Idea: turn kube/supervisor.yaml into a ksonnet (or helm? jsonnet? JSON-e?) template
