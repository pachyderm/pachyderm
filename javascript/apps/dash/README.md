# Dash

## Development

### Against the mock gRPC
To run the development environment against the mock gRPC server, run:

`make launch-dev`

This will start the UI server, API server, mock gRPC server, and a mock IDP. You can access the UI in the browser at `localhost:4000`.

### Using 2.0 Pachyderm Cluster 🚧
This feature is still under development, and will likely change.

1. Ensure that your cluster is deployed with `2.0.0-alpha.2` or later. Users of brew can install with `brew install pachyderm/tap/pachctl@2.0`.
1. [Deploy a pachyderm cluster locally](https://docs.pachyderm.com/latest/getting_started/local_installation/), or create a workspace using Hub.
1. Create a `.env.test.local` file in your frontend directory. In this file, add the variable `REACT_APP_PACHD_ADDRESS`.
1. Fill `REACT_APP_PACHD_ADDRESS` with the address of the gRPC API server in your pach cluster. For local clusters, this is `localhost:30650`. (2.0 clusters are not yet available to be deployed on Hub)
1. [Create a Github OAuth app](https://docs.github.com/en/developers/apps/creating-an-oauth-app). For local clusters, set your callback url to `http://localhost:30658/callback`. Make sure to save the secret key, you'll need it for the next step.
1. Generate an enterprise key for the next step: https://enterprise-token-gen.pachyderm.io/dev. For Max OS users, `echo '<your-enterprise-token-here>' | pachctl enterprise activate` might be required due to input buffer limits.
1. Run the steps in the following gist:
https://gist.github.com/actgardner/d11545796112003c637882571df4b357. DONT RUN `pachctl auth login`, as that will change your user to a non-admin user.
1. Create the pachyderm client using: `pachctl idp create-client --id dash --name dash --redirectUris http://localhost:4000/oauth/callback/\?inline\=true`. Save the client-secret!
1. Create a `.env.test.local` file in the backend directory, and create a variable called `OAUTH_CLIENT_SECRET` with the value set to the dash client secret.
1. Update the `pachd` to allow dash to issue tokens on it's behalf using `pachctl idp update-client pachd --trustedPeers dash`
1. Run `pachctl port-forward`
1. Run `npm run start:dev` from the /backend directory, and `npm run start` from the /frontend directory.
1. Make sure to delete any existing `auth-token` from the mock-server in `localStorage`.
1. Navigate to `localhost:4000` in the browser.

## Running the production server

`make launch-prod`

This will start the production server at `localhost:3000`. Additionally, if you'd like to test the production UI/API against the mock gRPC & Auth server, you can run `npm run start:mock` from /backend and add a `.env.production.local` file that replicates the variables found in `.env.test`.
## Graphql GQL/TS generation

In the /frontend directory, `generate-operations` will output a `operations.gql` file to /generated, which includes all of the operations that we use in the client code. These are used when writing backend unit tests.

In the /backend directory, `generate-types` will output  `types.ts` to /backend/generated which is a single set of Typescript types used in both the Frontend and Backend projects. NOTE: The Frontend references these types using `@graphqlTypes`.
