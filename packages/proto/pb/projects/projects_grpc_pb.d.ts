// package: mock
// file: projects/projects.proto

/* tslint:disable */
/* eslint-disable */

import * as grpc from "@grpc/grpc-js";
import {handleClientStreamingCall} from "@grpc/grpc-js/build/src/server-call";
import * as projects_projects_pb from "../projects/projects_pb";
import * as google_protobuf_empty_pb from "google-protobuf/google/protobuf/empty_pb";
import * as google_protobuf_timestamp_pb from "google-protobuf/google/protobuf/timestamp_pb";

interface IAPIService extends grpc.ServiceDefinition<grpc.UntypedServiceImplementation> {
    inspectProject: IAPIService_IInspectProject;
    listProject: IAPIService_IListProject;
}

interface IAPIService_IInspectProject extends grpc.MethodDefinition<projects_projects_pb.ProjectRequest, projects_projects_pb.Project> {
    path: "/mock.API/InspectProject";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<projects_projects_pb.ProjectRequest>;
    requestDeserialize: grpc.deserialize<projects_projects_pb.ProjectRequest>;
    responseSerialize: grpc.serialize<projects_projects_pb.Project>;
    responseDeserialize: grpc.deserialize<projects_projects_pb.Project>;
}
interface IAPIService_IListProject extends grpc.MethodDefinition<google_protobuf_empty_pb.Empty, projects_projects_pb.Projects> {
    path: "/mock.API/ListProject";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<google_protobuf_empty_pb.Empty>;
    requestDeserialize: grpc.deserialize<google_protobuf_empty_pb.Empty>;
    responseSerialize: grpc.serialize<projects_projects_pb.Projects>;
    responseDeserialize: grpc.deserialize<projects_projects_pb.Projects>;
}

export const APIService: IAPIService;

export interface IAPIServer extends grpc.UntypedServiceImplementation {
    inspectProject: grpc.handleUnaryCall<projects_projects_pb.ProjectRequest, projects_projects_pb.Project>;
    listProject: grpc.handleUnaryCall<google_protobuf_empty_pb.Empty, projects_projects_pb.Projects>;
}

export interface IAPIClient {
    inspectProject(request: projects_projects_pb.ProjectRequest, callback: (error: grpc.ServiceError | null, response: projects_projects_pb.Project) => void): grpc.ClientUnaryCall;
    inspectProject(request: projects_projects_pb.ProjectRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: projects_projects_pb.Project) => void): grpc.ClientUnaryCall;
    inspectProject(request: projects_projects_pb.ProjectRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: projects_projects_pb.Project) => void): grpc.ClientUnaryCall;
    listProject(request: google_protobuf_empty_pb.Empty, callback: (error: grpc.ServiceError | null, response: projects_projects_pb.Projects) => void): grpc.ClientUnaryCall;
    listProject(request: google_protobuf_empty_pb.Empty, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: projects_projects_pb.Projects) => void): grpc.ClientUnaryCall;
    listProject(request: google_protobuf_empty_pb.Empty, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: projects_projects_pb.Projects) => void): grpc.ClientUnaryCall;
}

export class APIClient extends grpc.Client implements IAPIClient {
    constructor(address: string, credentials: grpc.ChannelCredentials, options?: Partial<grpc.ClientOptions>);
    public inspectProject(request: projects_projects_pb.ProjectRequest, callback: (error: grpc.ServiceError | null, response: projects_projects_pb.Project) => void): grpc.ClientUnaryCall;
    public inspectProject(request: projects_projects_pb.ProjectRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: projects_projects_pb.Project) => void): grpc.ClientUnaryCall;
    public inspectProject(request: projects_projects_pb.ProjectRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: projects_projects_pb.Project) => void): grpc.ClientUnaryCall;
    public listProject(request: google_protobuf_empty_pb.Empty, callback: (error: grpc.ServiceError | null, response: projects_projects_pb.Projects) => void): grpc.ClientUnaryCall;
    public listProject(request: google_protobuf_empty_pb.Empty, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: projects_projects_pb.Projects) => void): grpc.ClientUnaryCall;
    public listProject(request: google_protobuf_empty_pb.Empty, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: projects_projects_pb.Projects) => void): grpc.ClientUnaryCall;
}
