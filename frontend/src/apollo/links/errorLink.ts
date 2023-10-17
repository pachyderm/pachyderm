import {onError} from '@apollo/client/link/error';
import {captureException} from '@sentry/react';
import {GraphQLError} from 'graphql';

import logout from '@dash-frontend/lib/logout';

export const useErrorLink = (
  setApolloError: React.Dispatch<
    React.SetStateAction<GraphQLError | undefined>
  >,
) => {
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
        !['exchangeCode', 'authConfig'].includes(operation.operationName)
      ) {
        logout();
      }

      if (
        graphQLErrors.some(
          (error) =>
            error.extensions?.code === 'NOT_FOUND' &&
            !['createPipelineV2', 'job'].some((p) => error?.path?.includes(p)),
        )
      ) {
        setApolloError(graphQLErrors[0]);
      }

      if (
        graphQLErrors.some(
          (error) =>
            error.extensions?.code === 'INTERNAL_SERVER_ERROR' &&
            (!error.path || error.path?.includes('project')),
        )
      ) {
        setApolloError(graphQLErrors[0]);
      }
    } else if (networkError) {
      captureException(`[Network error]: ${networkError}`);
    }
  });
};
