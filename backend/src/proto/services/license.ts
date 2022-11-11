import {Timestamp} from 'google-protobuf/google/protobuf/timestamp_pb';

import {ServiceArgs} from '../lib/types';
import {APIClient} from '../proto/license/license_grpc_pb';
import {
  ActivateRequest,
  ActivateResponse,
  AddClusterRequest,
  AddClusterResponse,
  DeleteAllRequest,
  DeleteAllResponse,
  DeleteClusterRequest,
  DeleteClusterResponse,
  GetActivationCodeRequest,
  GetActivationCodeResponse,
  ListClustersRequest,
  ListClustersResponse,
  ListUserClustersRequest,
  ListUserClustersResponse,
  UpdateClusterRequest,
  UpdateClusterResponse,
} from '../proto/license/license_pb';

const license = ({
  pachdAddress,
  channelCredentials,
  credentialMetadata,
}: ServiceArgs) => {
  const client = new APIClient(pachdAddress, channelCredentials);
  return {
    activateLicense: (activationCode: string, expires?: Timestamp) => {
      return new Promise<ActivateResponse.AsObject>((resolve, reject) => {
        const request = new ActivateRequest()
          .setActivationCode(activationCode)
          .setExpires(expires);
        client.activate(request, credentialMetadata, (err, res) => {
          if (err) {
            reject(err);
          } else {
            resolve(res.toObject());
          }
        });
      });
    },
    addCluster: (
      id: string,
      address: string,
      secret: string,
      clusterID = '',
      userAddress = '',
      entepriseServer = false,
    ) => {
      return new Promise<AddClusterResponse.AsObject>((resolve, reject) => {
        const request = new AddClusterRequest()
          .setId(id)
          .setAddress(address)
          .setSecret(secret)
          .setUserAddress(userAddress)
          .setClusterDeploymentId(clusterID)
          .setEnterpriseServer(entepriseServer);

        client.addCluster(request, credentialMetadata, (err, res) => {
          if (err) reject(err);
          else {
            resolve(res.toObject());
          }
        });
      });
    },
    updateCluster: (
      id: string,
      address: string,
      userAddress = '',
      clusterID = '',
    ) => {
      return new Promise<UpdateClusterResponse.AsObject>((resolve, reject) => {
        const request = new UpdateClusterRequest()
          .setId(id)
          .setAddress(address)
          .setUserAddress(userAddress)
          .setClusterDeploymentId(clusterID);

        client.updateCluster(request, credentialMetadata, (err, res) => {
          if (err) reject(err);
          resolve(res.toObject());
        });
      });
    },
    deleteCluster: (id: string) => {
      return new Promise<DeleteClusterResponse.AsObject>((resolve, reject) => {
        const request = new DeleteClusterRequest().setId(id);

        client.deleteCluster(request, credentialMetadata, (err, res) => {
          if (err) reject(err);
          else resolve(res.toObject());
        });
      });
    },
    listClusters: () => {
      return new Promise<ListClustersResponse.AsObject>((resolve, reject) => {
        client.listClusters(
          new ListClustersRequest(),
          credentialMetadata,
          (err, res) => {
            if (err) reject(err);
            else resolve(res.toObject());
          },
        );
      });
    },
    getActivationCode: () => {
      return new Promise<GetActivationCodeResponse.AsObject>(
        (resolve, reject) => {
          client.getActivationCode(
            new GetActivationCodeRequest(),
            credentialMetadata,
            (err, res) => {
              if (err) reject(err);
              else resolve(res.toObject());
            },
          );
        },
      );
    },
    deleteAll: () => {
      return new Promise<DeleteAllResponse.AsObject>((resolve, reject) => {
        client.deleteAll(
          new DeleteAllRequest(),
          credentialMetadata,
          (err, res) => {
            if (err) reject(err);
            else resolve(res.toObject());
          },
        );
      });
    },
    listUserClusters: () => {
      return new Promise<ListUserClustersResponse.AsObject>(
        (resolve, reject) => {
          client.listUserClusters(
            new ListUserClustersRequest(),
            credentialMetadata,
            (err, res) => {
              if (err) reject(err);
              else resolve(res.toObject());
            },
          );
        },
      );
    },
  };
};

export default license;
