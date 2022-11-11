// package: license_v2
// file: license/license.proto

/* tslint:disable */
/* eslint-disable */

import * as grpc from "@grpc/grpc-js";
import * as license_license_pb from "../license/license_pb";
import * as google_protobuf_timestamp_pb from "google-protobuf/google/protobuf/timestamp_pb";
import * as gogoproto_gogo_pb from "../gogoproto/gogo_pb";
import * as enterprise_enterprise_pb from "../enterprise/enterprise_pb";

interface IAPIService extends grpc.ServiceDefinition<grpc.UntypedServiceImplementation> {
    activate: IAPIService_IActivate;
    getActivationCode: IAPIService_IGetActivationCode;
    deleteAll: IAPIService_IDeleteAll;
    addCluster: IAPIService_IAddCluster;
    deleteCluster: IAPIService_IDeleteCluster;
    listClusters: IAPIService_IListClusters;
    updateCluster: IAPIService_IUpdateCluster;
    heartbeat: IAPIService_IHeartbeat;
    listUserClusters: IAPIService_IListUserClusters;
}

interface IAPIService_IActivate extends grpc.MethodDefinition<license_license_pb.ActivateRequest, license_license_pb.ActivateResponse> {
    path: "/license_v2.API/Activate";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<license_license_pb.ActivateRequest>;
    requestDeserialize: grpc.deserialize<license_license_pb.ActivateRequest>;
    responseSerialize: grpc.serialize<license_license_pb.ActivateResponse>;
    responseDeserialize: grpc.deserialize<license_license_pb.ActivateResponse>;
}
interface IAPIService_IGetActivationCode extends grpc.MethodDefinition<license_license_pb.GetActivationCodeRequest, license_license_pb.GetActivationCodeResponse> {
    path: "/license_v2.API/GetActivationCode";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<license_license_pb.GetActivationCodeRequest>;
    requestDeserialize: grpc.deserialize<license_license_pb.GetActivationCodeRequest>;
    responseSerialize: grpc.serialize<license_license_pb.GetActivationCodeResponse>;
    responseDeserialize: grpc.deserialize<license_license_pb.GetActivationCodeResponse>;
}
interface IAPIService_IDeleteAll extends grpc.MethodDefinition<license_license_pb.DeleteAllRequest, license_license_pb.DeleteAllResponse> {
    path: "/license_v2.API/DeleteAll";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<license_license_pb.DeleteAllRequest>;
    requestDeserialize: grpc.deserialize<license_license_pb.DeleteAllRequest>;
    responseSerialize: grpc.serialize<license_license_pb.DeleteAllResponse>;
    responseDeserialize: grpc.deserialize<license_license_pb.DeleteAllResponse>;
}
interface IAPIService_IAddCluster extends grpc.MethodDefinition<license_license_pb.AddClusterRequest, license_license_pb.AddClusterResponse> {
    path: "/license_v2.API/AddCluster";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<license_license_pb.AddClusterRequest>;
    requestDeserialize: grpc.deserialize<license_license_pb.AddClusterRequest>;
    responseSerialize: grpc.serialize<license_license_pb.AddClusterResponse>;
    responseDeserialize: grpc.deserialize<license_license_pb.AddClusterResponse>;
}
interface IAPIService_IDeleteCluster extends grpc.MethodDefinition<license_license_pb.DeleteClusterRequest, license_license_pb.DeleteClusterResponse> {
    path: "/license_v2.API/DeleteCluster";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<license_license_pb.DeleteClusterRequest>;
    requestDeserialize: grpc.deserialize<license_license_pb.DeleteClusterRequest>;
    responseSerialize: grpc.serialize<license_license_pb.DeleteClusterResponse>;
    responseDeserialize: grpc.deserialize<license_license_pb.DeleteClusterResponse>;
}
interface IAPIService_IListClusters extends grpc.MethodDefinition<license_license_pb.ListClustersRequest, license_license_pb.ListClustersResponse> {
    path: "/license_v2.API/ListClusters";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<license_license_pb.ListClustersRequest>;
    requestDeserialize: grpc.deserialize<license_license_pb.ListClustersRequest>;
    responseSerialize: grpc.serialize<license_license_pb.ListClustersResponse>;
    responseDeserialize: grpc.deserialize<license_license_pb.ListClustersResponse>;
}
interface IAPIService_IUpdateCluster extends grpc.MethodDefinition<license_license_pb.UpdateClusterRequest, license_license_pb.UpdateClusterResponse> {
    path: "/license_v2.API/UpdateCluster";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<license_license_pb.UpdateClusterRequest>;
    requestDeserialize: grpc.deserialize<license_license_pb.UpdateClusterRequest>;
    responseSerialize: grpc.serialize<license_license_pb.UpdateClusterResponse>;
    responseDeserialize: grpc.deserialize<license_license_pb.UpdateClusterResponse>;
}
interface IAPIService_IHeartbeat extends grpc.MethodDefinition<license_license_pb.HeartbeatRequest, license_license_pb.HeartbeatResponse> {
    path: "/license_v2.API/Heartbeat";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<license_license_pb.HeartbeatRequest>;
    requestDeserialize: grpc.deserialize<license_license_pb.HeartbeatRequest>;
    responseSerialize: grpc.serialize<license_license_pb.HeartbeatResponse>;
    responseDeserialize: grpc.deserialize<license_license_pb.HeartbeatResponse>;
}
interface IAPIService_IListUserClusters extends grpc.MethodDefinition<license_license_pb.ListUserClustersRequest, license_license_pb.ListUserClustersResponse> {
    path: "/license_v2.API/ListUserClusters";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<license_license_pb.ListUserClustersRequest>;
    requestDeserialize: grpc.deserialize<license_license_pb.ListUserClustersRequest>;
    responseSerialize: grpc.serialize<license_license_pb.ListUserClustersResponse>;
    responseDeserialize: grpc.deserialize<license_license_pb.ListUserClustersResponse>;
}

