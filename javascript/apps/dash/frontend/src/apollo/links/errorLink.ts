import {onError} from '@apollo/client/link/error';

import logout from '@dash-frontend/lib/logout';

export const errorLink = () =>
  onError(({graphQLErrors, operation}) => {
    if (
      graphQLErrors &&
      graphQLErrors.some(
        (error) => error.extensions?.code === 'UNAUTHENTICATED',
      ) &&
      operation.operationName !== 'exchangeCode'
    ) {
      logout();
    }
  });
