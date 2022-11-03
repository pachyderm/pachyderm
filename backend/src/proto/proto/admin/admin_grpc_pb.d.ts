// package: admin_v2
// file: admin/admin.proto

/* tslint:disable */
/* eslint-disable */

import * as grpc from "@grpc/grpc-js";
import * as admin_admin_pb from "../admin/admin_pb";
import * as google_protobuf_empty_pb from "google-protobuf/google/protobuf/empty_pb";
import * as gogoproto_gogo_pb from "../gogoproto/gogo_pb";

interface IAPIService extends grpc.ServiceDefinition<grpc.UntypedServiceImplementation> {
    inspectCluster: IAPIService_IInspectCluster;
}

interface IAPIService_IInspectCluster extends grpc.MethodDefinition<google_protobuf_empty_pb.Empty, admin_admin_pb.ClusterInfo> {
    path: "/admin_v2.API/InspectCluster";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<google_protobuf_empty_pb.Empty>;
    requestDeserialize: grpc.deserialize<google_protobuf_empty_pb.Empty>;
    responseSerialize: grpc.serialize<admin_admin_pb.ClusterInfo>;
    responseDeserialize: grpc.deserialize<admin_admin_pb.ClusterInfo>;
}

export const APIService: IAPIService;

export interface IAPIServer extends grpc.UntypedServiceImplementation {
    inspectCluster: grpc.handleUnaryCall<google_protobuf_empty_pb.Empty, admin_admin_pb.ClusterInfo>;
}

export interface IAPIClient {
    inspectCluster(request: google_protobuf_empty_pb.Empty, callback: (error: grpc.ServiceError | null, response: admin_admin_pb.ClusterInfo) => void): grpc.ClientUnaryCall;
    inspectCluster(request: google_protobuf_empty_pb.Empty, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: admin_admin_pb.ClusterInfo) => void): grpc.ClientUnaryCall;
    inspectCluster(request: google_protobuf_empty_pb.Empty, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: admin_admin_pb.ClusterInfo) => void): grpc.ClientUnaryCall;
}

export class APIClient extends grpc.Client implements IAPIClient {
    constructor(address: string, credentials: grpc.ChannelCredentials, options?: Partial<grpc.ClientOptions>);
    public inspectCluster(request: google_protobuf_empty_pb.Empty, callback: (error: grpc.ServiceError | null, response: admin_admin_pb.ClusterInfo) => void): grpc.ClientUnaryCall;
    public inspectCluster(request: google_protobuf_empty_pb.Empty, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: admin_admin_pb.ClusterInfo) => void): grpc.ClientUnaryCall;
    public inspectCluster(request: google_protobuf_empty_pb.Empty, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: admin_admin_pb.ClusterInfo) => void): grpc.ClientUnaryCall;
}
