import {Empty} from 'google-protobuf/google/protobuf/empty_pb';

import {APIClient} from '../proto/version/versionpb/version_grpc_pb';
import {Version} from '../proto/version/versionpb/version_pb';
import {grpcApiConstructorArgs} from '../utils/createGrpcApiClient';

let client: APIClient;

/**
 * 1. Creates GRPC API Client. The connection is lazily created. The first RPC
 *    that fires will be cold.
 * 2. Builds RPC protobuff request based on the arguments to the function
 * 3. Fires off the RPC.
 */
const versionServiceRpcHandler = () => {
  client = client ?? new APIClient(...grpcApiConstructorArgs());

  return {
    getVersion: () => {
      return new Promise<Version.AsObject>((resolve, reject) => {
        client.getVersion(new Empty(), (error, res) => {
          if (error) {
            return reject(error);
          }
          return resolve(res.toObject());
        });
      });
    },
  };
};

export default versionServiceRpcHandler;
