import {captureException} from '@sentry/react';
import {
  QueryClient,
  QueryClientProvider as ReactQueryClientProvider,
  QueryClientConfig,
  QueryCache,
  MutationCache,
  Query,
  Mutation,
} from '@tanstack/react-query';
import {ReactQueryDevtools} from '@tanstack/react-query-devtools';
import React from 'react';

import {
  constructMessageFromError,
  isAuthDisabled,
  isAuthExpired,
} from '@dash-frontend/api/utils/error';
import {DEFAULT_POLLING_INTERVAL_MS} from '@dash-frontend/constants/pollIntervals';
import logout from '@dash-frontend/lib/logout';

const onQueryError = (
  error: unknown,
  query: Query<unknown, unknown, unknown>,
) => {
  if (isAuthExpired(error)) {
    return logout();
  }

  if (isAuthDisabled(error)) {
    return;
  }

  const message = constructMessageFromError(error);
  const hash = query.queryHash;

  captureException(
    `[Query Error]: ${message}${hash && `  - QueryHash: ${hash}`}`,
  );
};

const onMutationError = (
  error: unknown,
  _variables: unknown,
  _context: unknown,
  mutation: Mutation<unknown, unknown, unknown>,
) => {
  if (isAuthExpired(error)) {
    return logout();
  }

  if (isAuthDisabled(error)) {
    return;
  }

  const message = constructMessageFromError(error);
  const key = mutation.options.mutationKey;

  captureException(
    `[Mutation Error]: ${message}${key && `  - MutationKey: ${key}`}`,
  );
};

export const getQueryClientConfig = (): QueryClientConfig => {
  return {
    defaultOptions: {
      queries: {
        refetchInterval: DEFAULT_POLLING_INTERVAL_MS, // TODO: come up with a refetchInterval that makes sense
        refetchOnWindowFocus: false, // TODO: figure out if we want this enabled, probably yes
        retry: false,
        throwOnError: true,
      },
    },
    queryCache: new QueryCache({
      onError: onQueryError,
    }),
    mutationCache: new MutationCache({
      onError: onMutationError,
    }),
  };
};

const client = new QueryClient(getQueryClientConfig());

const QueryClientProvider = ({children}: {children?: React.ReactNode}) => {
  return (
    <ReactQueryClientProvider client={client}>
      {children}
      <ReactQueryDevtools initialIsOpen={false} buttonPosition="bottom-left" />
    </ReactQueryClientProvider>
  );
};

export default QueryClientProvider;
