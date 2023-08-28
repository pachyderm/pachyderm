// GENERATED CODE -- DO NOT EDIT!

'use strict';
var grpc = require('@grpc/grpc-js');
var admin_admin_pb = require('../admin/admin_pb.js');
var version_versionpb_version_pb = require('../version/versionpb/version_pb.js');
var pfs_pfs_pb = require('../pfs/pfs_pb.js');

function serialize_admin_v2_ClusterInfo(arg) {
  if (!(arg instanceof admin_admin_pb.ClusterInfo)) {
    throw new Error('Expected argument of type admin_v2.ClusterInfo');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_admin_v2_ClusterInfo(buffer_arg) {
  return admin_admin_pb.ClusterInfo.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_admin_v2_InspectClusterRequest(arg) {
  if (!(arg instanceof admin_admin_pb.InspectClusterRequest)) {
    throw new Error('Expected argument of type admin_v2.InspectClusterRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_admin_v2_InspectClusterRequest(buffer_arg) {
  return admin_admin_pb.InspectClusterRequest.deserializeBinary(new Uint8Array(buffer_arg));
}


var APIService = exports.APIService = {
  inspectCluster: {
    path: '/admin_v2.API/InspectCluster',
    requestStream: false,
    responseStream: false,
    requestType: admin_admin_pb.InspectClusterRequest,
    responseType: admin_admin_pb.ClusterInfo,
    requestSerialize: serialize_admin_v2_InspectClusterRequest,
    requestDeserialize: deserialize_admin_v2_InspectClusterRequest,
    responseSerialize: serialize_admin_v2_ClusterInfo,
    responseDeserialize: deserialize_admin_v2_ClusterInfo,
  },
};

exports.APIClient = grpc.makeGenericClientConstructor(APIService);
