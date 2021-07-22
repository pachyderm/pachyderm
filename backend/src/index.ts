import {Server} from 'http';
import {AddressInfo} from 'net';
import path from 'path';

import * as Sentry from '@sentry/node';
import cookieParser from 'cookie-parser';
import {renderFile} from 'ejs';
import express, {Express} from 'express';

import gqlServer from '@dash-backend/gqlServer';
import handleFileDownload from '@dash-backend/handlers/handleFileDownload';
import log from '@dash-backend/lib/log';

import createWebsocketServer from './createWebsocketServer';

const PORT = process.env.GRAPHQL_PORT || '3000';
const FE_BUILD_DIRECTORY =
  process.env.FE_BUILD_DIRECTORY ||
  path.join(__dirname, '../../frontend/build');

const attachWebServer = (app: Express) => {
  // Attach all environment variables prefixed with REACT_APP
  const env = Object.keys(process.env)
    .filter((key) => key.startsWith('REACT_APP') || key === 'NODE_ENV')
    .reduce<{[key: string]: string}>((env, key) => {
      env[key] = process.env[key] || '';

      return env;
    }, {});

  app.set('views', FE_BUILD_DIRECTORY);
  app.set('view options', {delimiter: '?'});
  app.engine('html', renderFile);

  app.get('/', (_, res) => {
    res.render('index.html', {
      PACH_DASH_CONFIG: env,
    });
  });

  // eslint-disable-next-line import/no-named-as-default-member
  app.use(express.static(FE_BUILD_DIRECTORY));

  app.get('/*', (_, res) => {
    res.render('index.html', {
      PACH_DASH_CONFIG: env,
    });
  });
};

const attachDownloadHandler = (app: Express) => {
  if (process.env.NODE_ENV === 'development') {
    app.use((_req, res, next) => {
      res.setHeader('Access-Control-Allow-Origin', 'http://localhost:4000');
      res.setHeader('Access-Control-Allow-Methods', 'GET');
      res.setHeader('Access-Control-Allow-Credentials', 'true');
      next();
    });
  }

  app.use(cookieParser());
  app.get('/download/:repoName/:branchName/:commitId/*', handleFileDownload);
};

const createServer = () => {
  const app = express();

  Sentry.init({
    dsn: process.env.SENTRY_DSN,
    tracesSampleRate: 0.5,
  });

  gqlServer.applyMiddleware({app});

  attachDownloadHandler(app);

  if (process.env.NODE_ENV !== 'development') {
    attachWebServer(app);
  }

  return {
    start: async () => {
      return new Promise<string>((res) => {
        app.locals.server = app.listen(PORT, () => {
          const address: AddressInfo = app.locals.server.address();
          const host = address.address === '::' ? 'localhost' : address.address;
          const port = address.port;

          log.info(
            `Server ready at ${
              process.env.NODE_ENV === 'production' ? 'https' : 'http'
            }://${host}:${port}${gqlServer.graphqlPath}`,
          );

          createWebsocketServer(app.locals.server);

          log.info(
            `Websocket server ready at ${
              process.env.NODE_ENV === 'production' ? 'wss' : 'ws'
            }://${host}:${port}${gqlServer.subscriptionsPath}`,
          );

          res(String(address.port));
        });
      });
    },
    stop: async () => {
      return new Promise((res, rej) => {
        if (app.locals.server) {
          const server = app.locals.server as Server;

          server.close((err) => {
            if (err) {
              rej(err);
            }
            res(null);
          });
        } else {
          res(null);
        }
      });
    },
  };
};

const server = createServer();

if (require.main === module) {
  server.start();
}

export default server;
