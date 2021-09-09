import {ServiceError, status} from '@grpc/grpc-js';
import {AuthIAPIServer, AuthenticateResponse} from '@pachyderm/node-pachyderm';

import {getAccountFromIdToken} from '@dash-backend/lib/auth';

const defaultState: {error: ServiceError | null} = {
  error: null,
};

const auth = () => {
  let state = {...defaultState};

  return {
    getService: (): Pick<AuthIAPIServer, 'authenticate'> => {
      return {
        authenticate: async ({request}, callback) => {
          if (state.error?.code === status.UNAUTHENTICATED) {
            return callback(new Error('Invalid ID token'), null);
          }

          const {id} = await getAccountFromIdToken(request.getIdToken());

          return callback(null, new AuthenticateResponse().setPachToken(id));
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
