// package: pps_v2
// file: pps/pps.proto

/* tslint:disable */
/* eslint-disable */

import * as grpc from "@grpc/grpc-js";
import * as pps_pps_pb from "../pps/pps_pb";
import * as google_protobuf_empty_pb from "google-protobuf/google/protobuf/empty_pb";
import * as google_protobuf_timestamp_pb from "google-protobuf/google/protobuf/timestamp_pb";
import * as google_protobuf_duration_pb from "google-protobuf/google/protobuf/duration_pb";
import * as google_protobuf_wrappers_pb from "google-protobuf/google/protobuf/wrappers_pb";
import * as pfs_pfs_pb from "../pfs/pfs_pb";
import * as task_task_pb from "../task/task_pb";
import * as protoextensions_json_schema_options_pb from "../protoextensions/json-schema-options_pb";
import * as protoextensions_validate_pb from "../protoextensions/validate_pb";

interface IAPIService extends grpc.ServiceDefinition<grpc.UntypedServiceImplementation> {
    inspectJob: IAPIService_IInspectJob;
    inspectJobSet: IAPIService_IInspectJobSet;
    listJob: IAPIService_IListJob;
    listJobSet: IAPIService_IListJobSet;
    subscribeJob: IAPIService_ISubscribeJob;
    deleteJob: IAPIService_IDeleteJob;
    stopJob: IAPIService_IStopJob;
    inspectDatum: IAPIService_IInspectDatum;
    listDatum: IAPIService_IListDatum;
    restartDatum: IAPIService_IRestartDatum;
    createPipeline: IAPIService_ICreatePipeline;
    createPipelineV2: IAPIService_ICreatePipelineV2;
    inspectPipeline: IAPIService_IInspectPipeline;
    listPipeline: IAPIService_IListPipeline;
    deletePipeline: IAPIService_IDeletePipeline;
    deletePipelines: IAPIService_IDeletePipelines;
    startPipeline: IAPIService_IStartPipeline;
    stopPipeline: IAPIService_IStopPipeline;
    runPipeline: IAPIService_IRunPipeline;
    runCron: IAPIService_IRunCron;
    createSecret: IAPIService_ICreateSecret;
    deleteSecret: IAPIService_IDeleteSecret;
    listSecret: IAPIService_IListSecret;
    inspectSecret: IAPIService_IInspectSecret;
    deleteAll: IAPIService_IDeleteAll;
    getLogs: IAPIService_IGetLogs;
    activateAuth: IAPIService_IActivateAuth;
    updateJobState: IAPIService_IUpdateJobState;
    runLoadTest: IAPIService_IRunLoadTest;
    runLoadTestDefault: IAPIService_IRunLoadTestDefault;
    renderTemplate: IAPIService_IRenderTemplate;
    listTask: IAPIService_IListTask;
    getKubeEvents: IAPIService_IGetKubeEvents;
    queryLoki: IAPIService_IQueryLoki;
    getClusterDefaults: IAPIService_IGetClusterDefaults;
    setClusterDefaults: IAPIService_ISetClusterDefaults;
}

