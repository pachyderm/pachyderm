// GENERATED CODE -- DO NOT EDIT!

'use strict';
var grpc = require('@grpc/grpc-js');
var mock_mock_pb = require('../mock/mock_pb.js');
var google_protobuf_empty_pb = require('google-protobuf/google/protobuf/empty_pb.js');
var google_protobuf_timestamp_pb = require('google-protobuf/google/protobuf/timestamp_pb.js');

function serialize_google_protobuf_Empty(arg) {
  if (!(arg instanceof google_protobuf_empty_pb.Empty)) {
    throw new Error('Expected argument of type google.protobuf.Empty');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_google_protobuf_Empty(buffer_arg) {
  return google_protobuf_empty_pb.Empty.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_mock_Project(arg) {
  if (!(arg instanceof mock_mock_pb.Project)) {
    throw new Error('Expected argument of type mock.Project');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_mock_Project(buffer_arg) {
  return mock_mock_pb.Project.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_mock_ProjectRequest(arg) {
  if (!(arg instanceof mock_mock_pb.ProjectRequest)) {
    throw new Error('Expected argument of type mock.ProjectRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_mock_ProjectRequest(buffer_arg) {
  return mock_mock_pb.ProjectRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_mock_Projects(arg) {
  if (!(arg instanceof mock_mock_pb.Projects)) {
    throw new Error('Expected argument of type mock.Projects');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_mock_Projects(buffer_arg) {
  return mock_mock_pb.Projects.deserializeBinary(new Uint8Array(buffer_arg));
}


var APIService = exports.APIService = {
  inspectProject: {
    path: '/mock.API/InspectProject',
    requestStream: false,
    responseStream: false,
    requestType: mock_mock_pb.ProjectRequest,
    responseType: mock_mock_pb.Project,
    requestSerialize: serialize_mock_ProjectRequest,
    requestDeserialize: deserialize_mock_ProjectRequest,
    responseSerialize: serialize_mock_Project,
    responseDeserialize: deserialize_mock_Project,
  },
  listProject: {
    path: '/mock.API/ListProject',
    requestStream: false,
    responseStream: false,
    requestType: google_protobuf_empty_pb.Empty,
    responseType: mock_mock_pb.Projects,
    requestSerialize: serialize_google_protobuf_Empty,
    requestDeserialize: deserialize_google_protobuf_Empty,
    responseSerialize: serialize_mock_Projects,
    responseDeserialize: deserialize_mock_Projects,
  },
};

exports.APIClient = grpc.makeGenericClientConstructor(APIService);
