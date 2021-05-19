// GENERATED CODE -- DO NOT EDIT!

'use strict';
var grpc = require('@grpc/grpc-js');
var pps_pps_pb = require('../pps/pps_pb.js');
var google_protobuf_empty_pb = require('google-protobuf/google/protobuf/empty_pb.js');
var google_protobuf_timestamp_pb = require('google-protobuf/google/protobuf/timestamp_pb.js');
var google_protobuf_duration_pb = require('google-protobuf/google/protobuf/duration_pb.js');
var gogoproto_gogo_pb = require('../gogoproto/gogo_pb.js');
var pfs_pfs_pb = require('../pfs/pfs_pb.js');

function serialize_google_protobuf_Empty(arg) {
  if (!(arg instanceof google_protobuf_empty_pb.Empty)) {
    throw new Error('Expected argument of type google.protobuf.Empty');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_google_protobuf_Empty(buffer_arg) {
  return google_protobuf_empty_pb.Empty.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pps_ActivateAuthRequest(arg) {
  if (!(arg instanceof pps_pps_pb.ActivateAuthRequest)) {
    throw new Error('Expected argument of type pps.ActivateAuthRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pps_ActivateAuthRequest(buffer_arg) {
  return pps_pps_pb.ActivateAuthRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pps_ActivateAuthResponse(arg) {
  if (!(arg instanceof pps_pps_pb.ActivateAuthResponse)) {
    throw new Error('Expected argument of type pps.ActivateAuthResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pps_ActivateAuthResponse(buffer_arg) {
  return pps_pps_pb.ActivateAuthResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pps_CreatePipelineJobRequest(arg) {
  if (!(arg instanceof pps_pps_pb.CreatePipelineJobRequest)) {
    throw new Error('Expected argument of type pps.CreatePipelineJobRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pps_CreatePipelineJobRequest(buffer_arg) {
  return pps_pps_pb.CreatePipelineJobRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pps_CreatePipelineRequest(arg) {
  if (!(arg instanceof pps_pps_pb.CreatePipelineRequest)) {
    throw new Error('Expected argument of type pps.CreatePipelineRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pps_CreatePipelineRequest(buffer_arg) {
  return pps_pps_pb.CreatePipelineRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pps_CreateSecretRequest(arg) {
  if (!(arg instanceof pps_pps_pb.CreateSecretRequest)) {
    throw new Error('Expected argument of type pps.CreateSecretRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pps_CreateSecretRequest(buffer_arg) {
  return pps_pps_pb.CreateSecretRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pps_DatumInfo(arg) {
  if (!(arg instanceof pps_pps_pb.DatumInfo)) {
    throw new Error('Expected argument of type pps.DatumInfo');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pps_DatumInfo(buffer_arg) {
  return pps_pps_pb.DatumInfo.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pps_DeletePipelineJobRequest(arg) {
  if (!(arg instanceof pps_pps_pb.DeletePipelineJobRequest)) {
    throw new Error('Expected argument of type pps.DeletePipelineJobRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pps_DeletePipelineJobRequest(buffer_arg) {
  return pps_pps_pb.DeletePipelineJobRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pps_DeletePipelineRequest(arg) {
  if (!(arg instanceof pps_pps_pb.DeletePipelineRequest)) {
    throw new Error('Expected argument of type pps.DeletePipelineRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pps_DeletePipelineRequest(buffer_arg) {
  return pps_pps_pb.DeletePipelineRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pps_DeleteSecretRequest(arg) {
  if (!(arg instanceof pps_pps_pb.DeleteSecretRequest)) {
    throw new Error('Expected argument of type pps.DeleteSecretRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pps_DeleteSecretRequest(buffer_arg) {
  return pps_pps_pb.DeleteSecretRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pps_FlushPipelineJobRequest(arg) {
  if (!(arg instanceof pps_pps_pb.FlushPipelineJobRequest)) {
    throw new Error('Expected argument of type pps.FlushPipelineJobRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pps_FlushPipelineJobRequest(buffer_arg) {
  return pps_pps_pb.FlushPipelineJobRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pps_GetLogsRequest(arg) {
  if (!(arg instanceof pps_pps_pb.GetLogsRequest)) {
    throw new Error('Expected argument of type pps.GetLogsRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pps_GetLogsRequest(buffer_arg) {
  return pps_pps_pb.GetLogsRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pps_InspectDatumRequest(arg) {
  if (!(arg instanceof pps_pps_pb.InspectDatumRequest)) {
    throw new Error('Expected argument of type pps.InspectDatumRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pps_InspectDatumRequest(buffer_arg) {
  return pps_pps_pb.InspectDatumRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pps_InspectPipelineJobRequest(arg) {
  if (!(arg instanceof pps_pps_pb.InspectPipelineJobRequest)) {
    throw new Error('Expected argument of type pps.InspectPipelineJobRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pps_InspectPipelineJobRequest(buffer_arg) {
  return pps_pps_pb.InspectPipelineJobRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pps_InspectPipelineRequest(arg) {
  if (!(arg instanceof pps_pps_pb.InspectPipelineRequest)) {
    throw new Error('Expected argument of type pps.InspectPipelineRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pps_InspectPipelineRequest(buffer_arg) {
  return pps_pps_pb.InspectPipelineRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pps_InspectSecretRequest(arg) {
  if (!(arg instanceof pps_pps_pb.InspectSecretRequest)) {
    throw new Error('Expected argument of type pps.InspectSecretRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pps_InspectSecretRequest(buffer_arg) {
  return pps_pps_pb.InspectSecretRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pps_ListDatumRequest(arg) {
  if (!(arg instanceof pps_pps_pb.ListDatumRequest)) {
    throw new Error('Expected argument of type pps.ListDatumRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pps_ListDatumRequest(buffer_arg) {
  return pps_pps_pb.ListDatumRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pps_ListPipelineJobRequest(arg) {
  if (!(arg instanceof pps_pps_pb.ListPipelineJobRequest)) {
    throw new Error('Expected argument of type pps.ListPipelineJobRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pps_ListPipelineJobRequest(buffer_arg) {
  return pps_pps_pb.ListPipelineJobRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pps_ListPipelineRequest(arg) {
  if (!(arg instanceof pps_pps_pb.ListPipelineRequest)) {
    throw new Error('Expected argument of type pps.ListPipelineRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pps_ListPipelineRequest(buffer_arg) {
  return pps_pps_pb.ListPipelineRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pps_LogMessage(arg) {
  if (!(arg instanceof pps_pps_pb.LogMessage)) {
    throw new Error('Expected argument of type pps.LogMessage');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pps_LogMessage(buffer_arg) {
  return pps_pps_pb.LogMessage.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pps_PipelineInfo(arg) {
  if (!(arg instanceof pps_pps_pb.PipelineInfo)) {
    throw new Error('Expected argument of type pps.PipelineInfo');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pps_PipelineInfo(buffer_arg) {
  return pps_pps_pb.PipelineInfo.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pps_PipelineInfos(arg) {
  if (!(arg instanceof pps_pps_pb.PipelineInfos)) {
    throw new Error('Expected argument of type pps.PipelineInfos');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pps_PipelineInfos(buffer_arg) {
  return pps_pps_pb.PipelineInfos.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pps_PipelineJob(arg) {
  if (!(arg instanceof pps_pps_pb.PipelineJob)) {
    throw new Error('Expected argument of type pps.PipelineJob');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pps_PipelineJob(buffer_arg) {
  return pps_pps_pb.PipelineJob.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pps_PipelineJobInfo(arg) {
  if (!(arg instanceof pps_pps_pb.PipelineJobInfo)) {
    throw new Error('Expected argument of type pps.PipelineJobInfo');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pps_PipelineJobInfo(buffer_arg) {
  return pps_pps_pb.PipelineJobInfo.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pps_RestartDatumRequest(arg) {
  if (!(arg instanceof pps_pps_pb.RestartDatumRequest)) {
    throw new Error('Expected argument of type pps.RestartDatumRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pps_RestartDatumRequest(buffer_arg) {
  return pps_pps_pb.RestartDatumRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pps_RunCronRequest(arg) {
  if (!(arg instanceof pps_pps_pb.RunCronRequest)) {
    throw new Error('Expected argument of type pps.RunCronRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pps_RunCronRequest(buffer_arg) {
  return pps_pps_pb.RunCronRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pps_RunPipelineRequest(arg) {
  if (!(arg instanceof pps_pps_pb.RunPipelineRequest)) {
    throw new Error('Expected argument of type pps.RunPipelineRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pps_RunPipelineRequest(buffer_arg) {
  return pps_pps_pb.RunPipelineRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pps_SecretInfo(arg) {
  if (!(arg instanceof pps_pps_pb.SecretInfo)) {
    throw new Error('Expected argument of type pps.SecretInfo');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pps_SecretInfo(buffer_arg) {
  return pps_pps_pb.SecretInfo.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pps_SecretInfos(arg) {
  if (!(arg instanceof pps_pps_pb.SecretInfos)) {
    throw new Error('Expected argument of type pps.SecretInfos');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pps_SecretInfos(buffer_arg) {
  return pps_pps_pb.SecretInfos.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pps_StartPipelineRequest(arg) {
  if (!(arg instanceof pps_pps_pb.StartPipelineRequest)) {
    throw new Error('Expected argument of type pps.StartPipelineRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pps_StartPipelineRequest(buffer_arg) {
  return pps_pps_pb.StartPipelineRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pps_StopPipelineJobRequest(arg) {
  if (!(arg instanceof pps_pps_pb.StopPipelineJobRequest)) {
    throw new Error('Expected argument of type pps.StopPipelineJobRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pps_StopPipelineJobRequest(buffer_arg) {
  return pps_pps_pb.StopPipelineJobRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pps_StopPipelineRequest(arg) {
  if (!(arg instanceof pps_pps_pb.StopPipelineRequest)) {
    throw new Error('Expected argument of type pps.StopPipelineRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pps_StopPipelineRequest(buffer_arg) {
  return pps_pps_pb.StopPipelineRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pps_UpdatePipelineJobStateRequest(arg) {
  if (!(arg instanceof pps_pps_pb.UpdatePipelineJobStateRequest)) {
    throw new Error('Expected argument of type pps.UpdatePipelineJobStateRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pps_UpdatePipelineJobStateRequest(buffer_arg) {
  return pps_pps_pb.UpdatePipelineJobStateRequest.deserializeBinary(new Uint8Array(buffer_arg));
}


var APIService = exports.APIService = {
  createPipelineJob: {
    path: '/pps.API/CreatePipelineJob',
    requestStream: false,
    responseStream: false,
    requestType: pps_pps_pb.CreatePipelineJobRequest,
    responseType: pps_pps_pb.PipelineJob,
    requestSerialize: serialize_pps_CreatePipelineJobRequest,
    requestDeserialize: deserialize_pps_CreatePipelineJobRequest,
    responseSerialize: serialize_pps_PipelineJob,
    responseDeserialize: deserialize_pps_PipelineJob,
  },
  inspectPipelineJob: {
    path: '/pps.API/InspectPipelineJob',
    requestStream: false,
    responseStream: false,
    requestType: pps_pps_pb.InspectPipelineJobRequest,
    responseType: pps_pps_pb.PipelineJobInfo,
    requestSerialize: serialize_pps_InspectPipelineJobRequest,
    requestDeserialize: deserialize_pps_InspectPipelineJobRequest,
    responseSerialize: serialize_pps_PipelineJobInfo,
    responseDeserialize: deserialize_pps_PipelineJobInfo,
  },
  // ListPipelineJob returns information about current and past Pachyderm jobs.
listPipelineJob: {
    path: '/pps.API/ListPipelineJob',
    requestStream: false,
    responseStream: true,
    requestType: pps_pps_pb.ListPipelineJobRequest,
    responseType: pps_pps_pb.PipelineJobInfo,
    requestSerialize: serialize_pps_ListPipelineJobRequest,
    requestDeserialize: deserialize_pps_ListPipelineJobRequest,
    responseSerialize: serialize_pps_PipelineJobInfo,
    responseDeserialize: deserialize_pps_PipelineJobInfo,
  },
  flushPipelineJob: {
    path: '/pps.API/FlushPipelineJob',
    requestStream: false,
    responseStream: true,
    requestType: pps_pps_pb.FlushPipelineJobRequest,
    responseType: pps_pps_pb.PipelineJobInfo,
    requestSerialize: serialize_pps_FlushPipelineJobRequest,
    requestDeserialize: deserialize_pps_FlushPipelineJobRequest,
    responseSerialize: serialize_pps_PipelineJobInfo,
    responseDeserialize: deserialize_pps_PipelineJobInfo,
  },
  deletePipelineJob: {
    path: '/pps.API/DeletePipelineJob',
    requestStream: false,
    responseStream: false,
    requestType: pps_pps_pb.DeletePipelineJobRequest,
    responseType: google_protobuf_empty_pb.Empty,
    requestSerialize: serialize_pps_DeletePipelineJobRequest,
    requestDeserialize: deserialize_pps_DeletePipelineJobRequest,
    responseSerialize: serialize_google_protobuf_Empty,
    responseDeserialize: deserialize_google_protobuf_Empty,
  },
  stopPipelineJob: {
    path: '/pps.API/StopPipelineJob',
    requestStream: false,
    responseStream: false,
    requestType: pps_pps_pb.StopPipelineJobRequest,
    responseType: google_protobuf_empty_pb.Empty,
    requestSerialize: serialize_pps_StopPipelineJobRequest,
    requestDeserialize: deserialize_pps_StopPipelineJobRequest,
    responseSerialize: serialize_google_protobuf_Empty,
    responseDeserialize: deserialize_google_protobuf_Empty,
  },
  inspectDatum: {
    path: '/pps.API/InspectDatum',
    requestStream: false,
    responseStream: false,
    requestType: pps_pps_pb.InspectDatumRequest,
    responseType: pps_pps_pb.DatumInfo,
    requestSerialize: serialize_pps_InspectDatumRequest,
    requestDeserialize: deserialize_pps_InspectDatumRequest,
    responseSerialize: serialize_pps_DatumInfo,
    responseDeserialize: deserialize_pps_DatumInfo,
  },
  // ListDatum returns information about each datum fed to a Pachyderm job
listDatum: {
    path: '/pps.API/ListDatum',
    requestStream: false,
    responseStream: true,
    requestType: pps_pps_pb.ListDatumRequest,
    responseType: pps_pps_pb.DatumInfo,
    requestSerialize: serialize_pps_ListDatumRequest,
    requestDeserialize: deserialize_pps_ListDatumRequest,
    responseSerialize: serialize_pps_DatumInfo,
    responseDeserialize: deserialize_pps_DatumInfo,
  },
  restartDatum: {
    path: '/pps.API/RestartDatum',
    requestStream: false,
    responseStream: false,
    requestType: pps_pps_pb.RestartDatumRequest,
    responseType: google_protobuf_empty_pb.Empty,
    requestSerialize: serialize_pps_RestartDatumRequest,
    requestDeserialize: deserialize_pps_RestartDatumRequest,
    responseSerialize: serialize_google_protobuf_Empty,
    responseDeserialize: deserialize_google_protobuf_Empty,
  },
  createPipeline: {
    path: '/pps.API/CreatePipeline',
    requestStream: false,
    responseStream: false,
    requestType: pps_pps_pb.CreatePipelineRequest,
    responseType: google_protobuf_empty_pb.Empty,
    requestSerialize: serialize_pps_CreatePipelineRequest,
    requestDeserialize: deserialize_pps_CreatePipelineRequest,
    responseSerialize: serialize_google_protobuf_Empty,
    responseDeserialize: deserialize_google_protobuf_Empty,
  },
  inspectPipeline: {
    path: '/pps.API/InspectPipeline',
    requestStream: false,
    responseStream: false,
    requestType: pps_pps_pb.InspectPipelineRequest,
    responseType: pps_pps_pb.PipelineInfo,
    requestSerialize: serialize_pps_InspectPipelineRequest,
    requestDeserialize: deserialize_pps_InspectPipelineRequest,
    responseSerialize: serialize_pps_PipelineInfo,
    responseDeserialize: deserialize_pps_PipelineInfo,
  },
  listPipeline: {
    path: '/pps.API/ListPipeline',
    requestStream: false,
    responseStream: false,
    requestType: pps_pps_pb.ListPipelineRequest,
    responseType: pps_pps_pb.PipelineInfos,
    requestSerialize: serialize_pps_ListPipelineRequest,
    requestDeserialize: deserialize_pps_ListPipelineRequest,
    responseSerialize: serialize_pps_PipelineInfos,
    responseDeserialize: deserialize_pps_PipelineInfos,
  },
  deletePipeline: {
    path: '/pps.API/DeletePipeline',
    requestStream: false,
    responseStream: false,
    requestType: pps_pps_pb.DeletePipelineRequest,
    responseType: google_protobuf_empty_pb.Empty,
    requestSerialize: serialize_pps_DeletePipelineRequest,
    requestDeserialize: deserialize_pps_DeletePipelineRequest,
    responseSerialize: serialize_google_protobuf_Empty,
    responseDeserialize: deserialize_google_protobuf_Empty,
  },
  startPipeline: {
    path: '/pps.API/StartPipeline',
    requestStream: false,
    responseStream: false,
    requestType: pps_pps_pb.StartPipelineRequest,
    responseType: google_protobuf_empty_pb.Empty,
    requestSerialize: serialize_pps_StartPipelineRequest,
    requestDeserialize: deserialize_pps_StartPipelineRequest,
    responseSerialize: serialize_google_protobuf_Empty,
    responseDeserialize: deserialize_google_protobuf_Empty,
  },
  stopPipeline: {
    path: '/pps.API/StopPipeline',
    requestStream: false,
    responseStream: false,
    requestType: pps_pps_pb.StopPipelineRequest,
    responseType: google_protobuf_empty_pb.Empty,
    requestSerialize: serialize_pps_StopPipelineRequest,
    requestDeserialize: deserialize_pps_StopPipelineRequest,
    responseSerialize: serialize_google_protobuf_Empty,
    responseDeserialize: deserialize_google_protobuf_Empty,
  },
  runPipeline: {
    path: '/pps.API/RunPipeline',
    requestStream: false,
    responseStream: false,
    requestType: pps_pps_pb.RunPipelineRequest,
    responseType: google_protobuf_empty_pb.Empty,
    requestSerialize: serialize_pps_RunPipelineRequest,
    requestDeserialize: deserialize_pps_RunPipelineRequest,
    responseSerialize: serialize_google_protobuf_Empty,
    responseDeserialize: deserialize_google_protobuf_Empty,
  },
  runCron: {
    path: '/pps.API/RunCron',
    requestStream: false,
    responseStream: false,
    requestType: pps_pps_pb.RunCronRequest,
    responseType: google_protobuf_empty_pb.Empty,
    requestSerialize: serialize_pps_RunCronRequest,
    requestDeserialize: deserialize_pps_RunCronRequest,
    responseSerialize: serialize_google_protobuf_Empty,
    responseDeserialize: deserialize_google_protobuf_Empty,
  },
  createSecret: {
    path: '/pps.API/CreateSecret',
    requestStream: false,
    responseStream: false,
    requestType: pps_pps_pb.CreateSecretRequest,
    responseType: google_protobuf_empty_pb.Empty,
    requestSerialize: serialize_pps_CreateSecretRequest,
    requestDeserialize: deserialize_pps_CreateSecretRequest,
    responseSerialize: serialize_google_protobuf_Empty,
    responseDeserialize: deserialize_google_protobuf_Empty,
  },
  deleteSecret: {
    path: '/pps.API/DeleteSecret',
    requestStream: false,
    responseStream: false,
    requestType: pps_pps_pb.DeleteSecretRequest,
    responseType: google_protobuf_empty_pb.Empty,
    requestSerialize: serialize_pps_DeleteSecretRequest,
    requestDeserialize: deserialize_pps_DeleteSecretRequest,
    responseSerialize: serialize_google_protobuf_Empty,
    responseDeserialize: deserialize_google_protobuf_Empty,
  },
  listSecret: {
    path: '/pps.API/ListSecret',
    requestStream: false,
    responseStream: false,
    requestType: google_protobuf_empty_pb.Empty,
    responseType: pps_pps_pb.SecretInfos,
    requestSerialize: serialize_google_protobuf_Empty,
    requestDeserialize: deserialize_google_protobuf_Empty,
    responseSerialize: serialize_pps_SecretInfos,
    responseDeserialize: deserialize_pps_SecretInfos,
  },
  inspectSecret: {
    path: '/pps.API/InspectSecret',
    requestStream: false,
    responseStream: false,
    requestType: pps_pps_pb.InspectSecretRequest,
    responseType: pps_pps_pb.SecretInfo,
    requestSerialize: serialize_pps_InspectSecretRequest,
    requestDeserialize: deserialize_pps_InspectSecretRequest,
    responseSerialize: serialize_pps_SecretInfo,
    responseDeserialize: deserialize_pps_SecretInfo,
  },
  // DeleteAll deletes everything
deleteAll: {
    path: '/pps.API/DeleteAll',
    requestStream: false,
    responseStream: false,
    requestType: google_protobuf_empty_pb.Empty,
    responseType: google_protobuf_empty_pb.Empty,
    requestSerialize: serialize_google_protobuf_Empty,
    requestDeserialize: deserialize_google_protobuf_Empty,
    responseSerialize: serialize_google_protobuf_Empty,
    responseDeserialize: deserialize_google_protobuf_Empty,
  },
  getLogs: {
    path: '/pps.API/GetLogs',
    requestStream: false,
    responseStream: true,
    requestType: pps_pps_pb.GetLogsRequest,
    responseType: pps_pps_pb.LogMessage,
    requestSerialize: serialize_pps_GetLogsRequest,
    requestDeserialize: deserialize_pps_GetLogsRequest,
    responseSerialize: serialize_pps_LogMessage,
    responseDeserialize: deserialize_pps_LogMessage,
  },
  // An internal call that causes PPS to put itself into an auth-enabled state
// (all pipeline have tokens, correct permissions, etcd)
activateAuth: {
    path: '/pps.API/ActivateAuth',
    requestStream: false,
    responseStream: false,
    requestType: pps_pps_pb.ActivateAuthRequest,
    responseType: pps_pps_pb.ActivateAuthResponse,
    requestSerialize: serialize_pps_ActivateAuthRequest,
    requestDeserialize: deserialize_pps_ActivateAuthRequest,
    responseSerialize: serialize_pps_ActivateAuthResponse,
    responseDeserialize: deserialize_pps_ActivateAuthResponse,
  },
  // An internal call used to move a job from one state to another
updatePipelineJobState: {
    path: '/pps.API/UpdatePipelineJobState',
    requestStream: false,
    responseStream: false,
    requestType: pps_pps_pb.UpdatePipelineJobStateRequest,
    responseType: google_protobuf_empty_pb.Empty,
    requestSerialize: serialize_pps_UpdatePipelineJobStateRequest,
    requestDeserialize: deserialize_pps_UpdatePipelineJobStateRequest,
    responseSerialize: serialize_google_protobuf_Empty,
    responseDeserialize: deserialize_google_protobuf_Empty,
  },
};

exports.APIClient = grpc.makeGenericClientConstructor(APIService);
