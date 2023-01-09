// GENERATED CODE -- DO NOT EDIT!

'use strict';
var grpc = require('@grpc/grpc-js');
var version_versionpb_version_pb = require('../../version/versionpb/version_pb.js');
var google_protobuf_empty_pb = require('google-protobuf/google/protobuf/empty_pb.js');

function serialize_google_protobuf_Empty(arg) {
  if (!(arg instanceof google_protobuf_empty_pb.Empty)) {
    throw new Error('Expected argument of type google.protobuf.Empty');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_google_protobuf_Empty(buffer_arg) {
  return google_protobuf_empty_pb.Empty.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_versionpb_v2_Version(arg) {
  if (!(arg instanceof version_versionpb_version_pb.Version)) {
    throw new Error('Expected argument of type versionpb_v2.Version');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_versionpb_v2_Version(buffer_arg) {
  return version_versionpb_version_pb.Version.deserializeBinary(new Uint8Array(buffer_arg));
}


var APIService = exports.APIService = {
  getVersion: {
    path: '/versionpb_v2.API/GetVersion',
    requestStream: false,
    responseStream: false,
    requestType: google_protobuf_empty_pb.Empty,
    responseType: version_versionpb_version_pb.Version,
    requestSerialize: serialize_google_protobuf_Empty,
    requestDeserialize: deserialize_google_protobuf_Empty,
    responseSerialize: serialize_versionpb_v2_Version,
    responseDeserialize: deserialize_versionpb_v2_Version,
  },
};

exports.APIClient = grpc.makeGenericClientConstructor(APIService);
