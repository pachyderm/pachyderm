import fs from 'fs';
import path from 'path';

import {ApolloServer, gql} from 'apollo-server-express';

import resolvers from 'resolvers';

const gqlServer = new ApolloServer({
  context: ({req}) => {
    return {
      authToken: req.headers['auth-token'],
      pachdAddress: req.headers['pachd-address'],
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
