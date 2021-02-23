import {ChannelCredentials} from '@grpc/grpc-js';
import {APIClient} from '@pachyderm/proto/pb/auth/auth_grpc_pb';
import {AuthenticateRequest} from '@pachyderm/proto/pb/auth/auth_pb';

const auth = (pachdAddress: string, channelCredentials: ChannelCredentials) => {
  const client = new APIClient(pachdAddress, channelCredentials);

  return {
    authenticate: (idToken: string) => {
      return new Promise<string>((resolve, reject) => {
        client.authenticate(
          new AuthenticateRequest().setIdToken(idToken),
          (err, res) => {
            if (err) return reject(err);

            return resolve(res.toObject().pachToken);
          },
        );
      });
    },
  };
};

export default auth;
