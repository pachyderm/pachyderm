# Dash

## Install Dependencies
```
make install
```

## Running tests
```
make test
```

## Graphql GQL/TS generation
```
make graphql
```


## Development

### Against the mock gRPC
To run the development environment against the mock gRPC server, run:

`make launch-mock`

This will start the UI server, API server, mock gRPC server, and a mock IDP. You can access the UI in the browser at `localhost:4000`.

#### Setting account
To switch between mock accounts, you can use our devtools in the JS console of the browser.

```
> devtools.setAccount('2'); // will set your account to Barret Wallace
```

### Using 2.0 Pachyderm Cluster 🚧
This feature is still under development, and will likely change.

#### Deploy locally without the Console pod
1. [Install helm](https://helm.sh/docs/intro/install/).
1. Grab the latest pachyderm helm chart: `helm repo add pachyderm https://pachyderm.github.io/helmchart`
1. If you haven't already, start up minikube
1. Install pachyderm locally: `helm install pachyderm --set deployTarget=LOCAL --version 2.0.0-rc.2 pachyderm/pachyderm`
1. Delete your existing pre-2.0 pachctl config file: `rm ~/.pachyderm/config.json`

#### Configuring Auth
1. [Create a Github OAuth app](https://docs.github.com/en/developers/apps/creating-an-oauth-app). For local clusters, set your callback url to `http://localhost:30658/callback`. Make sure to save the secret key, you'll need it for the next step.
1. Generate an enterprise key for the next step: https://enterprise-token-gen.pachyderm.io/dev. For Mac OS users, `echo '<your-enterprise-token-here>' | pachctl license activate`.
1. Run `make setup-auth`. This will walk you through the setup for your local cluster.
1. Run `pachctl port-forward`
1. (Optional) Use `pachctl auth login` to login via Github. If you opt not to do this, you will continue as the root user when creating resources.

#### Accessing Console
1. Run `make launch-dev`.
1. Make sure to delete any existing `auth-token` from the mock-server in `localStorage`.
1. Navigate to `localhost:4000` in the browser.

## Running the production server

`make launch-prod`

This will start the production server at `localhost:3000`. Additionally, if you'd like to test the production UI/API against the mock gRPC & Auth server, you can run `npm run start:mock` from /backend and add a `.env.production.local` file that replicates the variables found in `.env.test`.

## Working with environment variables
All variables are stored in `.env.${environmentName}`

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

## Deploying Dash

### Environment variables
You can find a list of the default variables configured in the dash container in the `.env.production` file. The variables you _may_ need to override in your deployment are:

- PACHD_ADDRESS
- OAUTH_CLIENT_ID
- OAUTH_CLIENT_SECRET
- OAUTH_REDIRECT_URI
- ISSUER_URI


### Staging and Mock Staging Deploy
Whenever you create or update a PR, a GitHub Action runs that generates a docker image based on the code in the branch. It gets pushed to our public docker repository and is tagged as `pachyderm/haberdashery:dev-pr-${{ github.sha }}`. The action will comment the image tag on the PR once the image has been published. To get your changes into staging or mock staging, do the following:
1. Make sure you have internal admin permissions in the environment for which you are trying to deploy to.
2. Create a workspace in that environment and connect to the workspace through kubectl using the command from the CLI button in hub.
4. Edit the console deployment: `kubectl edit deployments console`. 
5. Update the image tag in the template to match the tag you want to deploy.
6. Wait for the new console pod to spin up. You should see your changes.


We have plans to add a console docker tag to the workspace create modal in the near future as a way to make this process easier.
