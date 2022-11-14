// GENERATED CODE -- DO NOT EDIT!

'use strict';
var grpc = require('@grpc/grpc-js');
var enterprise_enterprise_pb = require('../enterprise/enterprise_pb.js');
var google_protobuf_timestamp_pb = require('google-protobuf/google/protobuf/timestamp_pb.js');

function serialize_enterprise_v2_ActivateRequest(arg) {
  if (!(arg instanceof enterprise_enterprise_pb.ActivateRequest)) {
    throw new Error('Expected argument of type enterprise_v2.ActivateRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_enterprise_v2_ActivateRequest(buffer_arg) {
  return enterprise_enterprise_pb.ActivateRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_enterprise_v2_ActivateResponse(arg) {
  if (!(arg instanceof enterprise_enterprise_pb.ActivateResponse)) {
    throw new Error('Expected argument of type enterprise_v2.ActivateResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_enterprise_v2_ActivateResponse(buffer_arg) {
  return enterprise_enterprise_pb.ActivateResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_enterprise_v2_DeactivateRequest(arg) {
  if (!(arg instanceof enterprise_enterprise_pb.DeactivateRequest)) {
    throw new Error('Expected argument of type enterprise_v2.DeactivateRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_enterprise_v2_DeactivateRequest(buffer_arg) {
  return enterprise_enterprise_pb.DeactivateRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_enterprise_v2_DeactivateResponse(arg) {
  if (!(arg instanceof enterprise_enterprise_pb.DeactivateResponse)) {
    throw new Error('Expected argument of type enterprise_v2.DeactivateResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_enterprise_v2_DeactivateResponse(buffer_arg) {
  return enterprise_enterprise_pb.DeactivateResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_enterprise_v2_GetActivationCodeRequest(arg) {
  if (!(arg instanceof enterprise_enterprise_pb.GetActivationCodeRequest)) {
    throw new Error('Expected argument of type enterprise_v2.GetActivationCodeRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_enterprise_v2_GetActivationCodeRequest(buffer_arg) {
  return enterprise_enterprise_pb.GetActivationCodeRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_enterprise_v2_GetActivationCodeResponse(arg) {
  if (!(arg instanceof enterprise_enterprise_pb.GetActivationCodeResponse)) {
    throw new Error('Expected argument of type enterprise_v2.GetActivationCodeResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_enterprise_v2_GetActivationCodeResponse(buffer_arg) {
  return enterprise_enterprise_pb.GetActivationCodeResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_enterprise_v2_GetStateRequest(arg) {
  if (!(arg instanceof enterprise_enterprise_pb.GetStateRequest)) {
    throw new Error('Expected argument of type enterprise_v2.GetStateRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_enterprise_v2_GetStateRequest(buffer_arg) {
  return enterprise_enterprise_pb.GetStateRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_enterprise_v2_GetStateResponse(arg) {
  if (!(arg instanceof enterprise_enterprise_pb.GetStateResponse)) {
    throw new Error('Expected argument of type enterprise_v2.GetStateResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_enterprise_v2_GetStateResponse(buffer_arg) {
  return enterprise_enterprise_pb.GetStateResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_enterprise_v2_HeartbeatRequest(arg) {
  if (!(arg instanceof enterprise_enterprise_pb.HeartbeatRequest)) {
    throw new Error('Expected argument of type enterprise_v2.HeartbeatRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_enterprise_v2_HeartbeatRequest(buffer_arg) {
  return enterprise_enterprise_pb.HeartbeatRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_enterprise_v2_HeartbeatResponse(arg) {
  if (!(arg instanceof enterprise_enterprise_pb.HeartbeatResponse)) {
    throw new Error('Expected argument of type enterprise_v2.HeartbeatResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_enterprise_v2_HeartbeatResponse(buffer_arg) {
  return enterprise_enterprise_pb.HeartbeatResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_enterprise_v2_PauseRequest(arg) {
  if (!(arg instanceof enterprise_enterprise_pb.PauseRequest)) {
    throw new Error('Expected argument of type enterprise_v2.PauseRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_enterprise_v2_PauseRequest(buffer_arg) {
  return enterprise_enterprise_pb.PauseRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_enterprise_v2_PauseResponse(arg) {
  if (!(arg instanceof enterprise_enterprise_pb.PauseResponse)) {
    throw new Error('Expected argument of type enterprise_v2.PauseResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_enterprise_v2_PauseResponse(buffer_arg) {
  return enterprise_enterprise_pb.PauseResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_enterprise_v2_PauseStatusRequest(arg) {
  if (!(arg instanceof enterprise_enterprise_pb.PauseStatusRequest)) {
    throw new Error('Expected argument of type enterprise_v2.PauseStatusRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_enterprise_v2_PauseStatusRequest(buffer_arg) {
  return enterprise_enterprise_pb.PauseStatusRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_enterprise_v2_PauseStatusResponse(arg) {
  if (!(arg instanceof enterprise_enterprise_pb.PauseStatusResponse)) {
    throw new Error('Expected argument of type enterprise_v2.PauseStatusResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_enterprise_v2_PauseStatusResponse(buffer_arg) {
  return enterprise_enterprise_pb.PauseStatusResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_enterprise_v2_UnpauseRequest(arg) {
  if (!(arg instanceof enterprise_enterprise_pb.UnpauseRequest)) {
    throw new Error('Expected argument of type enterprise_v2.UnpauseRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_enterprise_v2_UnpauseRequest(buffer_arg) {
  return enterprise_enterprise_pb.UnpauseRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_enterprise_v2_UnpauseResponse(arg) {
  if (!(arg instanceof enterprise_enterprise_pb.UnpauseResponse)) {
    throw new Error('Expected argument of type enterprise_v2.UnpauseResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_enterprise_v2_UnpauseResponse(buffer_arg) {
  return enterprise_enterprise_pb.UnpauseResponse.deserializeBinary(new Uint8Array(buffer_arg));
}


var APIService = exports.APIService = {
  // Provide a Pachyderm enterprise token, enabling Pachyderm enterprise
// features, such as the Pachyderm Dashboard and Auth system
activate: {
    path: '/enterprise_v2.API/Activate',
    requestStream: false,
    responseStream: false,
    requestType: enterprise_enterprise_pb.ActivateRequest,
    responseType: enterprise_enterprise_pb.ActivateResponse,
    requestSerialize: serialize_enterprise_v2_ActivateRequest,
    requestDeserialize: deserialize_enterprise_v2_ActivateRequest,
    responseSerialize: serialize_enterprise_v2_ActivateResponse,
    responseDeserialize: deserialize_enterprise_v2_ActivateResponse,
  },
  getState: {
    path: '/enterprise_v2.API/GetState',
    requestStream: false,
    responseStream: false,
    requestType: enterprise_enterprise_pb.GetStateRequest,
    responseType: enterprise_enterprise_pb.GetStateResponse,
    requestSerialize: serialize_enterprise_v2_GetStateRequest,
    requestDeserialize: deserialize_enterprise_v2_GetStateRequest,
    responseSerialize: serialize_enterprise_v2_GetStateResponse,
    responseDeserialize: deserialize_enterprise_v2_GetStateResponse,
  },
  getActivationCode: {
    path: '/enterprise_v2.API/GetActivationCode',
    requestStream: false,
    responseStream: false,
    requestType: enterprise_enterprise_pb.GetActivationCodeRequest,
    responseType: enterprise_enterprise_pb.GetActivationCodeResponse,
    requestSerialize: serialize_enterprise_v2_GetActivationCodeRequest,
    requestDeserialize: deserialize_enterprise_v2_GetActivationCodeRequest,
    responseSerialize: serialize_enterprise_v2_GetActivationCodeResponse,
    responseDeserialize: deserialize_enterprise_v2_GetActivationCodeResponse,
  },
  // Heartbeat is used in testing to trigger a heartbeat on demand. Normally this happens
// on a timer.
heartbeat: {
    path: '/enterprise_v2.API/Heartbeat',
    requestStream: false,
    responseStream: false,
    requestType: enterprise_enterprise_pb.HeartbeatRequest,
    responseType: enterprise_enterprise_pb.HeartbeatResponse,
    requestSerialize: serialize_enterprise_v2_HeartbeatRequest,
    requestDeserialize: deserialize_enterprise_v2_HeartbeatRequest,
    responseSerialize: serialize_enterprise_v2_HeartbeatResponse,
    responseDeserialize: deserialize_enterprise_v2_HeartbeatResponse,
  },
  // Deactivate removes a cluster's enterprise activation
// token and sets its enterprise state to NONE.
deactivate: {
    path: '/enterprise_v2.API/Deactivate',
    requestStream: false,
    responseStream: false,
    requestType: enterprise_enterprise_pb.DeactivateRequest,
    responseType: enterprise_enterprise_pb.DeactivateResponse,
    requestSerialize: serialize_enterprise_v2_DeactivateRequest,
    requestDeserialize: deserialize_enterprise_v2_DeactivateRequest,
    responseSerialize: serialize_enterprise_v2_DeactivateResponse,
    responseDeserialize: deserialize_enterprise_v2_DeactivateResponse,
  },
  // Pause pauses the cluster.
pause: {
    path: '/enterprise_v2.API/Pause',
    requestStream: false,
    responseStream: false,
    requestType: enterprise_enterprise_pb.PauseRequest,
    responseType: enterprise_enterprise_pb.PauseResponse,
    requestSerialize: serialize_enterprise_v2_PauseRequest,
    requestDeserialize: deserialize_enterprise_v2_PauseRequest,
    responseSerialize: serialize_enterprise_v2_PauseResponse,
    responseDeserialize: deserialize_enterprise_v2_PauseResponse,
  },
  // Unpause unpauses the cluser.
unpause: {
    path: '/enterprise_v2.API/Unpause',
    requestStream: false,
    responseStream: false,
    requestType: enterprise_enterprise_pb.UnpauseRequest,
    responseType: enterprise_enterprise_pb.UnpauseResponse,
    requestSerialize: serialize_enterprise_v2_UnpauseRequest,
    requestDeserialize: deserialize_enterprise_v2_UnpauseRequest,
    responseSerialize: serialize_enterprise_v2_UnpauseResponse,
    responseDeserialize: deserialize_enterprise_v2_UnpauseResponse,
  },
  pauseStatus: {
    path: '/enterprise_v2.API/PauseStatus',
    requestStream: false,
    responseStream: false,
    requestType: enterprise_enterprise_pb.PauseStatusRequest,
    responseType: enterprise_enterprise_pb.PauseStatusResponse,
    requestSerialize: serialize_enterprise_v2_PauseStatusRequest,
    requestDeserialize: deserialize_enterprise_v2_PauseStatusRequest,
    responseSerialize: serialize_enterprise_v2_PauseStatusResponse,
    responseDeserialize: deserialize_enterprise_v2_PauseStatusResponse,
  },
};

exports.APIClient = grpc.makeGenericClientConstructor(APIService);
