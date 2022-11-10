/* eslint-disable @typescript-eslint/no-explicit-any */

import {SUBSCRIPTION_INTERVAL} from '@dash-backend/constants/subscription';
import {generateIdTokenForAccount} from '@dash-backend/testHelpers';
import {Account} from '@graphqlTypes';
import {act} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import React, {ReactElement} from 'react';
import {BrowserRouter} from 'react-router-dom';

import ApolloProvider from '@dash-frontend/providers/ApolloProvider';

import {UrlState} from './hooks/useUrlQueryState';
import LoggedInProvider from './providers/LoggedInProvider';

export {default as server} from '@dash-backend/index';
export {default as mockServer} from '@dash-backend/mock';

// Two timeout intervals should ensure that new data
// is hydrated if render is performed between events
export const SUBSCRIPTION_TIMEOUT = SUBSCRIPTION_INTERVAL * 2;

export const withContextProviders = (
  Component: React.ElementType,
): ((props: any) => ReactElement) => {
  // eslint-disable-next-line react/display-name
  return (props: any): any => {
    return (
      <BrowserRouter>
        <LoggedInProvider>
          <ApolloProvider>
            <Component {...props} />
          </ApolloProvider>
        </LoggedInProvider>
      </BrowserRouter>
    );
  };
};

export const setIdTokenForAccount = (account: Account) => {
  window.localStorage.setItem('id-token', generateIdTokenForAccount(account));
};

export const click: typeof userEvent.click = async (...args) => {
  await act(() => userEvent.click(...args));
};

export const hover: typeof userEvent.hover = async (...args) => {
  await act(() => userEvent.hover(...args));
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

export const getUrlState = (): UrlState => {
  const searchParams = new URLSearchParams(window.location.search);
  return JSON.parse(atob(searchParams.get('view') || '{}'));
};

export const generateTutorialView = (tutorialId: string) => {
  return btoa(JSON.stringify({tutorialId}));
};
