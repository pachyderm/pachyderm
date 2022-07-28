// GENERATED CODE -- DO NOT EDIT!

'use strict';
var grpc = require('@grpc/grpc-js');
var license_license_pb = require('../license/license_pb.js');
var google_protobuf_timestamp_pb = require('google-protobuf/google/protobuf/timestamp_pb.js');
var gogoproto_gogo_pb = require('../gogoproto/gogo_pb.js');
var enterprise_enterprise_pb = require('../enterprise/enterprise_pb.js');

function serialize_license_v2_ActivateRequest(arg) {
  if (!(arg instanceof license_license_pb.ActivateRequest)) {
    throw new Error('Expected argument of type license_v2.ActivateRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_license_v2_ActivateRequest(buffer_arg) {
  return license_license_pb.ActivateRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_license_v2_ActivateResponse(arg) {
  if (!(arg instanceof license_license_pb.ActivateResponse)) {
    throw new Error('Expected argument of type license_v2.ActivateResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_license_v2_ActivateResponse(buffer_arg) {
  return license_license_pb.ActivateResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_license_v2_AddClusterRequest(arg) {
  if (!(arg instanceof license_license_pb.AddClusterRequest)) {
    throw new Error('Expected argument of type license_v2.AddClusterRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_license_v2_AddClusterRequest(buffer_arg) {
  return license_license_pb.AddClusterRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_license_v2_AddClusterResponse(arg) {
  if (!(arg instanceof license_license_pb.AddClusterResponse)) {
    throw new Error('Expected argument of type license_v2.AddClusterResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_license_v2_AddClusterResponse(buffer_arg) {
  return license_license_pb.AddClusterResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_license_v2_DeleteAllRequest(arg) {
  if (!(arg instanceof license_license_pb.DeleteAllRequest)) {
    throw new Error('Expected argument of type license_v2.DeleteAllRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_license_v2_DeleteAllRequest(buffer_arg) {
  return license_license_pb.DeleteAllRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_license_v2_DeleteAllResponse(arg) {
  if (!(arg instanceof license_license_pb.DeleteAllResponse)) {
    throw new Error('Expected argument of type license_v2.DeleteAllResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_license_v2_DeleteAllResponse(buffer_arg) {
  return license_license_pb.DeleteAllResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_license_v2_DeleteClusterRequest(arg) {
  if (!(arg instanceof license_license_pb.DeleteClusterRequest)) {
    throw new Error('Expected argument of type license_v2.DeleteClusterRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_license_v2_DeleteClusterRequest(buffer_arg) {
  return license_license_pb.DeleteClusterRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_license_v2_DeleteClusterResponse(arg) {
  if (!(arg instanceof license_license_pb.DeleteClusterResponse)) {
    throw new Error('Expected argument of type license_v2.DeleteClusterResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_license_v2_DeleteClusterResponse(buffer_arg) {
  return license_license_pb.DeleteClusterResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_license_v2_GetActivationCodeRequest(arg) {
  if (!(arg instanceof license_license_pb.GetActivationCodeRequest)) {
    throw new Error('Expected argument of type license_v2.GetActivationCodeRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_license_v2_GetActivationCodeRequest(buffer_arg) {
  return license_license_pb.GetActivationCodeRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_license_v2_GetActivationCodeResponse(arg) {
  if (!(arg instanceof license_license_pb.GetActivationCodeResponse)) {
    throw new Error('Expected argument of type license_v2.GetActivationCodeResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_license_v2_GetActivationCodeResponse(buffer_arg) {
  return license_license_pb.GetActivationCodeResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_license_v2_HeartbeatRequest(arg) {
  if (!(arg instanceof license_license_pb.HeartbeatRequest)) {
    throw new Error('Expected argument of type license_v2.HeartbeatRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_license_v2_HeartbeatRequest(buffer_arg) {
  return license_license_pb.HeartbeatRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_license_v2_HeartbeatResponse(arg) {
  if (!(arg instanceof license_license_pb.HeartbeatResponse)) {
    throw new Error('Expected argument of type license_v2.HeartbeatResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_license_v2_HeartbeatResponse(buffer_arg) {
  return license_license_pb.HeartbeatResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_license_v2_ListClustersRequest(arg) {
  if (!(arg instanceof license_license_pb.ListClustersRequest)) {
    throw new Error('Expected argument of type license_v2.ListClustersRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_license_v2_ListClustersRequest(buffer_arg) {
  return license_license_pb.ListClustersRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_license_v2_ListClustersResponse(arg) {
  if (!(arg instanceof license_license_pb.ListClustersResponse)) {
    throw new Error('Expected argument of type license_v2.ListClustersResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_license_v2_ListClustersResponse(buffer_arg) {
  return license_license_pb.ListClustersResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_license_v2_ListUserClustersRequest(arg) {
  if (!(arg instanceof license_license_pb.ListUserClustersRequest)) {
    throw new Error('Expected argument of type license_v2.ListUserClustersRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_license_v2_ListUserClustersRequest(buffer_arg) {
  return license_license_pb.ListUserClustersRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_license_v2_ListUserClustersResponse(arg) {
  if (!(arg instanceof license_license_pb.ListUserClustersResponse)) {
    throw new Error('Expected argument of type license_v2.ListUserClustersResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_license_v2_ListUserClustersResponse(buffer_arg) {
  return license_license_pb.ListUserClustersResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_license_v2_UpdateClusterRequest(arg) {
  if (!(arg instanceof license_license_pb.UpdateClusterRequest)) {
    throw new Error('Expected argument of type license_v2.UpdateClusterRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_license_v2_UpdateClusterRequest(buffer_arg) {
  return license_license_pb.UpdateClusterRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_license_v2_UpdateClusterResponse(arg) {
  if (!(arg instanceof license_license_pb.UpdateClusterResponse)) {
    throw new Error('Expected argument of type license_v2.UpdateClusterResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_license_v2_UpdateClusterResponse(buffer_arg) {
  return license_license_pb.UpdateClusterResponse.deserializeBinary(new Uint8Array(buffer_arg));
}


var APIService = exports.APIService = {
  // Activate enables the license service by setting the enterprise activation
// code to serve.
activate: {
    path: '/license_v2.API/Activate',
    requestStream: false,
    responseStream: false,
    requestType: license_license_pb.ActivateRequest,
    responseType: license_license_pb.ActivateResponse,
    requestSerialize: serialize_license_v2_ActivateRequest,
    requestDeserialize: deserialize_license_v2_ActivateRequest,
    responseSerialize: serialize_license_v2_ActivateResponse,
    responseDeserialize: deserialize_license_v2_ActivateResponse,
  },
  getActivationCode: {
    path: '/license_v2.API/GetActivationCode',
    requestStream: false,
    responseStream: false,
    requestType: license_license_pb.GetActivationCodeRequest,
    responseType: license_license_pb.GetActivationCodeResponse,
    requestSerialize: serialize_license_v2_GetActivationCodeRequest,
    requestDeserialize: deserialize_license_v2_GetActivationCodeRequest,
    responseSerialize: serialize_license_v2_GetActivationCodeResponse,
    responseDeserialize: deserialize_license_v2_GetActivationCodeResponse,
  },
  // DeleteAll deactivates the server and removes all data.
deleteAll: {
    path: '/license_v2.API/DeleteAll',
    requestStream: false,
    responseStream: false,
    requestType: license_license_pb.DeleteAllRequest,
    responseType: license_license_pb.DeleteAllResponse,
    requestSerialize: serialize_license_v2_DeleteAllRequest,
    requestDeserialize: deserialize_license_v2_DeleteAllRequest,
    responseSerialize: serialize_license_v2_DeleteAllResponse,
    responseDeserialize: deserialize_license_v2_DeleteAllResponse,
  },
  // CRUD operations for the pachds registered with this server.
addCluster: {
    path: '/license_v2.API/AddCluster',
    requestStream: false,
    responseStream: false,
    requestType: license_license_pb.AddClusterRequest,
    responseType: license_license_pb.AddClusterResponse,
    requestSerialize: serialize_license_v2_AddClusterRequest,
    requestDeserialize: deserialize_license_v2_AddClusterRequest,
    responseSerialize: serialize_license_v2_AddClusterResponse,
    responseDeserialize: deserialize_license_v2_AddClusterResponse,
  },
  deleteCluster: {
    path: '/license_v2.API/DeleteCluster',
    requestStream: false,
    responseStream: false,
    requestType: license_license_pb.DeleteClusterRequest,
    responseType: license_license_pb.DeleteClusterResponse,
    requestSerialize: serialize_license_v2_DeleteClusterRequest,
    requestDeserialize: deserialize_license_v2_DeleteClusterRequest,
    responseSerialize: serialize_license_v2_DeleteClusterResponse,
    responseDeserialize: deserialize_license_v2_DeleteClusterResponse,
  },
  listClusters: {
    path: '/license_v2.API/ListClusters',
    requestStream: false,
    responseStream: false,
    requestType: license_license_pb.ListClustersRequest,
    responseType: license_license_pb.ListClustersResponse,
    requestSerialize: serialize_license_v2_ListClustersRequest,
    requestDeserialize: deserialize_license_v2_ListClustersRequest,
    responseSerialize: serialize_license_v2_ListClustersResponse,
    responseDeserialize: deserialize_license_v2_ListClustersResponse,
  },
  updateCluster: {
    path: '/license_v2.API/UpdateCluster',
    requestStream: false,
    responseStream: false,
    requestType: license_license_pb.UpdateClusterRequest,
    responseType: license_license_pb.UpdateClusterResponse,
    requestSerialize: serialize_license_v2_UpdateClusterRequest,
    requestDeserialize: deserialize_license_v2_UpdateClusterRequest,
    responseSerialize: serialize_license_v2_UpdateClusterResponse,
    responseDeserialize: deserialize_license_v2_UpdateClusterResponse,
  },
  // Heartbeat is the RPC registered pachds make to the license server
// to communicate their status and fetch updates.
heartbeat: {
    path: '/license_v2.API/Heartbeat',
    requestStream: false,
    responseStream: false,
    requestType: license_license_pb.HeartbeatRequest,
    responseType: license_license_pb.HeartbeatResponse,
    requestSerialize: serialize_license_v2_HeartbeatRequest,
    requestDeserialize: deserialize_license_v2_HeartbeatRequest,
    responseSerialize: serialize_license_v2_HeartbeatResponse,
    responseDeserialize: deserialize_license_v2_HeartbeatResponse,
  },
  // Lists all clusters available to user
listUserClusters: {
    path: '/license_v2.API/ListUserClusters',
    requestStream: false,
    responseStream: false,
    requestType: license_license_pb.ListUserClustersRequest,
    responseType: license_license_pb.ListUserClustersResponse,
    requestSerialize: serialize_license_v2_ListUserClustersRequest,
    requestDeserialize: deserialize_license_v2_ListUserClustersRequest,
    responseSerialize: serialize_license_v2_ListUserClustersResponse,
    responseDeserialize: deserialize_license_v2_ListUserClustersResponse,
  },
};

exports.APIClient = grpc.makeGenericClientConstructor(APIService);
