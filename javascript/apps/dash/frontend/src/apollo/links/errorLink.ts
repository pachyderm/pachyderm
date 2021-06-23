import {onError} from '@apollo/client/link/error';
import {History as BrowserHistory} from 'history';

import logout from '@dash-frontend/lib/logout';

export const errorLink = (browserHistory: BrowserHistory) =>
  onError(({graphQLErrors, operation}) => {
    if (graphQLErrors) {
      if (
        graphQLErrors.some(
          (error) => error.extensions?.code === 'UNAUTHENTICATED',
        ) &&
        operation.operationName !== 'exchangeCode'
      ) {
        logout();
      }

      if (
        graphQLErrors.some((error) => error.extensions?.code === 'NOT_FOUND') &&
        operation.operationName !== 'job'
      ) {
        browserHistory.push('/not-found');
      }

      // How do we want to handle other error codes?
      if (
        graphQLErrors.some(
          (error) => error.extensions?.code === 'INTERNAL_SERVER_ERROR',
        )
      ) {
        browserHistory.push('/error');
      }
    }
  });
