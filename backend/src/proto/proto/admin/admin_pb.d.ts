// package: admin_v2
// file: admin/admin.proto

/* tslint:disable */
/* eslint-disable */

import * as jspb from "google-protobuf";
import * as gogoproto_gogo_pb from "../gogoproto/gogo_pb";
import * as version_versionpb_version_pb from "../version/versionpb/version_pb";

export class ClusterInfo extends jspb.Message { 
    getId(): string;
    setId(value: string): ClusterInfo;
    getDeploymentId(): string;
    setDeploymentId(value: string): ClusterInfo;
    getVersionWarningsOk(): boolean;
    setVersionWarningsOk(value: boolean): ClusterInfo;
    clearVersionWarningsList(): void;
    getVersionWarningsList(): Array<string>;
    setVersionWarningsList(value: Array<string>): ClusterInfo;
    addVersionWarnings(value: string, index?: number): string;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ClusterInfo.AsObject;
    static toObject(includeInstance: boolean, msg: ClusterInfo): ClusterInfo.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ClusterInfo, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ClusterInfo;
    static deserializeBinaryFromReader(message: ClusterInfo, reader: jspb.BinaryReader): ClusterInfo;
}

export namespace ClusterInfo {
    export type AsObject = {
        id: string,
        deploymentId: string,
        versionWarningsOk: boolean,
        versionWarningsList: Array<string>,
    }
}

export class InspectClusterRequest extends jspb.Message { 

    hasClientVersion(): boolean;
    clearClientVersion(): void;
    getClientVersion(): version_versionpb_version_pb.Version | undefined;
    setClientVersion(value?: version_versionpb_version_pb.Version): InspectClusterRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): InspectClusterRequest.AsObject;
    static toObject(includeInstance: boolean, msg: InspectClusterRequest): InspectClusterRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: InspectClusterRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): InspectClusterRequest;
    static deserializeBinaryFromReader(message: InspectClusterRequest, reader: jspb.BinaryReader): InspectClusterRequest;
}

export namespace InspectClusterRequest {
    export type AsObject = {
        clientVersion?: version_versionpb_version_pb.Version.AsObject,
    }
}
