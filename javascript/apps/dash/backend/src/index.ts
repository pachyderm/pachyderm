import {Server} from 'http';

import express from 'express';

import gqlServer from 'gqlServer';

const PORT = process.env.PORT || '3000';

const createServer = () => {
  const app = express();

  gqlServer.applyMiddleware({app, path: '/graphql'});

  return {
    port: PORT,
    start: async () => {
      return new Promise<string>((res) => {
        app.locals.server = app.listen({port: PORT}, () => {
          console.log(
            `Server ready at http://localhost:${PORT}${gqlServer.graphqlPath}`,
          );
          res(PORT);
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
