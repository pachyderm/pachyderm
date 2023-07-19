import 'cross-fetch';
import '@testing-library/jest-dom';
import '@testing-library/jest-dom/extend-expect';

import {randomBytes} from 'crypto';

import {generateIdTokenForAccount} from '@dash-backend/testHelpers';
import {configure} from '@testing-library/react';
import {enableFetchMocks} from 'jest-fetch-mock';

import {server, mockServer} from '@dash-frontend/testHelpers';

configure({asyncUtilTimeout: 5000});

enableFetchMocks();

// This is needed for grpc-js to run in the mock server in jest jsdom mode.
global.setImmediate =
  global.setImmediate ||
  function (callback: any) {
    return setTimeout(callback, 0);
  };

Object.defineProperty(window, 'crypto', {
  value: {
    getRandomValues: jest.fn((arr: number[]) => randomBytes(arr.length)),
  },
  configurable: true,
  writable: true,
});

class MockObserver implements IntersectionObserver {
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
  value: MockObserver,
});

Object.defineProperty(global, 'IntersectionObserver', {
  writable: true,
  configurable: true,
  value: MockObserver,
});

Object.defineProperty(global, 'ResizeObserver', {
  writable: true,
  configurable: true,
  value: MockObserver,
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

// This method needs an implementation due to CodeMirror
// https://github.com/jsdom/jsdom/pull/3533
// https://github.com/jsdom/jsdom/issues/3002
// https://github.com/jsdom/jsdom/issues/2751
Object.defineProperty(window.Range.prototype, 'getClientRects', {
  value: jest.fn(() => [
    {
      bottom: 0,
      height: 0,
      left: 0,
      right: 0,
      top: 0,
      width: 0,
    },
  ]),
  configurable: true,
  writable: true,
});

window.HTMLElement.prototype.scrollIntoView = jest.fn();

beforeAll(async () => {
  const [grpcPort, authPort] = await mockServer.start();
  const graphqlPort = await server.start();

  process.env.GRPC_PORT = String(grpcPort);
  process.env.AUTH_PORT = String(authPort);
  process.env.ISSUER_URI = `http://localhost:${authPort}`;
  process.env.REACT_APP_RUNTIME_ISSUER_URI = `http://localhost:${authPort}`;
  process.env.REACT_APP_BACKEND_PREFIX = `http://localhost:${graphqlPort}/`;
  process.env.REACT_APP_BACKEND_GRAPHQL_PREFIX = `http://localhost:${graphqlPort}/graphql`;
  process.env.PACHD_ADDRESS = `localhost:${grpcPort}`;
  process.env.REACT_APP_RUNTIME_SUBSCRIPTIONS_PREFIX = `:${graphqlPort}/graphql`;
  process.env.REACT_APP_RUNTIME_PACHYDERM_PUBLIC_HOST = 'localhost';
  process.env.REACT_APP_RUNTIME_PACHYDERM_PUBLIC_TLS = 'false';
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
