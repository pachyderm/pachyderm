import {IAPIServer} from '@pachyderm/proto/pb/auth/auth_grpc_pb';
import {AuthenticateResponse} from '@pachyderm/proto/pb/auth/auth_pb';

const defaultState = {
  hasInvalidIdToken: false,
};

const auth = () => {
  let state = {...defaultState};

  return {
    getService: (): Pick<IAPIServer, 'authenticate'> => {
      return {
        authenticate: (_, callback) => {
          if (state.hasInvalidIdToken) {
            return callback(new Error('Invalid ID token'), null);
          }

          return callback(null, new AuthenticateResponse().setPachToken('xyz'));
        },
      };
    },
    setHasInvalidIdToken: (hasInvalidToken: boolean) => {
      state.hasInvalidIdToken = hasInvalidToken;
    },
    resetState: () => {
      state = {...defaultState};
    },
  };
};

export default auth();
