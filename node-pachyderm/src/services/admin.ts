import {Empty} from 'google-protobuf/google/protobuf/empty_pb';

import {ServiceArgs} from '../lib/types';
import {APIClient} from '../proto/admin/admin_grpc_pb';
import {ClusterInfo} from '../proto/admin/admin_pb';

const admin = ({
  pachdAddress,
  channelCredentials,
  credentialMetadata,
}: ServiceArgs) => {
  const client = new APIClient(pachdAddress, channelCredentials);

  const adminService = {
    inspectCluster: () => {
      return new Promise<ClusterInfo.AsObject>((resolve, reject) => {
        client.inspectCluster(new Empty(), credentialMetadata, (error, res) => {
          if (error) {
            return reject(error);
          }
          return resolve(res.toObject());
        });
      });
    },
  };

  return adminService;
};

export default admin;
