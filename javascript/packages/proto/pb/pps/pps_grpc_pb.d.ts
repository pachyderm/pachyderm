// package: pps
// file: pps/pps.proto

/* tslint:disable */
/* eslint-disable */

import * as grpc from "@grpc/grpc-js";
import {handleClientStreamingCall} from "@grpc/grpc-js/build/src/server-call";
import * as pps_pps_pb from "../pps/pps_pb";
import * as google_protobuf_empty_pb from "google-protobuf/google/protobuf/empty_pb";
import * as google_protobuf_timestamp_pb from "google-protobuf/google/protobuf/timestamp_pb";
import * as google_protobuf_duration_pb from "google-protobuf/google/protobuf/duration_pb";
import * as gogoproto_gogo_pb from "../gogoproto/gogo_pb";
import * as pfs_pfs_pb from "../pfs/pfs_pb";

interface IAPIService extends grpc.ServiceDefinition<grpc.UntypedServiceImplementation> {
    createJob: IAPIService_ICreateJob;
    inspectJob: IAPIService_IInspectJob;
    listJob: IAPIService_IListJob;
    flushJob: IAPIService_IFlushJob;
    deleteJob: IAPIService_IDeleteJob;
    stopJob: IAPIService_IStopJob;
    inspectDatum: IAPIService_IInspectDatum;
    listDatum: IAPIService_IListDatum;
    restartDatum: IAPIService_IRestartDatum;
    createPipeline: IAPIService_ICreatePipeline;
    inspectPipeline: IAPIService_IInspectPipeline;
    listPipeline: IAPIService_IListPipeline;
    deletePipeline: IAPIService_IDeletePipeline;
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
}

