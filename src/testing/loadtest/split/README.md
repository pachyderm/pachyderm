# Split Load Test

The "split" load test constists of running several large jobs in which each
datum writes to a large number of output files, stressing Pachyderm's merge
operation. This is roughly analogous to building an inverted index of several
database records.

## Building the Load Test
Build the benchmark containers with `make docker-build`. This runs
`build/docker-build.sh` which mounts the Pachyderm repo in a `golang:1.11`
docker image, builds both the pipeline and supervisor components, puts them in
docker containers (`pachyderm/split-loadtest-supervisor` and
`pachyderm/split-loadtest-pipeline`, respectively), and pushes them to
dockerhub.

## Running the Load Test
1. Create a pachyderm cluster
1. Start a pod using the `pachyderm/split-loadtest-supervisor` docker image:
   ```
   make docker-build
   kubectl apply -f kube/supervisor.yaml
   ```

## To Do
- Deployment could be more ergonomic, especially if we want to run multiple
  load tests, or the same load test with multiple parameter configurations in
  multiple cloud providers
  - Idea: turn kube/supervisor.yaml into a ksonnet (or helm? jsonnet? JSON-e?) template
- The benchmark currently doesn't validate the output (e.g.:
  - check that output files are the same sum total size as the input files
  - check that all output lines are unique
  - check that `num_input_files/keys_per_file` distinct input file values
    appear in every output file
  Obviously this would be helpful when testing major rewrites of PPS/worker
- A flag for more/less verbose logging (right now logging is fairly verbose)

