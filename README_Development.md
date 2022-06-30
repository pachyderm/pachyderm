# Development Guide

1. [Coding Standards](https://github.com/pachyderm/company/blob/master/handbook/frontend.md)
1. [Writing PRs](#writing-prs)
1. [Working with GraphQL](#working-with-graphql)

<br/>

## Writing PRs

Your PRs should be concise and descriptive to the contents of your changeset. If following a Jira ticket, your branch should be named after the ticket like `FRON-100`. Our current process is to wait for at least **1** reviewer before merging the PR. Try to break down larger PRs into digestible chunks, if it can be separated into smaller changes. Try to exclude large procedural changes like moving a directory or applying linter changes from new features, and instead publish those changes as clearly marked, independent PRs.

### Do:
- Try to provide a brief description of what changes you're bringing in, and why, if necessary.
- Provide screenshots as an easy glimpse into the contents of the PR if there are significant visual components, such as bringing in new UI elements.
- Include any additional details required to be able to see and run the changeset. E.g. Any preliminary setup steps, necessary configurations, or helpful tips.
- Include any details about changes external to the PR. E.g. A link to changes in CI, an example of a bot in action, or a link to a cloud console.

### Pulumi Preview
Whenever you create or update a PR, a GitHub Action runs that generates a docker image based on the code in the branch. It will be deployed by [Pulumi](https://www.pulumi.com/) and accessible from the PR comment using `admin` and `password` credentials.

<br/>

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