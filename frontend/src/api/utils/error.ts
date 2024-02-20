import {ResponseComposition, RestContext} from 'msw';

import {CODES} from '../googleTypes';

/**
 * Detects an error thrown by triggering an AbortController signal
 */

export const isAbortSignalError = (e: Error | unknown) =>
  e instanceof DOMException && e?.name === 'AbortError';

// Request Error checking functions
export type StreamingRequestError = {
  error: RequestError;
};

export const streamingError = (
  res: ResponseComposition<StreamingRequestError>,
  ctx: RestContext,
  json: RequestError,
  statusCode = 400,
) =>
  res(
    ctx.status(statusCode),
    ctx.json({
      error: json,
    }),
  );

// This is used for a non-stream error
export type RequestError = {
  code?: number;
  details?: string[];
  message?: string;
};
export const isErrorWithCode = (
  error: unknown,
): error is Required<Pick<RequestError, 'code'>> => {
  return error !== null && typeof error === 'object' && 'code' in error;
};

export const isErrorWithMessage = (
  error: unknown,
): error is Required<Pick<RequestError, 'message'>> => {
  return error !== null && typeof error === 'object' && 'message' in error;
};

export const isErrorWithDetails = (
  error: unknown,
): error is Required<Pick<RequestError, 'details'>> => {
  return error !== null && typeof error === 'object' && 'details' in error;
};

export const isUnknown = (error: unknown) => {
  return isErrorWithCode(error) && error.code === CODES.Unknown;
};

export const isAuthDisabled = (error: unknown) => {
  return (
    isNotConnected(error) ||
    (isErrorWithCode(error) && error.code === CODES.Unimplemented)
  );
};

export const isAuthExpired = (error: unknown) => {
  return (
    isErrorWithCode(error) &&
    // token expiration in the past uses an Internal code while token corrupted uses an Unauthenticated code
    (error.code === CODES.Unauthenticated || error.code === CODES.Internal)
  );
};

export const isPermissionDenied = (error: unknown) => {
  return isErrorWithCode(error) && error.code === CODES.PermissionDenied;
};

export const isNotFound = (error: unknown) => {
  return isErrorWithCode(error) && error.code === CODES.NotFound;
};

export const isNotConnected = (error: unknown) => {
  return (
    isErrorWithMessage(error) &&
    (error.message.includes('no healthy upstream') ||
      error.message.includes('upstream connect error'))
  );
};

export const constructMessageFromError = (error: unknown) => {
  let message = '';
  if (
    isErrorWithCode(error) ||
    isErrorWithDetails(error) ||
    isErrorWithMessage(error)
  ) {
    message = `
    CODE: ${isErrorWithCode(error) && error?.code}, MESSAGE: ${
      isErrorWithMessage(error) && error?.message
    }, DETAILS: ${isErrorWithDetails(error) && error?.details}`;
  } else {
    message = error as string;
  }
  return message;
};
