import {onError} from '@apollo/client/link/error';
import {captureException} from '@sentry/react';
import {useHistory} from 'react-router';

import logout from '@dash-frontend/lib/logout';

export const useErrorLink = () => {
  const browserHistory = useHistory();

  return onError(({graphQLErrors, networkError, operation}) => {
    if (graphQLErrors) {
      graphQLErrors.forEach(({message, locations, path}) => {
        if (message || locations || path) {
          captureException(
            `[GraphQL error]: Message: ${message}, Location: ${locations}, Path: ${path}`,
          );
        }
      });

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
    } else if (networkError) {
      captureException(`[Network error]: ${networkError}`);
    }
  });
};
