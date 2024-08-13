# Console

## Installation

The following command will install dependencies in `./`, `./frontend`, and `./backend`.

```bash
make install
```

If you are on an M1 Mac, first run these commands or installation will fail:

<!-- https://github.com/Automattic/node-canvas/issues/1733 -->

```bash
brew install pkg-config cairo pango libpng jpeg giflib librsvg pixman
```
or
```bash
port install pkgconfig cairo pango libpng jpeg giflib librsvg
```

### Running console locally

1. [Deploy Pachyderm locally](./README_Pachyderm.md) in either Enterprise or Community Edition.
1. Run the following command:

   ```bash
   make launch-dev
   ```

## Logs

If you want pretty server logs [use the bunyan cli tool](https://github.com/trentm/node-bunyan#cli-usage).

While in the root folder run:

```bash
make launch-dev | make bunyan
```

You can filter only records above a certain level by adding `-l warn` to the bunyan make reference:

Inside of root the root Makefile, find the entry for bunyan. Change it to `npm exec --prefix backend bunyan -- -l warn`.

## Testing

Console frontend tests consist of:

- Jest frontend unit tests against mocked endpoints using MSW (mock service worker), and component library tests.
- Cypress E2E tests against a real Pachyderm cluster in EE (with auth), and in CE (unauth).
- Jest backend unit tests

### Running unit / integration tests

The following command runs unit and integration tests in `./frontend` and `./backend`.

```bash
make test
```

You may want to be more precise in either frontend or backend tests. To run jest against a single test, `cd` into the appropriate directory and provide a pattern for the test runner.

```bash
cd frontend
npm run test:frontend TestName.test.ts
npm run test:components Button
```

If you want to only run one of the tests outlined above refer to the `package.json` found in the subdirectory `frontend` or `backend`.

### E2E tests

We use Cypress for E2E tests.

There are three distinct sets of tests. They run against:

1. Community Edition.
1. Enterprise Edition.

Before running one of the above sets of tests you must configure your local pachyderm and console setup to match what the test is expecting to test against.

**Otherwise, to run either of the authenticated (Enterprise)or unauthenticated (Community Edition) test suites:**

1. [Run a local Pachyderm cluster](./README_Pachyderm.md) in either Enterprise or Community Edition.
1. Run Console locally with `make launch-dev`.
1. Use one of the following commands to start Cypress:

**To run the authenticated (Enterprise) test suite:**

```bash
make e2e-auth
```

**To run the unauthenticated (Community Edition) test suite:**

1. Note: Ensure you add `PACHYDERM_ENTERPRISE_KEY` env variable with a valid [Enterprise key](https://enterprise-token-gen.pachyderm.io/dev) to your `.env.development.local`. Otherwise one test will fail.

2. Run the following command:

   ```bash
   make e2e
   ```

## Running the production server

```bash
make launch-prod
```

This will start the production server at `localhost:3000`.

## Environment variables

All variables are stored in `.env.${environmentName}`. You can find a list of the default variables configured in the Console container in the `.env.production` file.

### React build-time variables

Build-time variables are embedded into the JavaScript files during the build process. Any variable starting with the `REACT_APP` prefix (without `_RUNTIME`) are build-time variables.

Our build tool, Vite, first looks at the `REACT_APP` variables present in the `.env` or environment, then scans the code for any occurrences of `process.env.WHATEVER` during the build process. It then statically replaces these instances with the value of the environment variable.

```ts
// .env.production
REACT_APP_PORT = 3000;

// React code
const port = process.env.REACT_APP_PORT;

// Built React code
const port = 3000;
```

⚠️ Be aware, since `process` is not available in the browser, errors may occur if a `REACT_APP` env var exists in your code but is not defined in all `.env.ENVIRONMENT` files. The call to `process.env.WHATEVER` will not be replaced, leading to an error when the browser attempts to run it.

ℹ️ In the Node.js runtime (on the server, and during test execution), these variables can be accessed like other environment variables on `process.env`, without referring to pachDashConfig.

### React runtime variables

Variables prefixed with the `REACT_APP_RUNTIME` are runtime variables. In a similiar manner to build-time variables, Vite statically replaces these with references to `window.pachDashConfig.WHATEVER` during the build process. The `pachDashConfig` object is then injected into the served index.html by our Express server at runtime.

The helper module at [runtimeVariables.ts](./frontend/src/lib/runtimeVariables.ts) handles these variables in the frontend, retrieving these variables correctly depending on the current environment (dev or prod).

```ts
// .env.production
REACT_APP_RUNTIME_PORT = 3000;

// React code
const port = process.env.REACT_APP_RUNTIME_PORT;

// Built React code
const port = window.pachDashConfig.REACT_APP_RUNTIME_PORT;
```

ℹ️ These are needed so that end-users can inject env vars from the Helm chart when they deploy Pachyderm. Adding a runtime variable to the helm chart is a manual process.

### Shared variables

Variables prefixed with `REACT_APP` or `_RUNTIME` are both available in the backend server. If you need to share an environment variable between the client and the server code, use the `REACT_APP` or `REACT_APP_RUNTIME` prefixes.

```ts
// can be used both on frontend and backend
const apiEndpoint = process.env.REACT_APP_REPORTING_API_ENDPOINT;
```

### Server-only variables

Finally, variables without `REACT_APP` or `REACT_APP_RUNTIME` prefixes are server-only variables and are only accessible in the server-side code.

```ts
// .env.production
PORT = 3000;

// server code
const port = process.env.PORT;
```

## Debugging Tests

To debug tests, add `--node-options=--inspect-brk` to npx. Like the following command:

```shell
npx --node-options=--inspect-brk jest
```

This command will execute Jest and pause its execution until you attach a debugger. You can then use the debugger to inspect and debug your test code.

You can attach the VSCode debugger by opening the command pallete and using `Debug: Attach to Node Process`.

## How to add icons

We typically export icons provided to us from the design team on Figma. The final SVG file should have one `svg` parent and one path child with no ids set, and `fill="currentcolor"` on the parent.

1. Install this SVG export extension for Figma <https://www.figma.com/community/plugin/814345141907543603/SVG-Export>
2. Set the default options "Use currentcolor as fill" and "Remove all fills" to true
3. Export your icons and add them under the `SVG` component in this repo
4. Update `index.ts` and `IconPreview.js` as appropriate.

## Updating backend protobuf code

1. Install [jq](https://stedolan.github.io/jq/download/)
1. Update the pachyderm version in `version.json`
1. Change directories to `backend/src/proto`
1. Install with `npm i`
1. Generate new protos with `npm run build:proto`

**_Note: You may want to delete your `node_modules` folder generated in this step after building new protos_**