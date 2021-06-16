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
  context: async ({req, connection}) => {
    let idToken: string | undefined;
    let authToken: string | undefined;
    let projectId: string | undefined;
    let account: Account | undefined;
    let host = '';

    if (connection) {
      idToken = connection.context['id-token'];
      authToken = connection.context['auth-token'];
      projectId = connection.variables?.args?.projectId;
    } else {
      idToken = req.header('id-token');
      authToken = req.header('auth-token');
      projectId = req.body?.variables?.args?.projectId;
      host = `//${req.get('host')}`;
    }

    if (idToken) {
      account = await getAccountFromIdToken(idToken);
    }

    const pachdAddress = process.env.PACHD_ADDRESS;
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
      log,
      pachdAddress,
      projectId,
    });

    return {
      account,
      authToken,
      host,
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
  subscriptions: {
    path: '/subscriptions',
  },
});

export default gqlServer;
