import {Empty} from 'google-protobuf/google/protobuf/empty_pb';

import {ServiceArgs} from '../lib/types';
import {APIClient} from '../proto/version/versionpb/version_grpc_pb';
import {Version} from '../proto/version/versionpb/version_pb';

const version = ({pachdAddress, channelCredentials}: ServiceArgs) => {
  const client = new APIClient(pachdAddress, channelCredentials);

  const versionService = {
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

  return versionService;
};

export default version;
