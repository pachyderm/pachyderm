import {createServiceError, status} from '@dash-backend/testHelpers';
import {render, waitFor} from '@testing-library/react';
import React from 'react';

import AuthenticatedRoute from '@dash-frontend/components/AuthenticatedRoute';
import {mockServer, withContextProviders} from '@dash-frontend/testHelpers';

const windowLocation = window.location;

describe('AuthenticatedRoute', () => {
  const TestBed = withContextProviders(
    AuthenticatedRoute(() => <>Authenticated!</>),
  );

  beforeEach(() => {
    window.history.pushState('', '', '/authenticated');
  });

  it('should allow logged in users to reach authenticated routes', async () => {
    window.localStorage.setItem('auth-token', '123');

    const {findByText} = render(<TestBed />);

    expect(await findByText('Authenticated!')).toBeVisible();
  });

  it('should show an error page if there is an issue in the oauth flow', async () => {
    window.localStorage.setItem('oauthError', 'error');

    const {findByText} = render(<TestBed />);

    expect(
      await findByText(`Looks like this API call can't be completed.`),
    ).toBeInTheDocument();
  });

  it('should show an error page if there is an issue with redeeming the auth code', async () => {
    const error = createServiceError({code: status.UNAUTHENTICATED});
    mockServer.setError(error);
    window.localStorage.setItem('oauthCode', 'code');

    const {findByText} = render(<TestBed />);

    expect(
      await findByText('Unable to authenticate. Try again later.'),
    ).toBeInTheDocument();
  });

  it('should exchange the oauth code for users returning from the oauth flow', async () => {
    window.localStorage.setItem('oauthCode', 'code');

    const {findByText} = render(<TestBed />);

    expect(await findByText('Authenticated!')).toBeVisible();
  });

  it('should show an error page if the OIDC provider is misconfigured', async () => {
    window.localStorage.removeItem('auth-token');
    window.localStorage.removeItem('id-token');
    mockServer.setAuthConfigurationError(true);

    const {findByText} = render(<TestBed />);

    expect(
      await findByText('Unable to authenticate. Try again later.'),
    ).toBeInTheDocument();
  });

  describe('external navigation', () => {
    beforeEach(() => {
      window.history.pushState(
        '',
        '',
        '/authenticated?connection=github&login_hint=test@pachyderm.com',
      );
    });

    afterAll(() => {
      window.location = windowLocation;
    });

    it('should redirect unauthenticated users through the oauth flow', async () => {
      Object.defineProperty(window, 'location', {
        configurable: true,
        writable: true,
        value: {
          ...windowLocation,
          assign: jest.fn(),
        },
      });

      window.localStorage.removeItem('auth-token');
      window.localStorage.removeItem('id-token');

      render(<TestBed />);

      await waitFor(() =>
        expect(window.location.assign).toHaveBeenCalledWith(
          [
            `http://localhost:${mockServer.state.authPort}/auth`,
            `?client_id=${process.env.OAUTH_CLIENT_ID}`,
            '&redirect_uri=http://localhost/oauth/callback/?inline=true',
            '&response_type=code',
            `&scope=openid+email+profile+groups+audience:server:client_id:${process.env.OAUTH_PACHD_CLIENT_ID}`,
            '&state=AAAAAAAAAAAAAAAAAAAA',
            '&connection=github',
            '&login_hint=test@pachyderm.com',
          ].join(''),
        ),
      );
    });
  });
});
