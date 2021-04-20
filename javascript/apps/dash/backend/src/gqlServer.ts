import fs from 'fs';
import path from 'path';

import {ApolloServer, gql} from 'apollo-server-express';
import {v4 as uuid} from 'uuid';

import loggingPlugin from '@dash-backend/apollo/plugins/loggingPlugin';
import client from '@dash-backend/grpc/client';
import baseLogger from '@dash-backend/lib/log';
import resolvers from '@dash-backend/resolvers';
import {Account} from '@graphqlTypes';

import {getAccountFromIdToken} from './lib/auth';

const gqlServer = new ApolloServer({
  context: async ({req}) => {
    const idToken = req.header('id-token');

    let account: Account | undefined;
    if (idToken) {
      account = await getAccountFromIdToken(idToken);
    }

    const pachdAddress = process.env.PACHD_ADDRESS;
    const authToken = req.header('auth-token');
    const projectId = req.body?.variables?.args?.projectId;

    const log = baseLogger.child({
      pachdAddress,
      operationId: uuid(),
      account: {
        id: account?.id,
        email: account?.email,
      },
    });

    const pachClient = client({
      authToken,
      pachdAddress,
      projectId,
      log,
    });

    return {
      account,
      authToken,
      log,
      pachClient,
      pachdAddress,
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
