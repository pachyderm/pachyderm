import {render, act, waitFor, screen} from '@testing-library/react';
import React, {useState} from 'react';

import useLoginWindow from '../useLoginWindow';

const googleConnectionOpts = {
  authUrl: 'http://test.com',
  clientId: 'app',
  redirect: 'http://app.com/oauth/redirect',
  connection: 'google',
  loginHint: 'test@pachyderm.com',
};

const oidcConnectionOpts = {
  authUrl: 'http://dex.com',
  clientId: 'app',
  redirect: 'http://app.com/oauth/redirect',
  openWindow: false,
  scope: 'openid+email+audience:server:client_id:otherApp',
  connection: 'google',
  loginHint: 'test@pachyderm.com',
};

const windowOpen = window.open;

describe('useLoginWindow', () => {
  const oauthCode = '123';
  const onSuccess = jest.fn();

  const Login: React.FC = () => {
    const {initiateOauthFlow, loginWindowError, loginWindowSucceeded} =
      useLoginWindow({onSuccess});

    const [firstRenderError] = useState(loginWindowError);
    const [firstRenderSucceeded] = useState(loginWindowSucceeded);

    return (
      <>
        <button onClick={() => initiateOauthFlow(googleConnectionOpts)}>
          Login with google
        </button>
        <button onClick={() => initiateOauthFlow(oidcConnectionOpts)}>
          Login with username/password
        </button>

        {loginWindowError && <span>Error: {loginWindowError}</span>}

        {firstRenderError && (
          <span>First render error: {firstRenderError}</span>
        )}
        {firstRenderSucceeded && <span>First render succeeded</span>}
      </>
    );
  };

  beforeEach(() => {
    jest.useFakeTimers();
    Object.defineProperty(window, 'open', windowOpen);
    window.localStorage.clear();
  });

  afterEach(() => {
    jest.useRealTimers();
  });

  it('should correctly handle oauth connection based login flows', async () => {
    Object.defineProperty(window, 'open', {
      value: jest.fn(() => ({
        closed: false,
      })),
    });

    render(<Login />);

    const loginWithGoogle = screen.getByText('Login with google');

    act(() => {
      loginWithGoogle.click();
    });

    expect(window.open).toHaveBeenCalledWith(
      [
        googleConnectionOpts.authUrl,
        `?client_id=${googleConnectionOpts.clientId}`,
        `&redirect_uri=${googleConnectionOpts.redirect}`,
        '&response_type=code',
        '&scope=openid+profile+email+user_id',
        '&state=AAAAAAAAAAAAAAAAAAAA',
        `&connection=${googleConnectionOpts.connection}`,
        `&login_hint=${googleConnectionOpts.loginHint}`,
      ].join(''),
      '',
      'width=500,height=500,left=262,top=107.2',
    );

    // ensure flow does not succeed before interval established
    act(() => {
      jest.advanceTimersByTime(500);
    });

    window.localStorage.setItem('oauthCode', oauthCode);

    act(() => {
      jest.advanceTimersByTime(500);
    });

    await waitFor(() => expect(onSuccess).toHaveBeenCalledWith(oauthCode));
  });

  it('should correctly handle oidc based connection flows', async () => {
    Object.defineProperty(window, 'location', {
      value: {
        assign: jest.fn(),
      },
    });

    // Needs to be set prior to render, as inline flows will catch
    // the code on-mount
    window.localStorage.setItem('oauthCode', oauthCode);

    render(<Login />);

    const loginWithOidc = screen.getByText('Login with username/password');
    loginWithOidc.click();

    expect(window.location.assign).toHaveBeenCalledWith(
      [
        oidcConnectionOpts.authUrl,
        `?client_id=${oidcConnectionOpts.clientId}`,
        `&redirect_uri=${oidcConnectionOpts.redirect}?inline=true`,
        '&response_type=code',
        `&scope=${oidcConnectionOpts.scope}`,
        '&state=AAAAAAAAAAAAAAAAAAAA',
        `&connection=${googleConnectionOpts.connection}`,
        `&login_hint=${googleConnectionOpts.loginHint}`,
      ].join(''),
    );

    await waitFor(() => expect(onSuccess).toHaveBeenCalledWith(oauthCode));
  });

  it('should return error code if login fails', async () => {
    Object.defineProperty(window, 'location', {
      value: {
        assign: jest.fn(),
      },
    });

    const oauthError = 'Error!';
    window.localStorage.setItem('oauthError', oauthError);

    render(<Login />);

    const loginWithOidc = screen.getByText('Login with username/password');
    loginWithOidc.click();

    expect(window.location.assign).toHaveBeenCalledWith(
      [
        oidcConnectionOpts.authUrl,
        `?client_id=${oidcConnectionOpts.clientId}`,
        `&redirect_uri=${oidcConnectionOpts.redirect}?inline=true`,
        '&response_type=code',
        `&scope=${oidcConnectionOpts.scope}`,
        '&state=AAAAAAAAAAAAAAAAAAAA',
        `&connection=${googleConnectionOpts.connection}`,
        `&login_hint=${googleConnectionOpts.loginHint}`,
      ].join(''),
    );

    expect(await screen.findByText(`Error: ${oauthError}`)).toBeVisible();
  });

  it('should set succeeded and error flags on first render', () => {
    window.localStorage.setItem('oauthCode', 'code');
    window.localStorage.setItem('oauthError', 'error');

    render(<Login />);

    expect(screen.getByText('First render succeeded')).toBeVisible();
    expect(screen.getByText('First render error: error')).toBeVisible();
  });
});
