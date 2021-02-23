import {IAPIServer} from '@pachyderm/proto/pb/auth/auth_grpc_pb';
import {AuthenticateResponse} from '@pachyderm/proto/pb/auth/auth_pb';

const auth: Pick<IAPIServer, 'authenticate'> = {
  authenticate: (_, callback) => {
    callback(null, new AuthenticateResponse().setPachToken('xyz'));
  },
};

export default auth;
