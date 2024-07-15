# Test Dags
This folder contains a collection of dags that can be useful for console development. The commands in the `Makefile` will set up interesting dags in a running pachyderm cluster.

## Commands

`setup-datums` - Sets up a dag that is helpful for testing with a large list of datums.

`generate-datums` - Generates datums and commit them to the repo created up in the setup-datums command.

`setup-opencv` - Sets up the openCV getting started tutorial found on our website.

`setup-project-tutorial` - Creates the tutorial DAG in a project named tutorial

`setup-cross-project` - Creates two cross-project DAGs across two projects each, for a total of four projects.

`generate-commits` - generates lots of commits

`setup-service-pipeline` - Sets up a basic service pipeline for testing.

`setup-spout-pipeline` - Pachyderm Spout example from [here](https://github.com/pachyderm/pachyderm/tree/master/examples/spouts/spout101).
