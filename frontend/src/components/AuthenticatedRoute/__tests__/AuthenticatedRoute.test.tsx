import {render, waitFor, screen} from '@testing-library/react';
import {setupServer} from 'msw/node';
import React from 'react';

import AuthenticatedRoute from '@dash-frontend/components/AuthenticatedRoute';
import {
  mockAuthenticate,
  mockAuthConfig,
  mockExchangeCode,
  mockGetAccountAuth,
  mockGetEnterpriseInfo,
  mockPipelines,
  mockWhoAmINotActivated,
  mockAuthConfigError,
  mockAuthExchanceError,
} from '@dash-frontend/mocks';
import {withContextProviders, loginUser} from '@dash-frontend/testHelpers';

const windowLocation = window.location;

describe('AuthenticatedRoute', () => {
  const server = setupServer();

  const TestBed = withContextProviders(
    AuthenticatedRoute(() => <>Authenticated!</>),
  );

  beforeAll(() => {
    server.listen();
    server.use(mockGetAccountAuth());
    server.use(mockGetEnterpriseInfo());
    server.use(mockPipelines());
    server.use(mockAuthConfig());
    server.use(mockWhoAmINotActivated());
  });

  beforeEach(() => {
    window.history.pushState('', '', '/authenticated');
  });

  afterEach(() => {
    window.localStorage.clear();
  });

  afterAll(() => server.close());

  it('should allow logged in users to reach authenticated routes', async () => {
    loginUser();

    render(<TestBed />);

    expect(await screen.findByText('Authenticated!')).toBeVisible();
  });

  it('should redirect unauthenticated users through the oauth flow', async () => {
    window.history.pushState(
      '',
      '',
      '/authenticated?connection=github&login_hint=test@pachyderm.com',
    );
    Object.defineProperty(window, 'location', {
      configurable: true,
      writable: true,
      value: {
        ...windowLocation,
        assign: jest.fn(),
      },
    });

    render(<TestBed />);

    await waitFor(() =>
      expect(window.location.assign).toHaveBeenCalledWith(
        [
          'http://localhost/dex/auth',
          '?client_id=console-test',
          '&redirect_uri=http://localhost/oauth/callback/?inline=true',
          '&response_type=code',
          '&scope=openid+email+profile+groups+audience:server:client_id:pachd',
          '&state=AAAAAAAAAAAAAAAAAAAA',
          '&connection=github',
          '&login_hint=test@pachyderm.com',
        ].join(''),
      ),
    );
  });

  it('should show an error page if there is an issue in the oauth flow', async () => {
    window.localStorage.setItem('oauthError', 'error');

    render(<TestBed />);

    expect(
      await screen.findByText("Looks like this API call can't be completed."),
    ).toBeInTheDocument();
  });

  it('should exchange the oauth code for users returning from the oauth flow', async () => {
    server.use(mockExchangeCode());
    server.use(mockAuthenticate());
    window.localStorage.setItem('oauthCode', 'code');

    render(<TestBed />);

    expect(await screen.findByText('Authenticated!')).toBeVisible();
  });

  it('should show an error page if there is an issue with redeeming the auth code', async () => {
    server.use(mockAuthExchanceError());
    window.localStorage.setItem('oauthCode', 'code');

    render(<TestBed />);

    expect(
      await screen.findByText('Unable to authenticate. Try again later.'),
    ).toBeInTheDocument();
  });

  it('should show an error page if the OIDC provider is misconfigured', async () => {
    server.use(mockAuthConfigError());

    render(<TestBed />);

    expect(
      await screen.findByText('Unable to authenticate. Try again later.'),
    ).toBeInTheDocument();
  });
});