interface IAPIService_ICreateJob extends grpc.MethodDefinition<pps_pps_pb.CreateJobRequest, pps_pps_pb.Job> {
    path: "/pps.API/CreateJob";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<pps_pps_pb.CreateJobRequest>;
    requestDeserialize: grpc.deserialize<pps_pps_pb.CreateJobRequest>;
    responseSerialize: grpc.serialize<pps_pps_pb.Job>;
    responseDeserialize: grpc.deserialize<pps_pps_pb.Job>;
}
interface IAPIService_IInspectJob extends grpc.MethodDefinition<pps_pps_pb.InspectJobRequest, pps_pps_pb.JobInfo> {
    path: "/pps.API/InspectJob";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<pps_pps_pb.InspectJobRequest>;
    requestDeserialize: grpc.deserialize<pps_pps_pb.InspectJobRequest>;
    responseSerialize: grpc.serialize<pps_pps_pb.JobInfo>;
    responseDeserialize: grpc.deserialize<pps_pps_pb.JobInfo>;
}
interface IAPIService_IListJob extends grpc.MethodDefinition<pps_pps_pb.ListJobRequest, pps_pps_pb.JobInfo> {
    path: "/pps.API/ListJob";
    requestStream: false;
    responseStream: true;
    requestSerialize: grpc.serialize<pps_pps_pb.ListJobRequest>;
    requestDeserialize: grpc.deserialize<pps_pps_pb.ListJobRequest>;
    responseSerialize: grpc.serialize<pps_pps_pb.JobInfo>;
    responseDeserialize: grpc.deserialize<pps_pps_pb.JobInfo>;
}
interface IAPIService_IFlushJob extends grpc.MethodDefinition<pps_pps_pb.FlushJobRequest, pps_pps_pb.JobInfo> {
    path: "/pps.API/FlushJob";
    requestStream: false;
    responseStream: true;
    requestSerialize: grpc.serialize<pps_pps_pb.FlushJobRequest>;
    requestDeserialize: grpc.deserialize<pps_pps_pb.FlushJobRequest>;
    responseSerialize: grpc.serialize<pps_pps_pb.JobInfo>;
    responseDeserialize: grpc.deserialize<pps_pps_pb.JobInfo>;
}
interface IAPIService_IDeleteJob extends grpc.MethodDefinition<pps_pps_pb.DeleteJobRequest, google_protobuf_empty_pb.Empty> {
    path: "/pps.API/DeleteJob";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<pps_pps_pb.DeleteJobRequest>;
    requestDeserialize: grpc.deserialize<pps_pps_pb.DeleteJobRequest>;
    responseSerialize: grpc.serialize<google_protobuf_empty_pb.Empty>;
    responseDeserialize: grpc.deserialize<google_protobuf_empty_pb.Empty>;
}
interface IAPIService_IStopJob extends grpc.MethodDefinition<pps_pps_pb.StopJobRequest, google_protobuf_empty_pb.Empty> {
    path: "/pps.API/StopJob";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<pps_pps_pb.StopJobRequest>;
    requestDeserialize: grpc.deserialize<pps_pps_pb.StopJobRequest>;
    responseSerialize: grpc.serialize<google_protobuf_empty_pb.Empty>;
    responseDeserialize: grpc.deserialize<google_protobuf_empty_pb.Empty>;
}
interface IAPIService_IInspectDatum extends grpc.MethodDefinition<pps_pps_pb.InspectDatumRequest, pps_pps_pb.DatumInfo> {
    path: "/pps.API/InspectDatum";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<pps_pps_pb.InspectDatumRequest>;
    requestDeserialize: grpc.deserialize<pps_pps_pb.InspectDatumRequest>;
    responseSerialize: grpc.serialize<pps_pps_pb.DatumInfo>;
    responseDeserialize: grpc.deserialize<pps_pps_pb.DatumInfo>;
}
interface IAPIService_IListDatum extends grpc.MethodDefinition<pps_pps_pb.ListDatumRequest, pps_pps_pb.DatumInfo> {
    path: "/pps.API/ListDatum";
    requestStream: false;
    responseStream: true;
    requestSerialize: grpc.serialize<pps_pps_pb.ListDatumRequest>;
    requestDeserialize: grpc.deserialize<pps_pps_pb.ListDatumRequest>;
    responseSerialize: grpc.serialize<pps_pps_pb.DatumInfo>;
    responseDeserialize: grpc.deserialize<pps_pps_pb.DatumInfo>;
}
interface IAPIService_IRestartDatum extends grpc.MethodDefinition<pps_pps_pb.RestartDatumRequest, google_protobuf_empty_pb.Empty> {
    path: "/pps.API/RestartDatum";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<pps_pps_pb.RestartDatumRequest>;
    requestDeserialize: grpc.deserialize<pps_pps_pb.RestartDatumRequest>;
    responseSerialize: grpc.serialize<google_protobuf_empty_pb.Empty>;
    responseDeserialize: grpc.deserialize<google_protobuf_empty_pb.Empty>;
}
interface IAPIService_ICreatePipeline extends grpc.MethodDefinition<pps_pps_pb.CreatePipelineRequest, google_protobuf_empty_pb.Empty> {
    path: "/pps.API/CreatePipeline";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<pps_pps_pb.CreatePipelineRequest>;
    requestDeserialize: grpc.deserialize<pps_pps_pb.CreatePipelineRequest>;
    responseSerialize: grpc.serialize<google_protobuf_empty_pb.Empty>;
    responseDeserialize: grpc.deserialize<google_protobuf_empty_pb.Empty>;
}
interface IAPIService_IInspectPipeline extends grpc.MethodDefinition<pps_pps_pb.InspectPipelineRequest, pps_pps_pb.PipelineInfo> {
    path: "/pps.API/InspectPipeline";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<pps_pps_pb.InspectPipelineRequest>;
    requestDeserialize: grpc.deserialize<pps_pps_pb.InspectPipelineRequest>;
    responseSerialize: grpc.serialize<pps_pps_pb.PipelineInfo>;
    responseDeserialize: grpc.deserialize<pps_pps_pb.PipelineInfo>;
}
interface IAPIService_IListPipeline extends grpc.MethodDefinition<pps_pps_pb.ListPipelineRequest, pps_pps_pb.PipelineInfos> {
    path: "/pps.API/ListPipeline";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<pps_pps_pb.ListPipelineRequest>;
    requestDeserialize: grpc.deserialize<pps_pps_pb.ListPipelineRequest>;
    responseSerialize: grpc.serialize<pps_pps_pb.PipelineInfos>;
    responseDeserialize: grpc.deserialize<pps_pps_pb.PipelineInfos>;
}
interface IAPIService_IDeletePipeline extends grpc.MethodDefinition<pps_pps_pb.DeletePipelineRequest, google_protobuf_empty_pb.Empty> {
    path: "/pps.API/DeletePipeline";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<pps_pps_pb.DeletePipelineRequest>;
    requestDeserialize: grpc.deserialize<pps_pps_pb.DeletePipelineRequest>;
    responseSerialize: grpc.serialize<google_protobuf_empty_pb.Empty>;
    responseDeserialize: grpc.deserialize<google_protobuf_empty_pb.Empty>;
}
interface IAPIService_IStartPipeline extends grpc.MethodDefinition<pps_pps_pb.StartPipelineRequest, google_protobuf_empty_pb.Empty> {
    path: "/pps.API/StartPipeline";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<pps_pps_pb.StartPipelineRequest>;
    requestDeserialize: grpc.deserialize<pps_pps_pb.StartPipelineRequest>;
    responseSerialize: grpc.serialize<google_protobuf_empty_pb.Empty>;
    responseDeserialize: grpc.deserialize<google_protobuf_empty_pb.Empty>;
}
interface IAPIService_IStopPipeline extends grpc.MethodDefinition<pps_pps_pb.StopPipelineRequest, google_protobuf_empty_pb.Empty> {
    path: "/pps.API/StopPipeline";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<pps_pps_pb.StopPipelineRequest>;
    requestDeserialize: grpc.deserialize<pps_pps_pb.StopPipelineRequest>;
    responseSerialize: grpc.serialize<google_protobuf_empty_pb.Empty>;
    responseDeserialize: grpc.deserialize<google_protobuf_empty_pb.Empty>;
}
interface IAPIService_IRunPipeline extends grpc.MethodDefinition<pps_pps_pb.RunPipelineRequest, google_protobuf_empty_pb.Empty> {
    path: "/pps.API/RunPipeline";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<pps_pps_pb.RunPipelineRequest>;
    requestDeserialize: grpc.deserialize<pps_pps_pb.RunPipelineRequest>;
    responseSerialize: grpc.serialize<google_protobuf_empty_pb.Empty>;
    responseDeserialize: grpc.deserialize<google_protobuf_empty_pb.Empty>;
}
interface IAPIService_IRunCron extends grpc.MethodDefinition<pps_pps_pb.RunCronRequest, google_protobuf_empty_pb.Empty> {
    path: "/pps.API/RunCron";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<pps_pps_pb.RunCronRequest>;
    requestDeserialize: grpc.deserialize<pps_pps_pb.RunCronRequest>;
    responseSerialize: grpc.serialize<google_protobuf_empty_pb.Empty>;
    responseDeserialize: grpc.deserialize<google_protobuf_empty_pb.Empty>;
}
interface IAPIService_ICreateSecret extends grpc.MethodDefinition<pps_pps_pb.CreateSecretRequest, google_protobuf_empty_pb.Empty> {
    path: "/pps.API/CreateSecret";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<pps_pps_pb.CreateSecretRequest>;
    requestDeserialize: grpc.deserialize<pps_pps_pb.CreateSecretRequest>;
    responseSerialize: grpc.serialize<google_protobuf_empty_pb.Empty>;
    responseDeserialize: grpc.deserialize<google_protobuf_empty_pb.Empty>;
}
interface IAPIService_IDeleteSecret extends grpc.MethodDefinition<pps_pps_pb.DeleteSecretRequest, google_protobuf_empty_pb.Empty> {
    path: "/pps.API/DeleteSecret";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<pps_pps_pb.DeleteSecretRequest>;
    requestDeserialize: grpc.deserialize<pps_pps_pb.DeleteSecretRequest>;
    responseSerialize: grpc.serialize<google_protobuf_empty_pb.Empty>;
    responseDeserialize: grpc.deserialize<google_protobuf_empty_pb.Empty>;
}
interface IAPIService_IListSecret extends grpc.MethodDefinition<google_protobuf_empty_pb.Empty, pps_pps_pb.SecretInfos> {
    path: "/pps.API/ListSecret";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<google_protobuf_empty_pb.Empty>;
    requestDeserialize: grpc.deserialize<google_protobuf_empty_pb.Empty>;
    responseSerialize: grpc.serialize<pps_pps_pb.SecretInfos>;
    responseDeserialize: grpc.deserialize<pps_pps_pb.SecretInfos>;
}
interface IAPIService_IInspectSecret extends grpc.MethodDefinition<pps_pps_pb.InspectSecretRequest, pps_pps_pb.SecretInfo> {
    path: "/pps.API/InspectSecret";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<pps_pps_pb.InspectSecretRequest>;
    requestDeserialize: grpc.deserialize<pps_pps_pb.InspectSecretRequest>;
    responseSerialize: grpc.serialize<pps_pps_pb.SecretInfo>;
    responseDeserialize: grpc.deserialize<pps_pps_pb.SecretInfo>;
}
interface IAPIService_IDeleteAll extends grpc.MethodDefinition<google_protobuf_empty_pb.Empty, google_protobuf_empty_pb.Empty> {
    path: "/pps.API/DeleteAll";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<google_protobuf_empty_pb.Empty>;
    requestDeserialize: grpc.deserialize<google_protobuf_empty_pb.Empty>;
    responseSerialize: grpc.serialize<google_protobuf_empty_pb.Empty>;
    responseDeserialize: grpc.deserialize<google_protobuf_empty_pb.Empty>;
}
interface IAPIService_IGetLogs extends grpc.MethodDefinition<pps_pps_pb.GetLogsRequest, pps_pps_pb.LogMessage> {
    path: "/pps.API/GetLogs";
    requestStream: false;
    responseStream: true;
    requestSerialize: grpc.serialize<pps_pps_pb.GetLogsRequest>;
    requestDeserialize: grpc.deserialize<pps_pps_pb.GetLogsRequest>;
    responseSerialize: grpc.serialize<pps_pps_pb.LogMessage>;
    responseDeserialize: grpc.deserialize<pps_pps_pb.LogMessage>;
}
interface IAPIService_IActivateAuth extends grpc.MethodDefinition<pps_pps_pb.ActivateAuthRequest, pps_pps_pb.ActivateAuthResponse> {
    path: "/pps.API/ActivateAuth";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<pps_pps_pb.ActivateAuthRequest>;
    requestDeserialize: grpc.deserialize<pps_pps_pb.ActivateAuthRequest>;
    responseSerialize: grpc.serialize<pps_pps_pb.ActivateAuthResponse>;
    responseDeserialize: grpc.deserialize<pps_pps_pb.ActivateAuthResponse>;
}
interface IAPIService_IUpdateJobState extends grpc.MethodDefinition<pps_pps_pb.UpdateJobStateRequest, google_protobuf_empty_pb.Empty> {
    path: "/pps.API/UpdateJobState";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<pps_pps_pb.UpdateJobStateRequest>;
    requestDeserialize: grpc.deserialize<pps_pps_pb.UpdateJobStateRequest>;
    responseSerialize: grpc.serialize<google_protobuf_empty_pb.Empty>;
    responseDeserialize: grpc.deserialize<google_protobuf_empty_pb.Empty>;
}

