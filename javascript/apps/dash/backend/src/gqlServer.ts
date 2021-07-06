import fs from 'fs';
import path from 'path';

import {ApolloServer, gql} from 'apollo-server-express';

import loggingPlugin from '@dash-backend/apollo/plugins/loggingPlugin';
import resolvers from '@dash-backend/resolvers';

import createContext from './lib/createContext';

const gqlServer = new ApolloServer({
  uploads: false,
  context: async ({req}) => {
    const idToken = req.header('id-token') || '';
    const authToken = req.header('auth-token') || '';
    const projectId: string = req.body?.variables?.args?.projectId || '';
    const host = `//${req.get('host')}` || '';
    return createContext({idToken, authToken, projectId, host});
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
