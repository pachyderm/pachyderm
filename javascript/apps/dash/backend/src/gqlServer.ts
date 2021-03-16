import fs from 'fs';
import path from 'path';

import {ApolloServer, gql} from 'apollo-server-express';
import {v4 as uuid} from 'uuid';

import loggingPlugin from '@dash-backend/apollo/plugins/loggingPlugin';
import baseLogger from '@dash-backend/lib/log';
import resolvers from '@dash-backend/resolvers';

const gqlServer = new ApolloServer({
  context: ({req}) => {
    const log = baseLogger.child({
      pachdAddress: req.header('pachd-address'),
      operationId: uuid(),
    });

    return {
      authToken: req.header('auth-token'),
      pachdAddress: req.header('pachd-address'),
      log,
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
  plugins: [loggingPlugin],
});

export default gqlServer;