export const APIService: IAPIService;

export interface IAPIServer {
    createJob: grpc.handleUnaryCall<pps_pps_pb.CreateJobRequest, pps_pps_pb.Job>;
    inspectJob: grpc.handleUnaryCall<pps_pps_pb.InspectJobRequest, pps_pps_pb.JobInfo>;
    listJob: grpc.handleServerStreamingCall<pps_pps_pb.ListJobRequest, pps_pps_pb.JobInfo>;
    flushJob: grpc.handleServerStreamingCall<pps_pps_pb.FlushJobRequest, pps_pps_pb.JobInfo>;
    deleteJob: grpc.handleUnaryCall<pps_pps_pb.DeleteJobRequest, google_protobuf_empty_pb.Empty>;
    stopJob: grpc.handleUnaryCall<pps_pps_pb.StopJobRequest, google_protobuf_empty_pb.Empty>;
    inspectDatum: grpc.handleUnaryCall<pps_pps_pb.InspectDatumRequest, pps_pps_pb.DatumInfo>;
    listDatum: grpc.handleServerStreamingCall<pps_pps_pb.ListDatumRequest, pps_pps_pb.DatumInfo>;
    restartDatum: grpc.handleUnaryCall<pps_pps_pb.RestartDatumRequest, google_protobuf_empty_pb.Empty>;
    createPipeline: grpc.handleUnaryCall<pps_pps_pb.CreatePipelineRequest, google_protobuf_empty_pb.Empty>;
    inspectPipeline: grpc.handleUnaryCall<pps_pps_pb.InspectPipelineRequest, pps_pps_pb.PipelineInfo>;
    listPipeline: grpc.handleUnaryCall<pps_pps_pb.ListPipelineRequest, pps_pps_pb.PipelineInfos>;
    deletePipeline: grpc.handleUnaryCall<pps_pps_pb.DeletePipelineRequest, google_protobuf_empty_pb.Empty>;
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
}

