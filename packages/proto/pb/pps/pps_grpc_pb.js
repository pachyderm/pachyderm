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

function serialize_pps_v2_ActivateAuthRequest(arg) {
  if (!(arg instanceof pps_pps_pb.ActivateAuthRequest)) {
    throw new Error('Expected argument of type pps_v2.ActivateAuthRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pps_v2_ActivateAuthRequest(buffer_arg) {
  return pps_pps_pb.ActivateAuthRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pps_v2_ActivateAuthResponse(arg) {
  if (!(arg instanceof pps_pps_pb.ActivateAuthResponse)) {
    throw new Error('Expected argument of type pps_v2.ActivateAuthResponse');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pps_v2_ActivateAuthResponse(buffer_arg) {
  return pps_pps_pb.ActivateAuthResponse.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pps_v2_CreatePipelineRequest(arg) {
  if (!(arg instanceof pps_pps_pb.CreatePipelineRequest)) {
    throw new Error('Expected argument of type pps_v2.CreatePipelineRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pps_v2_CreatePipelineRequest(buffer_arg) {
  return pps_pps_pb.CreatePipelineRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pps_v2_CreateSecretRequest(arg) {
  if (!(arg instanceof pps_pps_pb.CreateSecretRequest)) {
    throw new Error('Expected argument of type pps_v2.CreateSecretRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pps_v2_CreateSecretRequest(buffer_arg) {
  return pps_pps_pb.CreateSecretRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pps_v2_DatumInfo(arg) {
  if (!(arg instanceof pps_pps_pb.DatumInfo)) {
    throw new Error('Expected argument of type pps_v2.DatumInfo');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pps_v2_DatumInfo(buffer_arg) {
  return pps_pps_pb.DatumInfo.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pps_v2_DeleteJobRequest(arg) {
  if (!(arg instanceof pps_pps_pb.DeleteJobRequest)) {
    throw new Error('Expected argument of type pps_v2.DeleteJobRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pps_v2_DeleteJobRequest(buffer_arg) {
  return pps_pps_pb.DeleteJobRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pps_v2_DeletePipelineRequest(arg) {
  if (!(arg instanceof pps_pps_pb.DeletePipelineRequest)) {
    throw new Error('Expected argument of type pps_v2.DeletePipelineRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pps_v2_DeletePipelineRequest(buffer_arg) {
  return pps_pps_pb.DeletePipelineRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pps_v2_DeleteSecretRequest(arg) {
  if (!(arg instanceof pps_pps_pb.DeleteSecretRequest)) {
    throw new Error('Expected argument of type pps_v2.DeleteSecretRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pps_v2_DeleteSecretRequest(buffer_arg) {
  return pps_pps_pb.DeleteSecretRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pps_v2_GetLogsRequest(arg) {
  if (!(arg instanceof pps_pps_pb.GetLogsRequest)) {
    throw new Error('Expected argument of type pps_v2.GetLogsRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pps_v2_GetLogsRequest(buffer_arg) {
  return pps_pps_pb.GetLogsRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pps_v2_InspectDatumRequest(arg) {
  if (!(arg instanceof pps_pps_pb.InspectDatumRequest)) {
    throw new Error('Expected argument of type pps_v2.InspectDatumRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pps_v2_InspectDatumRequest(buffer_arg) {
  return pps_pps_pb.InspectDatumRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pps_v2_InspectJobRequest(arg) {
  if (!(arg instanceof pps_pps_pb.InspectJobRequest)) {
    throw new Error('Expected argument of type pps_v2.InspectJobRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pps_v2_InspectJobRequest(buffer_arg) {
  return pps_pps_pb.InspectJobRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pps_v2_InspectJobSetRequest(arg) {
  if (!(arg instanceof pps_pps_pb.InspectJobSetRequest)) {
    throw new Error('Expected argument of type pps_v2.InspectJobSetRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pps_v2_InspectJobSetRequest(buffer_arg) {
  return pps_pps_pb.InspectJobSetRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pps_v2_InspectPipelineRequest(arg) {
  if (!(arg instanceof pps_pps_pb.InspectPipelineRequest)) {
    throw new Error('Expected argument of type pps_v2.InspectPipelineRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pps_v2_InspectPipelineRequest(buffer_arg) {
  return pps_pps_pb.InspectPipelineRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pps_v2_InspectSecretRequest(arg) {
  if (!(arg instanceof pps_pps_pb.InspectSecretRequest)) {
    throw new Error('Expected argument of type pps_v2.InspectSecretRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pps_v2_InspectSecretRequest(buffer_arg) {
  return pps_pps_pb.InspectSecretRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pps_v2_JobInfo(arg) {
  if (!(arg instanceof pps_pps_pb.JobInfo)) {
    throw new Error('Expected argument of type pps_v2.JobInfo');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pps_v2_JobInfo(buffer_arg) {
  return pps_pps_pb.JobInfo.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pps_v2_JobSetInfo(arg) {
  if (!(arg instanceof pps_pps_pb.JobSetInfo)) {
    throw new Error('Expected argument of type pps_v2.JobSetInfo');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pps_v2_JobSetInfo(buffer_arg) {
  return pps_pps_pb.JobSetInfo.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pps_v2_ListDatumRequest(arg) {
  if (!(arg instanceof pps_pps_pb.ListDatumRequest)) {
    throw new Error('Expected argument of type pps_v2.ListDatumRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pps_v2_ListDatumRequest(buffer_arg) {
  return pps_pps_pb.ListDatumRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pps_v2_ListJobRequest(arg) {
  if (!(arg instanceof pps_pps_pb.ListJobRequest)) {
    throw new Error('Expected argument of type pps_v2.ListJobRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pps_v2_ListJobRequest(buffer_arg) {
  return pps_pps_pb.ListJobRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pps_v2_ListJobSetRequest(arg) {
  if (!(arg instanceof pps_pps_pb.ListJobSetRequest)) {
    throw new Error('Expected argument of type pps_v2.ListJobSetRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pps_v2_ListJobSetRequest(buffer_arg) {
  return pps_pps_pb.ListJobSetRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pps_v2_ListPipelineRequest(arg) {
  if (!(arg instanceof pps_pps_pb.ListPipelineRequest)) {
    throw new Error('Expected argument of type pps_v2.ListPipelineRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pps_v2_ListPipelineRequest(buffer_arg) {
  return pps_pps_pb.ListPipelineRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pps_v2_LogMessage(arg) {
  if (!(arg instanceof pps_pps_pb.LogMessage)) {
    throw new Error('Expected argument of type pps_v2.LogMessage');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pps_v2_LogMessage(buffer_arg) {
  return pps_pps_pb.LogMessage.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pps_v2_PipelineInfo(arg) {
  if (!(arg instanceof pps_pps_pb.PipelineInfo)) {
    throw new Error('Expected argument of type pps_v2.PipelineInfo');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pps_v2_PipelineInfo(buffer_arg) {
  return pps_pps_pb.PipelineInfo.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pps_v2_RestartDatumRequest(arg) {
  if (!(arg instanceof pps_pps_pb.RestartDatumRequest)) {
    throw new Error('Expected argument of type pps_v2.RestartDatumRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pps_v2_RestartDatumRequest(buffer_arg) {
  return pps_pps_pb.RestartDatumRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pps_v2_RunCronRequest(arg) {
  if (!(arg instanceof pps_pps_pb.RunCronRequest)) {
    throw new Error('Expected argument of type pps_v2.RunCronRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pps_v2_RunCronRequest(buffer_arg) {
  return pps_pps_pb.RunCronRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pps_v2_RunPipelineRequest(arg) {
  if (!(arg instanceof pps_pps_pb.RunPipelineRequest)) {
    throw new Error('Expected argument of type pps_v2.RunPipelineRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pps_v2_RunPipelineRequest(buffer_arg) {
  return pps_pps_pb.RunPipelineRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pps_v2_SecretInfo(arg) {
  if (!(arg instanceof pps_pps_pb.SecretInfo)) {
    throw new Error('Expected argument of type pps_v2.SecretInfo');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pps_v2_SecretInfo(buffer_arg) {
  return pps_pps_pb.SecretInfo.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pps_v2_SecretInfos(arg) {
  if (!(arg instanceof pps_pps_pb.SecretInfos)) {
    throw new Error('Expected argument of type pps_v2.SecretInfos');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pps_v2_SecretInfos(buffer_arg) {
  return pps_pps_pb.SecretInfos.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pps_v2_StartPipelineRequest(arg) {
  if (!(arg instanceof pps_pps_pb.StartPipelineRequest)) {
    throw new Error('Expected argument of type pps_v2.StartPipelineRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pps_v2_StartPipelineRequest(buffer_arg) {
  return pps_pps_pb.StartPipelineRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pps_v2_StopJobRequest(arg) {
  if (!(arg instanceof pps_pps_pb.StopJobRequest)) {
    throw new Error('Expected argument of type pps_v2.StopJobRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pps_v2_StopJobRequest(buffer_arg) {
  return pps_pps_pb.StopJobRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pps_v2_StopPipelineRequest(arg) {
  if (!(arg instanceof pps_pps_pb.StopPipelineRequest)) {
    throw new Error('Expected argument of type pps_v2.StopPipelineRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pps_v2_StopPipelineRequest(buffer_arg) {
  return pps_pps_pb.StopPipelineRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pps_v2_SubscribeJobRequest(arg) {
  if (!(arg instanceof pps_pps_pb.SubscribeJobRequest)) {
    throw new Error('Expected argument of type pps_v2.SubscribeJobRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pps_v2_SubscribeJobRequest(buffer_arg) {
  return pps_pps_pb.SubscribeJobRequest.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pps_v2_UpdateJobStateRequest(arg) {
  if (!(arg instanceof pps_pps_pb.UpdateJobStateRequest)) {
    throw new Error('Expected argument of type pps_v2.UpdateJobStateRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pps_v2_UpdateJobStateRequest(buffer_arg) {
  return pps_pps_pb.UpdateJobStateRequest.deserializeBinary(new Uint8Array(buffer_arg));
}


var APIService = exports.APIService = {
  inspectJob: {
    path: '/pps_v2.API/InspectJob',
    requestStream: false,
    responseStream: false,
    requestType: pps_pps_pb.InspectJobRequest,
    responseType: pps_pps_pb.JobInfo,
    requestSerialize: serialize_pps_v2_InspectJobRequest,
    requestDeserialize: deserialize_pps_v2_InspectJobRequest,
    responseSerialize: serialize_pps_v2_JobInfo,
    responseDeserialize: deserialize_pps_v2_JobInfo,
  },
  inspectJobSet: {
    path: '/pps_v2.API/InspectJobSet',
    requestStream: false,
    responseStream: true,
    requestType: pps_pps_pb.InspectJobSetRequest,
    responseType: pps_pps_pb.JobInfo,
    requestSerialize: serialize_pps_v2_InspectJobSetRequest,
    requestDeserialize: deserialize_pps_v2_InspectJobSetRequest,
    responseSerialize: serialize_pps_v2_JobInfo,
    responseDeserialize: deserialize_pps_v2_JobInfo,
  },
  // ListJob returns information about current and past Pachyderm jobs.
listJob: {
    path: '/pps_v2.API/ListJob',
    requestStream: false,
    responseStream: true,
    requestType: pps_pps_pb.ListJobRequest,
    responseType: pps_pps_pb.JobInfo,
    requestSerialize: serialize_pps_v2_ListJobRequest,
    requestDeserialize: deserialize_pps_v2_ListJobRequest,
    responseSerialize: serialize_pps_v2_JobInfo,
    responseDeserialize: deserialize_pps_v2_JobInfo,
  },
  listJobSet: {
    path: '/pps_v2.API/ListJobSet',
    requestStream: false,
    responseStream: true,
    requestType: pps_pps_pb.ListJobSetRequest,
    responseType: pps_pps_pb.JobSetInfo,
    requestSerialize: serialize_pps_v2_ListJobSetRequest,
    requestDeserialize: deserialize_pps_v2_ListJobSetRequest,
    responseSerialize: serialize_pps_v2_JobSetInfo,
    responseDeserialize: deserialize_pps_v2_JobSetInfo,
  },
  subscribeJob: {
    path: '/pps_v2.API/SubscribeJob',
    requestStream: false,
    responseStream: true,
    requestType: pps_pps_pb.SubscribeJobRequest,
    responseType: pps_pps_pb.JobInfo,
    requestSerialize: serialize_pps_v2_SubscribeJobRequest,
    requestDeserialize: deserialize_pps_v2_SubscribeJobRequest,
    responseSerialize: serialize_pps_v2_JobInfo,
    responseDeserialize: deserialize_pps_v2_JobInfo,
  },
  deleteJob: {
    path: '/pps_v2.API/DeleteJob',
    requestStream: false,
    responseStream: false,
    requestType: pps_pps_pb.DeleteJobRequest,
    responseType: google_protobuf_empty_pb.Empty,
    requestSerialize: serialize_pps_v2_DeleteJobRequest,
    requestDeserialize: deserialize_pps_v2_DeleteJobRequest,
    responseSerialize: serialize_google_protobuf_Empty,
    responseDeserialize: deserialize_google_protobuf_Empty,
  },
  stopJob: {
    path: '/pps_v2.API/StopJob',
    requestStream: false,
    responseStream: false,
    requestType: pps_pps_pb.StopJobRequest,
    responseType: google_protobuf_empty_pb.Empty,
    requestSerialize: serialize_pps_v2_StopJobRequest,
    requestDeserialize: deserialize_pps_v2_StopJobRequest,
    responseSerialize: serialize_google_protobuf_Empty,
    responseDeserialize: deserialize_google_protobuf_Empty,
  },
  inspectDatum: {
    path: '/pps_v2.API/InspectDatum',
    requestStream: false,
    responseStream: false,
    requestType: pps_pps_pb.InspectDatumRequest,
    responseType: pps_pps_pb.DatumInfo,
    requestSerialize: serialize_pps_v2_InspectDatumRequest,
    requestDeserialize: deserialize_pps_v2_InspectDatumRequest,
    responseSerialize: serialize_pps_v2_DatumInfo,
    responseDeserialize: deserialize_pps_v2_DatumInfo,
  },
  // ListDatum returns information about each datum fed to a Pachyderm job
listDatum: {
    path: '/pps_v2.API/ListDatum',
    requestStream: false,
    responseStream: true,
    requestType: pps_pps_pb.ListDatumRequest,
    responseType: pps_pps_pb.DatumInfo,
    requestSerialize: serialize_pps_v2_ListDatumRequest,
    requestDeserialize: deserialize_pps_v2_ListDatumRequest,
    responseSerialize: serialize_pps_v2_DatumInfo,
    responseDeserialize: deserialize_pps_v2_DatumInfo,
  },
  restartDatum: {
    path: '/pps_v2.API/RestartDatum',
    requestStream: false,
    responseStream: false,
    requestType: pps_pps_pb.RestartDatumRequest,
    responseType: google_protobuf_empty_pb.Empty,
    requestSerialize: serialize_pps_v2_RestartDatumRequest,
    requestDeserialize: deserialize_pps_v2_RestartDatumRequest,
    responseSerialize: serialize_google_protobuf_Empty,
    responseDeserialize: deserialize_google_protobuf_Empty,
  },
  createPipeline: {
    path: '/pps_v2.API/CreatePipeline',
    requestStream: false,
    responseStream: false,
    requestType: pps_pps_pb.CreatePipelineRequest,
    responseType: google_protobuf_empty_pb.Empty,
    requestSerialize: serialize_pps_v2_CreatePipelineRequest,
    requestDeserialize: deserialize_pps_v2_CreatePipelineRequest,
    responseSerialize: serialize_google_protobuf_Empty,
    responseDeserialize: deserialize_google_protobuf_Empty,
  },
  inspectPipeline: {
    path: '/pps_v2.API/InspectPipeline',
    requestStream: false,
    responseStream: false,
    requestType: pps_pps_pb.InspectPipelineRequest,
    responseType: pps_pps_pb.PipelineInfo,
    requestSerialize: serialize_pps_v2_InspectPipelineRequest,
    requestDeserialize: deserialize_pps_v2_InspectPipelineRequest,
    responseSerialize: serialize_pps_v2_PipelineInfo,
    responseDeserialize: deserialize_pps_v2_PipelineInfo,
  },
  listPipeline: {
    path: '/pps_v2.API/ListPipeline',
    requestStream: false,
    responseStream: true,
    requestType: pps_pps_pb.ListPipelineRequest,
    responseType: pps_pps_pb.PipelineInfo,
    requestSerialize: serialize_pps_v2_ListPipelineRequest,
    requestDeserialize: deserialize_pps_v2_ListPipelineRequest,
    responseSerialize: serialize_pps_v2_PipelineInfo,
    responseDeserialize: deserialize_pps_v2_PipelineInfo,
  },
  deletePipeline: {
    path: '/pps_v2.API/DeletePipeline',
    requestStream: false,
    responseStream: false,
    requestType: pps_pps_pb.DeletePipelineRequest,
    responseType: google_protobuf_empty_pb.Empty,
    requestSerialize: serialize_pps_v2_DeletePipelineRequest,
    requestDeserialize: deserialize_pps_v2_DeletePipelineRequest,
    responseSerialize: serialize_google_protobuf_Empty,
    responseDeserialize: deserialize_google_protobuf_Empty,
  },
  startPipeline: {
    path: '/pps_v2.API/StartPipeline',
    requestStream: false,
    responseStream: false,
    requestType: pps_pps_pb.StartPipelineRequest,
    responseType: google_protobuf_empty_pb.Empty,
    requestSerialize: serialize_pps_v2_StartPipelineRequest,
    requestDeserialize: deserialize_pps_v2_StartPipelineRequest,
    responseSerialize: serialize_google_protobuf_Empty,
    responseDeserialize: deserialize_google_protobuf_Empty,
  },
  stopPipeline: {
    path: '/pps_v2.API/StopPipeline',
    requestStream: false,
    responseStream: false,
    requestType: pps_pps_pb.StopPipelineRequest,
    responseType: google_protobuf_empty_pb.Empty,
    requestSerialize: serialize_pps_v2_StopPipelineRequest,
    requestDeserialize: deserialize_pps_v2_StopPipelineRequest,
    responseSerialize: serialize_google_protobuf_Empty,
    responseDeserialize: deserialize_google_protobuf_Empty,
  },
  runPipeline: {
    path: '/pps_v2.API/RunPipeline',
    requestStream: false,
    responseStream: false,
    requestType: pps_pps_pb.RunPipelineRequest,
    responseType: google_protobuf_empty_pb.Empty,
    requestSerialize: serialize_pps_v2_RunPipelineRequest,
    requestDeserialize: deserialize_pps_v2_RunPipelineRequest,
    responseSerialize: serialize_google_protobuf_Empty,
    responseDeserialize: deserialize_google_protobuf_Empty,
  },
  runCron: {
    path: '/pps_v2.API/RunCron',
    requestStream: false,
    responseStream: false,
    requestType: pps_pps_pb.RunCronRequest,
    responseType: google_protobuf_empty_pb.Empty,
    requestSerialize: serialize_pps_v2_RunCronRequest,
    requestDeserialize: deserialize_pps_v2_RunCronRequest,
    responseSerialize: serialize_google_protobuf_Empty,
    responseDeserialize: deserialize_google_protobuf_Empty,
  },
  createSecret: {
    path: '/pps_v2.API/CreateSecret',
    requestStream: false,
    responseStream: false,
    requestType: pps_pps_pb.CreateSecretRequest,
    responseType: google_protobuf_empty_pb.Empty,
    requestSerialize: serialize_pps_v2_CreateSecretRequest,
    requestDeserialize: deserialize_pps_v2_CreateSecretRequest,
    responseSerialize: serialize_google_protobuf_Empty,
    responseDeserialize: deserialize_google_protobuf_Empty,
  },
  deleteSecret: {
    path: '/pps_v2.API/DeleteSecret',
    requestStream: false,
    responseStream: false,
    requestType: pps_pps_pb.DeleteSecretRequest,
    responseType: google_protobuf_empty_pb.Empty,
    requestSerialize: serialize_pps_v2_DeleteSecretRequest,
    requestDeserialize: deserialize_pps_v2_DeleteSecretRequest,
    responseSerialize: serialize_google_protobuf_Empty,
    responseDeserialize: deserialize_google_protobuf_Empty,
  },
  listSecret: {
    path: '/pps_v2.API/ListSecret',
    requestStream: false,
    responseStream: false,
    requestType: google_protobuf_empty_pb.Empty,
    responseType: pps_pps_pb.SecretInfos,
    requestSerialize: serialize_google_protobuf_Empty,
    requestDeserialize: deserialize_google_protobuf_Empty,
    responseSerialize: serialize_pps_v2_SecretInfos,
    responseDeserialize: deserialize_pps_v2_SecretInfos,
  },
  inspectSecret: {
    path: '/pps_v2.API/InspectSecret',
    requestStream: false,
    responseStream: false,
    requestType: pps_pps_pb.InspectSecretRequest,
    responseType: pps_pps_pb.SecretInfo,
    requestSerialize: serialize_pps_v2_InspectSecretRequest,
    requestDeserialize: deserialize_pps_v2_InspectSecretRequest,
    responseSerialize: serialize_pps_v2_SecretInfo,
    responseDeserialize: deserialize_pps_v2_SecretInfo,
  },
  // DeleteAll deletes everything
deleteAll: {
    path: '/pps_v2.API/DeleteAll',
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
    path: '/pps_v2.API/GetLogs',
    requestStream: false,
    responseStream: true,
    requestType: pps_pps_pb.GetLogsRequest,
    responseType: pps_pps_pb.LogMessage,
    requestSerialize: serialize_pps_v2_GetLogsRequest,
    requestDeserialize: deserialize_pps_v2_GetLogsRequest,
    responseSerialize: serialize_pps_v2_LogMessage,
    responseDeserialize: deserialize_pps_v2_LogMessage,
  },
  // An internal call that causes PPS to put itself into an auth-enabled state
// (all pipeline have tokens, correct permissions, etcd)
activateAuth: {
    path: '/pps_v2.API/ActivateAuth',
    requestStream: false,
    responseStream: false,
    requestType: pps_pps_pb.ActivateAuthRequest,
    responseType: pps_pps_pb.ActivateAuthResponse,
    requestSerialize: serialize_pps_v2_ActivateAuthRequest,
    requestDeserialize: deserialize_pps_v2_ActivateAuthRequest,
    responseSerialize: serialize_pps_v2_ActivateAuthResponse,
    responseDeserialize: deserialize_pps_v2_ActivateAuthResponse,
  },
  // An internal call used to move a job from one state to another
updateJobState: {
    path: '/pps_v2.API/UpdateJobState',
    requestStream: false,
    responseStream: false,
    requestType: pps_pps_pb.UpdateJobStateRequest,
    responseType: google_protobuf_empty_pb.Empty,
    requestSerialize: serialize_pps_v2_UpdateJobStateRequest,
    requestDeserialize: deserialize_pps_v2_UpdateJobStateRequest,
    responseSerialize: serialize_google_protobuf_Empty,
    responseDeserialize: deserialize_google_protobuf_Empty,
  },
};

exports.APIClient = grpc.makeGenericClientConstructor(APIService);