export const APIService: IAPIService;

export interface IAPIServer extends grpc.UntypedServiceImplementation {
    activate: grpc.handleUnaryCall<license_license_pb.ActivateRequest, license_license_pb.ActivateResponse>;
    getActivationCode: grpc.handleUnaryCall<license_license_pb.GetActivationCodeRequest, license_license_pb.GetActivationCodeResponse>;
    deleteAll: grpc.handleUnaryCall<license_license_pb.DeleteAllRequest, license_license_pb.DeleteAllResponse>;
    addCluster: grpc.handleUnaryCall<license_license_pb.AddClusterRequest, license_license_pb.AddClusterResponse>;
    deleteCluster: grpc.handleUnaryCall<license_license_pb.DeleteClusterRequest, license_license_pb.DeleteClusterResponse>;
    listClusters: grpc.handleUnaryCall<license_license_pb.ListClustersRequest, license_license_pb.ListClustersResponse>;
    updateCluster: grpc.handleUnaryCall<license_license_pb.UpdateClusterRequest, license_license_pb.UpdateClusterResponse>;
    heartbeat: grpc.handleUnaryCall<license_license_pb.HeartbeatRequest, license_license_pb.HeartbeatResponse>;
    listUserClusters: grpc.handleUnaryCall<license_license_pb.ListUserClustersRequest, license_license_pb.ListUserClustersResponse>;
}

