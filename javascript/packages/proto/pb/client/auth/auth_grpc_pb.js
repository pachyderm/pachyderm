// GENERATED CODE -- DO NOT EDIT!

'use strict';
var grpc = require('@grpc/grpc-js');
var client_auth_auth_pb = require('../../client/auth/auth_pb.js');
var gogoproto_gogo_pb = require('../../gogoproto/gogo_pb.js');
var google_protobuf_timestamp_pb = require('google-protobuf/google/protobuf/timestamp_pb.js');

function serialize_auth_ActivateRequest(arg) {
  if (!(arg instanceof client_auth_auth_pb.ActivateRequest)) {
    throw new Error('Expected argument of type auth.ActivateRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_ActivateRequest(buffer_arg) {
  return client_auth_auth_pb.ActivateRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_ActivateResponse(arg) {
  if (!(arg instanceof client_auth_auth_pb.ActivateResponse)) {
    throw new Error('Expected argument of type auth.ActivateResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_ActivateResponse(buffer_arg) {
  return client_auth_auth_pb.ActivateResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_AuthenticateRequest(arg) {
  if (!(arg instanceof client_auth_auth_pb.AuthenticateRequest)) {
    throw new Error('Expected argument of type auth.AuthenticateRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_AuthenticateRequest(buffer_arg) {
  return client_auth_auth_pb.AuthenticateRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_AuthenticateResponse(arg) {
  if (!(arg instanceof client_auth_auth_pb.AuthenticateResponse)) {
    throw new Error('Expected argument of type auth.AuthenticateResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_AuthenticateResponse(buffer_arg) {
  return client_auth_auth_pb.AuthenticateResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_AuthorizeRequest(arg) {
  if (!(arg instanceof client_auth_auth_pb.AuthorizeRequest)) {
    throw new Error('Expected argument of type auth.AuthorizeRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_AuthorizeRequest(buffer_arg) {
  return client_auth_auth_pb.AuthorizeRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_AuthorizeResponse(arg) {
  if (!(arg instanceof client_auth_auth_pb.AuthorizeResponse)) {
    throw new Error('Expected argument of type auth.AuthorizeResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_AuthorizeResponse(buffer_arg) {
  return client_auth_auth_pb.AuthorizeResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_DeactivateRequest(arg) {
  if (!(arg instanceof client_auth_auth_pb.DeactivateRequest)) {
    throw new Error('Expected argument of type auth.DeactivateRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_DeactivateRequest(buffer_arg) {
  return client_auth_auth_pb.DeactivateRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_DeactivateResponse(arg) {
  if (!(arg instanceof client_auth_auth_pb.DeactivateResponse)) {
    throw new Error('Expected argument of type auth.DeactivateResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_DeactivateResponse(buffer_arg) {
  return client_auth_auth_pb.DeactivateResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_ExtendAuthTokenRequest(arg) {
  if (!(arg instanceof client_auth_auth_pb.ExtendAuthTokenRequest)) {
    throw new Error('Expected argument of type auth.ExtendAuthTokenRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_ExtendAuthTokenRequest(buffer_arg) {
  return client_auth_auth_pb.ExtendAuthTokenRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_ExtendAuthTokenResponse(arg) {
  if (!(arg instanceof client_auth_auth_pb.ExtendAuthTokenResponse)) {
    throw new Error('Expected argument of type auth.ExtendAuthTokenResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_ExtendAuthTokenResponse(buffer_arg) {
  return client_auth_auth_pb.ExtendAuthTokenResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_GetACLRequest(arg) {
  if (!(arg instanceof client_auth_auth_pb.GetACLRequest)) {
    throw new Error('Expected argument of type auth.GetACLRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_GetACLRequest(buffer_arg) {
  return client_auth_auth_pb.GetACLRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_GetACLResponse(arg) {
  if (!(arg instanceof client_auth_auth_pb.GetACLResponse)) {
    throw new Error('Expected argument of type auth.GetACLResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_GetACLResponse(buffer_arg) {
  return client_auth_auth_pb.GetACLResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_GetAdminsRequest(arg) {
  if (!(arg instanceof client_auth_auth_pb.GetAdminsRequest)) {
    throw new Error('Expected argument of type auth.GetAdminsRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_GetAdminsRequest(buffer_arg) {
  return client_auth_auth_pb.GetAdminsRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_GetAdminsResponse(arg) {
  if (!(arg instanceof client_auth_auth_pb.GetAdminsResponse)) {
    throw new Error('Expected argument of type auth.GetAdminsResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_GetAdminsResponse(buffer_arg) {
  return client_auth_auth_pb.GetAdminsResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_GetAuthTokenRequest(arg) {
  if (!(arg instanceof client_auth_auth_pb.GetAuthTokenRequest)) {
    throw new Error('Expected argument of type auth.GetAuthTokenRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_GetAuthTokenRequest(buffer_arg) {
  return client_auth_auth_pb.GetAuthTokenRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_GetAuthTokenResponse(arg) {
  if (!(arg instanceof client_auth_auth_pb.GetAuthTokenResponse)) {
    throw new Error('Expected argument of type auth.GetAuthTokenResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_GetAuthTokenResponse(buffer_arg) {
  return client_auth_auth_pb.GetAuthTokenResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_GetClusterRoleBindingsRequest(arg) {
  if (!(arg instanceof client_auth_auth_pb.GetClusterRoleBindingsRequest)) {
    throw new Error('Expected argument of type auth.GetClusterRoleBindingsRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_GetClusterRoleBindingsRequest(buffer_arg) {
  return client_auth_auth_pb.GetClusterRoleBindingsRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_GetClusterRoleBindingsResponse(arg) {
  if (!(arg instanceof client_auth_auth_pb.GetClusterRoleBindingsResponse)) {
    throw new Error('Expected argument of type auth.GetClusterRoleBindingsResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_GetClusterRoleBindingsResponse(buffer_arg) {
  return client_auth_auth_pb.GetClusterRoleBindingsResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_GetConfigurationRequest(arg) {
  if (!(arg instanceof client_auth_auth_pb.GetConfigurationRequest)) {
    throw new Error('Expected argument of type auth.GetConfigurationRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_GetConfigurationRequest(buffer_arg) {
  return client_auth_auth_pb.GetConfigurationRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_GetConfigurationResponse(arg) {
  if (!(arg instanceof client_auth_auth_pb.GetConfigurationResponse)) {
    throw new Error('Expected argument of type auth.GetConfigurationResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_GetConfigurationResponse(buffer_arg) {
  return client_auth_auth_pb.GetConfigurationResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_GetGroupsRequest(arg) {
  if (!(arg instanceof client_auth_auth_pb.GetGroupsRequest)) {
    throw new Error('Expected argument of type auth.GetGroupsRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_GetGroupsRequest(buffer_arg) {
  return client_auth_auth_pb.GetGroupsRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_GetGroupsResponse(arg) {
  if (!(arg instanceof client_auth_auth_pb.GetGroupsResponse)) {
    throw new Error('Expected argument of type auth.GetGroupsResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_GetGroupsResponse(buffer_arg) {
  return client_auth_auth_pb.GetGroupsResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_GetOIDCLoginRequest(arg) {
  if (!(arg instanceof client_auth_auth_pb.GetOIDCLoginRequest)) {
    throw new Error('Expected argument of type auth.GetOIDCLoginRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_GetOIDCLoginRequest(buffer_arg) {
  return client_auth_auth_pb.GetOIDCLoginRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_GetOIDCLoginResponse(arg) {
  if (!(arg instanceof client_auth_auth_pb.GetOIDCLoginResponse)) {
    throw new Error('Expected argument of type auth.GetOIDCLoginResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_GetOIDCLoginResponse(buffer_arg) {
  return client_auth_auth_pb.GetOIDCLoginResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_GetOneTimePasswordRequest(arg) {
  if (!(arg instanceof client_auth_auth_pb.GetOneTimePasswordRequest)) {
    throw new Error('Expected argument of type auth.GetOneTimePasswordRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_GetOneTimePasswordRequest(buffer_arg) {
  return client_auth_auth_pb.GetOneTimePasswordRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_GetOneTimePasswordResponse(arg) {
  if (!(arg instanceof client_auth_auth_pb.GetOneTimePasswordResponse)) {
    throw new Error('Expected argument of type auth.GetOneTimePasswordResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_GetOneTimePasswordResponse(buffer_arg) {
  return client_auth_auth_pb.GetOneTimePasswordResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_GetScopeRequest(arg) {
  if (!(arg instanceof client_auth_auth_pb.GetScopeRequest)) {
    throw new Error('Expected argument of type auth.GetScopeRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_GetScopeRequest(buffer_arg) {
  return client_auth_auth_pb.GetScopeRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_GetScopeResponse(arg) {
  if (!(arg instanceof client_auth_auth_pb.GetScopeResponse)) {
    throw new Error('Expected argument of type auth.GetScopeResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_GetScopeResponse(buffer_arg) {
  return client_auth_auth_pb.GetScopeResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_GetUsersRequest(arg) {
  if (!(arg instanceof client_auth_auth_pb.GetUsersRequest)) {
    throw new Error('Expected argument of type auth.GetUsersRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_GetUsersRequest(buffer_arg) {
  return client_auth_auth_pb.GetUsersRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_GetUsersResponse(arg) {
  if (!(arg instanceof client_auth_auth_pb.GetUsersResponse)) {
    throw new Error('Expected argument of type auth.GetUsersResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_GetUsersResponse(buffer_arg) {
  return client_auth_auth_pb.GetUsersResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_ModifyAdminsRequest(arg) {
  if (!(arg instanceof client_auth_auth_pb.ModifyAdminsRequest)) {
    throw new Error('Expected argument of type auth.ModifyAdminsRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_ModifyAdminsRequest(buffer_arg) {
  return client_auth_auth_pb.ModifyAdminsRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_ModifyAdminsResponse(arg) {
  if (!(arg instanceof client_auth_auth_pb.ModifyAdminsResponse)) {
    throw new Error('Expected argument of type auth.ModifyAdminsResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_ModifyAdminsResponse(buffer_arg) {
  return client_auth_auth_pb.ModifyAdminsResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_ModifyClusterRoleBindingRequest(arg) {
  if (!(arg instanceof client_auth_auth_pb.ModifyClusterRoleBindingRequest)) {
    throw new Error('Expected argument of type auth.ModifyClusterRoleBindingRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_ModifyClusterRoleBindingRequest(buffer_arg) {
  return client_auth_auth_pb.ModifyClusterRoleBindingRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_ModifyClusterRoleBindingResponse(arg) {
  if (!(arg instanceof client_auth_auth_pb.ModifyClusterRoleBindingResponse)) {
    throw new Error('Expected argument of type auth.ModifyClusterRoleBindingResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_ModifyClusterRoleBindingResponse(buffer_arg) {
  return client_auth_auth_pb.ModifyClusterRoleBindingResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_ModifyMembersRequest(arg) {
  if (!(arg instanceof client_auth_auth_pb.ModifyMembersRequest)) {
    throw new Error('Expected argument of type auth.ModifyMembersRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_ModifyMembersRequest(buffer_arg) {
  return client_auth_auth_pb.ModifyMembersRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_ModifyMembersResponse(arg) {
  if (!(arg instanceof client_auth_auth_pb.ModifyMembersResponse)) {
    throw new Error('Expected argument of type auth.ModifyMembersResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_ModifyMembersResponse(buffer_arg) {
  return client_auth_auth_pb.ModifyMembersResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_RevokeAuthTokenRequest(arg) {
  if (!(arg instanceof client_auth_auth_pb.RevokeAuthTokenRequest)) {
    throw new Error('Expected argument of type auth.RevokeAuthTokenRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_RevokeAuthTokenRequest(buffer_arg) {
  return client_auth_auth_pb.RevokeAuthTokenRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_RevokeAuthTokenResponse(arg) {
  if (!(arg instanceof client_auth_auth_pb.RevokeAuthTokenResponse)) {
    throw new Error('Expected argument of type auth.RevokeAuthTokenResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_RevokeAuthTokenResponse(buffer_arg) {
  return client_auth_auth_pb.RevokeAuthTokenResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_SetACLRequest(arg) {
  if (!(arg instanceof client_auth_auth_pb.SetACLRequest)) {
    throw new Error('Expected argument of type auth.SetACLRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_SetACLRequest(buffer_arg) {
  return client_auth_auth_pb.SetACLRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_SetACLResponse(arg) {
  if (!(arg instanceof client_auth_auth_pb.SetACLResponse)) {
    throw new Error('Expected argument of type auth.SetACLResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_SetACLResponse(buffer_arg) {
  return client_auth_auth_pb.SetACLResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_SetConfigurationRequest(arg) {
  if (!(arg instanceof client_auth_auth_pb.SetConfigurationRequest)) {
    throw new Error('Expected argument of type auth.SetConfigurationRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_SetConfigurationRequest(buffer_arg) {
  return client_auth_auth_pb.SetConfigurationRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_SetConfigurationResponse(arg) {
  if (!(arg instanceof client_auth_auth_pb.SetConfigurationResponse)) {
    throw new Error('Expected argument of type auth.SetConfigurationResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_SetConfigurationResponse(buffer_arg) {
  return client_auth_auth_pb.SetConfigurationResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_SetGroupsForUserRequest(arg) {
  if (!(arg instanceof client_auth_auth_pb.SetGroupsForUserRequest)) {
    throw new Error('Expected argument of type auth.SetGroupsForUserRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_SetGroupsForUserRequest(buffer_arg) {
  return client_auth_auth_pb.SetGroupsForUserRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_SetGroupsForUserResponse(arg) {
  if (!(arg instanceof client_auth_auth_pb.SetGroupsForUserResponse)) {
    throw new Error('Expected argument of type auth.SetGroupsForUserResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_SetGroupsForUserResponse(buffer_arg) {
  return client_auth_auth_pb.SetGroupsForUserResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_SetScopeRequest(arg) {
  if (!(arg instanceof client_auth_auth_pb.SetScopeRequest)) {
    throw new Error('Expected argument of type auth.SetScopeRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_SetScopeRequest(buffer_arg) {
  return client_auth_auth_pb.SetScopeRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_SetScopeResponse(arg) {
  if (!(arg instanceof client_auth_auth_pb.SetScopeResponse)) {
    throw new Error('Expected argument of type auth.SetScopeResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_SetScopeResponse(buffer_arg) {
  return client_auth_auth_pb.SetScopeResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_WhoAmIRequest(arg) {
  if (!(arg instanceof client_auth_auth_pb.WhoAmIRequest)) {
    throw new Error('Expected argument of type auth.WhoAmIRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_WhoAmIRequest(buffer_arg) {
  return client_auth_auth_pb.WhoAmIRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_auth_WhoAmIResponse(arg) {
  if (!(arg instanceof client_auth_auth_pb.WhoAmIResponse)) {
    throw new Error('Expected argument of type auth.WhoAmIResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_auth_WhoAmIResponse(buffer_arg) {
  return client_auth_auth_pb.WhoAmIResponse.deserializeBinary(new Uint8Array(buffer_arg));
}


var APIService = exports.APIService = {
  // Activate/Deactivate the auth API. 'Activate' sets an initial set of admins
// for the Pachyderm cluster, and 'Deactivate' removes all ACLs, tokens, and
// admins from the Pachyderm cluster, making all data publicly accessable
activate: {
    path: '/auth.API/Activate',
    requestStream: false,
    responseStream: false,
    requestType: client_auth_auth_pb.ActivateRequest,
    responseType: client_auth_auth_pb.ActivateResponse,
    requestSerialize: serialize_auth_ActivateRequest,
    requestDeserialize: deserialize_auth_ActivateRequest,
    responseSerialize: serialize_auth_ActivateResponse,
    responseDeserialize: deserialize_auth_ActivateResponse,
  },
  deactivate: {
    path: '/auth.API/Deactivate',
    requestStream: false,
    responseStream: false,
    requestType: client_auth_auth_pb.DeactivateRequest,
    responseType: client_auth_auth_pb.DeactivateResponse,
    requestSerialize: serialize_auth_DeactivateRequest,
    requestDeserialize: deserialize_auth_DeactivateRequest,
    responseSerialize: serialize_auth_DeactivateResponse,
    responseDeserialize: deserialize_auth_DeactivateResponse,
  },
  getConfiguration: {
    path: '/auth.API/GetConfiguration',
    requestStream: false,
    responseStream: false,
    requestType: client_auth_auth_pb.GetConfigurationRequest,
    responseType: client_auth_auth_pb.GetConfigurationResponse,
    requestSerialize: serialize_auth_GetConfigurationRequest,
    requestDeserialize: deserialize_auth_GetConfigurationRequest,
    responseSerialize: serialize_auth_GetConfigurationResponse,
    responseDeserialize: deserialize_auth_GetConfigurationResponse,
  },
  setConfiguration: {
    path: '/auth.API/SetConfiguration',
    requestStream: false,
    responseStream: false,
    requestType: client_auth_auth_pb.SetConfigurationRequest,
    responseType: client_auth_auth_pb.SetConfigurationResponse,
    requestSerialize: serialize_auth_SetConfigurationRequest,
    requestDeserialize: deserialize_auth_SetConfigurationRequest,
    responseSerialize: serialize_auth_SetConfigurationResponse,
    responseDeserialize: deserialize_auth_SetConfigurationResponse,
  },
  // Deprecated. GetAdmins returns the current list of cluster super admins
getAdmins: {
    path: '/auth.API/GetAdmins',
    requestStream: false,
    responseStream: false,
    requestType: client_auth_auth_pb.GetAdminsRequest,
    responseType: client_auth_auth_pb.GetAdminsResponse,
    requestSerialize: serialize_auth_GetAdminsRequest,
    requestDeserialize: deserialize_auth_GetAdminsRequest,
    responseSerialize: serialize_auth_GetAdminsResponse,
    responseDeserialize: deserialize_auth_GetAdminsResponse,
  },
  // Deprecated. ModifyAdmins adds or removes super admins from the cluster
modifyAdmins: {
    path: '/auth.API/ModifyAdmins',
    requestStream: false,
    responseStream: false,
    requestType: client_auth_auth_pb.ModifyAdminsRequest,
    responseType: client_auth_auth_pb.ModifyAdminsResponse,
    requestSerialize: serialize_auth_ModifyAdminsRequest,
    requestDeserialize: deserialize_auth_ModifyAdminsRequest,
    responseSerialize: serialize_auth_ModifyAdminsResponse,
    responseDeserialize: deserialize_auth_ModifyAdminsResponse,
  },
  // GetClusterRoleBindings returns the current set of cluster role bindings
getClusterRoleBindings: {
    path: '/auth.API/GetClusterRoleBindings',
    requestStream: false,
    responseStream: false,
    requestType: client_auth_auth_pb.GetClusterRoleBindingsRequest,
    responseType: client_auth_auth_pb.GetClusterRoleBindingsResponse,
    requestSerialize: serialize_auth_GetClusterRoleBindingsRequest,
    requestDeserialize: deserialize_auth_GetClusterRoleBindingsRequest,
    responseSerialize: serialize_auth_GetClusterRoleBindingsResponse,
    responseDeserialize: deserialize_auth_GetClusterRoleBindingsResponse,
  },
  // ModifyAdmin sets the list of admin roles for a principal
modifyClusterRoleBinding: {
    path: '/auth.API/ModifyClusterRoleBinding',
    requestStream: false,
    responseStream: false,
    requestType: client_auth_auth_pb.ModifyClusterRoleBindingRequest,
    responseType: client_auth_auth_pb.ModifyClusterRoleBindingResponse,
    requestSerialize: serialize_auth_ModifyClusterRoleBindingRequest,
    requestDeserialize: deserialize_auth_ModifyClusterRoleBindingRequest,
    responseSerialize: serialize_auth_ModifyClusterRoleBindingResponse,
    responseDeserialize: deserialize_auth_ModifyClusterRoleBindingResponse,
  },
  authenticate: {
    path: '/auth.API/Authenticate',
    requestStream: false,
    responseStream: false,
    requestType: client_auth_auth_pb.AuthenticateRequest,
    responseType: client_auth_auth_pb.AuthenticateResponse,
    requestSerialize: serialize_auth_AuthenticateRequest,
    requestDeserialize: deserialize_auth_AuthenticateRequest,
    responseSerialize: serialize_auth_AuthenticateResponse,
    responseDeserialize: deserialize_auth_AuthenticateResponse,
  },
  authorize: {
    path: '/auth.API/Authorize',
    requestStream: false,
    responseStream: false,
    requestType: client_auth_auth_pb.AuthorizeRequest,
    responseType: client_auth_auth_pb.AuthorizeResponse,
    requestSerialize: serialize_auth_AuthorizeRequest,
    requestDeserialize: deserialize_auth_AuthorizeRequest,
    responseSerialize: serialize_auth_AuthorizeResponse,
    responseDeserialize: deserialize_auth_AuthorizeResponse,
  },
  whoAmI: {
    path: '/auth.API/WhoAmI',
    requestStream: false,
    responseStream: false,
    requestType: client_auth_auth_pb.WhoAmIRequest,
    responseType: client_auth_auth_pb.WhoAmIResponse,
    requestSerialize: serialize_auth_WhoAmIRequest,
    requestDeserialize: deserialize_auth_WhoAmIRequest,
    responseSerialize: serialize_auth_WhoAmIResponse,
    responseDeserialize: deserialize_auth_WhoAmIResponse,
  },
  getScope: {
    path: '/auth.API/GetScope',
    requestStream: false,
    responseStream: false,
    requestType: client_auth_auth_pb.GetScopeRequest,
    responseType: client_auth_auth_pb.GetScopeResponse,
    requestSerialize: serialize_auth_GetScopeRequest,
    requestDeserialize: deserialize_auth_GetScopeRequest,
    responseSerialize: serialize_auth_GetScopeResponse,
    responseDeserialize: deserialize_auth_GetScopeResponse,
  },
  setScope: {
    path: '/auth.API/SetScope',
    requestStream: false,
    responseStream: false,
    requestType: client_auth_auth_pb.SetScopeRequest,
    responseType: client_auth_auth_pb.SetScopeResponse,
    requestSerialize: serialize_auth_SetScopeRequest,
    requestDeserialize: deserialize_auth_SetScopeRequest,
    responseSerialize: serialize_auth_SetScopeResponse,
    responseDeserialize: deserialize_auth_SetScopeResponse,
  },
  getACL: {
    path: '/auth.API/GetACL',
    requestStream: false,
    responseStream: false,
    requestType: client_auth_auth_pb.GetACLRequest,
    responseType: client_auth_auth_pb.GetACLResponse,
    requestSerialize: serialize_auth_GetACLRequest,
    requestDeserialize: deserialize_auth_GetACLRequest,
    responseSerialize: serialize_auth_GetACLResponse,
    responseDeserialize: deserialize_auth_GetACLResponse,
  },
  setACL: {
    path: '/auth.API/SetACL',
    requestStream: false,
    responseStream: false,
    requestType: client_auth_auth_pb.SetACLRequest,
    responseType: client_auth_auth_pb.SetACLResponse,
    requestSerialize: serialize_auth_SetACLRequest,
    requestDeserialize: deserialize_auth_SetACLRequest,
    responseSerialize: serialize_auth_SetACLResponse,
    responseDeserialize: deserialize_auth_SetACLResponse,
  },
  getOIDCLogin: {
    path: '/auth.API/GetOIDCLogin',
    requestStream: false,
    responseStream: false,
    requestType: client_auth_auth_pb.GetOIDCLoginRequest,
    responseType: client_auth_auth_pb.GetOIDCLoginResponse,
    requestSerialize: serialize_auth_GetOIDCLoginRequest,
    requestDeserialize: deserialize_auth_GetOIDCLoginRequest,
    responseSerialize: serialize_auth_GetOIDCLoginResponse,
    responseDeserialize: deserialize_auth_GetOIDCLoginResponse,
  },
  getAuthToken: {
    path: '/auth.API/GetAuthToken',
    requestStream: false,
    responseStream: false,
    requestType: client_auth_auth_pb.GetAuthTokenRequest,
    responseType: client_auth_auth_pb.GetAuthTokenResponse,
    requestSerialize: serialize_auth_GetAuthTokenRequest,
    requestDeserialize: deserialize_auth_GetAuthTokenRequest,
    responseSerialize: serialize_auth_GetAuthTokenResponse,
    responseDeserialize: deserialize_auth_GetAuthTokenResponse,
  },
  extendAuthToken: {
    path: '/auth.API/ExtendAuthToken',
    requestStream: false,
    responseStream: false,
    requestType: client_auth_auth_pb.ExtendAuthTokenRequest,
    responseType: client_auth_auth_pb.ExtendAuthTokenResponse,
    requestSerialize: serialize_auth_ExtendAuthTokenRequest,
    requestDeserialize: deserialize_auth_ExtendAuthTokenRequest,
    responseSerialize: serialize_auth_ExtendAuthTokenResponse,
    responseDeserialize: deserialize_auth_ExtendAuthTokenResponse,
  },
  revokeAuthToken: {
    path: '/auth.API/RevokeAuthToken',
    requestStream: false,
    responseStream: false,
    requestType: client_auth_auth_pb.RevokeAuthTokenRequest,
    responseType: client_auth_auth_pb.RevokeAuthTokenResponse,
    requestSerialize: serialize_auth_RevokeAuthTokenRequest,
    requestDeserialize: deserialize_auth_RevokeAuthTokenRequest,
    responseSerialize: serialize_auth_RevokeAuthTokenResponse,
    responseDeserialize: deserialize_auth_RevokeAuthTokenResponse,
  },
  setGroupsForUser: {
    path: '/auth.API/SetGroupsForUser',
    requestStream: false,
    responseStream: false,
    requestType: client_auth_auth_pb.SetGroupsForUserRequest,
    responseType: client_auth_auth_pb.SetGroupsForUserResponse,
    requestSerialize: serialize_auth_SetGroupsForUserRequest,
    requestDeserialize: deserialize_auth_SetGroupsForUserRequest,
    responseSerialize: serialize_auth_SetGroupsForUserResponse,
    responseDeserialize: deserialize_auth_SetGroupsForUserResponse,
  },
  modifyMembers: {
    path: '/auth.API/ModifyMembers',
    requestStream: false,
    responseStream: false,
    requestType: client_auth_auth_pb.ModifyMembersRequest,
    responseType: client_auth_auth_pb.ModifyMembersResponse,
    requestSerialize: serialize_auth_ModifyMembersRequest,
    requestDeserialize: deserialize_auth_ModifyMembersRequest,
    responseSerialize: serialize_auth_ModifyMembersResponse,
    responseDeserialize: deserialize_auth_ModifyMembersResponse,
  },
  getGroups: {
    path: '/auth.API/GetGroups',
    requestStream: false,
    responseStream: false,
    requestType: client_auth_auth_pb.GetGroupsRequest,
    responseType: client_auth_auth_pb.GetGroupsResponse,
    requestSerialize: serialize_auth_GetGroupsRequest,
    requestDeserialize: deserialize_auth_GetGroupsRequest,
    responseSerialize: serialize_auth_GetGroupsResponse,
    responseDeserialize: deserialize_auth_GetGroupsResponse,
  },
  getUsers: {
    path: '/auth.API/GetUsers',
    requestStream: false,
    responseStream: false,
    requestType: client_auth_auth_pb.GetUsersRequest,
    responseType: client_auth_auth_pb.GetUsersResponse,
    requestSerialize: serialize_auth_GetUsersRequest,
    requestDeserialize: deserialize_auth_GetUsersRequest,
    responseSerialize: serialize_auth_GetUsersResponse,
    responseDeserialize: deserialize_auth_GetUsersResponse,
  },
  getOneTimePassword: {
    path: '/auth.API/GetOneTimePassword',
    requestStream: false,
    responseStream: false,
    requestType: client_auth_auth_pb.GetOneTimePasswordRequest,
    responseType: client_auth_auth_pb.GetOneTimePasswordResponse,
    requestSerialize: serialize_auth_GetOneTimePasswordRequest,
    requestDeserialize: deserialize_auth_GetOneTimePasswordRequest,
    responseSerialize: serialize_auth_GetOneTimePasswordResponse,
    responseDeserialize: deserialize_auth_GetOneTimePasswordResponse,
  },
};

exports.APIClient = grpc.makeGenericClientConstructor(APIService);
