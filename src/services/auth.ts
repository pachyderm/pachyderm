import {ServiceArgs} from '../lib/types';
import {APIClient} from '../proto/auth/auth_grpc_pb';
import {AuthenticateRequest} from '../proto/auth/auth_pb';

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
