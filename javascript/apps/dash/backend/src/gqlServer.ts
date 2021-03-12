import fs from 'fs';
import path from 'path';

import {ApolloServer, gql} from 'apollo-server-express';

import resolvers from '@dash-backend/resolvers';

const gqlServer = new ApolloServer({
  context: ({req}) => {
    return {
      authToken: req.header('auth-token'),
      pachdAddress: req.header('pachd-address'),
    };
  },
  // TODO: Maybe move this and add global error messaging
  formatError: (error) => {
    if (error.extensions?.code === 'INTERNAL_SERVER_ERROR') {
      error.message = 'Something went wrong';
    }
    return error;
  },
  resolvers,
  typeDefs: gql(
    fs.readFileSync(path.join(__dirname, '../src/schema.graphqls'), 'utf8'),
  ),
});

export default gqlServer;