interface IAPIService_IInspectJob extends grpc.MethodDefinition<pps_pps_pb.InspectJobRequest, pps_pps_pb.JobInfo> {
    path: "/pps_v2.API/InspectJob";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<pps_pps_pb.InspectJobRequest>;
    requestDeserialize: grpc.deserialize<pps_pps_pb.InspectJobRequest>;
    responseSerialize: grpc.serialize<pps_pps_pb.JobInfo>;
    responseDeserialize: grpc.deserialize<pps_pps_pb.JobInfo>;
}
interface IAPIService_IInspectJobSet extends grpc.MethodDefinition<pps_pps_pb.InspectJobSetRequest, pps_pps_pb.JobInfo> {
    path: "/pps_v2.API/InspectJobSet";
    requestStream: false;
    responseStream: true;
    requestSerialize: grpc.serialize<pps_pps_pb.InspectJobSetRequest>;
    requestDeserialize: grpc.deserialize<pps_pps_pb.InspectJobSetRequest>;
    responseSerialize: grpc.serialize<pps_pps_pb.JobInfo>;
    responseDeserialize: grpc.deserialize<pps_pps_pb.JobInfo>;
}
interface IAPIService_IListJob extends grpc.MethodDefinition<pps_pps_pb.ListJobRequest, pps_pps_pb.JobInfo> {
    path: "/pps_v2.API/ListJob";
    requestStream: false;
    responseStream: true;
    requestSerialize: grpc.serialize<pps_pps_pb.ListJobRequest>;
    requestDeserialize: grpc.deserialize<pps_pps_pb.ListJobRequest>;
    responseSerialize: grpc.serialize<pps_pps_pb.JobInfo>;
    responseDeserialize: grpc.deserialize<pps_pps_pb.JobInfo>;
}
interface IAPIService_IListJobSet extends grpc.MethodDefinition<pps_pps_pb.ListJobSetRequest, pps_pps_pb.JobSetInfo> {
    path: "/pps_v2.API/ListJobSet";
    requestStream: false;
    responseStream: true;
    requestSerialize: grpc.serialize<pps_pps_pb.ListJobSetRequest>;
    requestDeserialize: grpc.deserialize<pps_pps_pb.ListJobSetRequest>;
    responseSerialize: grpc.serialize<pps_pps_pb.JobSetInfo>;
    responseDeserialize: grpc.deserialize<pps_pps_pb.JobSetInfo>;
}
interface IAPIService_ISubscribeJob extends grpc.MethodDefinition<pps_pps_pb.SubscribeJobRequest, pps_pps_pb.JobInfo> {
    path: "/pps_v2.API/SubscribeJob";
    requestStream: false;
    responseStream: true;
    requestSerialize: grpc.serialize<pps_pps_pb.SubscribeJobRequest>;
    requestDeserialize: grpc.deserialize<pps_pps_pb.SubscribeJobRequest>;
    responseSerialize: grpc.serialize<pps_pps_pb.JobInfo>;
    responseDeserialize: grpc.deserialize<pps_pps_pb.JobInfo>;
}
interface IAPIService_IDeleteJob extends grpc.MethodDefinition<pps_pps_pb.DeleteJobRequest, google_protobuf_empty_pb.Empty> {
    path: "/pps_v2.API/DeleteJob";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<pps_pps_pb.DeleteJobRequest>;
    requestDeserialize: grpc.deserialize<pps_pps_pb.DeleteJobRequest>;
    responseSerialize: grpc.serialize<google_protobuf_empty_pb.Empty>;
    responseDeserialize: grpc.deserialize<google_protobuf_empty_pb.Empty>;
}
interface IAPIService_IStopJob extends grpc.MethodDefinition<pps_pps_pb.StopJobRequest, google_protobuf_empty_pb.Empty> {
    path: "/pps_v2.API/StopJob";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<pps_pps_pb.StopJobRequest>;
    requestDeserialize: grpc.deserialize<pps_pps_pb.StopJobRequest>;
    responseSerialize: grpc.serialize<google_protobuf_empty_pb.Empty>;
    responseDeserialize: grpc.deserialize<google_protobuf_empty_pb.Empty>;
}
interface IAPIService_IInspectDatum extends grpc.MethodDefinition<pps_pps_pb.InspectDatumRequest, pps_pps_pb.DatumInfo> {
    path: "/pps_v2.API/InspectDatum";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<pps_pps_pb.InspectDatumRequest>;
    requestDeserialize: grpc.deserialize<pps_pps_pb.InspectDatumRequest>;
    responseSerialize: grpc.serialize<pps_pps_pb.DatumInfo>;
    responseDeserialize: grpc.deserialize<pps_pps_pb.DatumInfo>;
}
interface IAPIService_IListDatum extends grpc.MethodDefinition<pps_pps_pb.ListDatumRequest, pps_pps_pb.DatumInfo> {
    path: "/pps_v2.API/ListDatum";
    requestStream: false;
    responseStream: true;
    requestSerialize: grpc.serialize<pps_pps_pb.ListDatumRequest>;
    requestDeserialize: grpc.deserialize<pps_pps_pb.ListDatumRequest>;
    responseSerialize: grpc.serialize<pps_pps_pb.DatumInfo>;
    responseDeserialize: grpc.deserialize<pps_pps_pb.DatumInfo>;
}
interface IAPIService_IRestartDatum extends grpc.MethodDefinition<pps_pps_pb.RestartDatumRequest, google_protobuf_empty_pb.Empty> {
    path: "/pps_v2.API/RestartDatum";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<pps_pps_pb.RestartDatumRequest>;
    requestDeserialize: grpc.deserialize<pps_pps_pb.RestartDatumRequest>;
    responseSerialize: grpc.serialize<google_protobuf_empty_pb.Empty>;
    responseDeserialize: grpc.deserialize<google_protobuf_empty_pb.Empty>;
}
interface IAPIService_ICreatePipeline extends grpc.MethodDefinition<pps_pps_pb.CreatePipelineRequest, google_protobuf_empty_pb.Empty> {
    path: "/pps_v2.API/CreatePipeline";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<pps_pps_pb.CreatePipelineRequest>;
    requestDeserialize: grpc.deserialize<pps_pps_pb.CreatePipelineRequest>;
    responseSerialize: grpc.serialize<google_protobuf_empty_pb.Empty>;
    responseDeserialize: grpc.deserialize<google_protobuf_empty_pb.Empty>;
}
interface IAPIService_ICreatePipelineV2 extends grpc.MethodDefinition<pps_pps_pb.CreatePipelineV2Request, pps_pps_pb.CreatePipelineV2Response> {
    path: "/pps_v2.API/CreatePipelineV2";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<pps_pps_pb.CreatePipelineV2Request>;
    requestDeserialize: grpc.deserialize<pps_pps_pb.CreatePipelineV2Request>;
    responseSerialize: grpc.serialize<pps_pps_pb.CreatePipelineV2Response>;
    responseDeserialize: grpc.deserialize<pps_pps_pb.CreatePipelineV2Response>;
}
interface IAPIService_IInspectPipeline extends grpc.MethodDefinition<pps_pps_pb.InspectPipelineRequest, pps_pps_pb.PipelineInfo> {
    path: "/pps_v2.API/InspectPipeline";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<pps_pps_pb.InspectPipelineRequest>;
    requestDeserialize: grpc.deserialize<pps_pps_pb.InspectPipelineRequest>;
    responseSerialize: grpc.serialize<pps_pps_pb.PipelineInfo>;
    responseDeserialize: grpc.deserialize<pps_pps_pb.PipelineInfo>;
}
interface IAPIService_IListPipeline extends grpc.MethodDefinition<pps_pps_pb.ListPipelineRequest, pps_pps_pb.PipelineInfo> {
    path: "/pps_v2.API/ListPipeline";
    requestStream: false;
    responseStream: true;
    requestSerialize: grpc.serialize<pps_pps_pb.ListPipelineRequest>;
    requestDeserialize: grpc.deserialize<pps_pps_pb.ListPipelineRequest>;
    responseSerialize: grpc.serialize<pps_pps_pb.PipelineInfo>;
    responseDeserialize: grpc.deserialize<pps_pps_pb.PipelineInfo>;
}
interface IAPIService_IDeletePipeline extends grpc.MethodDefinition<pps_pps_pb.DeletePipelineRequest, google_protobuf_empty_pb.Empty> {
    path: "/pps_v2.API/DeletePipeline";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<pps_pps_pb.DeletePipelineRequest>;
    requestDeserialize: grpc.deserialize<pps_pps_pb.DeletePipelineRequest>;
    responseSerialize: grpc.serialize<google_protobuf_empty_pb.Empty>;
    responseDeserialize: grpc.deserialize<google_protobuf_empty_pb.Empty>;
}
interface IAPIService_IDeletePipelines extends grpc.MethodDefinition<pps_pps_pb.DeletePipelinesRequest, pps_pps_pb.DeletePipelinesResponse> {
    path: "/pps_v2.API/DeletePipelines";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<pps_pps_pb.DeletePipelinesRequest>;
    requestDeserialize: grpc.deserialize<pps_pps_pb.DeletePipelinesRequest>;
    responseSerialize: grpc.serialize<pps_pps_pb.DeletePipelinesResponse>;
    responseDeserialize: grpc.deserialize<pps_pps_pb.DeletePipelinesResponse>;
}
interface IAPIService_IStartPipeline extends grpc.MethodDefinition<pps_pps_pb.StartPipelineRequest, google_protobuf_empty_pb.Empty> {
    path: "/pps_v2.API/StartPipeline";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<pps_pps_pb.StartPipelineRequest>;
    requestDeserialize: grpc.deserialize<pps_pps_pb.StartPipelineRequest>;
    responseSerialize: grpc.serialize<google_protobuf_empty_pb.Empty>;
    responseDeserialize: grpc.deserialize<google_protobuf_empty_pb.Empty>;
}
interface IAPIService_IStopPipeline extends grpc.MethodDefinition<pps_pps_pb.StopPipelineRequest, google_protobuf_empty_pb.Empty> {
    path: "/pps_v2.API/StopPipeline";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<pps_pps_pb.StopPipelineRequest>;
    requestDeserialize: grpc.deserialize<pps_pps_pb.StopPipelineRequest>;
    responseSerialize: grpc.serialize<google_protobuf_empty_pb.Empty>;
    responseDeserialize: grpc.deserialize<google_protobuf_empty_pb.Empty>;
}
interface IAPIService_IRunPipeline extends grpc.MethodDefinition<pps_pps_pb.RunPipelineRequest, google_protobuf_empty_pb.Empty> {
    path: "/pps_v2.API/RunPipeline";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<pps_pps_pb.RunPipelineRequest>;
    requestDeserialize: grpc.deserialize<pps_pps_pb.RunPipelineRequest>;
    responseSerialize: grpc.serialize<google_protobuf_empty_pb.Empty>;
    responseDeserialize: grpc.deserialize<google_protobuf_empty_pb.Empty>;
}
interface IAPIService_IRunCron extends grpc.MethodDefinition<pps_pps_pb.RunCronRequest, google_protobuf_empty_pb.Empty> {
    path: "/pps_v2.API/RunCron";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<pps_pps_pb.RunCronRequest>;
    requestDeserialize: grpc.deserialize<pps_pps_pb.RunCronRequest>;
    responseSerialize: grpc.serialize<google_protobuf_empty_pb.Empty>;
    responseDeserialize: grpc.deserialize<google_protobuf_empty_pb.Empty>;
}
interface IAPIService_ICreateSecret extends grpc.MethodDefinition<pps_pps_pb.CreateSecretRequest, google_protobuf_empty_pb.Empty> {
    path: "/pps_v2.API/CreateSecret";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<pps_pps_pb.CreateSecretRequest>;
    requestDeserialize: grpc.deserialize<pps_pps_pb.CreateSecretRequest>;
    responseSerialize: grpc.serialize<google_protobuf_empty_pb.Empty>;
    responseDeserialize: grpc.deserialize<google_protobuf_empty_pb.Empty>;
}
interface IAPIService_IDeleteSecret extends grpc.MethodDefinition<pps_pps_pb.DeleteSecretRequest, google_protobuf_empty_pb.Empty> {
    path: "/pps_v2.API/DeleteSecret";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<pps_pps_pb.DeleteSecretRequest>;
    requestDeserialize: grpc.deserialize<pps_pps_pb.DeleteSecretRequest>;
    responseSerialize: grpc.serialize<google_protobuf_empty_pb.Empty>;
    responseDeserialize: grpc.deserialize<google_protobuf_empty_pb.Empty>;
}
interface IAPIService_IListSecret extends grpc.MethodDefinition<google_protobuf_empty_pb.Empty, pps_pps_pb.SecretInfos> {
    path: "/pps_v2.API/ListSecret";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<google_protobuf_empty_pb.Empty>;
    requestDeserialize: grpc.deserialize<google_protobuf_empty_pb.Empty>;
    responseSerialize: grpc.serialize<pps_pps_pb.SecretInfos>;
    responseDeserialize: grpc.deserialize<pps_pps_pb.SecretInfos>;
}
interface IAPIService_IInspectSecret extends grpc.MethodDefinition<pps_pps_pb.InspectSecretRequest, pps_pps_pb.SecretInfo> {
    path: "/pps_v2.API/InspectSecret";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<pps_pps_pb.InspectSecretRequest>;
    requestDeserialize: grpc.deserialize<pps_pps_pb.InspectSecretRequest>;
    responseSerialize: grpc.serialize<pps_pps_pb.SecretInfo>;
    responseDeserialize: grpc.deserialize<pps_pps_pb.SecretInfo>;
}
interface IAPIService_IDeleteAll extends grpc.MethodDefinition<google_protobuf_empty_pb.Empty, google_protobuf_empty_pb.Empty> {
    path: "/pps_v2.API/DeleteAll";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<google_protobuf_empty_pb.Empty>;
    requestDeserialize: grpc.deserialize<google_protobuf_empty_pb.Empty>;
    responseSerialize: grpc.serialize<google_protobuf_empty_pb.Empty>;
    responseDeserialize: grpc.deserialize<google_protobuf_empty_pb.Empty>;
}
interface IAPIService_IGetLogs extends grpc.MethodDefinition<pps_pps_pb.GetLogsRequest, pps_pps_pb.LogMessage> {
    path: "/pps_v2.API/GetLogs";
    requestStream: false;
    responseStream: true;
    requestSerialize: grpc.serialize<pps_pps_pb.GetLogsRequest>;
    requestDeserialize: grpc.deserialize<pps_pps_pb.GetLogsRequest>;
    responseSerialize: grpc.serialize<pps_pps_pb.LogMessage>;
    responseDeserialize: grpc.deserialize<pps_pps_pb.LogMessage>;
}
interface IAPIService_IActivateAuth extends grpc.MethodDefinition<pps_pps_pb.ActivateAuthRequest, pps_pps_pb.ActivateAuthResponse> {
    path: "/pps_v2.API/ActivateAuth";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<pps_pps_pb.ActivateAuthRequest>;
    requestDeserialize: grpc.deserialize<pps_pps_pb.ActivateAuthRequest>;
    responseSerialize: grpc.serialize<pps_pps_pb.ActivateAuthResponse>;
    responseDeserialize: grpc.deserialize<pps_pps_pb.ActivateAuthResponse>;
}
interface IAPIService_IUpdateJobState extends grpc.MethodDefinition<pps_pps_pb.UpdateJobStateRequest, google_protobuf_empty_pb.Empty> {
    path: "/pps_v2.API/UpdateJobState";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<pps_pps_pb.UpdateJobStateRequest>;
    requestDeserialize: grpc.deserialize<pps_pps_pb.UpdateJobStateRequest>;
    responseSerialize: grpc.serialize<google_protobuf_empty_pb.Empty>;
    responseDeserialize: grpc.deserialize<google_protobuf_empty_pb.Empty>;
}
interface IAPIService_IRunLoadTest extends grpc.MethodDefinition<pps_pps_pb.RunLoadTestRequest, pps_pps_pb.RunLoadTestResponse> {
    path: "/pps_v2.API/RunLoadTest";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<pps_pps_pb.RunLoadTestRequest>;
    requestDeserialize: grpc.deserialize<pps_pps_pb.RunLoadTestRequest>;
    responseSerialize: grpc.serialize<pps_pps_pb.RunLoadTestResponse>;
    responseDeserialize: grpc.deserialize<pps_pps_pb.RunLoadTestResponse>;
}
interface IAPIService_IRunLoadTestDefault extends grpc.MethodDefinition<google_protobuf_empty_pb.Empty, pps_pps_pb.RunLoadTestResponse> {
    path: "/pps_v2.API/RunLoadTestDefault";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<google_protobuf_empty_pb.Empty>;
    requestDeserialize: grpc.deserialize<google_protobuf_empty_pb.Empty>;
    responseSerialize: grpc.serialize<pps_pps_pb.RunLoadTestResponse>;
    responseDeserialize: grpc.deserialize<pps_pps_pb.RunLoadTestResponse>;
}
interface IAPIService_IRenderTemplate extends grpc.MethodDefinition<pps_pps_pb.RenderTemplateRequest, pps_pps_pb.RenderTemplateResponse> {
    path: "/pps_v2.API/RenderTemplate";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<pps_pps_pb.RenderTemplateRequest>;
    requestDeserialize: grpc.deserialize<pps_pps_pb.RenderTemplateRequest>;
    responseSerialize: grpc.serialize<pps_pps_pb.RenderTemplateResponse>;
    responseDeserialize: grpc.deserialize<pps_pps_pb.RenderTemplateResponse>;
}
interface IAPIService_IListTask extends grpc.MethodDefinition<task_task_pb.ListTaskRequest, task_task_pb.TaskInfo> {
    path: "/pps_v2.API/ListTask";
    requestStream: false;
    responseStream: true;
    requestSerialize: grpc.serialize<task_task_pb.ListTaskRequest>;
    requestDeserialize: grpc.deserialize<task_task_pb.ListTaskRequest>;
    responseSerialize: grpc.serialize<task_task_pb.TaskInfo>;
    responseDeserialize: grpc.deserialize<task_task_pb.TaskInfo>;
}
interface IAPIService_IGetKubeEvents extends grpc.MethodDefinition<pps_pps_pb.LokiRequest, pps_pps_pb.LokiLogMessage> {
    path: "/pps_v2.API/GetKubeEvents";
    requestStream: false;
    responseStream: true;
    requestSerialize: grpc.serialize<pps_pps_pb.LokiRequest>;
    requestDeserialize: grpc.deserialize<pps_pps_pb.LokiRequest>;
    responseSerialize: grpc.serialize<pps_pps_pb.LokiLogMessage>;
    responseDeserialize: grpc.deserialize<pps_pps_pb.LokiLogMessage>;
}
interface IAPIService_IQueryLoki extends grpc.MethodDefinition<pps_pps_pb.LokiRequest, pps_pps_pb.LokiLogMessage> {
    path: "/pps_v2.API/QueryLoki";
    requestStream: false;
    responseStream: true;
    requestSerialize: grpc.serialize<pps_pps_pb.LokiRequest>;
    requestDeserialize: grpc.deserialize<pps_pps_pb.LokiRequest>;
    responseSerialize: grpc.serialize<pps_pps_pb.LokiLogMessage>;
    responseDeserialize: grpc.deserialize<pps_pps_pb.LokiLogMessage>;
}
interface IAPIService_IGetClusterDefaults extends grpc.MethodDefinition<pps_pps_pb.GetClusterDefaultsRequest, pps_pps_pb.GetClusterDefaultsResponse> {
    path: "/pps_v2.API/GetClusterDefaults";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<pps_pps_pb.GetClusterDefaultsRequest>;
    requestDeserialize: grpc.deserialize<pps_pps_pb.GetClusterDefaultsRequest>;
    responseSerialize: grpc.serialize<pps_pps_pb.GetClusterDefaultsResponse>;
    responseDeserialize: grpc.deserialize<pps_pps_pb.GetClusterDefaultsResponse>;
}
interface IAPIService_ISetClusterDefaults extends grpc.MethodDefinition<pps_pps_pb.SetClusterDefaultsRequest, pps_pps_pb.SetClusterDefaultsResponse> {
    path: "/pps_v2.API/SetClusterDefaults";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<pps_pps_pb.SetClusterDefaultsRequest>;
    requestDeserialize: grpc.deserialize<pps_pps_pb.SetClusterDefaultsRequest>;
    responseSerialize: grpc.serialize<pps_pps_pb.SetClusterDefaultsResponse>;
    responseDeserialize: grpc.deserialize<pps_pps_pb.SetClusterDefaultsResponse>;
}

