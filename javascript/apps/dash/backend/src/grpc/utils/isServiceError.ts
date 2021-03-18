import {ServiceError} from '@grpc/grpc-js';

// eslint-disable-next-line @typescript-eslint/no-explicit-any
const isServiceError = (object: any): object is ServiceError => {
  return 'code' in object && 'details' in object;
};

export default isServiceError;
