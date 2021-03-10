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

function serialize_pps_CreateJobRequest(arg) {
  if (!(arg instanceof pps_pps_pb.CreateJobRequest)) {
    throw new Error('Expected argument of type pps.CreateJobRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pps_CreateJobRequest(buffer_arg) {
  return pps_pps_pb.CreateJobRequest.deserializeBinary(new Uint8Array(buffer_arg));
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

function serialize_pps_DeleteJobRequest(arg) {
  if (!(arg instanceof pps_pps_pb.DeleteJobRequest)) {
    throw new Error('Expected argument of type pps.DeleteJobRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pps_DeleteJobRequest(buffer_arg) {
  return pps_pps_pb.DeleteJobRequest.deserializeBinary(new Uint8Array(buffer_arg));
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

function serialize_pps_FlushJobRequest(arg) {
  if (!(arg instanceof pps_pps_pb.FlushJobRequest)) {
    throw new Error('Expected argument of type pps.FlushJobRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pps_FlushJobRequest(buffer_arg) {
  return pps_pps_pb.FlushJobRequest.deserializeBinary(new Uint8Array(buffer_arg));
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

function serialize_pps_InspectJobRequest(arg) {
  if (!(arg instanceof pps_pps_pb.InspectJobRequest)) {
    throw new Error('Expected argument of type pps.InspectJobRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pps_InspectJobRequest(buffer_arg) {
  return pps_pps_pb.InspectJobRequest.deserializeBinary(new Uint8Array(buffer_arg));
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

function serialize_pps_Job(arg) {
  if (!(arg instanceof pps_pps_pb.Job)) {
    throw new Error('Expected argument of type pps.Job');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pps_Job(buffer_arg) {
  return pps_pps_pb.Job.deserializeBinary(new Uint8Array(buffer_arg));
}

function serialize_pps_JobInfo(arg) {
  if (!(arg instanceof pps_pps_pb.JobInfo)) {
    throw new Error('Expected argument of type pps.JobInfo');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pps_JobInfo(buffer_arg) {
  return pps_pps_pb.JobInfo.deserializeBinary(new Uint8Array(buffer_arg));
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

function serialize_pps_ListJobRequest(arg) {
  if (!(arg instanceof pps_pps_pb.ListJobRequest)) {
    throw new Error('Expected argument of type pps.ListJobRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pps_ListJobRequest(buffer_arg) {
  return pps_pps_pb.ListJobRequest.deserializeBinary(new Uint8Array(buffer_arg));
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

function serialize_pps_StopJobRequest(arg) {
  if (!(arg instanceof pps_pps_pb.StopJobRequest)) {
    throw new Error('Expected argument of type pps.StopJobRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pps_StopJobRequest(buffer_arg) {
  return pps_pps_pb.StopJobRequest.deserializeBinary(new Uint8Array(buffer_arg));
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

function serialize_pps_UpdateJobStateRequest(arg) {
  if (!(arg instanceof pps_pps_pb.UpdateJobStateRequest)) {
    throw new Error('Expected argument of type pps.UpdateJobStateRequest');
  }
  return Buffer.from(arg.serializeBinary());
}

function deserialize_pps_UpdateJobStateRequest(buffer_arg) {
  return pps_pps_pb.UpdateJobStateRequest.deserializeBinary(new Uint8Array(buffer_arg));
}


var APIService = exports.APIService = {
  createJob: {
    path: '/pps.API/CreateJob',
    requestStream: false,
    responseStream: false,
    requestType: pps_pps_pb.CreateJobRequest,
    responseType: pps_pps_pb.Job,
    requestSerialize: serialize_pps_CreateJobRequest,
    requestDeserialize: deserialize_pps_CreateJobRequest,
    responseSerialize: serialize_pps_Job,
    responseDeserialize: deserialize_pps_Job,
  },
  inspectJob: {
    path: '/pps.API/InspectJob',
    requestStream: false,
    responseStream: false,
    requestType: pps_pps_pb.InspectJobRequest,
    responseType: pps_pps_pb.JobInfo,
    requestSerialize: serialize_pps_InspectJobRequest,
    requestDeserialize: deserialize_pps_InspectJobRequest,
    responseSerialize: serialize_pps_JobInfo,
    responseDeserialize: deserialize_pps_JobInfo,
  },
  // ListJob returns information about current and past Pachyderm jobs.
listJob: {
    path: '/pps.API/ListJob',
    requestStream: false,
    responseStream: true,
    requestType: pps_pps_pb.ListJobRequest,
    responseType: pps_pps_pb.JobInfo,
    requestSerialize: serialize_pps_ListJobRequest,
    requestDeserialize: deserialize_pps_ListJobRequest,
    responseSerialize: serialize_pps_JobInfo,
    responseDeserialize: deserialize_pps_JobInfo,
  },
  flushJob: {
    path: '/pps.API/FlushJob',
    requestStream: false,
    responseStream: true,
    requestType: pps_pps_pb.FlushJobRequest,
    responseType: pps_pps_pb.JobInfo,
    requestSerialize: serialize_pps_FlushJobRequest,
    requestDeserialize: deserialize_pps_FlushJobRequest,
    responseSerialize: serialize_pps_JobInfo,
    responseDeserialize: deserialize_pps_JobInfo,
  },
  deleteJob: {
    path: '/pps.API/DeleteJob',
    requestStream: false,
    responseStream: false,
    requestType: pps_pps_pb.DeleteJobRequest,
    responseType: google_protobuf_empty_pb.Empty,
    requestSerialize: serialize_pps_DeleteJobRequest,
    requestDeserialize: deserialize_pps_DeleteJobRequest,
    responseSerialize: serialize_google_protobuf_Empty,
    responseDeserialize: deserialize_google_protobuf_Empty,
  },
  stopJob: {
    path: '/pps.API/StopJob',
    requestStream: false,
    responseStream: false,
    requestType: pps_pps_pb.StopJobRequest,
    responseType: google_protobuf_empty_pb.Empty,
    requestSerialize: serialize_pps_StopJobRequest,
    requestDeserialize: deserialize_pps_StopJobRequest,
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
updateJobState: {
    path: '/pps.API/UpdateJobState',
    requestStream: false,
    responseStream: false,
    requestType: pps_pps_pb.UpdateJobStateRequest,
    responseType: google_protobuf_empty_pb.Empty,
    requestSerialize: serialize_pps_UpdateJobStateRequest,
    requestDeserialize: deserialize_pps_UpdateJobStateRequest,
    responseSerialize: serialize_google_protobuf_Empty,
    responseDeserialize: deserialize_google_protobuf_Empty,
  },
};

exports.APIClient = grpc.makeGenericClientConstructor(APIService);
