import {ServiceArgs} from '../lib/types';
import {APIClient} from '../proto/auth/auth_grpc_pb';
import {AuthenticateRequest, WhoAmIRequest} from '../proto/auth/auth_pb';

const auth = ({
  pachdAddress,
  channelCredentials,
  credentialMetadata,
}: ServiceArgs) => {
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
    whoAmI: () => {
      return new Promise((resolve, reject) => {
        client.whoAmI(new WhoAmIRequest(), credentialMetadata, (error, res) => {
          if (error) {
            return reject(error);
          } else {
            return resolve(res.toObject());
          }
        });
      });
    },
  };
};

export default auth;
