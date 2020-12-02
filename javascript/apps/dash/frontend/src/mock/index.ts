import {Server} from 'http';
import {join} from 'path';

import {GraphQLFileLoader} from '@graphql-tools/graphql-file-loader';
import {loadSchemaSync} from '@graphql-tools/load';
import {mergeSchemas} from '@graphql-tools/merge';
import {ApolloServer, addResolveFunctionsToSchema} from 'apollo-server-express';
import {json} from 'body-parser';
import express, {Express} from 'express';
import reduce from 'lodash/reduce';
import destroyable from 'server-destroy';

import {Pach} from 'lib/graphqlTypes';

import {PachFixtures, pachs} from './fixtures/pach';
import {PachPipelineFixtures, pachPipelines} from './fixtures/pachPipeline';
import {PachRepoFixtures, pachRepos} from './fixtures/pachRepo';

interface MockData {
  currentPachId: Pach['id'];
  pachs: PachFixtures;
  pachPipelines: PachPipelineFixtures;
  pachRepos: PachRepoFixtures;
}
interface MockServerApp extends Express {
  locals: {
    allResponseStatus?: number;
    data: MockData;
    delay: number;
    mockAccountResponse?: boolean;
    port: number;
    server: Server;
    responseStatusMap: Record<string, number>;
  };
}

const initialData: MockData = {
  currentPachId: 'tutorial',
  pachs,
  pachPipelines,
  pachRepos,
};

const createMockServer = (mockPort = process.env.MOCK_PORT) => {
  const app: MockServerApp = express() as MockServerApp;

  app.use(json());
  app.locals.port = Number(mockPort) || 0;
  app.locals.responseStatusMap = {};
  app.locals.data = initialData;

  const mockServer = {
    start: () => {
      return new Promise((resolve, reject) => {
        app.locals.server = app.listen(app.locals.port, () => {
          const address = app?.locals?.server?.address();

          if (address && typeof address !== 'string' && address.port) {
            app.locals.port = address.port;

            return resolve(app.locals.port);
          } else {
            return reject(new Error('The server has been closed.'));
          }
        });
        destroyable(app.locals.server);
      });
    },

    stop: () => {
      return new Promise((resolve) => {
        if (app.locals.server) {
          app.locals.server.destroy(() => resolve(null));
        } else {
          resolve(null);
        }
      });
    },

    setAccount: (id: string) => {
      // TODO: Does this still make sense?
    },

    setPach: (id: string) => {
      app.locals.data = {
        ...app.locals.data,
        currentPachId: app.locals.data.pachs[id]
          ? id
          : initialData.currentPachId,
      };
    },

    setDelay: (delay: number) => {
      app.locals.delay = delay;
    },

    setAllResponseStatus: (
      responseStatus: number,
      mockAccountResponse = true,
    ) => {
      app.locals.allResponseStatus = responseStatus;
      app.locals.mockAccountResponse = mockAccountResponse;
    },

    setResponseStatus: (
      query: string,
      responseStatus: number,
      mockAccountResponse = false,
    ) => {
      app.locals.responseStatusMap[query] = responseStatus;
      app.locals.mockAccountResponse = mockAccountResponse;
    },
    clearResponseStatus: (query: string) => {
      app.locals.responseStatusMap = reduce(
        app.locals.responseStatusMap,
        (acc, val, key) => {
          if (key !== query) acc[key] = val;
          return acc;
        },
        {} as Record<string, number>,
      );
    },

    clearAllResponseStatus: () => {
      app.locals.allResponseStatus = undefined;
      app.locals.mockAccountResponse = false;
    },

    clearDelay: () => {
      app.locals.delay = 0;
    },

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    fetch: (resource: string | Request, ...rest: any) => {
      if (typeof resource === 'string' && resource.startsWith('/')) {
        return fetch(`http://localhost:${app.locals.port}${resource}`, ...rest);
      }

      return fetch(resource, ...rest);
    },

    resetData: () => {
      app.locals.data = initialData;
    },
  };

  // error middleware
  app.use((req, res, next) => {
    const queryPath =
      req.path === '/graphql' ? req.body.operationName : req.path;
    if (queryPath === '/mock/clear-all-response-status') {
      return next();
    }

    if (
      queryPath === '/v1/account' &&
      req.method === 'GET' &&
      !app.locals.mockAccountResponse
    ) {
      return next();
    }

    if (app.locals.allResponseStatus) {
      return res.sendStatus(app.locals.allResponseStatus);
    }

    if (app.locals.responseStatusMap[queryPath] === undefined) {
      return next();
    }

    const responseStatus = app.locals.responseStatusMap[queryPath];

    return res.sendStatus(responseStatus);
  });

  const schema = mergeSchemas({
    schemas: [
      loadSchemaSync(join(__dirname, './schema.graphqls'), {
        loaders: [new GraphQLFileLoader()],
      }),
      loadSchemaSync(join(__dirname, './mockSchema.graphqls'), {
        loaders: [new GraphQLFileLoader()],
      }),
    ],
  });

  const resolvers = {
    Query: {
      hello: () => {
        return {
          id: '1',
          message: 'Hello Dash!',
        };
      },
      pipelines: () =>
        app.locals.data.pachPipelines[app.locals.data.currentPachId],
      repos: () => app.locals.data.pachRepos[app.locals.data.currentPachId],
    },
    Mutation: {
      setPach: (_: unknown, {input: {id}}: {input: {id: string}}) => {
        mockServer.setPach(id);
        return true;
      },
    },
  };

  addResolveFunctionsToSchema({schema, resolvers});

  const apolloServer = new ApolloServer({schema});

  apolloServer.applyMiddleware({app});

  return mockServer;
};

if (require.main === module) {
  createMockServer().start();
}

export default createMockServer();
