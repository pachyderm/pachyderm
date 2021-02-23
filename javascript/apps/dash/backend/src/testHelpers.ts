import {gql, GraphQLClient} from 'graphql-request';

import mockServer from './mock';

import graphqlServer from '.';

const client = new GraphQLClient(
  `http://localhost:${graphqlServer.port}/graphql`,
);

const query = <T>(query: TemplateStringsArray) => {
  return client.request<T>(
    gql(query),
    {},
    {
      'auth-token': 'xyz',
      'pachd-address': `localhost:${mockServer.grpcPort}`,
    },
  );
};

export {mockServer, graphqlServer, query};
