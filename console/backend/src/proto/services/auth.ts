import {ServiceArgs} from '../lib/types';
import {APIClient} from '../proto/auth/auth_grpc_pb';
import {WhoAmIRequest, WhoAmIResponse} from '../proto/auth/auth_pb';
import {grpcApiConstructorArgs} from '../utils/createGrpcApiClient';

let client: APIClient;

const authServiceRpcHandler = ({
  credentialMetadata,
}: Pick<ServiceArgs, 'credentialMetadata'>) => {
  client = client ?? new APIClient(...grpcApiConstructorArgs());

  return {
    whoAmI: () => {
      return new Promise<WhoAmIResponse.AsObject>((resolve, reject) => {
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

export default authServiceRpcHandler;
