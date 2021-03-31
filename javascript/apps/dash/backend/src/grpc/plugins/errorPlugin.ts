import {Status} from '@grpc/grpc-js/build/src/constants';
import {AuthenticationError, ApolloError} from 'apollo-server-errors';

import {GRPCPlugin, NotFoundError} from '@dash-backend/lib/types';

import isServiceError from '../utils/isServiceError';

const errorPlugin: GRPCPlugin = {
  onError: ({error, requestName}) => {
    // grpc has a set of "canonical" error codes that we use in pachd,
    // which are defined on the "Status" enum
    // https://github.com/googleapis/googleapis/blob/master/google/rpc/code.proto
    if (isServiceError(error)) {
      if (error.code === Status.UNAUTHENTICATED) {
        throw new AuthenticationError(error.details);
      }

      if (error.code === Status.NOT_FOUND) {
        throw new NotFoundError(error.details);
      }

      // We can transform additional error types below.
      // Unhandled errors will be returned as an INTERNAL_SERVER_ERROR
      throw new ApolloError(error.details, '500', {
        ...error,
        grpcCode: error.code,
      });
    }

    throw new ApolloError(`Something went wrong requesting ${requestName}`);
  },
};

export default errorPlugin;
