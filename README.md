# Dash

## Install Dependencies
```
make install
```

## Running tests
```
make test
```

### e2e tests
1. Start up a [local pachyderm cluster](#pachyderm-cluster)
1. Setup auth to our e2e Auth0 client [like this](#auth0-config)
1. Start running console locally with `make launch-dev`

Finally in another terminal window:
```
make e2e
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

<a name="pachyderm-cluster"></a>
### Using 2.0 Pachyderm Cluster 🚧
This feature is still under development, and will likely change.

#### Deploy locally without the Console pod
1. [Install helm](https://helm.sh/docs/intro/install/).
1. Grab the latest pachyderm helm chart: `helm repo add pachyderm https://pachyderm.github.io/helmchart`
1. If you haven't already, start up minikube
1. Install pachyderm locally: `helm install pachyderm --set deployTarget=LOCAL --version 2.0.0 pachyderm/pachyderm`
1. Delete your existing pre-2.0 pachctl config file: `rm ~/.pachyderm/config.json`

#### Configuring Auth

<a name="auth0-config"></a>
##### With Auth0

1. Find the Auth0 console client ID and client secret in 1Password
1. Find the user login for e2e-testing@pachyderm.com in 1Password
1. Run `make setup-auth` and enter the values from 1Password when asked (they're written to .env.development.local so you shouldn't need to do this again)
1. Run `pachctl port-forward`

##### With Github
1. [Create a Github OAuth app](https://docs.github.com/en/developers/apps/creating-an-oauth-app). For local clusters, set your callback url to `http://localhost:30658/callback`. Make sure to save the secret key, you'll need it for the next step.
1. Generate an enterprise key for the next step: https://enterprise-token-gen.pachyderm.io/dev. For Mac OS users, `echo '<your-enterprise-token-here>' | pachctl license activate`.
1. Run `AUTH_CLIENT=github make setup-auth`. This will walk you through the setup for your local cluster.
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

## Working with GraphQL

`graphql-codegen` is leveraged in this project to generate both type definitions and client-side code for Console's GraphQL resources based on this codebase's graphql documents.

This repo defines a command `make graphql`, which will generate the Typescript code with `graphql-codegen`, and place that code in the `/generated` directory in both `/frontend` and `/backend`. **Note**: This command is also checked in CI to ensure that types that have been checked in match all of the graphql documents.

### Steps for defining new Types.
1. Update `/backend/src/schema.graphqls`.
    ```graphqls
    // ...
    type World {
      msg: String!
    }

    Query {
      // ...
      hello: World!
    }
    ```
1. Run `make graphql`
1. Implement the resolver. **Note**: You don't need to create a new resolver file per query/mutation. These files are created on a per-resource basis.
    ```ts
    // src/resolvers/World.ts

    import {QueryResolvers} from '@graphqlTypes';

    interface WorldResolver {
      Query: {
        hello: QueryResolvers['hello'];
      };
    }

    const worldResolver: WorldResolver = {
      Query: {
        hello: () => {
          // new code...
        }
      }
    }

    export default worldResolver;
    ```
1. If you've created a new resolver in the previous step, add that to the index resolver. Additionally, if the new resolver does not require authentication, add it to the `unauthenticated` array. This should be _very_ rare, and it should be clear if and when you need to use this escape hatch.
    ```ts
    // ...
    import worldResolver from './World';

    const resolvers: Resolvers = merge(
      // ...
      worldResolver,
      {},
    );
    ```
### Steps for defining new Queries and Mutations.
1. Add a new file for the query or mutation under `/frontend/src/<queries|mutations>`. This will be a Typescript file that uses `gql` to generate a parseable GraphQL document from a template string.

    ```ts
    // frontend/src/queries/hello.ts
    import {gql} from '@apollo/client';

    export const HELLO_QUERY = gql`
      query helloWorld {
        hello {
          msg
        }
      }
    `;
    ```
1. Run `make graphql`.
1. Create a wrapper React hook to abstract any `@apollo/client` specific logic.
    ```ts
    // frontend/src/hooks/useHello.ts
    import {useHelloQuery} from '@dash-frontend/generated/hooks';

    export const useHello = () => {
      const {data, error, loading} = useHelloQuery();

      return {
        error,
        msg: data?.msg || '',
        loading,
      };
    };
    ```
