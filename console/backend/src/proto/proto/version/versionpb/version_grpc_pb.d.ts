// package: versionpb_v2
// file: version/versionpb/version.proto

/* tslint:disable */
/* eslint-disable */

import * as grpc from "@grpc/grpc-js";
import * as version_versionpb_version_pb from "../../version/versionpb/version_pb";
import * as google_protobuf_empty_pb from "google-protobuf/google/protobuf/empty_pb";

interface IAPIService extends grpc.ServiceDefinition<grpc.UntypedServiceImplementation> {
    getVersion: IAPIService_IGetVersion;
}

interface IAPIService_IGetVersion extends grpc.MethodDefinition<google_protobuf_empty_pb.Empty, version_versionpb_version_pb.Version> {
    path: "/versionpb_v2.API/GetVersion";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<google_protobuf_empty_pb.Empty>;
    requestDeserialize: grpc.deserialize<google_protobuf_empty_pb.Empty>;
    responseSerialize: grpc.serialize<version_versionpb_version_pb.Version>;
    responseDeserialize: grpc.deserialize<version_versionpb_version_pb.Version>;
}

export const APIService: IAPIService;

export interface IAPIServer extends grpc.UntypedServiceImplementation {
    getVersion: grpc.handleUnaryCall<google_protobuf_empty_pb.Empty, version_versionpb_version_pb.Version>;
}

export interface IAPIClient {
    getVersion(request: google_protobuf_empty_pb.Empty, callback: (error: grpc.ServiceError | null, response: version_versionpb_version_pb.Version) => void): grpc.ClientUnaryCall;
    getVersion(request: google_protobuf_empty_pb.Empty, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: version_versionpb_version_pb.Version) => void): grpc.ClientUnaryCall;
    getVersion(request: google_protobuf_empty_pb.Empty, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: version_versionpb_version_pb.Version) => void): grpc.ClientUnaryCall;
}

export class APIClient extends grpc.Client implements IAPIClient {
    constructor(address: string, credentials: grpc.ChannelCredentials, options?: Partial<grpc.ClientOptions>);
    public getVersion(request: google_protobuf_empty_pb.Empty, callback: (error: grpc.ServiceError | null, response: version_versionpb_version_pb.Version) => void): grpc.ClientUnaryCall;
    public getVersion(request: google_protobuf_empty_pb.Empty, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: version_versionpb_version_pb.Version) => void): grpc.ClientUnaryCall;
    public getVersion(request: google_protobuf_empty_pb.Empty, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: version_versionpb_version_pb.Version) => void): grpc.ClientUnaryCall;
}
