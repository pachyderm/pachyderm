// GENERATED CODE -- DO NOT EDIT!

'use strict';
var grpc = require('@grpc/grpc-js');
var auth_auth_pb = require('../auth/auth_pb.js');
var gogoproto_gogo_pb = require('../gogoproto/gogo_pb.js');
var google_protobuf_timestamp_pb = require('google-protobuf/google/protobuf/timestamp_pb.js');

function serialize_auth_ActivateRequest(arg) {
  if (!(arg instanceof auth_auth_pb.ActivateRequest)) {
    throw new Error('Expected argument of type auth.ActivateRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_ActivateRequest(buffer_arg) {
  return auth_auth_pb.ActivateRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_ActivateResponse(arg) {
  if (!(arg instanceof auth_auth_pb.ActivateResponse)) {
    throw new Error('Expected argument of type auth.ActivateResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_ActivateResponse(buffer_arg) {
  return auth_auth_pb.ActivateResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_AuthenticateRequest(arg) {
  if (!(arg instanceof auth_auth_pb.AuthenticateRequest)) {
    throw new Error('Expected argument of type auth.AuthenticateRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_AuthenticateRequest(buffer_arg) {
  return auth_auth_pb.AuthenticateRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_AuthenticateResponse(arg) {
  if (!(arg instanceof auth_auth_pb.AuthenticateResponse)) {
    throw new Error('Expected argument of type auth.AuthenticateResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_AuthenticateResponse(buffer_arg) {
  return auth_auth_pb.AuthenticateResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_AuthorizeRequest(arg) {
  if (!(arg instanceof auth_auth_pb.AuthorizeRequest)) {
    throw new Error('Expected argument of type auth.AuthorizeRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_AuthorizeRequest(buffer_arg) {
  return auth_auth_pb.AuthorizeRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_AuthorizeResponse(arg) {
  if (!(arg instanceof auth_auth_pb.AuthorizeResponse)) {
    throw new Error('Expected argument of type auth.AuthorizeResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_AuthorizeResponse(buffer_arg) {
  return auth_auth_pb.AuthorizeResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_DeactivateRequest(arg) {
  if (!(arg instanceof auth_auth_pb.DeactivateRequest)) {
    throw new Error('Expected argument of type auth.DeactivateRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_DeactivateRequest(buffer_arg) {
  return auth_auth_pb.DeactivateRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_DeactivateResponse(arg) {
  if (!(arg instanceof auth_auth_pb.DeactivateResponse)) {
    throw new Error('Expected argument of type auth.DeactivateResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_DeactivateResponse(buffer_arg) {
  return auth_auth_pb.DeactivateResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_DeleteExpiredAuthTokensRequest(arg) {
  if (!(arg instanceof auth_auth_pb.DeleteExpiredAuthTokensRequest)) {
    throw new Error('Expected argument of type auth.DeleteExpiredAuthTokensRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_DeleteExpiredAuthTokensRequest(buffer_arg) {
  return auth_auth_pb.DeleteExpiredAuthTokensRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_DeleteExpiredAuthTokensResponse(arg) {
  if (!(arg instanceof auth_auth_pb.DeleteExpiredAuthTokensResponse)) {
    throw new Error('Expected argument of type auth.DeleteExpiredAuthTokensResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_DeleteExpiredAuthTokensResponse(buffer_arg) {
  return auth_auth_pb.DeleteExpiredAuthTokensResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_ExtractAuthTokensRequest(arg) {
  if (!(arg instanceof auth_auth_pb.ExtractAuthTokensRequest)) {
    throw new Error('Expected argument of type auth.ExtractAuthTokensRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_ExtractAuthTokensRequest(buffer_arg) {
  return auth_auth_pb.ExtractAuthTokensRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_ExtractAuthTokensResponse(arg) {
  if (!(arg instanceof auth_auth_pb.ExtractAuthTokensResponse)) {
    throw new Error('Expected argument of type auth.ExtractAuthTokensResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_ExtractAuthTokensResponse(buffer_arg) {
  return auth_auth_pb.ExtractAuthTokensResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_GetConfigurationRequest(arg) {
  if (!(arg instanceof auth_auth_pb.GetConfigurationRequest)) {
    throw new Error('Expected argument of type auth.GetConfigurationRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_GetConfigurationRequest(buffer_arg) {
  return auth_auth_pb.GetConfigurationRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_GetConfigurationResponse(arg) {
  if (!(arg instanceof auth_auth_pb.GetConfigurationResponse)) {
    throw new Error('Expected argument of type auth.GetConfigurationResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_GetConfigurationResponse(buffer_arg) {
  return auth_auth_pb.GetConfigurationResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_GetGroupsForPrincipalRequest(arg) {
  if (!(arg instanceof auth_auth_pb.GetGroupsForPrincipalRequest)) {
    throw new Error('Expected argument of type auth.GetGroupsForPrincipalRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_GetGroupsForPrincipalRequest(buffer_arg) {
  return auth_auth_pb.GetGroupsForPrincipalRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_GetGroupsRequest(arg) {
  if (!(arg instanceof auth_auth_pb.GetGroupsRequest)) {
    throw new Error('Expected argument of type auth.GetGroupsRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_GetGroupsRequest(buffer_arg) {
  return auth_auth_pb.GetGroupsRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_GetGroupsResponse(arg) {
  if (!(arg instanceof auth_auth_pb.GetGroupsResponse)) {
    throw new Error('Expected argument of type auth.GetGroupsResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_GetGroupsResponse(buffer_arg) {
  return auth_auth_pb.GetGroupsResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_GetOIDCLoginRequest(arg) {
  if (!(arg instanceof auth_auth_pb.GetOIDCLoginRequest)) {
    throw new Error('Expected argument of type auth.GetOIDCLoginRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_GetOIDCLoginRequest(buffer_arg) {
  return auth_auth_pb.GetOIDCLoginRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_GetOIDCLoginResponse(arg) {
  if (!(arg instanceof auth_auth_pb.GetOIDCLoginResponse)) {
    throw new Error('Expected argument of type auth.GetOIDCLoginResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_GetOIDCLoginResponse(buffer_arg) {
  return auth_auth_pb.GetOIDCLoginResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_GetPermissionsForPrincipalRequest(arg) {
  if (!(arg instanceof auth_auth_pb.GetPermissionsForPrincipalRequest)) {
    throw new Error('Expected argument of type auth.GetPermissionsForPrincipalRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_GetPermissionsForPrincipalRequest(buffer_arg) {
  return auth_auth_pb.GetPermissionsForPrincipalRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_GetPermissionsRequest(arg) {
  if (!(arg instanceof auth_auth_pb.GetPermissionsRequest)) {
    throw new Error('Expected argument of type auth.GetPermissionsRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_GetPermissionsRequest(buffer_arg) {
  return auth_auth_pb.GetPermissionsRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_GetPermissionsResponse(arg) {
  if (!(arg instanceof auth_auth_pb.GetPermissionsResponse)) {
    throw new Error('Expected argument of type auth.GetPermissionsResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_GetPermissionsResponse(buffer_arg) {
  return auth_auth_pb.GetPermissionsResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_GetRobotTokenRequest(arg) {
  if (!(arg instanceof auth_auth_pb.GetRobotTokenRequest)) {
    throw new Error('Expected argument of type auth.GetRobotTokenRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_GetRobotTokenRequest(buffer_arg) {
  return auth_auth_pb.GetRobotTokenRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_GetRobotTokenResponse(arg) {
  if (!(arg instanceof auth_auth_pb.GetRobotTokenResponse)) {
    throw new Error('Expected argument of type auth.GetRobotTokenResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_GetRobotTokenResponse(buffer_arg) {
  return auth_auth_pb.GetRobotTokenResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_GetRoleBindingRequest(arg) {
  if (!(arg instanceof auth_auth_pb.GetRoleBindingRequest)) {
    throw new Error('Expected argument of type auth.GetRoleBindingRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_GetRoleBindingRequest(buffer_arg) {
  return auth_auth_pb.GetRoleBindingRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_GetRoleBindingResponse(arg) {
  if (!(arg instanceof auth_auth_pb.GetRoleBindingResponse)) {
    throw new Error('Expected argument of type auth.GetRoleBindingResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_GetRoleBindingResponse(buffer_arg) {
  return auth_auth_pb.GetRoleBindingResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_GetUsersRequest(arg) {
  if (!(arg instanceof auth_auth_pb.GetUsersRequest)) {
    throw new Error('Expected argument of type auth.GetUsersRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_GetUsersRequest(buffer_arg) {
  return auth_auth_pb.GetUsersRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_GetUsersResponse(arg) {
  if (!(arg instanceof auth_auth_pb.GetUsersResponse)) {
    throw new Error('Expected argument of type auth.GetUsersResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_GetUsersResponse(buffer_arg) {
  return auth_auth_pb.GetUsersResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_ModifyMembersRequest(arg) {
  if (!(arg instanceof auth_auth_pb.ModifyMembersRequest)) {
    throw new Error('Expected argument of type auth.ModifyMembersRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_ModifyMembersRequest(buffer_arg) {
  return auth_auth_pb.ModifyMembersRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_ModifyMembersResponse(arg) {
  if (!(arg instanceof auth_auth_pb.ModifyMembersResponse)) {
    throw new Error('Expected argument of type auth.ModifyMembersResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_ModifyMembersResponse(buffer_arg) {
  return auth_auth_pb.ModifyMembersResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_ModifyRoleBindingRequest(arg) {
  if (!(arg instanceof auth_auth_pb.ModifyRoleBindingRequest)) {
    throw new Error('Expected argument of type auth.ModifyRoleBindingRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_ModifyRoleBindingRequest(buffer_arg) {
  return auth_auth_pb.ModifyRoleBindingRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_ModifyRoleBindingResponse(arg) {
  if (!(arg instanceof auth_auth_pb.ModifyRoleBindingResponse)) {
    throw new Error('Expected argument of type auth.ModifyRoleBindingResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_ModifyRoleBindingResponse(buffer_arg) {
  return auth_auth_pb.ModifyRoleBindingResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_RestoreAuthTokenRequest(arg) {
  if (!(arg instanceof auth_auth_pb.RestoreAuthTokenRequest)) {
    throw new Error('Expected argument of type auth.RestoreAuthTokenRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_RestoreAuthTokenRequest(buffer_arg) {
  return auth_auth_pb.RestoreAuthTokenRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_RestoreAuthTokenResponse(arg) {
  if (!(arg instanceof auth_auth_pb.RestoreAuthTokenResponse)) {
    throw new Error('Expected argument of type auth.RestoreAuthTokenResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_RestoreAuthTokenResponse(buffer_arg) {
  return auth_auth_pb.RestoreAuthTokenResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_RevokeAuthTokenRequest(arg) {
  if (!(arg instanceof auth_auth_pb.RevokeAuthTokenRequest)) {
    throw new Error('Expected argument of type auth.RevokeAuthTokenRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_RevokeAuthTokenRequest(buffer_arg) {
  return auth_auth_pb.RevokeAuthTokenRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_RevokeAuthTokenResponse(arg) {
  if (!(arg instanceof auth_auth_pb.RevokeAuthTokenResponse)) {
    throw new Error('Expected argument of type auth.RevokeAuthTokenResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_RevokeAuthTokenResponse(buffer_arg) {
  return auth_auth_pb.RevokeAuthTokenResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_RevokeAuthTokensForUserRequest(arg) {
  if (!(arg instanceof auth_auth_pb.RevokeAuthTokensForUserRequest)) {
    throw new Error('Expected argument of type auth.RevokeAuthTokensForUserRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_RevokeAuthTokensForUserRequest(buffer_arg) {
  return auth_auth_pb.RevokeAuthTokensForUserRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_RevokeAuthTokensForUserResponse(arg) {
  if (!(arg instanceof auth_auth_pb.RevokeAuthTokensForUserResponse)) {
    throw new Error('Expected argument of type auth.RevokeAuthTokensForUserResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_RevokeAuthTokensForUserResponse(buffer_arg) {
  return auth_auth_pb.RevokeAuthTokensForUserResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_SetConfigurationRequest(arg) {
  if (!(arg instanceof auth_auth_pb.SetConfigurationRequest)) {
    throw new Error('Expected argument of type auth.SetConfigurationRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_SetConfigurationRequest(buffer_arg) {
  return auth_auth_pb.SetConfigurationRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_SetConfigurationResponse(arg) {
  if (!(arg instanceof auth_auth_pb.SetConfigurationResponse)) {
    throw new Error('Expected argument of type auth.SetConfigurationResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_SetConfigurationResponse(buffer_arg) {
  return auth_auth_pb.SetConfigurationResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_SetGroupsForUserRequest(arg) {
  if (!(arg instanceof auth_auth_pb.SetGroupsForUserRequest)) {
    throw new Error('Expected argument of type auth.SetGroupsForUserRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_SetGroupsForUserRequest(buffer_arg) {
  return auth_auth_pb.SetGroupsForUserRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_SetGroupsForUserResponse(arg) {
  if (!(arg instanceof auth_auth_pb.SetGroupsForUserResponse)) {
    throw new Error('Expected argument of type auth.SetGroupsForUserResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_SetGroupsForUserResponse(buffer_arg) {
  return auth_auth_pb.SetGroupsForUserResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_WhoAmIRequest(arg) {
  if (!(arg instanceof auth_auth_pb.WhoAmIRequest)) {
    throw new Error('Expected argument of type auth.WhoAmIRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_WhoAmIRequest(buffer_arg) {
  return auth_auth_pb.WhoAmIRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_WhoAmIResponse(arg) {
  if (!(arg instanceof auth_auth_pb.WhoAmIResponse)) {
    throw new Error('Expected argument of type auth.WhoAmIResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_WhoAmIResponse(buffer_arg) {
  return auth_auth_pb.WhoAmIResponse.deserializeBinary(new Uint8Array(buffer_arg));
}


var APIService = exports.APIService = {
  // Activate/Deactivate the auth API. 'Activate' sets an initial set of admins
// for the Pachyderm cluster, and 'Deactivate' removes all ACLs, tokens, and
// admins from the Pachyderm cluster, making all data publicly accessable
activate: {
    path: '/auth.API/Activate',
    requestStream: false,
    responseStream: false,
    requestType: auth_auth_pb.ActivateRequest,
    responseType: auth_auth_pb.ActivateResponse,
    requestSerialize: serialize_auth_ActivateRequest,
    requestDeserialize: deserialize_auth_ActivateRequest,
    responseSerialize: serialize_auth_ActivateResponse,
    responseDeserialize: deserialize_auth_ActivateResponse,
  },
  deactivate: {
    path: '/auth.API/Deactivate',
    requestStream: false,
    responseStream: false,
    requestType: auth_auth_pb.DeactivateRequest,
    responseType: auth_auth_pb.DeactivateResponse,
    requestSerialize: serialize_auth_DeactivateRequest,
    requestDeserialize: deserialize_auth_DeactivateRequest,
    responseSerialize: serialize_auth_DeactivateResponse,
    responseDeserialize: deserialize_auth_DeactivateResponse,
  },
  getConfiguration: {
    path: '/auth.API/GetConfiguration',
    requestStream: false,
    responseStream: false,
    requestType: auth_auth_pb.GetConfigurationRequest,
    responseType: auth_auth_pb.GetConfigurationResponse,
    requestSerialize: serialize_auth_GetConfigurationRequest,
    requestDeserialize: deserialize_auth_GetConfigurationRequest,
    responseSerialize: serialize_auth_GetConfigurationResponse,
    responseDeserialize: deserialize_auth_GetConfigurationResponse,
  },
  setConfiguration: {
    path: '/auth.API/SetConfiguration',
    requestStream: false,
    responseStream: false,
    requestType: auth_auth_pb.SetConfigurationRequest,
    responseType: auth_auth_pb.SetConfigurationResponse,
    requestSerialize: serialize_auth_SetConfigurationRequest,
    requestDeserialize: deserialize_auth_SetConfigurationRequest,
    responseSerialize: serialize_auth_SetConfigurationResponse,
    responseDeserialize: deserialize_auth_SetConfigurationResponse,
  },
  authenticate: {
    path: '/auth.API/Authenticate',
    requestStream: false,
    responseStream: false,
    requestType: auth_auth_pb.AuthenticateRequest,
    responseType: auth_auth_pb.AuthenticateResponse,
    requestSerialize: serialize_auth_AuthenticateRequest,
    requestDeserialize: deserialize_auth_AuthenticateRequest,
    responseSerialize: serialize_auth_AuthenticateResponse,
    responseDeserialize: deserialize_auth_AuthenticateResponse,
  },
  authorize: {
    path: '/auth.API/Authorize',
    requestStream: false,
    responseStream: false,
    requestType: auth_auth_pb.AuthorizeRequest,
    responseType: auth_auth_pb.AuthorizeResponse,
    requestSerialize: serialize_auth_AuthorizeRequest,
    requestDeserialize: deserialize_auth_AuthorizeRequest,
    responseSerialize: serialize_auth_AuthorizeResponse,
    responseDeserialize: deserialize_auth_AuthorizeResponse,
  },
  getPermissions: {
    path: '/auth.API/GetPermissions',
    requestStream: false,
    responseStream: false,
    requestType: auth_auth_pb.GetPermissionsRequest,
    responseType: auth_auth_pb.GetPermissionsResponse,
    requestSerialize: serialize_auth_GetPermissionsRequest,
    requestDeserialize: deserialize_auth_GetPermissionsRequest,
    responseSerialize: serialize_auth_GetPermissionsResponse,
    responseDeserialize: deserialize_auth_GetPermissionsResponse,
  },
  getPermissionsForPrincipal: {
    path: '/auth.API/GetPermissionsForPrincipal',
    requestStream: false,
    responseStream: false,
    requestType: auth_auth_pb.GetPermissionsForPrincipalRequest,
    responseType: auth_auth_pb.GetPermissionsResponse,
    requestSerialize: serialize_auth_GetPermissionsForPrincipalRequest,
    requestDeserialize: deserialize_auth_GetPermissionsForPrincipalRequest,
    responseSerialize: serialize_auth_GetPermissionsResponse,
    responseDeserialize: deserialize_auth_GetPermissionsResponse,
  },
  whoAmI: {
    path: '/auth.API/WhoAmI',
    requestStream: false,
    responseStream: false,
    requestType: auth_auth_pb.WhoAmIRequest,
    responseType: auth_auth_pb.WhoAmIResponse,
    requestSerialize: serialize_auth_WhoAmIRequest,
    requestDeserialize: deserialize_auth_WhoAmIRequest,
    responseSerialize: serialize_auth_WhoAmIResponse,
    responseDeserialize: deserialize_auth_WhoAmIResponse,
  },
  modifyRoleBinding: {
    path: '/auth.API/ModifyRoleBinding',
    requestStream: false,
    responseStream: false,
    requestType: auth_auth_pb.ModifyRoleBindingRequest,
    responseType: auth_auth_pb.ModifyRoleBindingResponse,
    requestSerialize: serialize_auth_ModifyRoleBindingRequest,
    requestDeserialize: deserialize_auth_ModifyRoleBindingRequest,
    responseSerialize: serialize_auth_ModifyRoleBindingResponse,
    responseDeserialize: deserialize_auth_ModifyRoleBindingResponse,
  },
  getRoleBinding: {
    path: '/auth.API/GetRoleBinding',
    requestStream: false,
    responseStream: false,
    requestType: auth_auth_pb.GetRoleBindingRequest,
    responseType: auth_auth_pb.GetRoleBindingResponse,
    requestSerialize: serialize_auth_GetRoleBindingRequest,
    requestDeserialize: deserialize_auth_GetRoleBindingRequest,
    responseSerialize: serialize_auth_GetRoleBindingResponse,
    responseDeserialize: deserialize_auth_GetRoleBindingResponse,
  },
  getOIDCLogin: {
    path: '/auth.API/GetOIDCLogin',
    requestStream: false,
    responseStream: false,
    requestType: auth_auth_pb.GetOIDCLoginRequest,
    responseType: auth_auth_pb.GetOIDCLoginResponse,
    requestSerialize: serialize_auth_GetOIDCLoginRequest,
    requestDeserialize: deserialize_auth_GetOIDCLoginRequest,
    responseSerialize: serialize_auth_GetOIDCLoginResponse,
    responseDeserialize: deserialize_auth_GetOIDCLoginResponse,
  },
  getRobotToken: {
    path: '/auth.API/GetRobotToken',
    requestStream: false,
    responseStream: false,
    requestType: auth_auth_pb.GetRobotTokenRequest,
    responseType: auth_auth_pb.GetRobotTokenResponse,
    requestSerialize: serialize_auth_GetRobotTokenRequest,
    requestDeserialize: deserialize_auth_GetRobotTokenRequest,
    responseSerialize: serialize_auth_GetRobotTokenResponse,
    responseDeserialize: deserialize_auth_GetRobotTokenResponse,
  },
  revokeAuthToken: {
    path: '/auth.API/RevokeAuthToken',
    requestStream: false,
    responseStream: false,
    requestType: auth_auth_pb.RevokeAuthTokenRequest,
    responseType: auth_auth_pb.RevokeAuthTokenResponse,
    requestSerialize: serialize_auth_RevokeAuthTokenRequest,
    requestDeserialize: deserialize_auth_RevokeAuthTokenRequest,
    responseSerialize: serialize_auth_RevokeAuthTokenResponse,
    responseDeserialize: deserialize_auth_RevokeAuthTokenResponse,
  },
  revokeAuthTokensForUser: {
    path: '/auth.API/RevokeAuthTokensForUser',
    requestStream: false,
    responseStream: false,
    requestType: auth_auth_pb.RevokeAuthTokensForUserRequest,
    responseType: auth_auth_pb.RevokeAuthTokensForUserResponse,
    requestSerialize: serialize_auth_RevokeAuthTokensForUserRequest,
    requestDeserialize: deserialize_auth_RevokeAuthTokensForUserRequest,
    responseSerialize: serialize_auth_RevokeAuthTokensForUserResponse,
    responseDeserialize: deserialize_auth_RevokeAuthTokensForUserResponse,
  },
  setGroupsForUser: {
    path: '/auth.API/SetGroupsForUser',
    requestStream: false,
    responseStream: false,
    requestType: auth_auth_pb.SetGroupsForUserRequest,
    responseType: auth_auth_pb.SetGroupsForUserResponse,
    requestSerialize: serialize_auth_SetGroupsForUserRequest,
    requestDeserialize: deserialize_auth_SetGroupsForUserRequest,
    responseSerialize: serialize_auth_SetGroupsForUserResponse,
    responseDeserialize: deserialize_auth_SetGroupsForUserResponse,
  },
  modifyMembers: {
    path: '/auth.API/ModifyMembers',
    requestStream: false,
    responseStream: false,
    requestType: auth_auth_pb.ModifyMembersRequest,
    responseType: auth_auth_pb.ModifyMembersResponse,
    requestSerialize: serialize_auth_ModifyMembersRequest,
    requestDeserialize: deserialize_auth_ModifyMembersRequest,
    responseSerialize: serialize_auth_ModifyMembersResponse,
    responseDeserialize: deserialize_auth_ModifyMembersResponse,
  },
  getGroups: {
    path: '/auth.API/GetGroups',
    requestStream: false,
    responseStream: false,
    requestType: auth_auth_pb.GetGroupsRequest,
    responseType: auth_auth_pb.GetGroupsResponse,
    requestSerialize: serialize_auth_GetGroupsRequest,
    requestDeserialize: deserialize_auth_GetGroupsRequest,
    responseSerialize: serialize_auth_GetGroupsResponse,
    responseDeserialize: deserialize_auth_GetGroupsResponse,
  },
  getGroupsForPrincipal: {
    path: '/auth.API/GetGroupsForPrincipal',
    requestStream: false,
    responseStream: false,
    requestType: auth_auth_pb.GetGroupsForPrincipalRequest,
    responseType: auth_auth_pb.GetGroupsResponse,
    requestSerialize: serialize_auth_GetGroupsForPrincipalRequest,
    requestDeserialize: deserialize_auth_GetGroupsForPrincipalRequest,
    responseSerialize: serialize_auth_GetGroupsResponse,
    responseDeserialize: deserialize_auth_GetGroupsResponse,
  },
  getUsers: {
    path: '/auth.API/GetUsers',
    requestStream: false,
    responseStream: false,
    requestType: auth_auth_pb.GetUsersRequest,
    responseType: auth_auth_pb.GetUsersResponse,
    requestSerialize: serialize_auth_GetUsersRequest,
    requestDeserialize: deserialize_auth_GetUsersRequest,
    responseSerialize: serialize_auth_GetUsersResponse,
    responseDeserialize: deserialize_auth_GetUsersResponse,
  },
  extractAuthTokens: {
    path: '/auth.API/ExtractAuthTokens',
    requestStream: false,
    responseStream: false,
    requestType: auth_auth_pb.ExtractAuthTokensRequest,
    responseType: auth_auth_pb.ExtractAuthTokensResponse,
    requestSerialize: serialize_auth_ExtractAuthTokensRequest,
    requestDeserialize: deserialize_auth_ExtractAuthTokensRequest,
    responseSerialize: serialize_auth_ExtractAuthTokensResponse,
    responseDeserialize: deserialize_auth_ExtractAuthTokensResponse,
  },
  restoreAuthToken: {
    path: '/auth.API/RestoreAuthToken',
    requestStream: false,
    responseStream: false,
    requestType: auth_auth_pb.RestoreAuthTokenRequest,
    responseType: auth_auth_pb.RestoreAuthTokenResponse,
    requestSerialize: serialize_auth_RestoreAuthTokenRequest,
    requestDeserialize: deserialize_auth_RestoreAuthTokenRequest,
    responseSerialize: serialize_auth_RestoreAuthTokenResponse,
    responseDeserialize: deserialize_auth_RestoreAuthTokenResponse,
  },
  deleteExpiredAuthTokens: {
    path: '/auth.API/DeleteExpiredAuthTokens',
    requestStream: false,
    responseStream: false,
    requestType: auth_auth_pb.DeleteExpiredAuthTokensRequest,
    responseType: auth_auth_pb.DeleteExpiredAuthTokensResponse,
    requestSerialize: serialize_auth_DeleteExpiredAuthTokensRequest,
    requestDeserialize: deserialize_auth_DeleteExpiredAuthTokensRequest,
    responseSerialize: serialize_auth_DeleteExpiredAuthTokensResponse,
    responseDeserialize: deserialize_auth_DeleteExpiredAuthTokensResponse,
  },
};

exports.APIClient = grpc.makeGenericClientConstructor(APIService);
