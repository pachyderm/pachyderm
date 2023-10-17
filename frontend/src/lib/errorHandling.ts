import {ServerError, ServerParseError, ApolloError} from '@apollo/client';

export const StatusCodes = [
  'OK',
  'CANCELLED',
  'UNKNOWN',
  'INVALID_ARGUMENT',
  'DEADLINE_EXCEEDED',
  'NOT_FOUND',
  'ALREADY_EXISTS',
  'PERMISSION_DENIED',
  'RESOURCE_EXHAUSTED',
  'FAILED_PRECONDITION',
  'ABORTED',
  'OUT_OF_RANGE',
  'UNIMPLEMENTED',
  'INTERNAL',
  'UNAVAILABLE',
  'DATA_LOSS',
  'UNAUTHENTICATED',
];

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

export const getGRPCCode = (error: ApolloError | undefined) => {
  if (
    error?.graphQLErrors.length &&
    error?.graphQLErrors.length > 0 &&
    error?.graphQLErrors[0].extensions &&
    typeof error?.graphQLErrors[0].extensions.code === 'string'
  ) {
    return error?.graphQLErrors[0].extensions.code;
  } else if (
    ifServerError(error?.networkError) &&
    typeof error?.networkError?.result !== 'string'
  ) {
    try {
      return StatusCodes[
        parseInt(error?.networkError?.result?.errors[0].extensions.grpcCode)
      ];
    } catch {
      return undefined;
    }
  } else return undefined;
};

export default getServerErrorMessage;
