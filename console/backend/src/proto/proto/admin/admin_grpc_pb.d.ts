// package: admin_v2
// file: admin/admin.proto

/* tslint:disable */
/* eslint-disable */

import * as grpc from "@grpc/grpc-js";
import * as admin_admin_pb from "../admin/admin_pb";
import * as version_versionpb_version_pb from "../version/versionpb/version_pb";
import * as pfs_pfs_pb from "../pfs/pfs_pb";

interface IAPIService extends grpc.ServiceDefinition<grpc.UntypedServiceImplementation> {
    inspectCluster: IAPIService_IInspectCluster;
}

interface IAPIService_IInspectCluster extends grpc.MethodDefinition<admin_admin_pb.InspectClusterRequest, admin_admin_pb.ClusterInfo> {
    path: "/admin_v2.API/InspectCluster";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<admin_admin_pb.InspectClusterRequest>;
    requestDeserialize: grpc.deserialize<admin_admin_pb.InspectClusterRequest>;
    responseSerialize: grpc.serialize<admin_admin_pb.ClusterInfo>;
    responseDeserialize: grpc.deserialize<admin_admin_pb.ClusterInfo>;
}

export const APIService: IAPIService;

export interface IAPIServer extends grpc.UntypedServiceImplementation {
    inspectCluster: grpc.handleUnaryCall<admin_admin_pb.InspectClusterRequest, admin_admin_pb.ClusterInfo>;
}

export interface IAPIClient {
    inspectCluster(request: admin_admin_pb.InspectClusterRequest, callback: (error: grpc.ServiceError | null, response: admin_admin_pb.ClusterInfo) => void): grpc.ClientUnaryCall;
    inspectCluster(request: admin_admin_pb.InspectClusterRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: admin_admin_pb.ClusterInfo) => void): grpc.ClientUnaryCall;
    inspectCluster(request: admin_admin_pb.InspectClusterRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: admin_admin_pb.ClusterInfo) => void): grpc.ClientUnaryCall;
}

export class APIClient extends grpc.Client implements IAPIClient {
    constructor(address: string, credentials: grpc.ChannelCredentials, options?: Partial<grpc.ClientOptions>);
    public inspectCluster(request: admin_admin_pb.InspectClusterRequest, callback: (error: grpc.ServiceError | null, response: admin_admin_pb.ClusterInfo) => void): grpc.ClientUnaryCall;
    public inspectCluster(request: admin_admin_pb.InspectClusterRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: admin_admin_pb.ClusterInfo) => void): grpc.ClientUnaryCall;
    public inspectCluster(request: admin_admin_pb.InspectClusterRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: admin_admin_pb.ClusterInfo) => void): grpc.ClientUnaryCall;
}