export const APIService: IAPIService;

export interface IAPIServer extends grpc.UntypedServiceImplementation {
    inspectJob: grpc.handleUnaryCall<pps_pps_pb.InspectJobRequest, pps_pps_pb.JobInfo>;
    inspectJobSet: grpc.handleServerStreamingCall<pps_pps_pb.InspectJobSetRequest, pps_pps_pb.JobInfo>;
    listJob: grpc.handleServerStreamingCall<pps_pps_pb.ListJobRequest, pps_pps_pb.JobInfo>;
    listJobSet: grpc.handleServerStreamingCall<pps_pps_pb.ListJobSetRequest, pps_pps_pb.JobSetInfo>;
    subscribeJob: grpc.handleServerStreamingCall<pps_pps_pb.SubscribeJobRequest, pps_pps_pb.JobInfo>;
    deleteJob: grpc.handleUnaryCall<pps_pps_pb.DeleteJobRequest, google_protobuf_empty_pb.Empty>;
    stopJob: grpc.handleUnaryCall<pps_pps_pb.StopJobRequest, google_protobuf_empty_pb.Empty>;
    inspectDatum: grpc.handleUnaryCall<pps_pps_pb.InspectDatumRequest, pps_pps_pb.DatumInfo>;
    listDatum: grpc.handleServerStreamingCall<pps_pps_pb.ListDatumRequest, pps_pps_pb.DatumInfo>;
    restartDatum: grpc.handleUnaryCall<pps_pps_pb.RestartDatumRequest, google_protobuf_empty_pb.Empty>;
    createPipeline: grpc.handleUnaryCall<pps_pps_pb.CreatePipelineRequest, google_protobuf_empty_pb.Empty>;
    createPipelineV2: grpc.handleUnaryCall<pps_pps_pb.CreatePipelineV2Request, pps_pps_pb.CreatePipelineV2Response>;
    inspectPipeline: grpc.handleUnaryCall<pps_pps_pb.InspectPipelineRequest, pps_pps_pb.PipelineInfo>;
    listPipeline: grpc.handleServerStreamingCall<pps_pps_pb.ListPipelineRequest, pps_pps_pb.PipelineInfo>;
    deletePipeline: grpc.handleUnaryCall<pps_pps_pb.DeletePipelineRequest, google_protobuf_empty_pb.Empty>;
    deletePipelines: grpc.handleUnaryCall<pps_pps_pb.DeletePipelinesRequest, pps_pps_pb.DeletePipelinesResponse>;
    startPipeline: grpc.handleUnaryCall<pps_pps_pb.StartPipelineRequest, google_protobuf_empty_pb.Empty>;
    stopPipeline: grpc.handleUnaryCall<pps_pps_pb.StopPipelineRequest, google_protobuf_empty_pb.Empty>;
    runPipeline: grpc.handleUnaryCall<pps_pps_pb.RunPipelineRequest, google_protobuf_empty_pb.Empty>;
    runCron: grpc.handleUnaryCall<pps_pps_pb.RunCronRequest, google_protobuf_empty_pb.Empty>;
    createSecret: grpc.handleUnaryCall<pps_pps_pb.CreateSecretRequest, google_protobuf_empty_pb.Empty>;
    deleteSecret: grpc.handleUnaryCall<pps_pps_pb.DeleteSecretRequest, google_protobuf_empty_pb.Empty>;
    listSecret: grpc.handleUnaryCall<google_protobuf_empty_pb.Empty, pps_pps_pb.SecretInfos>;
    inspectSecret: grpc.handleUnaryCall<pps_pps_pb.InspectSecretRequest, pps_pps_pb.SecretInfo>;
    deleteAll: grpc.handleUnaryCall<google_protobuf_empty_pb.Empty, google_protobuf_empty_pb.Empty>;
    getLogs: grpc.handleServerStreamingCall<pps_pps_pb.GetLogsRequest, pps_pps_pb.LogMessage>;
    activateAuth: grpc.handleUnaryCall<pps_pps_pb.ActivateAuthRequest, pps_pps_pb.ActivateAuthResponse>;
    updateJobState: grpc.handleUnaryCall<pps_pps_pb.UpdateJobStateRequest, google_protobuf_empty_pb.Empty>;
    runLoadTest: grpc.handleUnaryCall<pps_pps_pb.RunLoadTestRequest, pps_pps_pb.RunLoadTestResponse>;
    runLoadTestDefault: grpc.handleUnaryCall<google_protobuf_empty_pb.Empty, pps_pps_pb.RunLoadTestResponse>;
    renderTemplate: grpc.handleUnaryCall<pps_pps_pb.RenderTemplateRequest, pps_pps_pb.RenderTemplateResponse>;
    listTask: grpc.handleServerStreamingCall<task_task_pb.ListTaskRequest, task_task_pb.TaskInfo>;
    getKubeEvents: grpc.handleServerStreamingCall<pps_pps_pb.LokiRequest, pps_pps_pb.LokiLogMessage>;
    queryLoki: grpc.handleServerStreamingCall<pps_pps_pb.LokiRequest, pps_pps_pb.LokiLogMessage>;
    getClusterDefaults: grpc.handleUnaryCall<pps_pps_pb.GetClusterDefaultsRequest, pps_pps_pb.GetClusterDefaultsResponse>;
    setClusterDefaults: grpc.handleUnaryCall<pps_pps_pb.SetClusterDefaultsRequest, pps_pps_pb.SetClusterDefaultsResponse>;
}

