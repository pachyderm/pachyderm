import {render, waitFor} from '@testing-library/react';
import React from 'react';
import {Route} from 'react-router';

import AuthenticatedRoute from '@dash-frontend/components/AuthenticatedRoute';
import {mockServer, withContextProviders} from '@dash-frontend/testHelpers';

const windowLocation = window.location;

describe('AuthenticatedRoute', () => {
  const TestBed = withContextProviders(() => {
    return (
      <>
        <Route
          path="/authenticated"
          component={AuthenticatedRoute(() => (
            <>Authenticated!</>
          ))}
        />

        <Route path="/error" component={() => <>Error</>} />
      </>
    );
  });

  beforeEach(() => {
    window.history.pushState('', '', '/authenticated');
  });

  afterEach(() => {
    window.location = windowLocation;
  });

  it('should allow logged in users to reach authenticated routes', async () => {
    window.localStorage.setItem('auth-token', '123');

    const {findByText} = render(<TestBed />);

    expect(await findByText('Authenticated!')).toBeVisible();
  });

  it('should redirect the user to the error page if there is an issue in the oauth flow', async () => {
    window.localStorage.setItem('oauthError', 'error');

    render(<TestBed />);

    await waitFor(() => expect(window.location.pathname).toBe('/error'));
  });

  it('should redirect the user to the error page if there is an issue with redeeming the auth code', async () => {
    mockServer.setHasInvalidIdToken(true);
    window.localStorage.setItem('oauthCode', 'code');

    render(<TestBed />);

    await waitFor(() => expect(window.location.pathname).toBe('/error'));
  });

  it('should exchange the oauth code for users returning from the oauth flow', async () => {
    window.localStorage.setItem('oauthCode', 'code');

    const {findByText} = render(<TestBed />);

    expect(await findByText('Authenticated!')).toBeVisible();
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
      expect(window.location.assign).toBeCalledWith(
        [
          'http://localhost:30658/auth',
          `?client_id=${process.env.REACT_APP_OAUTH_CLIENT_ID}`,
          '&redirect_uri=http://localhost/oauth/callback/?inline=true',
          '&response_type=code',
          '&scope=openid+email+profile+audience:server:client_id:pachd',
          '&state=AAAAAAAAAAAAAAAAAAAA',
        ].join(''),
      ),
    );
  });
});
