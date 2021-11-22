// GENERATED CODE -- DO NOT EDIT!

'use strict';
var grpc = require('@grpc/grpc-js');
var auth_auth_pb = require('../auth/auth_pb.js');
var gogoproto_gogo_pb = require('../gogoproto/gogo_pb.js');
var google_protobuf_timestamp_pb = require('google-protobuf/google/protobuf/timestamp_pb.js');

function serialize_auth_v2_ActivateRequest(arg) {
  if (!(arg instanceof auth_auth_pb.ActivateRequest)) {
    throw new Error('Expected argument of type auth_v2.ActivateRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v2_ActivateRequest(buffer_arg) {
  return auth_auth_pb.ActivateRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v2_ActivateResponse(arg) {
  if (!(arg instanceof auth_auth_pb.ActivateResponse)) {
    throw new Error('Expected argument of type auth_v2.ActivateResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v2_ActivateResponse(buffer_arg) {
  return auth_auth_pb.ActivateResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v2_AuthenticateRequest(arg) {
  if (!(arg instanceof auth_auth_pb.AuthenticateRequest)) {
    throw new Error('Expected argument of type auth_v2.AuthenticateRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v2_AuthenticateRequest(buffer_arg) {
  return auth_auth_pb.AuthenticateRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v2_AuthenticateResponse(arg) {
  if (!(arg instanceof auth_auth_pb.AuthenticateResponse)) {
    throw new Error('Expected argument of type auth_v2.AuthenticateResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v2_AuthenticateResponse(buffer_arg) {
  return auth_auth_pb.AuthenticateResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v2_AuthorizeRequest(arg) {
  if (!(arg instanceof auth_auth_pb.AuthorizeRequest)) {
    throw new Error('Expected argument of type auth_v2.AuthorizeRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v2_AuthorizeRequest(buffer_arg) {
  return auth_auth_pb.AuthorizeRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v2_AuthorizeResponse(arg) {
  if (!(arg instanceof auth_auth_pb.AuthorizeResponse)) {
    throw new Error('Expected argument of type auth_v2.AuthorizeResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v2_AuthorizeResponse(buffer_arg) {
  return auth_auth_pb.AuthorizeResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v2_DeactivateRequest(arg) {
  if (!(arg instanceof auth_auth_pb.DeactivateRequest)) {
    throw new Error('Expected argument of type auth_v2.DeactivateRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v2_DeactivateRequest(buffer_arg) {
  return auth_auth_pb.DeactivateRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v2_DeactivateResponse(arg) {
  if (!(arg instanceof auth_auth_pb.DeactivateResponse)) {
    throw new Error('Expected argument of type auth_v2.DeactivateResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v2_DeactivateResponse(buffer_arg) {
  return auth_auth_pb.DeactivateResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v2_DeleteExpiredAuthTokensRequest(arg) {
  if (!(arg instanceof auth_auth_pb.DeleteExpiredAuthTokensRequest)) {
    throw new Error('Expected argument of type auth_v2.DeleteExpiredAuthTokensRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v2_DeleteExpiredAuthTokensRequest(buffer_arg) {
  return auth_auth_pb.DeleteExpiredAuthTokensRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v2_DeleteExpiredAuthTokensResponse(arg) {
  if (!(arg instanceof auth_auth_pb.DeleteExpiredAuthTokensResponse)) {
    throw new Error('Expected argument of type auth_v2.DeleteExpiredAuthTokensResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v2_DeleteExpiredAuthTokensResponse(buffer_arg) {
  return auth_auth_pb.DeleteExpiredAuthTokensResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v2_ExtractAuthTokensRequest(arg) {
  if (!(arg instanceof auth_auth_pb.ExtractAuthTokensRequest)) {
    throw new Error('Expected argument of type auth_v2.ExtractAuthTokensRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v2_ExtractAuthTokensRequest(buffer_arg) {
  return auth_auth_pb.ExtractAuthTokensRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v2_ExtractAuthTokensResponse(arg) {
  if (!(arg instanceof auth_auth_pb.ExtractAuthTokensResponse)) {
    throw new Error('Expected argument of type auth_v2.ExtractAuthTokensResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v2_ExtractAuthTokensResponse(buffer_arg) {
  return auth_auth_pb.ExtractAuthTokensResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v2_GetConfigurationRequest(arg) {
  if (!(arg instanceof auth_auth_pb.GetConfigurationRequest)) {
    throw new Error('Expected argument of type auth_v2.GetConfigurationRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v2_GetConfigurationRequest(buffer_arg) {
  return auth_auth_pb.GetConfigurationRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v2_GetConfigurationResponse(arg) {
  if (!(arg instanceof auth_auth_pb.GetConfigurationResponse)) {
    throw new Error('Expected argument of type auth_v2.GetConfigurationResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v2_GetConfigurationResponse(buffer_arg) {
  return auth_auth_pb.GetConfigurationResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v2_GetGroupsForPrincipalRequest(arg) {
  if (!(arg instanceof auth_auth_pb.GetGroupsForPrincipalRequest)) {
    throw new Error('Expected argument of type auth_v2.GetGroupsForPrincipalRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v2_GetGroupsForPrincipalRequest(buffer_arg) {
  return auth_auth_pb.GetGroupsForPrincipalRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v2_GetGroupsRequest(arg) {
  if (!(arg instanceof auth_auth_pb.GetGroupsRequest)) {
    throw new Error('Expected argument of type auth_v2.GetGroupsRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v2_GetGroupsRequest(buffer_arg) {
  return auth_auth_pb.GetGroupsRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v2_GetGroupsResponse(arg) {
  if (!(arg instanceof auth_auth_pb.GetGroupsResponse)) {
    throw new Error('Expected argument of type auth_v2.GetGroupsResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v2_GetGroupsResponse(buffer_arg) {
  return auth_auth_pb.GetGroupsResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v2_GetOIDCLoginRequest(arg) {
  if (!(arg instanceof auth_auth_pb.GetOIDCLoginRequest)) {
    throw new Error('Expected argument of type auth_v2.GetOIDCLoginRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v2_GetOIDCLoginRequest(buffer_arg) {
  return auth_auth_pb.GetOIDCLoginRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v2_GetOIDCLoginResponse(arg) {
  if (!(arg instanceof auth_auth_pb.GetOIDCLoginResponse)) {
    throw new Error('Expected argument of type auth_v2.GetOIDCLoginResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v2_GetOIDCLoginResponse(buffer_arg) {
  return auth_auth_pb.GetOIDCLoginResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v2_GetPermissionsForPrincipalRequest(arg) {
  if (!(arg instanceof auth_auth_pb.GetPermissionsForPrincipalRequest)) {
    throw new Error('Expected argument of type auth_v2.GetPermissionsForPrincipalRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v2_GetPermissionsForPrincipalRequest(buffer_arg) {
  return auth_auth_pb.GetPermissionsForPrincipalRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v2_GetPermissionsRequest(arg) {
  if (!(arg instanceof auth_auth_pb.GetPermissionsRequest)) {
    throw new Error('Expected argument of type auth_v2.GetPermissionsRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v2_GetPermissionsRequest(buffer_arg) {
  return auth_auth_pb.GetPermissionsRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v2_GetPermissionsResponse(arg) {
  if (!(arg instanceof auth_auth_pb.GetPermissionsResponse)) {
    throw new Error('Expected argument of type auth_v2.GetPermissionsResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v2_GetPermissionsResponse(buffer_arg) {
  return auth_auth_pb.GetPermissionsResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v2_GetRobotTokenRequest(arg) {
  if (!(arg instanceof auth_auth_pb.GetRobotTokenRequest)) {
    throw new Error('Expected argument of type auth_v2.GetRobotTokenRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v2_GetRobotTokenRequest(buffer_arg) {
  return auth_auth_pb.GetRobotTokenRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v2_GetRobotTokenResponse(arg) {
  if (!(arg instanceof auth_auth_pb.GetRobotTokenResponse)) {
    throw new Error('Expected argument of type auth_v2.GetRobotTokenResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v2_GetRobotTokenResponse(buffer_arg) {
  return auth_auth_pb.GetRobotTokenResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v2_GetRoleBindingRequest(arg) {
  if (!(arg instanceof auth_auth_pb.GetRoleBindingRequest)) {
    throw new Error('Expected argument of type auth_v2.GetRoleBindingRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v2_GetRoleBindingRequest(buffer_arg) {
  return auth_auth_pb.GetRoleBindingRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v2_GetRoleBindingResponse(arg) {
  if (!(arg instanceof auth_auth_pb.GetRoleBindingResponse)) {
    throw new Error('Expected argument of type auth_v2.GetRoleBindingResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v2_GetRoleBindingResponse(buffer_arg) {
  return auth_auth_pb.GetRoleBindingResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v2_GetRolesForPermissionRequest(arg) {
  if (!(arg instanceof auth_auth_pb.GetRolesForPermissionRequest)) {
    throw new Error('Expected argument of type auth_v2.GetRolesForPermissionRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v2_GetRolesForPermissionRequest(buffer_arg) {
  return auth_auth_pb.GetRolesForPermissionRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v2_GetRolesForPermissionResponse(arg) {
  if (!(arg instanceof auth_auth_pb.GetRolesForPermissionResponse)) {
    throw new Error('Expected argument of type auth_v2.GetRolesForPermissionResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v2_GetRolesForPermissionResponse(buffer_arg) {
  return auth_auth_pb.GetRolesForPermissionResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v2_GetUsersRequest(arg) {
  if (!(arg instanceof auth_auth_pb.GetUsersRequest)) {
    throw new Error('Expected argument of type auth_v2.GetUsersRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v2_GetUsersRequest(buffer_arg) {
  return auth_auth_pb.GetUsersRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v2_GetUsersResponse(arg) {
  if (!(arg instanceof auth_auth_pb.GetUsersResponse)) {
    throw new Error('Expected argument of type auth_v2.GetUsersResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v2_GetUsersResponse(buffer_arg) {
  return auth_auth_pb.GetUsersResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v2_ModifyMembersRequest(arg) {
  if (!(arg instanceof auth_auth_pb.ModifyMembersRequest)) {
    throw new Error('Expected argument of type auth_v2.ModifyMembersRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v2_ModifyMembersRequest(buffer_arg) {
  return auth_auth_pb.ModifyMembersRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v2_ModifyMembersResponse(arg) {
  if (!(arg instanceof auth_auth_pb.ModifyMembersResponse)) {
    throw new Error('Expected argument of type auth_v2.ModifyMembersResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v2_ModifyMembersResponse(buffer_arg) {
  return auth_auth_pb.ModifyMembersResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v2_ModifyRoleBindingRequest(arg) {
  if (!(arg instanceof auth_auth_pb.ModifyRoleBindingRequest)) {
    throw new Error('Expected argument of type auth_v2.ModifyRoleBindingRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v2_ModifyRoleBindingRequest(buffer_arg) {
  return auth_auth_pb.ModifyRoleBindingRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v2_ModifyRoleBindingResponse(arg) {
  if (!(arg instanceof auth_auth_pb.ModifyRoleBindingResponse)) {
    throw new Error('Expected argument of type auth_v2.ModifyRoleBindingResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v2_ModifyRoleBindingResponse(buffer_arg) {
  return auth_auth_pb.ModifyRoleBindingResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v2_RestoreAuthTokenRequest(arg) {
  if (!(arg instanceof auth_auth_pb.RestoreAuthTokenRequest)) {
    throw new Error('Expected argument of type auth_v2.RestoreAuthTokenRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v2_RestoreAuthTokenRequest(buffer_arg) {
  return auth_auth_pb.RestoreAuthTokenRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v2_RestoreAuthTokenResponse(arg) {
  if (!(arg instanceof auth_auth_pb.RestoreAuthTokenResponse)) {
    throw new Error('Expected argument of type auth_v2.RestoreAuthTokenResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v2_RestoreAuthTokenResponse(buffer_arg) {
  return auth_auth_pb.RestoreAuthTokenResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v2_RevokeAuthTokenRequest(arg) {
  if (!(arg instanceof auth_auth_pb.RevokeAuthTokenRequest)) {
    throw new Error('Expected argument of type auth_v2.RevokeAuthTokenRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v2_RevokeAuthTokenRequest(buffer_arg) {
  return auth_auth_pb.RevokeAuthTokenRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v2_RevokeAuthTokenResponse(arg) {
  if (!(arg instanceof auth_auth_pb.RevokeAuthTokenResponse)) {
    throw new Error('Expected argument of type auth_v2.RevokeAuthTokenResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v2_RevokeAuthTokenResponse(buffer_arg) {
  return auth_auth_pb.RevokeAuthTokenResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v2_RevokeAuthTokensForUserRequest(arg) {
  if (!(arg instanceof auth_auth_pb.RevokeAuthTokensForUserRequest)) {
    throw new Error('Expected argument of type auth_v2.RevokeAuthTokensForUserRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v2_RevokeAuthTokensForUserRequest(buffer_arg) {
  return auth_auth_pb.RevokeAuthTokensForUserRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v2_RevokeAuthTokensForUserResponse(arg) {
  if (!(arg instanceof auth_auth_pb.RevokeAuthTokensForUserResponse)) {
    throw new Error('Expected argument of type auth_v2.RevokeAuthTokensForUserResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v2_RevokeAuthTokensForUserResponse(buffer_arg) {
  return auth_auth_pb.RevokeAuthTokensForUserResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v2_RotateRootTokenRequest(arg) {
  if (!(arg instanceof auth_auth_pb.RotateRootTokenRequest)) {
    throw new Error('Expected argument of type auth_v2.RotateRootTokenRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v2_RotateRootTokenRequest(buffer_arg) {
  return auth_auth_pb.RotateRootTokenRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v2_RotateRootTokenResponse(arg) {
  if (!(arg instanceof auth_auth_pb.RotateRootTokenResponse)) {
    throw new Error('Expected argument of type auth_v2.RotateRootTokenResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v2_RotateRootTokenResponse(buffer_arg) {
  return auth_auth_pb.RotateRootTokenResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v2_SetConfigurationRequest(arg) {
  if (!(arg instanceof auth_auth_pb.SetConfigurationRequest)) {
    throw new Error('Expected argument of type auth_v2.SetConfigurationRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v2_SetConfigurationRequest(buffer_arg) {
  return auth_auth_pb.SetConfigurationRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v2_SetConfigurationResponse(arg) {
  if (!(arg instanceof auth_auth_pb.SetConfigurationResponse)) {
    throw new Error('Expected argument of type auth_v2.SetConfigurationResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v2_SetConfigurationResponse(buffer_arg) {
  return auth_auth_pb.SetConfigurationResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v2_SetGroupsForUserRequest(arg) {
  if (!(arg instanceof auth_auth_pb.SetGroupsForUserRequest)) {
    throw new Error('Expected argument of type auth_v2.SetGroupsForUserRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v2_SetGroupsForUserRequest(buffer_arg) {
  return auth_auth_pb.SetGroupsForUserRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v2_SetGroupsForUserResponse(arg) {
  if (!(arg instanceof auth_auth_pb.SetGroupsForUserResponse)) {
    throw new Error('Expected argument of type auth_v2.SetGroupsForUserResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v2_SetGroupsForUserResponse(buffer_arg) {
  return auth_auth_pb.SetGroupsForUserResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v2_WhoAmIRequest(arg) {
  if (!(arg instanceof auth_auth_pb.WhoAmIRequest)) {
    throw new Error('Expected argument of type auth_v2.WhoAmIRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v2_WhoAmIRequest(buffer_arg) {
  return auth_auth_pb.WhoAmIRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_v2_WhoAmIResponse(arg) {
  if (!(arg instanceof auth_auth_pb.WhoAmIResponse)) {
    throw new Error('Expected argument of type auth_v2.WhoAmIResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_v2_WhoAmIResponse(buffer_arg) {
  return auth_auth_pb.WhoAmIResponse.deserializeBinary(new Uint8Array(buffer_arg));
}


var APIService = exports.APIService = {
  // Activate/Deactivate the auth API. 'Activate' sets an initial set of admins
// for the Pachyderm cluster, and 'Deactivate' removes all ACLs, tokens, and
// admins from the Pachyderm cluster, making all data publicly accessable
activate: {
    path: '/auth_v2.API/Activate',
    requestStream: false,
    responseStream: false,
    requestType: auth_auth_pb.ActivateRequest,
    responseType: auth_auth_pb.ActivateResponse,
    requestSerialize: serialize_auth_v2_ActivateRequest,
    requestDeserialize: deserialize_auth_v2_ActivateRequest,
    responseSerialize: serialize_auth_v2_ActivateResponse,
    responseDeserialize: deserialize_auth_v2_ActivateResponse,
  },
  deactivate: {
    path: '/auth_v2.API/Deactivate',
    requestStream: false,
    responseStream: false,
    requestType: auth_auth_pb.DeactivateRequest,
    responseType: auth_auth_pb.DeactivateResponse,
    requestSerialize: serialize_auth_v2_DeactivateRequest,
    requestDeserialize: deserialize_auth_v2_DeactivateRequest,
    responseSerialize: serialize_auth_v2_DeactivateResponse,
    responseDeserialize: deserialize_auth_v2_DeactivateResponse,
  },
  getConfiguration: {
    path: '/auth_v2.API/GetConfiguration',
    requestStream: false,
    responseStream: false,
    requestType: auth_auth_pb.GetConfigurationRequest,
    responseType: auth_auth_pb.GetConfigurationResponse,
    requestSerialize: serialize_auth_v2_GetConfigurationRequest,
    requestDeserialize: deserialize_auth_v2_GetConfigurationRequest,
    responseSerialize: serialize_auth_v2_GetConfigurationResponse,
    responseDeserialize: deserialize_auth_v2_GetConfigurationResponse,
  },
  setConfiguration: {
    path: '/auth_v2.API/SetConfiguration',
    requestStream: false,
    responseStream: false,
    requestType: auth_auth_pb.SetConfigurationRequest,
    responseType: auth_auth_pb.SetConfigurationResponse,
    requestSerialize: serialize_auth_v2_SetConfigurationRequest,
    requestDeserialize: deserialize_auth_v2_SetConfigurationRequest,
    responseSerialize: serialize_auth_v2_SetConfigurationResponse,
    responseDeserialize: deserialize_auth_v2_SetConfigurationResponse,
  },
  authenticate: {
    path: '/auth_v2.API/Authenticate',
    requestStream: false,
    responseStream: false,
    requestType: auth_auth_pb.AuthenticateRequest,
    responseType: auth_auth_pb.AuthenticateResponse,
    requestSerialize: serialize_auth_v2_AuthenticateRequest,
    requestDeserialize: deserialize_auth_v2_AuthenticateRequest,
    responseSerialize: serialize_auth_v2_AuthenticateResponse,
    responseDeserialize: deserialize_auth_v2_AuthenticateResponse,
  },
  authorize: {
    path: '/auth_v2.API/Authorize',
    requestStream: false,
    responseStream: false,
    requestType: auth_auth_pb.AuthorizeRequest,
    responseType: auth_auth_pb.AuthorizeResponse,
    requestSerialize: serialize_auth_v2_AuthorizeRequest,
    requestDeserialize: deserialize_auth_v2_AuthorizeRequest,
    responseSerialize: serialize_auth_v2_AuthorizeResponse,
    responseDeserialize: deserialize_auth_v2_AuthorizeResponse,
  },
  getPermissions: {
    path: '/auth_v2.API/GetPermissions',
    requestStream: false,
    responseStream: false,
    requestType: auth_auth_pb.GetPermissionsRequest,
    responseType: auth_auth_pb.GetPermissionsResponse,
    requestSerialize: serialize_auth_v2_GetPermissionsRequest,
    requestDeserialize: deserialize_auth_v2_GetPermissionsRequest,
    responseSerialize: serialize_auth_v2_GetPermissionsResponse,
    responseDeserialize: deserialize_auth_v2_GetPermissionsResponse,
  },
  getPermissionsForPrincipal: {
    path: '/auth_v2.API/GetPermissionsForPrincipal',
    requestStream: false,
    responseStream: false,
    requestType: auth_auth_pb.GetPermissionsForPrincipalRequest,
    responseType: auth_auth_pb.GetPermissionsResponse,
    requestSerialize: serialize_auth_v2_GetPermissionsForPrincipalRequest,
    requestDeserialize: deserialize_auth_v2_GetPermissionsForPrincipalRequest,
    responseSerialize: serialize_auth_v2_GetPermissionsResponse,
    responseDeserialize: deserialize_auth_v2_GetPermissionsResponse,
  },
  whoAmI: {
    path: '/auth_v2.API/WhoAmI',
    requestStream: false,
    responseStream: false,
    requestType: auth_auth_pb.WhoAmIRequest,
    responseType: auth_auth_pb.WhoAmIResponse,
    requestSerialize: serialize_auth_v2_WhoAmIRequest,
    requestDeserialize: deserialize_auth_v2_WhoAmIRequest,
    responseSerialize: serialize_auth_v2_WhoAmIResponse,
    responseDeserialize: deserialize_auth_v2_WhoAmIResponse,
  },
  getRolesForPermission: {
    path: '/auth_v2.API/GetRolesForPermission',
    requestStream: false,
    responseStream: false,
    requestType: auth_auth_pb.GetRolesForPermissionRequest,
    responseType: auth_auth_pb.GetRolesForPermissionResponse,
    requestSerialize: serialize_auth_v2_GetRolesForPermissionRequest,
    requestDeserialize: deserialize_auth_v2_GetRolesForPermissionRequest,
    responseSerialize: serialize_auth_v2_GetRolesForPermissionResponse,
    responseDeserialize: deserialize_auth_v2_GetRolesForPermissionResponse,
  },
  modifyRoleBinding: {
    path: '/auth_v2.API/ModifyRoleBinding',
    requestStream: false,
    responseStream: false,
    requestType: auth_auth_pb.ModifyRoleBindingRequest,
    responseType: auth_auth_pb.ModifyRoleBindingResponse,
    requestSerialize: serialize_auth_v2_ModifyRoleBindingRequest,
    requestDeserialize: deserialize_auth_v2_ModifyRoleBindingRequest,
    responseSerialize: serialize_auth_v2_ModifyRoleBindingResponse,
    responseDeserialize: deserialize_auth_v2_ModifyRoleBindingResponse,
  },
  getRoleBinding: {
    path: '/auth_v2.API/GetRoleBinding',
    requestStream: false,
    responseStream: false,
    requestType: auth_auth_pb.GetRoleBindingRequest,
    responseType: auth_auth_pb.GetRoleBindingResponse,
    requestSerialize: serialize_auth_v2_GetRoleBindingRequest,
    requestDeserialize: deserialize_auth_v2_GetRoleBindingRequest,
    responseSerialize: serialize_auth_v2_GetRoleBindingResponse,
    responseDeserialize: deserialize_auth_v2_GetRoleBindingResponse,
  },
  getOIDCLogin: {
    path: '/auth_v2.API/GetOIDCLogin',
    requestStream: false,
    responseStream: false,
    requestType: auth_auth_pb.GetOIDCLoginRequest,
    responseType: auth_auth_pb.GetOIDCLoginResponse,
    requestSerialize: serialize_auth_v2_GetOIDCLoginRequest,
    requestDeserialize: deserialize_auth_v2_GetOIDCLoginRequest,
    responseSerialize: serialize_auth_v2_GetOIDCLoginResponse,
    responseDeserialize: deserialize_auth_v2_GetOIDCLoginResponse,
  },
  getRobotToken: {
    path: '/auth_v2.API/GetRobotToken',
    requestStream: false,
    responseStream: false,
    requestType: auth_auth_pb.GetRobotTokenRequest,
    responseType: auth_auth_pb.GetRobotTokenResponse,
    requestSerialize: serialize_auth_v2_GetRobotTokenRequest,
    requestDeserialize: deserialize_auth_v2_GetRobotTokenRequest,
    responseSerialize: serialize_auth_v2_GetRobotTokenResponse,
    responseDeserialize: deserialize_auth_v2_GetRobotTokenResponse,
  },
  revokeAuthToken: {
    path: '/auth_v2.API/RevokeAuthToken',
    requestStream: false,
    responseStream: false,
    requestType: auth_auth_pb.RevokeAuthTokenRequest,
    responseType: auth_auth_pb.RevokeAuthTokenResponse,
    requestSerialize: serialize_auth_v2_RevokeAuthTokenRequest,
    requestDeserialize: deserialize_auth_v2_RevokeAuthTokenRequest,
    responseSerialize: serialize_auth_v2_RevokeAuthTokenResponse,
    responseDeserialize: deserialize_auth_v2_RevokeAuthTokenResponse,
  },
  revokeAuthTokensForUser: {
    path: '/auth_v2.API/RevokeAuthTokensForUser',
    requestStream: false,
    responseStream: false,
    requestType: auth_auth_pb.RevokeAuthTokensForUserRequest,
    responseType: auth_auth_pb.RevokeAuthTokensForUserResponse,
    requestSerialize: serialize_auth_v2_RevokeAuthTokensForUserRequest,
    requestDeserialize: deserialize_auth_v2_RevokeAuthTokensForUserRequest,
    responseSerialize: serialize_auth_v2_RevokeAuthTokensForUserResponse,
    responseDeserialize: deserialize_auth_v2_RevokeAuthTokensForUserResponse,
  },
  setGroupsForUser: {
    path: '/auth_v2.API/SetGroupsForUser',
    requestStream: false,
    responseStream: false,
    requestType: auth_auth_pb.SetGroupsForUserRequest,
    responseType: auth_auth_pb.SetGroupsForUserResponse,
    requestSerialize: serialize_auth_v2_SetGroupsForUserRequest,
    requestDeserialize: deserialize_auth_v2_SetGroupsForUserRequest,
    responseSerialize: serialize_auth_v2_SetGroupsForUserResponse,
    responseDeserialize: deserialize_auth_v2_SetGroupsForUserResponse,
  },
  modifyMembers: {
    path: '/auth_v2.API/ModifyMembers',
    requestStream: false,
    responseStream: false,
    requestType: auth_auth_pb.ModifyMembersRequest,
    responseType: auth_auth_pb.ModifyMembersResponse,
    requestSerialize: serialize_auth_v2_ModifyMembersRequest,
    requestDeserialize: deserialize_auth_v2_ModifyMembersRequest,
    responseSerialize: serialize_auth_v2_ModifyMembersResponse,
    responseDeserialize: deserialize_auth_v2_ModifyMembersResponse,
  },
  getGroups: {
    path: '/auth_v2.API/GetGroups',
    requestStream: false,
    responseStream: false,
    requestType: auth_auth_pb.GetGroupsRequest,
    responseType: auth_auth_pb.GetGroupsResponse,
    requestSerialize: serialize_auth_v2_GetGroupsRequest,
    requestDeserialize: deserialize_auth_v2_GetGroupsRequest,
    responseSerialize: serialize_auth_v2_GetGroupsResponse,
    responseDeserialize: deserialize_auth_v2_GetGroupsResponse,
  },
  getGroupsForPrincipal: {
    path: '/auth_v2.API/GetGroupsForPrincipal',
    requestStream: false,
    responseStream: false,
    requestType: auth_auth_pb.GetGroupsForPrincipalRequest,
    responseType: auth_auth_pb.GetGroupsResponse,
    requestSerialize: serialize_auth_v2_GetGroupsForPrincipalRequest,
    requestDeserialize: deserialize_auth_v2_GetGroupsForPrincipalRequest,
    responseSerialize: serialize_auth_v2_GetGroupsResponse,
    responseDeserialize: deserialize_auth_v2_GetGroupsResponse,
  },
  getUsers: {
    path: '/auth_v2.API/GetUsers',
    requestStream: false,
    responseStream: false,
    requestType: auth_auth_pb.GetUsersRequest,
    responseType: auth_auth_pb.GetUsersResponse,
    requestSerialize: serialize_auth_v2_GetUsersRequest,
    requestDeserialize: deserialize_auth_v2_GetUsersRequest,
    responseSerialize: serialize_auth_v2_GetUsersResponse,
    responseDeserialize: deserialize_auth_v2_GetUsersResponse,
  },
  extractAuthTokens: {
    path: '/auth_v2.API/ExtractAuthTokens',
    requestStream: false,
    responseStream: false,
    requestType: auth_auth_pb.ExtractAuthTokensRequest,
    responseType: auth_auth_pb.ExtractAuthTokensResponse,
    requestSerialize: serialize_auth_v2_ExtractAuthTokensRequest,
    requestDeserialize: deserialize_auth_v2_ExtractAuthTokensRequest,
    responseSerialize: serialize_auth_v2_ExtractAuthTokensResponse,
    responseDeserialize: deserialize_auth_v2_ExtractAuthTokensResponse,
  },
  restoreAuthToken: {
    path: '/auth_v2.API/RestoreAuthToken',
    requestStream: false,
    responseStream: false,
    requestType: auth_auth_pb.RestoreAuthTokenRequest,
    responseType: auth_auth_pb.RestoreAuthTokenResponse,
    requestSerialize: serialize_auth_v2_RestoreAuthTokenRequest,
    requestDeserialize: deserialize_auth_v2_RestoreAuthTokenRequest,
    responseSerialize: serialize_auth_v2_RestoreAuthTokenResponse,
    responseDeserialize: deserialize_auth_v2_RestoreAuthTokenResponse,
  },
  deleteExpiredAuthTokens: {
    path: '/auth_v2.API/DeleteExpiredAuthTokens',
    requestStream: false,
    responseStream: false,
    requestType: auth_auth_pb.DeleteExpiredAuthTokensRequest,
    responseType: auth_auth_pb.DeleteExpiredAuthTokensResponse,
    requestSerialize: serialize_auth_v2_DeleteExpiredAuthTokensRequest,
    requestDeserialize: deserialize_auth_v2_DeleteExpiredAuthTokensRequest,
    responseSerialize: serialize_auth_v2_DeleteExpiredAuthTokensResponse,
    responseDeserialize: deserialize_auth_v2_DeleteExpiredAuthTokensResponse,
  },
  rotateRootToken: {
    path: '/auth_v2.API/RotateRootToken',
    requestStream: false,
    responseStream: false,
    requestType: auth_auth_pb.RotateRootTokenRequest,
    responseType: auth_auth_pb.RotateRootTokenResponse,
    requestSerialize: serialize_auth_v2_RotateRootTokenRequest,
    requestDeserialize: deserialize_auth_v2_RotateRootTokenRequest,
    responseSerialize: serialize_auth_v2_RotateRootTokenResponse,
    responseDeserialize: deserialize_auth_v2_RotateRootTokenResponse,
  },
};

exports.APIClient = grpc.makeGenericClientConstructor(APIService);
