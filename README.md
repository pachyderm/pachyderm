# Console

## Installation
```
make install
```
This installs dependencies in `./`, `./frontend`, and `./backend`.

<br />

## Local Deploy

When running Console locally, you have the option of using Console on top of a real pachyderm cluster, or a mock server with existing fixtures. When running, Console will be live at `localhost:4000`

### Against the mock backend:

```
make launch-mock
```

This will start the UI server, API server, and mock gRPC server all in one. To switch between mock accounts, you can use our devtools in the JS console of the browser.

```
> devtools.setAccount('2'); // will switch your current account
```

### Against a real pachyderm cluster:
[Deploy Pachyderm locally,](./README_Pachyderm.md) in either Enterprise or Community Edition. Make sure Pachyderm is port-forwarding with `pachctl port-forward`, then run:

```
make launch-dev
```

<br />

## Testing
Console uses both unit tests in Jest against the mock backend, and E2E/integration tests with Cypress against a real Pachyderm cluster.

<br />

## Unit tests
The following command will run unit tests in `./frontend` and `./backend`.
```
make test
```
You may want to be more precise in either frontend or backend tests. To run jest against a single test, `cd` into the appropriate directory and provide a pattern for the test runner.

```
cd frontend
npm run test TestName.test.ts
```

<br />

## E2E tests
We use E2E tests for both Community Edition Console and Enterprise Console.

1. Make sure you have a port-forwarding [local pachyderm cluster](./README_Pachyderm.md) in either Enterprise or Community Edition.
1. Start running console locally with `make launch-dev`
1. Finally in another terminal window, use one of the following commands to start Cypress

To run the unauthenticated (Community Edition) test suites:
```
make e2e
```
And to run the authenticated (Enterprise) test suites:
```
make e2e-auth
```

Note: you will need a `PACHYDERM_ENTERPRISE_KEY` env variable with a valid [Enterprise key](https://enterprise-token-gen.pachyderm.io/dev) to be able to pass the authenticated test suite, as console tests the community edition upgrade flow.

<br />

## Graphql & TypeScript type generation

`graphql-codegen` is leveraged in this project to generate both type definitions and client-side code for Console's GraphQL resources based on this codebase's graphql documents. [Find out more about type generation here](./README_Development.md).

To update the generated types, run:

```
make graphql
```

<br />

## Running the production server

```
make launch-prod
```

This will start the production server at `localhost:3000`. Additionally, if you'd like to test the production UI/API against the mock gRPC & Auth server, you can run `npm run start:mock` from /backend and add a `.env.production.local` file that replicates the variables found in `.env.test`.

<br />

## Working with environment variables
All variables are stored in `.env.${environmentName}`. You can find a list of the default variables configured in the Console container in the `.env.production` file. 

### Client build-time variables
Any variables added with the `REACT_APP` prefix (no `_RUNTIME`) will be appended to
the `process.env` object in the client Javascript bundle.

```
// client-side code
const port = process.env.REACT_APP_PORT;
```

### Client runtime variables
Any variables with the `REACT_APP_RUNTIME` prefix will be added to a window object
called `pachDashConfig` with can be accessed in the client code:

```
// client-side code
const port = process.env.pachDashConfig.REACT_APP_RUNTIME_PORT;
```

These variables can also be accessed in the node runtime (on the server, and during test execution), but not via the `pachDashConfig` object. The reason for this, is that
the node runtime does not support maps as environment variables.

```
// node runtime
const port = process.env.REACT_APP_RUNTIME_PORT;
```

### Server-only variables
Any variables not prefixed with `REACT_APP` or `REACT_APP_RUNTIME` will only be accessible from the server __only__.

```
// server code
const port = process.env.PORT;
```

