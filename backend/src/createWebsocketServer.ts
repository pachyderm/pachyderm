import fs from 'fs';
import {Server} from 'http';
import path from 'path';

import {makeExecutableSchema} from '@graphql-tools/schema';
import * as Sentry from '@sentry/node';
import {useServer as createServer} from 'graphql-ws/lib/use/ws';
import {Server as wsServer} from 'ws';

import log from '@dash-backend/lib/log';

import createContext from './lib/createContext';
import resolvers from './resolvers';

const createWebsocketServer = (server: Server) => {
  const websocketServer = new wsServer({
    server,
    path: '/graphql',
  });

  const schema = makeExecutableSchema({
    typeDefs: fs.readFileSync(
      path.join(__dirname, '../src/schema.graphqls'),
      'utf8',
    ),
    resolvers,
  });

  createServer(
    {
      schema,

      onConnect: () => {
        log.info({
          EventSource: 'Websocket',
          Event: 'Connect',
        });
      },
      onSubscribe: async (ctx, msg) => {
        log.info({
          EventSource: 'Websocket',
          Event: 'Subscribe',
          meta: {
            id: msg.id,
            operationName: msg.payload.operationName,
            type: msg.type,
          },
        });
      },
      onNext: (_ctx, msg) => {
        log.info({
          EventSource: 'Websocket',
          Event: 'Next',
          meta: {id: msg.id, type: msg.type},
        });
      },
      onError: (_ctx, msg, errors) => {
        Sentry.captureException(
          `[WebSocketServer error]: Message: ${msg}, Errors: ${errors}`,
        );
        log.error({EventSource: 'Websocket', Event: 'Error', errors, msg});
      },
      onComplete: (_ctx, msg) => {
        log.info({
          EventSource: 'Websocket',
          Event: 'Complete',
          meta: {id: msg.id, type: msg.type},
        });
      },
      context: async (ctx, _msg, args) => {
        const idToken = ctx.connectionParams
          ? (ctx.connectionParams['id-token'] as string)
          : '';
        const authToken = ctx.connectionParams
          ? (ctx.connectionParams['auth-token'] as string)
          : '';
        const projectId: string =
          (args.variableValues?.args as Record<string, string> | undefined)
            ?.projectId || '';
        return createContext({idToken, authToken, projectId});
      },
    },
    websocketServer,
  );
};

export default createWebsocketServer;
