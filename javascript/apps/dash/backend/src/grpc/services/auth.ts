import {APIClient} from '@pachyderm/proto/pb/auth/auth_grpc_pb';
import {AuthenticateRequest} from '@pachyderm/proto/pb/auth/auth_pb';

import {ServiceArgs} from '@dash-backend/lib/types';

const auth = ({pachdAddress, channelCredentials, log}: ServiceArgs) => {
  const client = new APIClient(pachdAddress, channelCredentials);

  return {
    authenticate: (idToken: string) => {
      return new Promise<string>((resolve, reject) => {
        log.info('authenticate request');

        client.authenticate(
          new AuthenticateRequest().setIdToken(idToken),
          (error, res) => {
            if (error) {
              log.error({error: error.message}, 'authenticate request failed');
              return reject(error);
            }

            log.info('authenticate request succeeded');
            return resolve(res.toObject().pachToken);
          },
        );
      });
    },
  };
};

export default auth;
