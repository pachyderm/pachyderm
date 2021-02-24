// package: mock
// file: mock/mock.proto

/* tslint:disable */
/* eslint-disable */

import * as grpc from "@grpc/grpc-js";
import {handleClientStreamingCall} from "@grpc/grpc-js/build/src/server-call";
import * as mock_mock_pb from "../mock/mock_pb";
import * as google_protobuf_empty_pb from "google-protobuf/google/protobuf/empty_pb";
import * as google_protobuf_timestamp_pb from "google-protobuf/google/protobuf/timestamp_pb";

interface IAPIService extends grpc.ServiceDefinition<grpc.UntypedServiceImplementation> {
    inspectProject: IAPIService_IInspectProject;
    listProject: IAPIService_IListProject;
}

interface IAPIService_IInspectProject extends grpc.MethodDefinition<mock_mock_pb.ProjectRequest, mock_mock_pb.Project> {
    path: "/mock.API/InspectProject";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<mock_mock_pb.ProjectRequest>;
    requestDeserialize: grpc.deserialize<mock_mock_pb.ProjectRequest>;
    responseSerialize: grpc.serialize<mock_mock_pb.Project>;
    responseDeserialize: grpc.deserialize<mock_mock_pb.Project>;
}
interface IAPIService_IListProject extends grpc.MethodDefinition<google_protobuf_empty_pb.Empty, mock_mock_pb.Projects> {
    path: "/mock.API/ListProject";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<google_protobuf_empty_pb.Empty>;
    requestDeserialize: grpc.deserialize<google_protobuf_empty_pb.Empty>;
    responseSerialize: grpc.serialize<mock_mock_pb.Projects>;
    responseDeserialize: grpc.deserialize<mock_mock_pb.Projects>;
}

export const APIService: IAPIService;

export interface IAPIServer {
    inspectProject: grpc.handleUnaryCall<mock_mock_pb.ProjectRequest, mock_mock_pb.Project>;
    listProject: grpc.handleUnaryCall<google_protobuf_empty_pb.Empty, mock_mock_pb.Projects>;
}

export interface IAPIClient {
    inspectProject(request: mock_mock_pb.ProjectRequest, callback: (error: grpc.ServiceError | null, response: mock_mock_pb.Project) => void): grpc.ClientUnaryCall;
    inspectProject(request: mock_mock_pb.ProjectRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: mock_mock_pb.Project) => void): grpc.ClientUnaryCall;
    inspectProject(request: mock_mock_pb.ProjectRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: mock_mock_pb.Project) => void): grpc.ClientUnaryCall;
    listProject(request: google_protobuf_empty_pb.Empty, callback: (error: grpc.ServiceError | null, response: mock_mock_pb.Projects) => void): grpc.ClientUnaryCall;
    listProject(request: google_protobuf_empty_pb.Empty, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: mock_mock_pb.Projects) => void): grpc.ClientUnaryCall;
    listProject(request: google_protobuf_empty_pb.Empty, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: mock_mock_pb.Projects) => void): grpc.ClientUnaryCall;
}

export class APIClient extends grpc.Client implements IAPIClient {
    constructor(address: string, credentials: grpc.ChannelCredentials, options?: Partial<grpc.ClientOptions>);
    public inspectProject(request: mock_mock_pb.ProjectRequest, callback: (error: grpc.ServiceError | null, response: mock_mock_pb.Project) => void): grpc.ClientUnaryCall;
    public inspectProject(request: mock_mock_pb.ProjectRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: mock_mock_pb.Project) => void): grpc.ClientUnaryCall;
    public inspectProject(request: mock_mock_pb.ProjectRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: mock_mock_pb.Project) => void): grpc.ClientUnaryCall;
    public listProject(request: google_protobuf_empty_pb.Empty, callback: (error: grpc.ServiceError | null, response: mock_mock_pb.Projects) => void): grpc.ClientUnaryCall;
    public listProject(request: google_protobuf_empty_pb.Empty, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: mock_mock_pb.Projects) => void): grpc.ClientUnaryCall;
    public listProject(request: google_protobuf_empty_pb.Empty, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: mock_mock_pb.Projects) => void): grpc.ClientUnaryCall;
}
