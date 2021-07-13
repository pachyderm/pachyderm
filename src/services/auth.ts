import {APIClient} from '@pachyderm/proto/pb/auth/auth_grpc_pb';
import {AuthenticateRequest} from '@pachyderm/proto/pb/auth/auth_pb';

import {ServiceArgs} from '../lib/types';

const auth = ({pachdAddress, channelCredentials}: ServiceArgs) => {
  const client = new APIClient(pachdAddress, channelCredentials);

  return {
    authenticate: (idToken: string) => {
      return new Promise<string>((resolve, reject) => {
        client.authenticate(
          new AuthenticateRequest().setIdToken(idToken),
          (error, res) => {
            if (error) {
              return reject(error);
            }

            return resolve(res.toObject().pachToken);
          },
        );
      });
    },
  };
};

export default auth;
