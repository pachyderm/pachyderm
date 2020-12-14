import fs from 'fs';
import path from 'path';

import {ApolloServer, gql} from 'apollo-server-express';

import {Repo} from 'generated/types';
import pfs from 'grpc/pfs';

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
  // TODO: Move these to /resolvers folder
  resolvers: {
    Query: {
      repos: async (parent, args, context): Promise<Repo[]> => {
        const repos = await pfs(
          context.pachdAddress,
          context.authToken,
        ).listRepo();

        return repos.map((repo) => ({
          createdAt: repo.created?.seconds || 0,
          description: repo.description,
          isPipelineOutput: false, // TODO: How do we derive this?
          name: repo.repo?.name || '',
          sizeInBytes: repo.sizeBytes,
        }));
      },
    },
  },
  typeDefs: gql(
    fs.readFileSync(path.join(__dirname, '../src/schema.graphqls'), 'utf8'),
  ),
});

export default gqlServer;
