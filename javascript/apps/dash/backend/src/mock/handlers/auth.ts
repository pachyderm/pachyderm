import {ServiceError, status} from '@grpc/grpc-js';
import {IAPIServer} from '@pachyderm/proto/pb/auth/auth_grpc_pb';
import {AuthenticateResponse} from '@pachyderm/proto/pb/auth/auth_pb';

const defaultState: {error: ServiceError | null} = {
  error: null,
};

const auth = () => {
  let state = {...defaultState};

  return {
    getService: (): Pick<IAPIServer, 'authenticate'> => {
      return {
        authenticate: (_, callback) => {
          if (state.error?.code === status.UNAUTHENTICATED) {
            return callback(new Error('Invalid ID token'), null);
          }

          return callback(null, new AuthenticateResponse().setPachToken('xyz'));
        },
      };
    },
    setError: (error: ServiceError | null) => {
      state.error = error;
    },
    resetState: () => {
      state = {...defaultState};
    },
  };
};

export default auth();
