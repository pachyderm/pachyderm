import 'cross-fetch';
import {enableFetchMocks} from 'jest-fetch-mock';
import '@testing-library/jest-dom';
import '@testing-library/jest-dom/extend-expect';

import mockServer from '@dash-backend/mock';

enableFetchMocks();

beforeAll(async () => {
  const port = await mockServer.start();

  process.env.REACT_APP_BACKEND_PREFIX = `http://localhost:${port}/`;
  process.env.REACT_APP_BACKEND_GRAPHQL_PREFIX = `http://localhost:${port}/graphql`;
});
beforeEach(() => {
  fetchMock.dontMock();
  // mockServer.resetData(); TODO implement this
});
afterAll(() => mockServer.stop());
