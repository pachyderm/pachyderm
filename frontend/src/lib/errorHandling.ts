import {ServerError, ServerParseError, ApolloError} from '@apollo/client';

const ifServerError = (
  error: Error | ServerError | ServerParseError | null | undefined,
): error is ServerError => true;

const getServerErrorMessage = (error: ApolloError | undefined) => {
  return ifServerError(error?.networkError)
    ? error?.networkError.result.errors[0].extensions.details
    : error?.message;
};

export default getServerErrorMessage;
