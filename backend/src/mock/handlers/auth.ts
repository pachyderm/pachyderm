import {status} from '@grpc/grpc-js';
import {AuthIAPIServer, AuthenticateResponse} from '@pachyderm/node-pachyderm';

import {getAccountFromIdToken} from '@dash-backend/lib/auth';

import MockState from './MockState';

const auth = () => {
  return {
    getService: (): Pick<AuthIAPIServer, 'authenticate'> => {
      return {
        authenticate: async ({request}, callback) => {
          if (MockState.state.error?.code === status.UNAUTHENTICATED) {
            return callback(new Error('Invalid ID token'), null);
          }

          const {id} = await getAccountFromIdToken(request.getIdToken());

          return callback(null, new AuthenticateResponse().setPachToken(id));
        },
      };
    },
  };
};

export default auth();
