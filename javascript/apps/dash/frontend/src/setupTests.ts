import {randomBytes} from 'crypto';

import 'cross-fetch';
import {enableFetchMocks} from 'jest-fetch-mock';
import '@testing-library/jest-dom';
import '@testing-library/jest-dom/extend-expect';

import {generateIdTokenForAccount} from '@dash-backend/testHelpers';
import {server, mockServer} from '@dash-frontend/testHelpers';

enableFetchMocks();

Object.defineProperty(window, 'crypto', {
  value: {
    getRandomValues: jest.fn((arr: number[]) => randomBytes(arr.length)),
  },
  configurable: true,
  writable: true,
});

beforeAll(async () => {
  const [grpcPort, authPort] = await mockServer.start();
  const graphqlPort = await server.start();

  process.env.GRPC_PORT = String(grpcPort);
  process.env.AUTH_PORT = String(authPort);
  process.env.ISSUER_URI = `http://localhost:${authPort}`;
  process.env.REACT_APP_BACKEND_PREFIX = `http://localhost:${graphqlPort}/`;
  process.env.REACT_APP_BACKEND_GRAPHQL_PREFIX = `http://localhost:${graphqlPort}/graphql`;
  process.env.PACHD_ADDRESS = `localhost:${grpcPort}`;
});
beforeEach(() => {
  fetchMock.dontMock();
  mockServer.resetState();

  window.localStorage.setItem('auth-token', 'xyz');
  window.localStorage.setItem(
    'id-token',
    generateIdTokenForAccount(mockServer.state.account),
  );
});
afterAll(async () => {
  await mockServer.stop();
  await server.stop();
});
