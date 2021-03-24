/* eslint-disable @typescript-eslint/no-explicit-any */
import React, {ReactElement} from 'react';
import {BrowserRouter} from 'react-router-dom';

import {generateIdTokenForAccount} from '@dash-backend/testHelpers';
import ApolloProvider from '@dash-frontend/providers/ApolloProvider';
import {Account} from '@graphqlTypes';

export {default as server} from '@dash-backend/index';
export {default as mockServer} from '@dash-backend/mock';

export const withContextProviders = (
  Component: React.ElementType,
): ((props: any) => ReactElement) => {
  // eslint-disable-next-line react/display-name
  return (props: any): any => {
    return (
      <BrowserRouter>
        <ApolloProvider>
          <Component {...props} />
        </ApolloProvider>
      </BrowserRouter>
    );
  };
};

export const setIdTokenForAccount = (account: Account) => {
  window.localStorage.setItem('id-token', generateIdTokenForAccount(account));
};
