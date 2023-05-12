import {toProtoResourceType} from '@dash-backend/lib/gqlEnumMappers';

import {ServiceArgs} from '../lib/types';
import {APIClient} from '../proto/auth/auth_grpc_pb';
import {
  ActivateRequest,
  ActivateResponse,
  AuthenticateRequest,
  DeactivateRequest,
  DeactivateResponse,
  GetPermissionsRequest,
  GetPermissionsResponse,
  GetRoleBindingRequest,
  GetRoleBindingResponse,
  Resource,
  WhoAmIRequest,
  WhoAmIResponse,
} from '../proto/auth/auth_pb';

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
    activate: (rootToken?: string) => {
      return new Promise<ActivateResponse.AsObject>((resolve, reject) => {
        const request = new ActivateRequest();
        if (rootToken) request.setRootToken(rootToken);
        client.activate(request, credentialMetadata, (error, res) => {
          if (error) {
            return reject(error);
          } else {
            return resolve(res.toObject());
          }
        });
      });
    },
    deactivate: () => {
      return new Promise<DeactivateResponse.AsObject>((resolve, reject) => {
        client.deactivate(
          new DeactivateRequest(),
          credentialMetadata,
          (error, res) => {
            if (error) {
              return reject(error);
            } else {
              return resolve(res.toObject());
            }
          },
        );
      });
    },
    getRoleBinding: (args: GetRoleBindingRequest.AsObject) => {
      return new Promise<GetRoleBindingResponse.AsObject>((resolve, reject) => {
        const request = new GetRoleBindingRequest();
        request.setResource(
          new Resource()
            .setType(args.resource?.type || 0)
            .setName(args.resource?.name || ''),
        );
        client.getRoleBinding(request, credentialMetadata, (error, res) => {
          if (error) {
            return reject(error);
          } else {
            return resolve(res.toObject());
          }
        });
      });
    },
    getPermissions: (args: Required<GetPermissionsRequest.AsObject>) => {
      return new Promise<GetPermissionsResponse.AsObject>((resolve, reject) => {
        const request = new GetPermissionsRequest();
        const resource = new Resource().setName(args.resource.name);

        if (Number.isInteger(args?.resource?.type)) {
          resource.setType(args?.resource?.type);
        } else {
          resource.setType(toProtoResourceType(args.resource.type));
        }

        request.setResource(resource);
        client.getPermissions(request, credentialMetadata, (error, res) => {
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
