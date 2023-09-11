import http from 'http';
import {AddressInfo} from 'net';
import {URL} from 'url';

import {Server, ServerCredentials} from '@grpc/grpc-js';
import cors from 'cors';
import express from 'express';
import cloneDeep from 'lodash/cloneDeep';

import log from '@dash-backend/lib/log';
import accounts from '@dash-backend/mock/fixtures/accounts';
import keys from '@dash-backend/mock/fixtures/keys';
import openIdConfiguration from '@dash-backend/mock/fixtures/openIdConfiguration';
import {
  GRPC_MAX_MESSAGE_LENGTH,
  AuthAPIService,
  EnterpriseAPIService,
  PfsAPIService,
  PpsAPIService,
  AdminAPIService,
  VersionAPIService,
} from '@dash-backend/proto';
import {generateIdTokenForAccount} from '@dash-backend/testHelpers';

const defaultState = {
  tokenError: false,
  authConfigurationError: false,
  account: accounts['1'],
  authPort: 0,
};

const getAuthUrl = (path: string) => {
  const issuerUrl = new URL(process.env.ISSUER_URI || '');
  issuerUrl.pathname = path;

  return issuerUrl.toString();
};

export let grpcServer: Server;

const createServer = () => {
  grpcServer = new Server({
    'grpc.max_receive_message_length': GRPC_MAX_MESSAGE_LENGTH,
    'grpc.max_send_message_length': GRPC_MAX_MESSAGE_LENGTH,
  });
  const authApp = express();
  let authServer: http.Server;
  const state = cloneDeep(defaultState);

  // grpcServer.addService(PpsAPIService, pps.getService());
  // grpcServer.addService(PfsAPIService, pfs.getService());
  // grpcServer.addService(AuthAPIService, auth.getService());
  // grpcServer.addService(EnterpriseAPIService, enterprise.getService());
  // grpcServer.addService(AdminAPIService, admin.getService());
  // grpcServer.addService(VersionAPIService, version.getService());

  // allow cors request to dev auth server
  // for devtools
  authApp.use(cors());

  authApp.get('/.well-known/openid-configuration', (_, res) => {
    if (mockServer.state.authConfigurationError) {
      return res.send({});
    }

    return res.send({
      issuer: process.env.ISSUER_URI,
      authorization_endpoint: getAuthUrl('/auth'),
      token_endpoint: getAuthUrl('/token'),
      jwks_uri: getAuthUrl('/keys'),
      userinfo_endpoint: getAuthUrl('/userinfo'),
      device_authorization_endpoint: getAuthUrl('/device/code'),
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
    // setError: (error: ServiceError | null) => (MockState.state.error = error),
    setTokenError: (tokenError: boolean) =>
      (mockServer.state.tokenError = tokenError),
    setAuthConfigurationError: (authConfigurationError: boolean) =>
      (mockServer.state.authConfigurationError = authConfigurationError),
    setAccount: (id: string) => {
      mockServer.state.account = accounts[id] || accounts.default;
    },
    getAccount: () => mockServer.state.account,
    // getState: () => MockState.getState(),
    // resetState: () => {
    //   MockState.resetState();

    //   mockServer.state = {
    //     ...cloneDeep(defaultState),
    //     authPort: mockServer.state.authPort,
    //   };
    // },
    removeAllServices: () => {
      grpcServer.removeService(PpsAPIService);
      grpcServer.removeService(PfsAPIService);
      grpcServer.removeService(AuthAPIService);
      grpcServer.removeService(EnterpriseAPIService);
      grpcServer.removeService(AdminAPIService);
      grpcServer.removeService(VersionAPIService);
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
