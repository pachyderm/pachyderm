import {Server} from 'http';
import {AddressInfo} from 'net';
import path from 'path';

import * as Sentry from '@sentry/node';
import compression from 'compression';
import cookieParser from 'cookie-parser';
import {renderFile} from 'ejs';
import express, {Express, urlencoded, json} from 'express';

import log from '@dash-backend/lib/log';

import {
  exchangeCode,
  getAuthAccount,
  getAuthConfig,
} from './handlers/handleAuth';
import {encodeArchiveUrl} from './handlers/handleEncodeArchiveUrl';
import uploadsRouter from './handlers/handleFileUpload';
import proxyForward from './handlers/handleProxyForward';
import fileUploads from './lib/FileUploads';

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

const attachFileHandlerHeaders = (app: Express) => {
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
};

const createServer = () => {
  if (!process.env.PACHD_ADDRESS) {
    throw new Error("Can't start server missing PACHD_ADDRESS");
  }

  const app = express();
  app.use(json({limit: '500kb'}));
  app.use(urlencoded({extended: true}));
  app.use(cookieParser());

  app.use('/upload', uploadsRouter);
  app.post('/auth/account', getAuthAccount);
  app.get('/auth/config', getAuthConfig);
  app.post('/auth/exchange', exchangeCode);
  app.post('/encode/archive', encodeArchiveUrl);
  app.get('/proxyForward/*', proxyForward);

  attachFileHandlerHeaders(app);

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
            }://${host}:${port}`,
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