export interface IAPIClient {
    activate(request: license_license_pb.ActivateRequest, callback: (error: grpc.ServiceError | null, response: license_license_pb.ActivateResponse) => void): grpc.ClientUnaryCall;
    activate(request: license_license_pb.ActivateRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: license_license_pb.ActivateResponse) => void): grpc.ClientUnaryCall;
    activate(request: license_license_pb.ActivateRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: license_license_pb.ActivateResponse) => void): grpc.ClientUnaryCall;
    getActivationCode(request: license_license_pb.GetActivationCodeRequest, callback: (error: grpc.ServiceError | null, response: license_license_pb.GetActivationCodeResponse) => void): grpc.ClientUnaryCall;
    getActivationCode(request: license_license_pb.GetActivationCodeRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: license_license_pb.GetActivationCodeResponse) => void): grpc.ClientUnaryCall;
    getActivationCode(request: license_license_pb.GetActivationCodeRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: license_license_pb.GetActivationCodeResponse) => void): grpc.ClientUnaryCall;
    deleteAll(request: license_license_pb.DeleteAllRequest, callback: (error: grpc.ServiceError | null, response: license_license_pb.DeleteAllResponse) => void): grpc.ClientUnaryCall;
    deleteAll(request: license_license_pb.DeleteAllRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: license_license_pb.DeleteAllResponse) => void): grpc.ClientUnaryCall;
    deleteAll(request: license_license_pb.DeleteAllRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: license_license_pb.DeleteAllResponse) => void): grpc.ClientUnaryCall;
    addCluster(request: license_license_pb.AddClusterRequest, callback: (error: grpc.ServiceError | null, response: license_license_pb.AddClusterResponse) => void): grpc.ClientUnaryCall;
    addCluster(request: license_license_pb.AddClusterRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: license_license_pb.AddClusterResponse) => void): grpc.ClientUnaryCall;
    addCluster(request: license_license_pb.AddClusterRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: license_license_pb.AddClusterResponse) => void): grpc.ClientUnaryCall;
    deleteCluster(request: license_license_pb.DeleteClusterRequest, callback: (error: grpc.ServiceError | null, response: license_license_pb.DeleteClusterResponse) => void): grpc.ClientUnaryCall;
    deleteCluster(request: license_license_pb.DeleteClusterRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: license_license_pb.DeleteClusterResponse) => void): grpc.ClientUnaryCall;
    deleteCluster(request: license_license_pb.DeleteClusterRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: license_license_pb.DeleteClusterResponse) => void): grpc.ClientUnaryCall;
    listClusters(request: license_license_pb.ListClustersRequest, callback: (error: grpc.ServiceError | null, response: license_license_pb.ListClustersResponse) => void): grpc.ClientUnaryCall;
    listClusters(request: license_license_pb.ListClustersRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: license_license_pb.ListClustersResponse) => void): grpc.ClientUnaryCall;
    listClusters(request: license_license_pb.ListClustersRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: license_license_pb.ListClustersResponse) => void): grpc.ClientUnaryCall;
    updateCluster(request: license_license_pb.UpdateClusterRequest, callback: (error: grpc.ServiceError | null, response: license_license_pb.UpdateClusterResponse) => void): grpc.ClientUnaryCall;
    updateCluster(request: license_license_pb.UpdateClusterRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: license_license_pb.UpdateClusterResponse) => void): grpc.ClientUnaryCall;
    updateCluster(request: license_license_pb.UpdateClusterRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: license_license_pb.UpdateClusterResponse) => void): grpc.ClientUnaryCall;
    heartbeat(request: license_license_pb.HeartbeatRequest, callback: (error: grpc.ServiceError | null, response: license_license_pb.HeartbeatResponse) => void): grpc.ClientUnaryCall;
    heartbeat(request: license_license_pb.HeartbeatRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: license_license_pb.HeartbeatResponse) => void): grpc.ClientUnaryCall;
    heartbeat(request: license_license_pb.HeartbeatRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: license_license_pb.HeartbeatResponse) => void): grpc.ClientUnaryCall;
    listUserClusters(request: license_license_pb.ListUserClustersRequest, callback: (error: grpc.ServiceError | null, response: license_license_pb.ListUserClustersResponse) => void): grpc.ClientUnaryCall;
    listUserClusters(request: license_license_pb.ListUserClustersRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: license_license_pb.ListUserClustersResponse) => void): grpc.ClientUnaryCall;
    listUserClusters(request: license_license_pb.ListUserClustersRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: license_license_pb.ListUserClustersResponse) => void): grpc.ClientUnaryCall;
}

