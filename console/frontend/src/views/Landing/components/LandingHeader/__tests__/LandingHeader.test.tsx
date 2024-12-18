import {render, screen} from '@testing-library/react';
import {setupServer} from 'msw/node';
import React from 'react';

import {
  mockGetVersionInfo,
  mockGetAccountAuth,
  mockGetEnterpriseInfo,
} from '@dash-frontend/mocks';
import {withContextProviders, loginUser} from '@dash-frontend/testHelpers';

import LandingHeader from '../../LandingHeader';

describe('LandingHeader', () => {
  const server = setupServer();

  const Header = withContextProviders(LandingHeader);

  beforeAll(() => server.listen());

  beforeEach(() => {
    window.localStorage.clear();
    server.resetHandlers();
    server.use(mockGetVersionInfo());
  });

  afterAll(() => server.close());

  it('should show the hpe branding when enterprise is active', async () => {
    server.use(mockGetAccountAuth());
    server.use(mockGetEnterpriseInfo());
    loginUser();
    render(<Header />);

    await screen.findByRole('heading', {name: 'HPE ML Data Management'});
  });
});
