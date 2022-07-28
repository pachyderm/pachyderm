import {GraphQLError} from 'graphql';
import {useMemo} from 'react';
import {useLocation} from 'react-router';

export enum ErrorViewType {
  NOT_FOUND = 'NOT_FOUND',
  RESOURCE_NOT_FOUND = 'RESOURCE_NOT_FOUND',
  UNAUTHENTICATED = 'UNAUTHENTICATED',
  GENERIC = 'GENERIC',
}

const useErrorView = (graphQLError?: GraphQLError) => {
  const {pathname} = useLocation();

  const errorType = useMemo(() => {
    if (pathname === '/not-found') {
      return ErrorViewType.NOT_FOUND;
    }

    if (graphQLError?.extensions.code === ErrorViewType.UNAUTHENTICATED) {
      return ErrorViewType.UNAUTHENTICATED;
    }

    if (graphQLError?.extensions.code === ErrorViewType.NOT_FOUND) {
      return ErrorViewType.RESOURCE_NOT_FOUND;
    }

    const errorType = ErrorViewType.GENERIC;

    return errorType;
  }, [graphQLError?.extensions.code, pathname]);

  const errorMessage = useMemo(() => {
    switch (errorType) {
      case ErrorViewType.UNAUTHENTICATED:
        return 'Unable to authenticate. Try again later.';
      case ErrorViewType.NOT_FOUND:
        return 'Elephants never forget, so this page must not exist.';
      case ErrorViewType.RESOURCE_NOT_FOUND:
        return 'Unable to locate this resource, are you sure it exists?';
      default:
        return `Looks like this API call can't be completed.`;
    }
  }, [errorType]);

  const errorDetails = graphQLError?.extensions?.exception?.stacktrace
    ? graphQLError?.extensions?.exception?.stacktrace[0]
    : String(graphQLError?.extensions?.details);

  return {errorType, errorMessage, errorDetails};
};

export default useErrorView;
