import {ServiceArgs} from '../lib/types';
import {APIClient} from '../proto/admin/admin_grpc_pb';
import {ClusterInfo, InspectClusterRequest} from '../proto/admin/admin_pb';
import {grpcApiConstructorArgs} from '../utils/createGrpcApiClient';

let client: APIClient;

const admin = ({
  credentialMetadata,
}: Pick<ServiceArgs, 'credentialMetadata'>) => {
  client = client ?? new APIClient(...grpcApiConstructorArgs());

  return {
    inspectCluster: () => {
      return new Promise<ClusterInfo.AsObject>((resolve, reject) => {
        client.inspectCluster(
          new InspectClusterRequest(),
          credentialMetadata,
          (error, res) => {
            if (error) {
              return reject(error);
            }
            return resolve(res.toObject());
          },
        );
      });
    },
  };
};

export default admin;