export interface IAPIClient {
    createJob(request: pps_pps_pb.CreateJobRequest, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.Job) => void): grpc.ClientUnaryCall;
    createJob(request: pps_pps_pb.CreateJobRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.Job) => void): grpc.ClientUnaryCall;
    createJob(request: pps_pps_pb.CreateJobRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.Job) => void): grpc.ClientUnaryCall;
    inspectJob(request: pps_pps_pb.InspectJobRequest, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.JobInfo) => void): grpc.ClientUnaryCall;
    inspectJob(request: pps_pps_pb.InspectJobRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.JobInfo) => void): grpc.ClientUnaryCall;
    inspectJob(request: pps_pps_pb.InspectJobRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.JobInfo) => void): grpc.ClientUnaryCall;
    listJob(request: pps_pps_pb.ListJobRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pps_pps_pb.JobInfo>;
    listJob(request: pps_pps_pb.ListJobRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pps_pps_pb.JobInfo>;
    flushJob(request: pps_pps_pb.FlushJobRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pps_pps_pb.JobInfo>;
    flushJob(request: pps_pps_pb.FlushJobRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pps_pps_pb.JobInfo>;
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
    inspectPipeline(request: pps_pps_pb.InspectPipelineRequest, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.PipelineInfo) => void): grpc.ClientUnaryCall;
    inspectPipeline(request: pps_pps_pb.InspectPipelineRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.PipelineInfo) => void): grpc.ClientUnaryCall;
    inspectPipeline(request: pps_pps_pb.InspectPipelineRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.PipelineInfo) => void): grpc.ClientUnaryCall;
    listPipeline(request: pps_pps_pb.ListPipelineRequest, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.PipelineInfos) => void): grpc.ClientUnaryCall;
    listPipeline(request: pps_pps_pb.ListPipelineRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.PipelineInfos) => void): grpc.ClientUnaryCall;
    listPipeline(request: pps_pps_pb.ListPipelineRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.PipelineInfos) => void): grpc.ClientUnaryCall;
    deletePipeline(request: pps_pps_pb.DeletePipelineRequest, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    deletePipeline(request: pps_pps_pb.DeletePipelineRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    deletePipeline(request: pps_pps_pb.DeletePipelineRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
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
}

export class APIClient extends grpc.Client implements IAPIClient {
    constructor(address: string, credentials: grpc.ChannelCredentials, options?: Partial<grpc.ClientOptions>);
    public createJob(request: pps_pps_pb.CreateJobRequest, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.Job) => void): grpc.ClientUnaryCall;
    public createJob(request: pps_pps_pb.CreateJobRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.Job) => void): grpc.ClientUnaryCall;
    public createJob(request: pps_pps_pb.CreateJobRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.Job) => void): grpc.ClientUnaryCall;
    public inspectJob(request: pps_pps_pb.InspectJobRequest, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.JobInfo) => void): grpc.ClientUnaryCall;
    public inspectJob(request: pps_pps_pb.InspectJobRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.JobInfo) => void): grpc.ClientUnaryCall;
    public inspectJob(request: pps_pps_pb.InspectJobRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.JobInfo) => void): grpc.ClientUnaryCall;
    public listJob(request: pps_pps_pb.ListJobRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pps_pps_pb.JobInfo>;
    public listJob(request: pps_pps_pb.ListJobRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pps_pps_pb.JobInfo>;
    public flushJob(request: pps_pps_pb.FlushJobRequest, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pps_pps_pb.JobInfo>;
    public flushJob(request: pps_pps_pb.FlushJobRequest, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<pps_pps_pb.JobInfo>;
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
    public inspectPipeline(request: pps_pps_pb.InspectPipelineRequest, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.PipelineInfo) => void): grpc.ClientUnaryCall;
    public inspectPipeline(request: pps_pps_pb.InspectPipelineRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.PipelineInfo) => void): grpc.ClientUnaryCall;
    public inspectPipeline(request: pps_pps_pb.InspectPipelineRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.PipelineInfo) => void): grpc.ClientUnaryCall;
    public listPipeline(request: pps_pps_pb.ListPipelineRequest, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.PipelineInfos) => void): grpc.ClientUnaryCall;
    public listPipeline(request: pps_pps_pb.ListPipelineRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.PipelineInfos) => void): grpc.ClientUnaryCall;
    public listPipeline(request: pps_pps_pb.ListPipelineRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: pps_pps_pb.PipelineInfos) => void): grpc.ClientUnaryCall;
    public deletePipeline(request: pps_pps_pb.DeletePipelineRequest, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public deletePipeline(request: pps_pps_pb.DeletePipelineRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
    public deletePipeline(request: pps_pps_pb.DeletePipelineRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: google_protobuf_empty_pb.Empty) => void): grpc.ClientUnaryCall;
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
}
