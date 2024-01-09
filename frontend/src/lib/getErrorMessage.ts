const getErrorMessage = (error: unknown) => {
  if (!error) {
    return;
  }

  if (typeof error === 'object' && 'message' in error) {
    return error.message as string;
  }

  if (error instanceof Error) {
    return error.message as string;
  }

  return String(error);
};

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

export const getGRPCCode = (error: unknown) => {
  if (!error) {
    return;
  }

  if (typeof error === 'object' && 'code' in error) {
    return StatusCodes[parseInt(error.code as string)];
  }

  return String(error);
};

export default getErrorMessage;