export interface IAPIClient {
    inspectJob(request: pps_pps_pb.InspectJobRequest, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.JobInfo) => void): grpc.ClientUnaryCall;
    inspectJob(request: pps_pps_pb.InspectJobRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.JobInfo) => void): grpc.ClientUnaryCall;
    inspectJob(request: pps_pps_pb.InspectJobRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.JobInfo) => void): grpc.ClientUnaryCall;
    inspectJobSet(request: pps_pps_pb.InspectJobSetRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pps_pps_pb.JobInfo>;
    inspectJobSet(request: pps_pps_pb.InspectJobSetRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pps_pps_pb.JobInfo>;
    listJob(request: pps_pps_pb.ListJobRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pps_pps_pb.JobInfo>;
    listJob(request: pps_pps_pb.ListJobRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pps_pps_pb.JobInfo>;
    listJobSet(request: pps_pps_pb.ListJobSetRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pps_pps_pb.JobSetInfo>;
    listJobSet(request: pps_pps_pb.ListJobSetRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pps_pps_pb.JobSetInfo>;
    subscribeJob(request: pps_pps_pb.SubscribeJobRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pps_pps_pb.JobInfo>;
    subscribeJob(request: pps_pps_pb.SubscribeJobRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pps_pps_pb.JobInfo>;
    deleteJob(request: pps_pps_pb.DeleteJobRequest, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    deleteJob(request: pps_pps_pb.DeleteJobRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    deleteJob(request: pps_pps_pb.DeleteJobRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    stopJob(request: pps_pps_pb.StopJobRequest, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    stopJob(request: pps_pps_pb.StopJobRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    stopJob(request: pps_pps_pb.StopJobRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    inspectDatum(request: pps_pps_pb.InspectDatumRequest, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.DatumInfo) => void): grpc.ClientUnaryCall;
    inspectDatum(request: pps_pps_pb.InspectDatumRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.DatumInfo) => void): grpc.ClientUnaryCall;
    inspectDatum(request: pps_pps_pb.InspectDatumRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.DatumInfo) => void): grpc.ClientUnaryCall;
    listDatum(request: pps_pps_pb.ListDatumRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pps_pps_pb.DatumInfo>;
    listDatum(request: pps_pps_pb.ListDatumRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pps_pps_pb.DatumInfo>;
    restartDatum(request: pps_pps_pb.RestartDatumRequest, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    restartDatum(request: pps_pps_pb.RestartDatumRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    restartDatum(request: pps_pps_pb.RestartDatumRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    createPipeline(request: pps_pps_pb.CreatePipelineRequest, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    createPipeline(request: pps_pps_pb.CreatePipelineRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    createPipeline(request: pps_pps_pb.CreatePipelineRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    createPipelineV2(request: pps_pps_pb.CreatePipelineV2Request, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.CreatePipelineV2Response) => void): grpc.ClientUnaryCall;
    createPipelineV2(request: pps_pps_pb.CreatePipelineV2Request, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.CreatePipelineV2Response) => void): grpc.ClientUnaryCall;
    createPipelineV2(request: pps_pps_pb.CreatePipelineV2Request, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.CreatePipelineV2Response) => void): grpc.ClientUnaryCall;
    inspectPipeline(request: pps_pps_pb.InspectPipelineRequest, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.PipelineInfo) => void): grpc.ClientUnaryCall;
    inspectPipeline(request: pps_pps_pb.InspectPipelineRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.PipelineInfo) => void): grpc.ClientUnaryCall;
    inspectPipeline(request: pps_pps_pb.InspectPipelineRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.PipelineInfo) => void): grpc.ClientUnaryCall;
    listPipeline(request: pps_pps_pb.ListPipelineRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pps_pps_pb.PipelineInfo>;
    listPipeline(request: pps_pps_pb.ListPipelineRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pps_pps_pb.PipelineInfo>;
    deletePipeline(request: pps_pps_pb.DeletePipelineRequest, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    deletePipeline(request: pps_pps_pb.DeletePipelineRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    deletePipeline(request: pps_pps_pb.DeletePipelineRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    deletePipelines(request: pps_pps_pb.DeletePipelinesRequest, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.DeletePipelinesResponse) => void): grpc.ClientUnaryCall;
    deletePipelines(request: pps_pps_pb.DeletePipelinesRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.DeletePipelinesResponse) => void): grpc.ClientUnaryCall;
    deletePipelines(request: pps_pps_pb.DeletePipelinesRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.DeletePipelinesResponse) => void): grpc.ClientUnaryCall;
    startPipeline(request: pps_pps_pb.StartPipelineRequest, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    startPipeline(request: pps_pps_pb.StartPipelineRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    startPipeline(request: pps_pps_pb.StartPipelineRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    stopPipeline(request: pps_pps_pb.StopPipelineRequest, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    stopPipeline(request: pps_pps_pb.StopPipelineRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    stopPipeline(request: pps_pps_pb.StopPipelineRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    runPipeline(request: pps_pps_pb.RunPipelineRequest, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    runPipeline(request: pps_pps_pb.RunPipelineRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    runPipeline(request: pps_pps_pb.RunPipelineRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    runCron(request: pps_pps_pb.RunCronRequest, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    runCron(request: pps_pps_pb.RunCronRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    runCron(request: pps_pps_pb.RunCronRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    createSecret(request: pps_pps_pb.CreateSecretRequest, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    createSecret(request: pps_pps_pb.CreateSecretRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    createSecret(request: pps_pps_pb.CreateSecretRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    deleteSecret(request: pps_pps_pb.DeleteSecretRequest, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    deleteSecret(request: pps_pps_pb.DeleteSecretRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    deleteSecret(request: pps_pps_pb.DeleteSecretRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    listSecret(request: google_protobuf_empty_pb.Empty, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.SecretInfos) => void): grpc.ClientUnaryCall;
    listSecret(request: google_protobuf_empty_pb.Empty, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.SecretInfos) => void): grpc.ClientUnaryCall;
    listSecret(request: google_protobuf_empty_pb.Empty, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.SecretInfos) => void): grpc.ClientUnaryCall;
    inspectSecret(request: pps_pps_pb.InspectSecretRequest, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.SecretInfo) => void): grpc.ClientUnaryCall;
    inspectSecret(request: pps_pps_pb.InspectSecretRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.SecretInfo) => void): grpc.ClientUnaryCall;
    inspectSecret(request: pps_pps_pb.InspectSecretRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.SecretInfo) => void): grpc.ClientUnaryCall;
    deleteAll(request: google_protobuf_empty_pb.Empty, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    deleteAll(request: google_protobuf_empty_pb.Empty, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    deleteAll(request: google_protobuf_empty_pb.Empty, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    getLogs(request: pps_pps_pb.GetLogsRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pps_pps_pb.LogMessage>;
    getLogs(request: pps_pps_pb.GetLogsRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pps_pps_pb.LogMessage>;
    activateAuth(request: pps_pps_pb.ActivateAuthRequest, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.ActivateAuthResponse) => void): grpc.ClientUnaryCall;
    activateAuth(request: pps_pps_pb.ActivateAuthRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.ActivateAuthResponse) => void): grpc.ClientUnaryCall;
    activateAuth(request: pps_pps_pb.ActivateAuthRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.ActivateAuthResponse) => void): grpc.ClientUnaryCall;
    updateJobState(request: pps_pps_pb.UpdateJobStateRequest, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    updateJobState(request: pps_pps_pb.UpdateJobStateRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    updateJobState(request: pps_pps_pb.UpdateJobStateRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    runLoadTest(request: pps_pps_pb.RunLoadTestRequest, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.RunLoadTestResponse) => void): grpc.ClientUnaryCall;
    runLoadTest(request: pps_pps_pb.RunLoadTestRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.RunLoadTestResponse) => void): grpc.ClientUnaryCall;
    runLoadTest(request: pps_pps_pb.RunLoadTestRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.RunLoadTestResponse) => void): grpc.ClientUnaryCall;
    runLoadTestDefault(request: google_protobuf_empty_pb.Empty, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.RunLoadTestResponse) => void): grpc.ClientUnaryCall;
    runLoadTestDefault(request: google_protobuf_empty_pb.Empty, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.RunLoadTestResponse) => void): grpc.ClientUnaryCall;
    runLoadTestDefault(request: google_protobuf_empty_pb.Empty, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.RunLoadTestResponse) => void): grpc.ClientUnaryCall;
    renderTemplate(request: pps_pps_pb.RenderTemplateRequest, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.RenderTemplateResponse) => void): grpc.ClientUnaryCall;
    renderTemplate(request: pps_pps_pb.RenderTemplateRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.RenderTemplateResponse) => void): grpc.ClientUnaryCall;
    renderTemplate(request: pps_pps_pb.RenderTemplateRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.RenderTemplateResponse) => void): grpc.ClientUnaryCall;
    listTask(request: task_task_pb.ListTaskRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<task_task_pb.TaskInfo>;
    listTask(request: task_task_pb.ListTaskRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<task_task_pb.TaskInfo>;
    getKubeEvents(request: pps_pps_pb.LokiRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pps_pps_pb.LokiLogMessage>;
    getKubeEvents(request: pps_pps_pb.LokiRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pps_pps_pb.LokiLogMessage>;
    queryLoki(request: pps_pps_pb.LokiRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pps_pps_pb.LokiLogMessage>;
    queryLoki(request: pps_pps_pb.LokiRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pps_pps_pb.LokiLogMessage>;
    getClusterDefaults(request: pps_pps_pb.GetClusterDefaultsRequest, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.GetClusterDefaultsResponse) => void): grpc.ClientUnaryCall;
    getClusterDefaults(request: pps_pps_pb.GetClusterDefaultsRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.GetClusterDefaultsResponse) => void): grpc.ClientUnaryCall;
    getClusterDefaults(request: pps_pps_pb.GetClusterDefaultsRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.GetClusterDefaultsResponse) => void): grpc.ClientUnaryCall;
    setClusterDefaults(request: pps_pps_pb.SetClusterDefaultsRequest, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.SetClusterDefaultsResponse) => void): grpc.ClientUnaryCall;
    setClusterDefaults(request: pps_pps_pb.SetClusterDefaultsRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.SetClusterDefaultsResponse) => void): grpc.ClientUnaryCall;
    setClusterDefaults(request: pps_pps_pb.SetClusterDefaultsRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.SetClusterDefaultsResponse) => void): grpc.ClientUnaryCall;
}

export class APIClient extends grpc.Client implements IAPIClient {
    constructor(address: string, credentials: grpc.ChannelCredentials, options?: Partial<grpc.ClientOptions>);
    public inspectJob(request: pps_pps_pb.InspectJobRequest, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.JobInfo) => void): grpc.ClientUnaryCall;
    public inspectJob(request: pps_pps_pb.InspectJobRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.JobInfo) => void): grpc.ClientUnaryCall;
    public inspectJob(request: pps_pps_pb.InspectJobRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.JobInfo) => void): grpc.ClientUnaryCall;
    public inspectJobSet(request: pps_pps_pb.InspectJobSetRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pps_pps_pb.JobInfo>;
    public inspectJobSet(request: pps_pps_pb.InspectJobSetRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pps_pps_pb.JobInfo>;
    public listJob(request: pps_pps_pb.ListJobRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pps_pps_pb.JobInfo>;
    public listJob(request: pps_pps_pb.ListJobRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pps_pps_pb.JobInfo>;
    public listJobSet(request: pps_pps_pb.ListJobSetRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pps_pps_pb.JobSetInfo>;
    public listJobSet(request: pps_pps_pb.ListJobSetRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pps_pps_pb.JobSetInfo>;
    public subscribeJob(request: pps_pps_pb.SubscribeJobRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pps_pps_pb.JobInfo>;
    public subscribeJob(request: pps_pps_pb.SubscribeJobRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pps_pps_pb.JobInfo>;
    public deleteJob(request: pps_pps_pb.DeleteJobRequest, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public deleteJob(request: pps_pps_pb.DeleteJobRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public deleteJob(request: pps_pps_pb.DeleteJobRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public stopJob(request: pps_pps_pb.StopJobRequest, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public stopJob(request: pps_pps_pb.StopJobRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public stopJob(request: pps_pps_pb.StopJobRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public inspectDatum(request: pps_pps_pb.InspectDatumRequest, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.DatumInfo) => void): grpc.ClientUnaryCall;
    public inspectDatum(request: pps_pps_pb.InspectDatumRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.DatumInfo) => void): grpc.ClientUnaryCall;
    public inspectDatum(request: pps_pps_pb.InspectDatumRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.DatumInfo) => void): grpc.ClientUnaryCall;
    public listDatum(request: pps_pps_pb.ListDatumRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pps_pps_pb.DatumInfo>;
    public listDatum(request: pps_pps_pb.ListDatumRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pps_pps_pb.DatumInfo>;
    public restartDatum(request: pps_pps_pb.RestartDatumRequest, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public restartDatum(request: pps_pps_pb.RestartDatumRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public restartDatum(request: pps_pps_pb.RestartDatumRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public createPipeline(request: pps_pps_pb.CreatePipelineRequest, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public createPipeline(request: pps_pps_pb.CreatePipelineRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public createPipeline(request: pps_pps_pb.CreatePipelineRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public createPipelineV2(request: pps_pps_pb.CreatePipelineV2Request, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.CreatePipelineV2Response) => void): grpc.ClientUnaryCall;
    public createPipelineV2(request: pps_pps_pb.CreatePipelineV2Request, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.CreatePipelineV2Response) => void): grpc.ClientUnaryCall;
    public createPipelineV2(request: pps_pps_pb.CreatePipelineV2Request, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.CreatePipelineV2Response) => void): grpc.ClientUnaryCall;
    public inspectPipeline(request: pps_pps_pb.InspectPipelineRequest, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.PipelineInfo) => void): grpc.ClientUnaryCall;
    public inspectPipeline(request: pps_pps_pb.InspectPipelineRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.PipelineInfo) => void): grpc.ClientUnaryCall;
    public inspectPipeline(request: pps_pps_pb.InspectPipelineRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.PipelineInfo) => void): grpc.ClientUnaryCall;
    public listPipeline(request: pps_pps_pb.ListPipelineRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pps_pps_pb.PipelineInfo>;
    public listPipeline(request: pps_pps_pb.ListPipelineRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pps_pps_pb.PipelineInfo>;
    public deletePipeline(request: pps_pps_pb.DeletePipelineRequest, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public deletePipeline(request: pps_pps_pb.DeletePipelineRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public deletePipeline(request: pps_pps_pb.DeletePipelineRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public deletePipelines(request: pps_pps_pb.DeletePipelinesRequest, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.DeletePipelinesResponse) => void): grpc.ClientUnaryCall;
    public deletePipelines(request: pps_pps_pb.DeletePipelinesRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.DeletePipelinesResponse) => void): grpc.ClientUnaryCall;
    public deletePipelines(request: pps_pps_pb.DeletePipelinesRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.DeletePipelinesResponse) => void): grpc.ClientUnaryCall;
    public startPipeline(request: pps_pps_pb.StartPipelineRequest, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public startPipeline(request: pps_pps_pb.StartPipelineRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public startPipeline(request: pps_pps_pb.StartPipelineRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public stopPipeline(request: pps_pps_pb.StopPipelineRequest, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public stopPipeline(request: pps_pps_pb.StopPipelineRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public stopPipeline(request: pps_pps_pb.StopPipelineRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public runPipeline(request: pps_pps_pb.RunPipelineRequest, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public runPipeline(request: pps_pps_pb.RunPipelineRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public runPipeline(request: pps_pps_pb.RunPipelineRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public runCron(request: pps_pps_pb.RunCronRequest, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public runCron(request: pps_pps_pb.RunCronRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public runCron(request: pps_pps_pb.RunCronRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public createSecret(request: pps_pps_pb.CreateSecretRequest, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public createSecret(request: pps_pps_pb.CreateSecretRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public createSecret(request: pps_pps_pb.CreateSecretRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public deleteSecret(request: pps_pps_pb.DeleteSecretRequest, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public deleteSecret(request: pps_pps_pb.DeleteSecretRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public deleteSecret(request: pps_pps_pb.DeleteSecretRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public listSecret(request: google_protobuf_empty_pb.Empty, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.SecretInfos) => void): grpc.ClientUnaryCall;
    public listSecret(request: google_protobuf_empty_pb.Empty, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.SecretInfos) => void): grpc.ClientUnaryCall;
    public listSecret(request: google_protobuf_empty_pb.Empty, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.SecretInfos) => void): grpc.ClientUnaryCall;
    public inspectSecret(request: pps_pps_pb.InspectSecretRequest, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.SecretInfo) => void): grpc.ClientUnaryCall;
    public inspectSecret(request: pps_pps_pb.InspectSecretRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.SecretInfo) => void): grpc.ClientUnaryCall;
    public inspectSecret(request: pps_pps_pb.InspectSecretRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.SecretInfo) => void): grpc.ClientUnaryCall;
    public deleteAll(request: google_protobuf_empty_pb.Empty, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public deleteAll(request: google_protobuf_empty_pb.Empty, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public deleteAll(request: google_protobuf_empty_pb.Empty, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public getLogs(request: pps_pps_pb.GetLogsRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pps_pps_pb.LogMessage>;
    public getLogs(request: pps_pps_pb.GetLogsRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pps_pps_pb.LogMessage>;
    public activateAuth(request: pps_pps_pb.ActivateAuthRequest, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.ActivateAuthResponse) => void): grpc.ClientUnaryCall;
    public activateAuth(request: pps_pps_pb.ActivateAuthRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.ActivateAuthResponse) => void): grpc.ClientUnaryCall;
    public activateAuth(request: pps_pps_pb.ActivateAuthRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.ActivateAuthResponse) => void): grpc.ClientUnaryCall;
    public updateJobState(request: pps_pps_pb.UpdateJobStateRequest, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public updateJobState(request: pps_pps_pb.UpdateJobStateRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public updateJobState(request: pps_pps_pb.UpdateJobStateRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public runLoadTest(request: pps_pps_pb.RunLoadTestRequest, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.RunLoadTestResponse) => void): grpc.ClientUnaryCall;
    public runLoadTest(request: pps_pps_pb.RunLoadTestRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.RunLoadTestResponse) => void): grpc.ClientUnaryCall;
    public runLoadTest(request: pps_pps_pb.RunLoadTestRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.RunLoadTestResponse) => void): grpc.ClientUnaryCall;
    public runLoadTestDefault(request: google_protobuf_empty_pb.Empty, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.RunLoadTestResponse) => void): grpc.ClientUnaryCall;
    public runLoadTestDefault(request: google_protobuf_empty_pb.Empty, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.RunLoadTestResponse) => void): grpc.ClientUnaryCall;
    public runLoadTestDefault(request: google_protobuf_empty_pb.Empty, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.RunLoadTestResponse) => void): grpc.ClientUnaryCall;
    public renderTemplate(request: pps_pps_pb.RenderTemplateRequest, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.RenderTemplateResponse) => void): grpc.ClientUnaryCall;
    public renderTemplate(request: pps_pps_pb.RenderTemplateRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.RenderTemplateResponse) => void): grpc.ClientUnaryCall;
    public renderTemplate(request: pps_pps_pb.RenderTemplateRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.RenderTemplateResponse) => void): grpc.ClientUnaryCall;
    public listTask(request: task_task_pb.ListTaskRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<task_task_pb.TaskInfo>;
    public listTask(request: task_task_pb.ListTaskRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<task_task_pb.TaskInfo>;
    public getKubeEvents(request: pps_pps_pb.LokiRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pps_pps_pb.LokiLogMessage>;
    public getKubeEvents(request: pps_pps_pb.LokiRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pps_pps_pb.LokiLogMessage>;
    public queryLoki(request: pps_pps_pb.LokiRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pps_pps_pb.LokiLogMessage>;
    public queryLoki(request: pps_pps_pb.LokiRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pps_pps_pb.LokiLogMessage>;
    public getClusterDefaults(request: pps_pps_pb.GetClusterDefaultsRequest, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.GetClusterDefaultsResponse) => void): grpc.ClientUnaryCall;
    public getClusterDefaults(request: pps_pps_pb.GetClusterDefaultsRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.GetClusterDefaultsResponse) => void): grpc.ClientUnaryCall;
    public getClusterDefaults(request: pps_pps_pb.GetClusterDefaultsRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.GetClusterDefaultsResponse) => void): grpc.ClientUnaryCall;
    public setClusterDefaults(request: pps_pps_pb.SetClusterDefaultsRequest, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.SetClusterDefaultsResponse) => void): grpc.ClientUnaryCall;
    public setClusterDefaults(request: pps_pps_pb.SetClusterDefaultsRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.SetClusterDefaultsResponse) => void): grpc.ClientUnaryCall;
    public setClusterDefaults(request: pps_pps_pb.SetClusterDefaultsRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.SetClusterDefaultsResponse) => void): grpc.ClientUnaryCall;
}
