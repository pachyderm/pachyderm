# Test Suite for Pachyderm Integration with Label Studio

This folder is for internal use only. It contains Python tests to validate
that our Label Studio integration in [our fork](https://github.com/pachyderm/label-studio)
works with the latest version of Pachyderm in `master`. The tests run in CI.

## Run tests locally

### Prerequisites

1. Ensure you have Python installed before running test these tests locally.
2. Add the pachd address for the Pachyderm cluster that Label Studio will
   connect to in <mark>Makefile::run-tests</mark> section. This can be a fresh
   cluster.

### Running the tests

The Python tests assume Label Studio is running at localhost:8080. To change
this, simply update the URL [here](./tests/constants.py).

Two easy options to run the tests:

1. Docker
   - `make docker-test`
   - If you have a specific Docker image you'd like to test, edit the
     relevant commands to run that image instead.
2. Run program directly
   - Make sure you have [Label Studio installed](https://labelstud.io/guide/install.html)
   - `make direct-test`

To re-run the tests after Label Studio is already running, simply `make run-tests`.
