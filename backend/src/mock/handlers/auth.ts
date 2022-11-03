import {status} from '@grpc/grpc-js';
import {Status} from '@grpc/grpc-js/build/src/constants';

import {getAccountFromIdToken} from '@dash-backend/lib/auth';
import {AuthIAPIServer, AuthenticateResponse} from '@dash-backend/proto';
import {createServiceError} from '@dash-backend/testHelpers';

import MockState from './MockState';

const auth = () => {
  return {
    getService: (): Pick<AuthIAPIServer, 'authenticate' | 'whoAmI'> => {
      return {
        authenticate: async ({request}, callback) => {
          if (MockState.state.error?.code === status.UNAUTHENTICATED) {
            return callback(new Error('Invalid ID token'), null);
          }

          const {id} = await getAccountFromIdToken(request.getIdToken());

          return callback(null, new AuthenticateResponse().setPachToken(id));
        },
        whoAmI: async (_request, callback) => {
          return callback(createServiceError({code: Status.UNAUTHENTICATED}));
        },
      };
    },
  };
};

export default auth();
