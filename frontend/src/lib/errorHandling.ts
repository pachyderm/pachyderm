import {ServerError, ServerParseError, ApolloError} from '@apollo/client';

const ifServerError = (
  error: Error | ServerError | ServerParseError | null | undefined,
): error is ServerError => true;

const getServerErrorMessage = (error: ApolloError | undefined) => {
  if (
    ifServerError(error?.networkError) &&
    typeof error?.networkError?.result === 'string'
  ) {
    return error.networkError.result;
  } else if (
    ifServerError(error?.networkError) &&
    typeof error?.networkError?.result !== 'string' &&
    error?.networkError?.result?.errors[0].extensions.details
  ) {
    return error?.networkError.result?.errors[0].extensions.details;
  }

  return error?.message;
};

export default getServerErrorMessage;
