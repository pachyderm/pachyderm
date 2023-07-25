import {Empty} from 'google-protobuf/google/protobuf/empty_pb';

import {APIClient} from '../proto/version/versionpb/version_grpc_pb';
import {Version} from '../proto/version/versionpb/version_pb';
import {grpcApiConstructorArgs} from '../utils/createGrpcApiClient';

let client: APIClient;

const version = () => {
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

export default version;
