import {useMemo} from 'react';
import {useLocation} from 'react-router';

import {
  isErrorWithDetails,
  isErrorWithMessage,
} from '@dash-frontend/api/utils/error';

export enum ErrorViewType {
  NOT_FOUND = 'NOT_FOUND',
  RESOURCE_NOT_FOUND = 'RESOURCE_NOT_FOUND',
  UNAUTHENTICATED = 'UNAUTHENTICATED',
  ISSUER_ERROR = 'ISSUER_ERROR',
  GENERIC = 'GENERIC',
}

const useErrorView = (error?: unknown) => {
  const {pathname} = useLocation();

  const errorType = useMemo(() => {
    if (pathname === '/not-found') {
      return ErrorViewType.NOT_FOUND;
    }

    if (isErrorWithMessage(error)) {
      if (/authentication error/i.test(error.message)) {
        return ErrorViewType.UNAUTHENTICATED;
      }

      if (/not found/.test(error.message)) {
        return ErrorViewType.RESOURCE_NOT_FOUND;
      }
    }

    const errorType = ErrorViewType.GENERIC;

    return errorType;
  }, [error, pathname]);

  const errorMessage = useMemo(() => {
    switch (errorType) {
      case ErrorViewType.UNAUTHENTICATED:
        return (
          (isErrorWithMessage(error) && error.message) ||
          'Unable to authenticate. Try again later.'
        );
      case ErrorViewType.NOT_FOUND:
        return 'Elephants never forget, so this page must not exist.';
      case ErrorViewType.RESOURCE_NOT_FOUND:
        return 'Unable to locate this resource, are you sure it exists?';
      default:
        return `Looks like this API call can't be completed.`;
    }
  }, [errorType, error]);

  const errorDetails = isErrorWithDetails(error)
    ? error.details.reduce((acc, curr) => (acc = acc + '\n' + curr), '')
    : undefined;

  return {errorType, errorMessage, errorDetails};
};

export default useErrorView;
