import {Empty} from 'google-protobuf/google/protobuf/empty_pb';

import {ServiceArgs} from '../lib/types';
import {APIClient} from '../proto/enterprise/enterprise_grpc_pb';
import {
  ActivateRequest,
  GetStateRequest,
  GetStateResponse,
  ActivateResponse,
  DeactivateRequest,
  GetActivationCodeRequest,
  GetActivationCodeResponse,
} from '../proto/enterprise/enterprise_pb';
import {grpcApiConstructorArgs} from '../utils/createGrpcApiClient';

let client: APIClient;

const enterpriseServiceRpcHandler = ({
  credentialMetadata,
}: Pick<ServiceArgs, 'credentialMetadata'>) => {
  client = client ?? new APIClient(...grpcApiConstructorArgs());

  return {
    getState: () => {
      return new Promise<GetStateResponse.AsObject>((resolve, reject) => {
        client.getState(
          new GetStateRequest(),
          credentialMetadata,
          (err, res) => {
            if (err) {
              reject(err);
            } else {
              resolve(res.toObject());
            }
          },
        );
      });
    },
    activate: (licnseServer: string, id: string, secret: string) => {
      return new Promise<ActivateResponse.AsObject>((resolve, reject) => {
        const request = new ActivateRequest();
        request.setSecret(secret).setLicenseServer(licnseServer).setId(id);
        client.activate(request, credentialMetadata, (err, res) => {
          if (err) {
            reject(err);
          } else {
            resolve(res.toObject());
          }
        });
      });
    },
    deactivate: () => {
      return new Promise<Empty.AsObject>((resolve, reject) => {
        client.deactivate(
          new DeactivateRequest(),
          credentialMetadata,
          (err, _res) => {
            if (err) {
              reject(err);
            } else {
              resolve({});
            }
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
              if (err) {
                reject(err);
              } else {
                resolve(res.toObject());
              }
            },
          );
        },
      );
    },
  };
};

export default enterpriseServiceRpcHandler;
