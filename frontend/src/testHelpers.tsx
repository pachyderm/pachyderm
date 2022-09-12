/* eslint-disable @typescript-eslint/no-explicit-any */

import {SUBSCRIPTION_INTERVAL} from '@dash-backend/constants/subscription';
import {generateIdTokenForAccount} from '@dash-backend/testHelpers';
import {Account} from '@graphqlTypes';
import {act} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {typeOptions} from '@testing-library/user-event/dist/type/typeImplementation';
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

export const click: typeof userEvent.click = (...args) => {
  act(() => userEvent.click(...args));
};

export const hover: typeof userEvent.hover = (...args) => {
  act(() => userEvent.hover(...args));
};

export const paste: typeof userEvent.paste = (...args) => {
  act(() => userEvent.paste(...args));
};

export const tab: typeof userEvent.tab = (...args) => {
  act(() => userEvent.tab(...args));
};

// NOTE: The return type doesn't match userEvent.type,
// as our "type" command always returns a promise, whereas
// theirs only returns a promise if the input arguments contain a
// "delay". This has caused problems for us in the past,
// which is why we have a custom return type here.
export const type = async (
  element: Element,
  text: string,
  userOpts?: typeOptions & {delay?: 0 | undefined},
) => {
  await act(async () => {
    await userEvent.type(element, text, userOpts);
  });
};

export const getUrlState = (): UrlState => {
  const searchParams = new URLSearchParams(window.location.search);
  return JSON.parse(atob(searchParams.get('view') || '{}'));
};

export const generateTutorialView = (tutorialId: string) => {
  return btoa(JSON.stringify({tutorialId}));
};
