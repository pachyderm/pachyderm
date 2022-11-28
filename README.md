# Console

## Installation
The following command will install dependencies in `./`, `./frontend`, and `./backend`.

```
make install
```

<br />

## Local Deploy
You can run Console locally against either of the following:
1. A mock backend with useful fixtures.
1. A real Pachyderm cluster.

When Console runs it will be live at `localhost:4000`.

### Against the mock backend:

Note: Ensure you disable port-forwarding from a real pachyderm cluster.

The following command will start the UI server, API server, and mock gRPC server all-in-one.

```
make launch-mock
```

To switch between mock accounts use our devtools in the JS console of the browser.

```
> devtools.setAccount('2'); // will switch your current account
```

### Against a real Pachyderm cluster:
1. [Deploy Pachyderm locally](./README_Pachyderm.md) in either Enterprise or Community Edition. 
1. Ensure Pachyderm is port-forwarded with `pachctl port-forward`
1. Run the following command:

    ```
    make launch-dev
    ```

<br />

## Testing
### What tests do we have?

Console frontend tests consist of:
1. Jest unit tests against:
    1. The mock backend.
    1. A component library.

1. Cypress E2E tests against:
    1. A real Pachyderm cluster in EE (auth).
    1. A real Pachyderm cluster in CE (unauth).
    1. The mock backend (mock).

Console backend tests consist of:
1. Backend unit tests
1. Backend integration tests

### Running unit / integration tests
The following command runs unit and integration tests in `./frontend` and `./backend`.

```
make test
```

You may want to be more precise in either frontend or backend tests. To run jest against a single test, `cd` into the appropriate directory and provide a pattern for the test runner.

```
cd frontend
npm run frontend:test TestName.test.ts
npm run components:test Button
```

If you want to only run one of the tests outlined above refer to the `package.json` found in the subdirectory `frontend` or `backend`.

### E2E tests
We use Cypress for E2E tests.

There are three distinct sets of tests. They run against:
1. Community Edition.
1. Enterprise Edition.
1. The mock backend.

Before running one of the above sets of tests you must configure your local pachyderm and console setup to match what the test is expecting to test against.

**To run the mock test suite:**
1. Run `make launch-mock`
1. Run `npm run cypress:local-mock`

**Otherwise, to run either of the authenticated (Enterprise)or unauthenticated (Community Edition) test suites:**

1. [Run and port-forward a local Pachyderm cluster](./README_Pachyderm.md) in either Enterprise or Community Edition.
1. Run Console locally with `make launch-dev`.
1. Use one of the following commands to start Cypress:


**To run the authenticated (Enterprise) test suite:**

```
make e2e-auth
```

**To run the unauthenticated (Community Edition) test suite:**

1. Note: Ensure you add `PACHYDERM_ENTERPRISE_KEY` env variable with a valid [Enterprise key](https://enterprise-token-gen.pachyderm.io/dev) to your `.env.development.local`. Otherwise one test will fail.

2. Run the following command:
    ```
    make e2e
    ```

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

