import {Server} from 'http';
import path from 'path';
import {AddressInfo} from 'net';

import {renderFile} from 'ejs';
import express, {Express} from 'express';

import gqlServer from 'gqlServer';
import {isTest} from 'lib/isTest';

const PORT = process.env.PORT || '3000';
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

const createServer = () => {
  const app = express();

  gqlServer.applyMiddleware({app, path: '/graphql'});

  if (process.env.NODE_ENV !== 'development') {
    attachWebServer(app);
  }

  return {
    start: async () => {
      return new Promise<string>((res) => {
        app.locals.server = app.listen({port: PORT}, () => {
          const address: AddressInfo = app.locals.server.address();

          if (!isTest()) {
            console.log(
              `Server ready at http://localhost:${address.port}${gqlServer.graphqlPath}`,
            );
          }

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
