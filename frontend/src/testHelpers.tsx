/* eslint-disable @typescript-eslint/no-explicit-any */

import {QueryClient, QueryClientProvider} from '@tanstack/react-query';
import {act} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import React, {ReactElement} from 'react';
import {BrowserRouter} from 'react-router-dom';

import {getQueryClientConfig} from '@dash-frontend/providers/QueryClientProvider';

import {Account} from './api/auth';
import LoggedInProvider from './providers/LoggedInProvider';

// Two timeout intervals should ensure that new data
// is hydrated if render is performed between events
const SUBSCRIPTION_INTERVAL = 3000;
export const SUBSCRIPTION_TIMEOUT = SUBSCRIPTION_INTERVAL * 2;

export const withContextProviders = <
  T extends React.JSXElementConstructor<any> | keyof JSX.IntrinsicElements,
>(
  Component: React.ComponentType<React.ComponentProps<T>>,
): ((
  props: React.ComponentProps<T>,
) => ReactElement<React.ComponentProps<T>>) => {
  // eslint-disable-next-line react/display-name
  return (props: any): any => {
    const queryClient = new QueryClient(getQueryClientConfig());

    return (
      <BrowserRouter>
        <LoggedInProvider>
          <QueryClientProvider client={queryClient}>
            <Component {...props} />
          </QueryClientProvider>
        </LoggedInProvider>
      </BrowserRouter>
    );
  };
};

export const setIdTokenForAccount = (account: Account) => {
  window.localStorage.setItem(
    'id-token',
    `${account.id}-${account.email}-token`,
  );
};

export const loginUser = (
  id = 'id',
  email = 'user@email.com',
  authToken = '123',
) => {
  setIdTokenForAccount({id, email});
  window.localStorage.setItem('auth-token', authToken);
};

export const click: typeof userEvent.click = async (...args) => {
  await act(() => userEvent.click(...args));
};

export const hover: typeof userEvent.hover = async (...args) => {
  await act(() => userEvent.hover(...args));
};

export const unhover: typeof userEvent.unhover = async (...args) => {
  await act(() => userEvent.unhover(...args));
};

export const paste: typeof userEvent.paste = async (...args) => {
  await act(() => userEvent.paste(...args));
};

export const tab: typeof userEvent.tab = async (...args) => {
  await act(() => userEvent.tab(...args));
};

export const clear: typeof userEvent.clear = async (...args) => {
  await act(() => userEvent.clear(...args));
};

export const upload: typeof userEvent.upload = async (...args) => {
  await act(() => userEvent.upload(...args));
};

export const keyboard: typeof userEvent.keyboard = async (...args) => {
  return await act(() => userEvent.keyboard(...args));
};

// NOTE: The return type doesn't match userEvent.type,
// as our "type" command always returns a promise, whereas
// theirs only returns a promise if the input arguments contain a
// "delay". This has caused problems for us in the past,
// which is why we have a custom return type here.
export const type = async (element: Element, text: string) => {
  await act(async () => {
    await userEvent.type(element, text);
  });
};
