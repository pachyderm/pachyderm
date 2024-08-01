// package: admin_v2
// file: admin/admin.proto

/* tslint:disable */
/* eslint-disable */

import * as jspb from "google-protobuf";
import * as version_versionpb_version_pb from "../version/versionpb/version_pb";
import * as pfs_pfs_pb from "../pfs/pfs_pb";

export class ClusterInfo extends jspb.Message { 
    getId(): string;
    setId(value: string): ClusterInfo;
    getDeploymentId(): string;
    setDeploymentId(value: string): ClusterInfo;
    getWarningsOk(): boolean;
    setWarningsOk(value: boolean): ClusterInfo;
    clearWarningsList(): void;
    getWarningsList(): Array<string>;
    setWarningsList(value: Array<string>): ClusterInfo;
    addWarnings(value: string, index?: number): string;
    getProxyHost(): string;
    setProxyHost(value: string): ClusterInfo;
    getProxyTls(): boolean;
    setProxyTls(value: boolean): ClusterInfo;
    getPaused(): boolean;
    setPaused(value: boolean): ClusterInfo;

    hasWebResources(): boolean;
    clearWebResources(): void;
    getWebResources(): WebResource | undefined;
    setWebResources(value?: WebResource): ClusterInfo;

    getMetadataMap(): jspb.Map<string, string>;
    clearMetadataMap(): void;

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
        warningsOk: boolean,
        warningsList: Array<string>,
        proxyHost: string,
        proxyTls: boolean,
        paused: boolean,
        webResources?: WebResource.AsObject,

        metadataMap: Array<[string, string]>,
    }
}

export class InspectClusterRequest extends jspb.Message { 

    hasClientVersion(): boolean;
    clearClientVersion(): void;
    getClientVersion(): version_versionpb_version_pb.Version | undefined;
    setClientVersion(value?: version_versionpb_version_pb.Version): InspectClusterRequest;

    hasCurrentProject(): boolean;
    clearCurrentProject(): void;
    getCurrentProject(): pfs_pfs_pb.Project | undefined;
    setCurrentProject(value?: pfs_pfs_pb.Project): InspectClusterRequest;

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
        currentProject?: pfs_pfs_pb.Project.AsObject,
    }
}

export class WebResource extends jspb.Message { 
    getArchiveDownloadBaseUrl(): string;
    setArchiveDownloadBaseUrl(value: string): WebResource;
    getCreatePipelineRequestJsonSchemaUrl(): string;
    setCreatePipelineRequestJsonSchemaUrl(value: string): WebResource;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): WebResource.AsObject;
    static toObject(includeInstance: boolean, msg: WebResource): WebResource.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: WebResource, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): WebResource;
    static deserializeBinaryFromReader(message: WebResource, reader: jspb.BinaryReader): WebResource;
}

export namespace WebResource {
    export type AsObject = {
        archiveDownloadBaseUrl: string,
        createPipelineRequestJsonSchemaUrl: string,
    }
}
