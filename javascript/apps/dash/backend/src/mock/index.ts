import fs from 'fs';
import http from 'http';
import path from 'path';
import {URL} from 'url';

import {Server, ServerCredentials} from '@grpc/grpc-js';
import {APIService as AuthService} from '@pachyderm/proto/pb/auth/auth_grpc_pb';
import {APIService as PFSService} from '@pachyderm/proto/pb/pfs/pfs_grpc_pb';
import {APIService as PPSService} from '@pachyderm/proto/pb/pps/pps_grpc_pb';
import {APIService as ProjectsService} from '@pachyderm/proto/pb/projects/projects_grpc_pb';
import express from 'express';
import {sign} from 'jsonwebtoken';

import keys from './fixtures/keys';
import openIdConfiguration from './fixtures/openIdConfiguration';
import auth from './handlers/auth';
import pfs from './handlers/pfs';
import pps from './handlers/pps';
import projects from './handlers/projects';

const grpcPort = 50051;
const authPort = 30658;

const createServer = () => {
  const grpcServer = new Server();
  const authApp = express();
  let authServer: http.Server;

  grpcServer.addService(PPSService, pps);
  grpcServer.addService(PFSService, pfs);
  grpcServer.addService(AuthService, auth);
  grpcServer.addService(ProjectsService, projects);

  authApp.get('/.well-known/openid-configuration', (_, res) => {
    return res.send(openIdConfiguration);
  });

  authApp.get('/keys', (_, res) => {
    return res.send(keys);
  });

  authApp.get('/auth', (req, res) => {
    const state = String(req.query.state);
    const redirectUri = String(req.query.redirect_uri);

    const url = new URL('', redirectUri);
    url.searchParams.append('state', state);
    url.searchParams.append('code', 'xyz');

    return res.redirect(url.toString());
  });

  authApp.post('/token', (_, res) => {
    const idToken = sign(
      {some: 'stuff', azp: 'dash'},
      fs.readFileSync(path.resolve(__dirname, 'mockPrivate.key')),
      {
        algorithm: 'RS256',
        issuer: openIdConfiguration.issuer,
        subject: 'user',
        audience: ['pachd', 'dash'],
        expiresIn: '30 days',
      },
    );

    // eslint-disable-next-line @typescript-eslint/naming-convention
    return res.send({id_token: idToken});
  });

  const mockServer = {
    grpcPort,
    authPort,
    start: () => {
      return new Promise<number>((res, rej) => {
        authServer = authApp.listen(authPort, () => {
          console.log(`auth server listening on ${authPort}`);
        });

        grpcServer.bindAsync(
          `localhost:${grpcPort}`,
          ServerCredentials.createInsecure(),
          (err: Error | null, port: number) => {
            if (err != null) {
              rej(err);
            }

            console.log(`mock grpc listening on ${port}`);
            grpcServer.start();

            res(grpcPort);
          },
        );
      });
    },

    stop: () => {
      return new Promise<null>((res, rej) => {
        grpcServer.tryShutdown((grpcError) => {
          let closeErr = grpcError;

          authServer.close((authError) => {
            closeErr = authError;
          });

          if (closeErr != null) {
            rej(closeErr);
          }

          res(null);
        });
      });
    },
  };

  return mockServer;
};

const server = createServer();

if (require.main === module) {
  server.start();
}

export default server;
