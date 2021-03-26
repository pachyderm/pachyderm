# Dash

## Install Dependencies
```
make install
```

## Development

### Against the mock gRPC
To run the development environment against the mock gRPC server, run:

`make launch-mock`

This will start the UI server, API server, mock gRPC server, and a mock IDP. You can access the UI in the browser at `localhost:4000`.

### Using 2.0 Pachyderm Cluster 🚧
This feature is still under development, and will likely change.

1. Ensure that your cluster is deployed with `2.0.0-alpha.9` or later. Users of brew can install with `brew install pachyderm/tap/pachctl@2.0`.
1. [Deploy a pachyderm cluster locally](https://docs.pachyderm.com/latest/getting_started/local_installation/), or create a workspace using Hub.
1. [Create a Github OAuth app](https://docs.github.com/en/developers/apps/creating-an-oauth-app). For local clusters, set your callback url to `http://localhost:30658/callback`. Make sure to save the secret key, you'll need it for the next step.
1. Generate an enterprise key for the next step: https://enterprise-token-gen.pachyderm.io/dev. For Mac OS users, `echo '<your-enterprise-token-here>' | pachctl enterprise activate`.
1. Run `make setup-auth`. This will walk you through the setup for your local cluster.
1. Run `pachctl port-forward`
1. (Optional) Use `pachctl auth login` to login via Github. If you opt not to do this, you will continue as the root user when creating resources.
2. Run `make launch-dev`.
3. Make sure to delete any existing `auth-token` from the mock-server in `localStorage`.
4. Navigate to `localhost:4000` in the browser.

## Running the production server

`make launch-prod`

This will start the production server at `localhost:3000`. Additionally, if you'd like to test the production UI/API against the mock gRPC & Auth server, you can run `npm run start:mock` from /backend and add a `.env.production.local` file that replicates the variables found in `.env.test`.
## Graphql GQL/TS generation

In the /frontend directory, `generate-operations` will output a `operations.gql` file to /generated, which includes all of the operations that we use in the client code. These are used when writing backend unit tests.

In the /backend directory, `generate-types` will output  `types.ts` to /backend/generated which is a single set of Typescript types used in both the Frontend and Backend projects. NOTE: The Frontend references these types using `@graphqlTypes`.
