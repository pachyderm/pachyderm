import {Server} from 'http';
import {AddressInfo} from 'net';
import path from 'path';

import Analytics from '@rudderstack/rudder-sdk-node';
import * as Sentry from '@sentry/node';
import compression from 'compression';
import cookieParser from 'cookie-parser';
import {renderFile} from 'ejs';
import express, {Express, urlencoded, json} from 'express';

import gqlServer from '@dash-backend/gqlServer';
import handleFileDownload from '@dash-backend/handlers/handleFileDownload';
import log from '@dash-backend/lib/log';

import createWebsocketServer from './createWebsocketServer';
import uploadsRouter from './handlers/handleFileUpload';
import fileUploads from './lib/FileUploads';
import retryAnalyticsContext from './lib/retryAnalyticsContext';

const PORT = process.env.GRAPHQL_PORT || '3000';
const FE_BUILD_DIRECTORY =
  process.env.FE_BUILD_DIRECTORY ||
  path.join(__dirname, '../../frontend/build');

const ENABLE_TELEMETRY =
  process.env.REACT_APP_RUNTIME_DISABLE_TELEMETRY !== 'true';

const attachWebServer = (app: Express) => {
  // Attach all environment variables prefixed with REACT_APP
  const env = Object.keys(process.env)
    .filter((key) => key.startsWith('REACT_APP') || key === 'NODE_ENV')
    .reduce<{[key: string]: string}>((env, key) => {
      env[key] = process.env[key] || '';

      return env;
    }, {});

  app.use(compression());
  app.set('views', FE_BUILD_DIRECTORY);
  app.set('view options', {delimiter: '?'});
  app.engine('html', renderFile);

  app.get('/health', (_, res) => {
    res.status(200).json({status: 'Healthy'});
  });

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

const attachFileHandlers = (app: Express) => {
  app.use((_req, res, next) => {
    if (
      process.env.NODE_ENV === 'development' ||
      process.env.NODE_ENV === 'test'
    ) {
      res.setHeader('Access-Control-Allow-Origin', 'http://localhost:4000');
    }
    res.setHeader('Access-Control-Allow-Methods', 'GET, POST');
    res.setHeader('Access-Control-Allow-Credentials', 'true');
    res.setHeader('Access-Control-Allow-Headers', 'Content-Type');
    next();
  });

  app.use(cookieParser());
  app.use('/upload', uploadsRouter);
  app.get(
    '/download/:projectId/:repoName/:branchName/:commitId/*',
    handleFileDownload,
  );
};

const attachAnalytics = async () => {
  if (!process.env.NODE_RUDDERSTACK_ID) {
    return;
  }
  const analyticsClient = new Analytics(
    process.env.NODE_RUDDERSTACK_ID,
    'https://pachyderm-dataplane.rudderstack.com/v1/batch',
  );
  try {
    const {anonymousId, expiration, enterpriseState} =
      await retryAnalyticsContext();

    if (anonymousId) {
      // Gather info about the cluster on launch
      analyticsClient.track({
        event: 'cluster_info',
        anonymousId,
        properties: {
          expiration,
          enterpriseState,
        },
      });
    }
  } catch {
    return;
  }
};

const createServer = () => {
  const app = express();
  app.use(json({limit: '500kb'}));
  app.use(urlencoded({extended: true}));

  attachFileHandlers(app);

  if (process.env.NODE_ENV !== 'development') {
    attachWebServer(app);
  }

  if (ENABLE_TELEMETRY) {
    if (process.env.SENTRY_DSN) {
      Sentry.init({
        dsn: process.env.SENTRY_DSN,
        tracesSampleRate: 0.5,
      });
    }
    attachAnalytics();
  }

  return {
    start: async () => {
      await gqlServer.start();
      gqlServer.applyMiddleware({app});
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
            }://${host}:${port}${gqlServer.graphqlPath}`,
          );

          res(String(address.port));
        });
      });
    },
    stop: async () => {
      return new Promise((res, rej) => {
        fileUploads.clearInterval();
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
