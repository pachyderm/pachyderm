// GENERATED CODE -- DO NOT EDIT!

'use strict';
var grpc = require('@grpc/grpc-js');
var admin_admin_pb = require('../admin/admin_pb.js');
var google_protobuf_empty_pb = require('google-protobuf/google/protobuf/empty_pb.js');
var gogoproto_gogo_pb = require('../gogoproto/gogo_pb.js');

function serialize_admin_v2_ClusterInfo(arg) {
  if (!(arg instanceof admin_admin_pb.ClusterInfo)) {
    throw new Error('Expected argument of type admin_v2.ClusterInfo');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_admin_v2_ClusterInfo(buffer_arg) {
  return admin_admin_pb.ClusterInfo.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_google_protobuf_Empty(arg) {
  if (!(arg instanceof google_protobuf_empty_pb.Empty)) {
    throw new Error('Expected argument of type google.protobuf.Empty');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_google_protobuf_Empty(buffer_arg) {
  return google_protobuf_empty_pb.Empty.deserializeBinary(new Uint8Array(buffer_arg));
}


var APIService = exports.APIService = {
  inspectCluster: {
    path: '/admin_v2.API/InspectCluster',
    requestStream: false,
    responseStream: false,
    requestType: google_protobuf_empty_pb.Empty,
    responseType: admin_admin_pb.ClusterInfo,
    requestSerialize: serialize_google_protobuf_Empty,
    requestDeserialize: deserialize_google_protobuf_Empty,
    responseSerialize: serialize_admin_v2_ClusterInfo,
    responseDeserialize: deserialize_admin_v2_ClusterInfo,
  },
};

exports.APIClient = grpc.makeGenericClientConstructor(APIService);
