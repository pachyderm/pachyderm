import {render, waitFor} from '@testing-library/react';
import {setupServer} from 'msw/node';
import React from 'react';

import {
  mockGetVersionInfo,
  mockGetAccountAuth,
  mockGetEnterpriseInfo,
} from '@dash-frontend/mocks';
import {withContextProviders, loginUser} from '@dash-frontend/testHelpers';

import BrandedTitleComponent from '../BrandedTitle';

describe('BrandedTitle', () => {
  const server = setupServer();

  const BrandedTitle = withContextProviders(() => (
    <BrandedTitleComponent title="Project" />
  ));

  beforeAll(() => {
    server.listen();
    server.use(mockGetVersionInfo());
  });

  afterEach(() => {
    server.resetHandlers();
    window.localStorage.clear();
  });

  afterAll(() => server.close());

  it('should show the pachyderm title when enterprise is inactive', async () => {
    render(<BrandedTitle title="Project" />);

    await waitFor(() => {
      expect(document.title).toBe('Project - Pachyderm Console');
    });
  });

  it('should show the HPE title when enterprise is active', async () => {
    server.use(mockGetAccountAuth());
    server.use(mockGetEnterpriseInfo());
    loginUser();

    render(<BrandedTitle title="Project" />);

    await waitFor(() => {
      expect(document.title).toBe('Project - HPE ML Data Management');
    });
  });
});
