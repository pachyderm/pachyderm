# Bazel + Console

### External Dependencies
If you are on an M-series Mac, you will need to install some external dependencies,
as the canvas module does not distribute ARM builds and therefore needs to be built
from source. See the [README.md](./README.md) for more details.

### PNPM
Prior to the introduction of the bazel tooling to this project, all developers 
used `npm` for project and dependency management. The
[JavaScript Rules developed by Aspect](https://docs.aspect.build/rules)
opt, instead, to use `pnpm` for dependency management. These tools are mostly
compatible but it is worth noting this change. `pnpm` generates and manages
`pnpm-lock.yaml` files as opposed to the `npm` managed `package-lock.json`
files. Additionally, the `node_modules/` directories will have a different
structure when built using `pnpm`. 

### Node Modules
The following command generates the `node_modules/` directory using bazel:
```
bazel run //:pnpm -- --dir $PWD install
```
when run within the `backend/` and `frontend/` subdirectories.

Once installed you can continue to run all commands through bazel using the
format:
```
bazel run //:pnpm -- --dir $PWD <command>
```
or you can use `npm`, if installed.

### Tests
The frontend and backend tests can be run, individually, with
```bash
bazel test //console/backend:backend_tests
bazel test //console/frontend:frontend_tests
```

#### Frontend Test Notes
Running the [PipelinesRuntimeChart tests](frontend/src/views/Project/components/PipelineList/components/PipelinesRuntimeChart/__tests__/PipelinesRuntimesChart.test.tsx)
requires the [canvas](https://github.com/Automattic/node-canvas) module,
which is problematic due to the fact that pre-built binaries are not published
for ARM platforms. Getting this module built and installed using bazel is still
not yet completed, so these tests are manually excluded from `bazel test`
invocations and the `canvas` dependency is manually excluded from the pnpm lock
file. To run these excluded tests locally, you must install the external
dependencies listed at the top of this document, and then run the following
commands from the `console/frontend` directory:
```bash
bazel run @nodejs//:npm -- --dir=$PWD install
bazel run @nodejs//:npm -- --dir=$PWD test -- src/views/Project/components/PipelineList/components/PipelinesRuntimeChart/__tests__/PipelinesRuntimesChart.test.tsx
```

#### Cypress (e2e) Tests
To run the cypress tests using the Bazel installed tooling, you must first have
a running Pachyderm instance, this can be done using the bazel pachdev command:
```bash
bazel run //src/testing/pachdev -- push
```
Then from the `console/` directory, you can start console with the following:
```bash
bazel run @nodejs//:npm -- --dir=$PWD run start:e2e
```
and start the cypress e2e tests with 
```bash
bazel run @nodejs//:npm -- --dir=$PWD run cypress:local
  --- or ---
bazel run @nodejs//:npm -- --dir=$PWD run cypress:local-auth
```
These cypress commands starts the cypress application and from there you
can run individual tests.
