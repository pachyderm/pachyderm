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

class MockIntersectionObserver implements IntersectionObserver {
  readonly root: Element | null = null;
  readonly rootMargin: string = '';
  readonly thresholds: ReadonlyArray<number> = [];
  disconnect: () => void = () => null;
  observe: (target: Element) => void = () => null;
  takeRecords: () => IntersectionObserverEntry[] = () => [];
  unobserve: (target: Element) => void = () => null;
}

Object.defineProperty(window, 'IntersectionObserver', {
  writable: true,
  configurable: true,
  value: MockIntersectionObserver,
});

Object.defineProperty(global, 'IntersectionObserver', {
  writable: true,
  configurable: true,
  value: MockIntersectionObserver,
});

Object.defineProperty(window.document, 'queryCommandSupported', {
  value: jest.fn(() => true),
  configurable: true,
  writable: true,
});

Object.defineProperty(window.document, 'execCommand', {
  value: jest.fn(),
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
  process.env.REACT_APP_BACKEND_SUBSCRIPTIONS_PREFIX = `:${graphqlPort}/subscriptions`;
  process.env.PACHD_ADDRESS = `localhost:${grpcPort}`;
});

beforeEach(() => {
  fetchMock.dontMock();
  mockServer.resetState();

  window.localStorage.setItem('auth-token', mockServer.getAccount().id);
  window.localStorage.setItem(
    'id-token',
    generateIdTokenForAccount(mockServer.getAccount()),
  );
});

afterAll(async () => {
  await mockServer.stop();
  await server.stop();
});
