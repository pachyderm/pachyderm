import 'cross-fetch';
import {enableFetchMocks} from 'jest-fetch-mock';
import '@testing-library/jest-dom';
import '@testing-library/jest-dom/extend-expect';

import server from '@dash-backend/index';
import mockServer from '@dash-backend/mock';

enableFetchMocks();

beforeAll(async () => {
  const [grpcPort, authPort] = await mockServer.start();
  const graphqlPort = await server.start();

  process.env.GRPC_PORT = String(grpcPort);
  process.env.AUTH_PORT = String(authPort);
  process.env.GRAPHQL_PORT = graphqlPort;
  process.env.ISSUER_URI = `http://localhost:${authPort}`;
  process.env.REACT_APP_BACKEND_PREFIX = `http://localhost:${graphqlPort}/`;
  process.env.REACT_APP_BACKEND_GRAPHQL_PREFIX = `http://localhost:${graphqlPort}/graphql`;
  process.env.REACT_APP_PACHD_ADDRESS = `localhost:${grpcPort}`;
});
beforeEach(() => {
  fetchMock.dontMock();
  mockServer.resetState();
});
afterAll(async () => {
  await mockServer.stop();
  await server.stop();
});