export class APIClient extends grpc.Client implements IAPIClient {
    constructor(address: string, credentials: grpc.ChannelCredentials, options?: Partial<grpc.ClientOptions>);
    public activate(request: license_license_pb.ActivateRequest, callback: (error: grpc.ServiceError | null, response: license_license_pb.ActivateResponse) => void): grpc.ClientUnaryCall;
    public activate(request: license_license_pb.ActivateRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: license_license_pb.ActivateResponse) => void): grpc.ClientUnaryCall;
    public activate(request: license_license_pb.ActivateRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: license_license_pb.ActivateResponse) => void): grpc.ClientUnaryCall;
    public getActivationCode(request: license_license_pb.GetActivationCodeRequest, callback: (error: grpc.ServiceError | null, response: license_license_pb.GetActivationCodeResponse) => void): grpc.ClientUnaryCall;
    public getActivationCode(request: license_license_pb.GetActivationCodeRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: license_license_pb.GetActivationCodeResponse) => void): grpc.ClientUnaryCall;
    public getActivationCode(request: license_license_pb.GetActivationCodeRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: license_license_pb.GetActivationCodeResponse) => void): grpc.ClientUnaryCall;
    public deleteAll(request: license_license_pb.DeleteAllRequest, callback: (error: grpc.ServiceError | null, response: license_license_pb.DeleteAllResponse) => void): grpc.ClientUnaryCall;
    public deleteAll(request: license_license_pb.DeleteAllRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: license_license_pb.DeleteAllResponse) => void): grpc.ClientUnaryCall;
    public deleteAll(request: license_license_pb.DeleteAllRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: license_license_pb.DeleteAllResponse) => void): grpc.ClientUnaryCall;
    public addCluster(request: license_license_pb.AddClusterRequest, callback: (error: grpc.ServiceError | null, response: license_license_pb.AddClusterResponse) => void): grpc.ClientUnaryCall;
    public addCluster(request: license_license_pb.AddClusterRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: license_license_pb.AddClusterResponse) => void): grpc.ClientUnaryCall;
    public addCluster(request: license_license_pb.AddClusterRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: license_license_pb.AddClusterResponse) => void): grpc.ClientUnaryCall;
    public deleteCluster(request: license_license_pb.DeleteClusterRequest, callback: (error: grpc.ServiceError | null, response: license_license_pb.DeleteClusterResponse) => void): grpc.ClientUnaryCall;
    public deleteCluster(request: license_license_pb.DeleteClusterRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: license_license_pb.DeleteClusterResponse) => void): grpc.ClientUnaryCall;
    public deleteCluster(request: license_license_pb.DeleteClusterRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: license_license_pb.DeleteClusterResponse) => void): grpc.ClientUnaryCall;
    public listClusters(request: license_license_pb.ListClustersRequest, callback: (error: grpc.ServiceError | null, response: license_license_pb.ListClustersResponse) => void): grpc.ClientUnaryCall;
    public listClusters(request: license_license_pb.ListClustersRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: license_license_pb.ListClustersResponse) => void): grpc.ClientUnaryCall;
    public listClusters(request: license_license_pb.ListClustersRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: license_license_pb.ListClustersResponse) => void): grpc.ClientUnaryCall;
    public updateCluster(request: license_license_pb.UpdateClusterRequest, callback: (error: grpc.ServiceError | null, response: license_license_pb.UpdateClusterResponse) => void): grpc.ClientUnaryCall;
    public updateCluster(request: license_license_pb.UpdateClusterRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: license_license_pb.UpdateClusterResponse) => void): grpc.ClientUnaryCall;
    public updateCluster(request: license_license_pb.UpdateClusterRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: license_license_pb.UpdateClusterResponse) => void): grpc.ClientUnaryCall;
    public heartbeat(request: license_license_pb.HeartbeatRequest, callback: (error: grpc.ServiceError | null, response: license_license_pb.HeartbeatResponse) => void): grpc.ClientUnaryCall;
    public heartbeat(request: license_license_pb.HeartbeatRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: license_license_pb.HeartbeatResponse) => void): grpc.ClientUnaryCall;
    public heartbeat(request: license_license_pb.HeartbeatRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: license_license_pb.HeartbeatResponse) => void): grpc.ClientUnaryCall;
    public listUserClusters(request: license_license_pb.ListUserClustersRequest, callback: (error: grpc.ServiceError | null, response: license_license_pb.ListUserClustersResponse) => void): grpc.ClientUnaryCall;
    public listUserClusters(request: license_license_pb.ListUserClustersRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: license_license_pb.ListUserClustersResponse) => void): grpc.ClientUnaryCall;
    public listUserClusters(request: license_license_pb.ListUserClustersRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: license_license_pb.ListUserClustersResponse) => void): grpc.ClientUnaryCall;
}
