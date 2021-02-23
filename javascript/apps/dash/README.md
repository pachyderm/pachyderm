# Dash

## Development

### Against the mock gRPC
To run the development environment against the mock gRPC server, run:

```
// In /backend
npm run mock-and-start
```

and...

```
// In /frontend
npm start
```

This will start the UI server, API server, mock gRPC server, and a mock IDP. You can access the UI in the browser at `localhost:4000`.

### Using 2.0 Pachyderm Cluster 🚧
This feature is still under development, and will likely change.

1. Ensure that your cluster is deployed with `2.0.0-alpha.2` or later.
2. Deploy a pachyderm cluster locally, or create a workspace using Hub.
3. Create a `.env.test.local` file in your frontend directory. In this file, add the variable `REACT_APP_PACHD_ADDRESS`.
4. Fill `REACT_APP_PACHD_ADDRESS` with the address of the gRPC API server in your pach cluster. For local clusters, this is `localhost:30650`. (2.0 clusters are not yet available to be deployed on Hub)
6. Create a Github Oauth app. Make sure to save the secret key, you'll need it for the next step.
7. Run the steps in the following gist:
https://gist.github.com/actgardner/d11545796112003c637882571df4b357. DONT RUN `pachctl auth login`, as that will change your user to a non-admin user.
8. Create the pachyderm client using: `pachctl idp create-client --id dash --name dash --redirectUris http://localhost:4000/oauth/callback/\?inline\=true`. Save the client-secret!
9. Create a `.env.test.local` file in the backend directory, and create a variable called `OAUTH_CLIENT_SECRET` with the value set to the dash client secret.
10. Update the `pachd` to allow dash to issue tokens on it's behalf using `pachctl idp update-client pachd --trustedPeers dash`
11. Run `pachctl port-forward`
11. Run `npm run start:dev` from the /backend directory, and `npm run start` from the /frontend directory.
12. Make sure to delete any existing `auth-token` from the mock-server in `localStorage`.
13. Navigate to `localhost:4000` in the browser.
