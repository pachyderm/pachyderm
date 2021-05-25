/* eslint-disable @typescript-eslint/naming-convention */
import http from 'http';
import {AddressInfo} from 'net';
import {URL} from 'url';

import {Server, ServerCredentials} from '@grpc/grpc-js';
import {APIService as AuthService} from '@pachyderm/proto/pb/auth/auth_grpc_pb';
import {APIService as PFSService} from '@pachyderm/proto/pb/pfs/pfs_grpc_pb';
import {APIService as PPSService} from '@pachyderm/proto/pb/pps/pps_grpc_pb';
import {APIService as ProjectsService} from '@pachyderm/proto/pb/projects/projects_grpc_pb';
import cors from 'cors';
import express from 'express';
import cloneDeep from 'lodash/cloneDeep';

import log from '@dash-backend/lib/log';
import {generateIdTokenForAccount} from '@dash-backend/testHelpers';

import accounts from './fixtures/accounts';
import keys from './fixtures/keys';
import openIdConfiguration from './fixtures/openIdConfiguration';
import auth from './handlers/auth';
import pfs from './handlers/pfs';
import pps from './handlers/pps';
import projects from './handlers/projects';

const defaultState = {
  tokenError: false,
  authConfigurationError: false,
  account: accounts['1'],
  authPort: 0,
};

const createServer = () => {
  const grpcServer = new Server();
  const authApp = express();
  let authServer: http.Server;
  const state = cloneDeep(defaultState);

  grpcServer.addService(PPSService, pps);
  grpcServer.addService(PFSService, pfs);
  grpcServer.addService(AuthService, auth.getService());
  grpcServer.addService(ProjectsService, projects.getService());

  // allow cors request to dev auth server
  // for devtools
  authApp.use(cors());

  authApp.get('/.well-known/openid-configuration', (_, res) => {
    const issuer = process.env.ISSUER_URI;

    if (mockServer.state.authConfigurationError) {
      return res.send({});
    }

    return res.send({
      issuer,
      authorization_endpoint: `${issuer}/auth`,
      token_endpoint: `${issuer}/token`,
      jwks_uri: `${issuer}/keys`,
      userinfo_endpoint: `${issuer}/userinfo`,
      device_authorization_endpoint: `${issuer}/device/code`,
      ...openIdConfiguration,
    });
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
    if (mockServer.state.tokenError) {
      throw new Error('Invalid Auth Code');
    }

    // eslint-disable-next-line @typescript-eslint/naming-convention
    return res.send({
      id_token: generateIdTokenForAccount(mockServer.state.account),
    });
  });

  const mockServer = {
    state,
    start: () => {
      return new Promise<[number, number]>((res, rej) => {
        authServer = authApp.listen(process.env.MOCK_AUTH_PORT || 0, () => {
          const address = authServer.address() as AddressInfo;

          log.info(`auth server listening on ${address.port}`);
          mockServer.state.authPort = address.port;

          grpcServer.bindAsync(
            `localhost:${process.env.MOCK_GRPC_PORT || 0}`,
            ServerCredentials.createInsecure(),
            (err: Error | null, grpcPort: number) => {
              if (err != null) {
                rej(err);
              }

              grpcServer.start();

              log.info(`mock grpc listening on ${grpcPort}`);

              res([grpcPort, address.port]);
            },
          );
        });
      });
    },

    stop: () => {
      return new Promise<null>((res, rej) => {
        grpcServer.tryShutdown((grpcError) => {
          let closeErr = grpcError;

          authServer.close((authError) => {
            closeErr = authError;

            if (closeErr != null) {
              rej(closeErr);
            }

            res(null);
          });
        });
      });
    },

    // mock methods
    setAuthError: auth.setError,
    setProjectsError: projects.setError,
    setTokenError: (tokenError: boolean) =>
      (mockServer.state.tokenError = tokenError),
    setAuthConfigurationError: (authConfigurationError: boolean) =>
      (mockServer.state.authConfigurationError = authConfigurationError),
    setAccount: (id: string) => {
      mockServer.state.account = accounts[id] || accounts.default;
    },
    getAccount: () => mockServer.state.account,
    resetState: () => {
      auth.resetState();
      projects.resetState();
      // Add additional handler resets here

      mockServer.state = {
        ...cloneDeep(defaultState),
        authPort: mockServer.state.authPort,
      };
    },
  };

  authApp.get('/set-account', (req, res) => {
    if (typeof req.query.id === 'string' && accounts[req.query.id]) {
      mockServer.setAccount(req.query.id);

      return res.send(mockServer.state.account);
    }

    return res
      .status(400)
      .send({error: 'An account fixture does not exist for the given id.'});
  });

  return mockServer;
};

const server = createServer();

if (require.main === module) {
  server.start();
}

export default server;
