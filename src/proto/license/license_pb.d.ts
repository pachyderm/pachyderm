// package: license_v2
// file: license/license.proto

/* tslint:disable */
/* eslint-disable */

import * as jspb from "google-protobuf";
import * as google_protobuf_timestamp_pb from "google-protobuf/google/protobuf/timestamp_pb";
import * as gogoproto_gogo_pb from "../gogoproto/gogo_pb";
import * as enterprise_enterprise_pb from "../enterprise/enterprise_pb";

export class ActivateRequest extends jspb.Message { 
    getActivationCode(): string;
    setActivationCode(value: string): ActivateRequest;

    hasExpires(): boolean;
    clearExpires(): void;
    getExpires(): google_protobuf_timestamp_pb.Timestamp | undefined;
    setExpires(value?: google_protobuf_timestamp_pb.Timestamp): ActivateRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ActivateRequest.AsObject;
    static toObject(includeInstance: boolean, msg: ActivateRequest): ActivateRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ActivateRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ActivateRequest;
    static deserializeBinaryFromReader(message: ActivateRequest, reader: jspb.BinaryReader): ActivateRequest;
}

export namespace ActivateRequest {
    export type AsObject = {
        activationCode: string,
        expires?: google_protobuf_timestamp_pb.Timestamp.AsObject,
    }
}

export class ActivateResponse extends jspb.Message { 

    hasInfo(): boolean;
    clearInfo(): void;
    getInfo(): enterprise_enterprise_pb.TokenInfo | undefined;
    setInfo(value?: enterprise_enterprise_pb.TokenInfo): ActivateResponse;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ActivateResponse.AsObject;
    static toObject(includeInstance: boolean, msg: ActivateResponse): ActivateResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ActivateResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ActivateResponse;
    static deserializeBinaryFromReader(message: ActivateResponse, reader: jspb.BinaryReader): ActivateResponse;
}

export namespace ActivateResponse {
    export type AsObject = {
        info?: enterprise_enterprise_pb.TokenInfo.AsObject,
    }
}

export class GetActivationCodeRequest extends jspb.Message { 

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GetActivationCodeRequest.AsObject;
    static toObject(includeInstance: boolean, msg: GetActivationCodeRequest): GetActivationCodeRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GetActivationCodeRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GetActivationCodeRequest;
    static deserializeBinaryFromReader(message: GetActivationCodeRequest, reader: jspb.BinaryReader): GetActivationCodeRequest;
}

export namespace GetActivationCodeRequest {
    export type AsObject = {
    }
}

export class GetActivationCodeResponse extends jspb.Message { 
    getState(): enterprise_enterprise_pb.State;
    setState(value: enterprise_enterprise_pb.State): GetActivationCodeResponse;

    hasInfo(): boolean;
    clearInfo(): void;
    getInfo(): enterprise_enterprise_pb.TokenInfo | undefined;
    setInfo(value?: enterprise_enterprise_pb.TokenInfo): GetActivationCodeResponse;
    getActivationCode(): string;
    setActivationCode(value: string): GetActivationCodeResponse;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GetActivationCodeResponse.AsObject;
    static toObject(includeInstance: boolean, msg: GetActivationCodeResponse): GetActivationCodeResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GetActivationCodeResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GetActivationCodeResponse;
    static deserializeBinaryFromReader(message: GetActivationCodeResponse, reader: jspb.BinaryReader): GetActivationCodeResponse;
}

export namespace GetActivationCodeResponse {
    export type AsObject = {
        state: enterprise_enterprise_pb.State,
        info?: enterprise_enterprise_pb.TokenInfo.AsObject,
        activationCode: string,
    }
}

export class DeactivateRequest extends jspb.Message { 

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): DeactivateRequest.AsObject;
    static toObject(includeInstance: boolean, msg: DeactivateRequest): DeactivateRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: DeactivateRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): DeactivateRequest;
    static deserializeBinaryFromReader(message: DeactivateRequest, reader: jspb.BinaryReader): DeactivateRequest;
}

export namespace DeactivateRequest {
    export type AsObject = {
    }
}

export class DeactivateResponse extends jspb.Message { 

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): DeactivateResponse.AsObject;
    static toObject(includeInstance: boolean, msg: DeactivateResponse): DeactivateResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: DeactivateResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): DeactivateResponse;
    static deserializeBinaryFromReader(message: DeactivateResponse, reader: jspb.BinaryReader): DeactivateResponse;
}

export namespace DeactivateResponse {
    export type AsObject = {
    }
}

export class AddClusterRequest extends jspb.Message { 
    getId(): string;
    setId(value: string): AddClusterRequest;
    getAddress(): string;
    setAddress(value: string): AddClusterRequest;
    getSecret(): string;
    setSecret(value: string): AddClusterRequest;
    getUserAddress(): string;
    setUserAddress(value: string): AddClusterRequest;
    getClusterDeploymentId(): string;
    setClusterDeploymentId(value: string): AddClusterRequest;
    getEnterpriseServer(): boolean;
    setEnterpriseServer(value: boolean): AddClusterRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): AddClusterRequest.AsObject;
    static toObject(includeInstance: boolean, msg: AddClusterRequest): AddClusterRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: AddClusterRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): AddClusterRequest;
    static deserializeBinaryFromReader(message: AddClusterRequest, reader: jspb.BinaryReader): AddClusterRequest;
}

export namespace AddClusterRequest {
    export type AsObject = {
        id: string,
        address: string,
        secret: string,
        userAddress: string,
        clusterDeploymentId: string,
        enterpriseServer: boolean,
    }
}

export class AddClusterResponse extends jspb.Message { 
    getSecret(): string;
    setSecret(value: string): AddClusterResponse;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): AddClusterResponse.AsObject;
    static toObject(includeInstance: boolean, msg: AddClusterResponse): AddClusterResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: AddClusterResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): AddClusterResponse;
    static deserializeBinaryFromReader(message: AddClusterResponse, reader: jspb.BinaryReader): AddClusterResponse;
}

export namespace AddClusterResponse {
    export type AsObject = {
        secret: string,
    }
}

export class DeleteClusterRequest extends jspb.Message { 
    getId(): string;
    setId(value: string): DeleteClusterRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): DeleteClusterRequest.AsObject;
    static toObject(includeInstance: boolean, msg: DeleteClusterRequest): DeleteClusterRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: DeleteClusterRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): DeleteClusterRequest;
    static deserializeBinaryFromReader(message: DeleteClusterRequest, reader: jspb.BinaryReader): DeleteClusterRequest;
}

export namespace DeleteClusterRequest {
    export type AsObject = {
        id: string,
    }
}

export class DeleteClusterResponse extends jspb.Message { 

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): DeleteClusterResponse.AsObject;
    static toObject(includeInstance: boolean, msg: DeleteClusterResponse): DeleteClusterResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: DeleteClusterResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): DeleteClusterResponse;
    static deserializeBinaryFromReader(message: DeleteClusterResponse, reader: jspb.BinaryReader): DeleteClusterResponse;
}

export namespace DeleteClusterResponse {
    export type AsObject = {
    }
}

export class ClusterStatus extends jspb.Message { 
    getId(): string;
    setId(value: string): ClusterStatus;
    getAddress(): string;
    setAddress(value: string): ClusterStatus;
    getVersion(): string;
    setVersion(value: string): ClusterStatus;
    getAuthEnabled(): boolean;
    setAuthEnabled(value: boolean): ClusterStatus;
    getClientId(): string;
    setClientId(value: string): ClusterStatus;

    hasLastHeartbeat(): boolean;
    clearLastHeartbeat(): void;
    getLastHeartbeat(): google_protobuf_timestamp_pb.Timestamp | undefined;
    setLastHeartbeat(value?: google_protobuf_timestamp_pb.Timestamp): ClusterStatus;

    hasCreatedAt(): boolean;
    clearCreatedAt(): void;
    getCreatedAt(): google_protobuf_timestamp_pb.Timestamp | undefined;
    setCreatedAt(value?: google_protobuf_timestamp_pb.Timestamp): ClusterStatus;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ClusterStatus.AsObject;
    static toObject(includeInstance: boolean, msg: ClusterStatus): ClusterStatus.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ClusterStatus, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ClusterStatus;
    static deserializeBinaryFromReader(message: ClusterStatus, reader: jspb.BinaryReader): ClusterStatus;
}

export namespace ClusterStatus {
    export type AsObject = {
        id: string,
        address: string,
        version: string,
        authEnabled: boolean,
        clientId: string,
        lastHeartbeat?: google_protobuf_timestamp_pb.Timestamp.AsObject,
        createdAt?: google_protobuf_timestamp_pb.Timestamp.AsObject,
    }
}

export class UpdateClusterRequest extends jspb.Message { 
    getId(): string;
    setId(value: string): UpdateClusterRequest;
    getAddress(): string;
    setAddress(value: string): UpdateClusterRequest;
    getUserAddress(): string;
    setUserAddress(value: string): UpdateClusterRequest;
    getClusterDeploymentId(): string;
    setClusterDeploymentId(value: string): UpdateClusterRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): UpdateClusterRequest.AsObject;
    static toObject(includeInstance: boolean, msg: UpdateClusterRequest): UpdateClusterRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: UpdateClusterRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): UpdateClusterRequest;
    static deserializeBinaryFromReader(message: UpdateClusterRequest, reader: jspb.BinaryReader): UpdateClusterRequest;
}

export namespace UpdateClusterRequest {
    export type AsObject = {
        id: string,
        address: string,
        userAddress: string,
        clusterDeploymentId: string,
    }
}

export class UpdateClusterResponse extends jspb.Message { 

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): UpdateClusterResponse.AsObject;
    static toObject(includeInstance: boolean, msg: UpdateClusterResponse): UpdateClusterResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: UpdateClusterResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): UpdateClusterResponse;
    static deserializeBinaryFromReader(message: UpdateClusterResponse, reader: jspb.BinaryReader): UpdateClusterResponse;
}

export namespace UpdateClusterResponse {
    export type AsObject = {
    }
}

export class ListClustersRequest extends jspb.Message { 

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ListClustersRequest.AsObject;
    static toObject(includeInstance: boolean, msg: ListClustersRequest): ListClustersRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ListClustersRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ListClustersRequest;
    static deserializeBinaryFromReader(message: ListClustersRequest, reader: jspb.BinaryReader): ListClustersRequest;
}

export namespace ListClustersRequest {
    export type AsObject = {
    }
}

export class ListClustersResponse extends jspb.Message { 
    clearClustersList(): void;
    getClustersList(): Array<ClusterStatus>;
    setClustersList(value: Array<ClusterStatus>): ListClustersResponse;
    addClusters(value?: ClusterStatus, index?: number): ClusterStatus;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ListClustersResponse.AsObject;
    static toObject(includeInstance: boolean, msg: ListClustersResponse): ListClustersResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ListClustersResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ListClustersResponse;
    static deserializeBinaryFromReader(message: ListClustersResponse, reader: jspb.BinaryReader): ListClustersResponse;
}

export namespace ListClustersResponse {
    export type AsObject = {
        clustersList: Array<ClusterStatus.AsObject>,
    }
}

export class DeleteAllRequest extends jspb.Message { 

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): DeleteAllRequest.AsObject;
    static toObject(includeInstance: boolean, msg: DeleteAllRequest): DeleteAllRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: DeleteAllRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): DeleteAllRequest;
    static deserializeBinaryFromReader(message: DeleteAllRequest, reader: jspb.BinaryReader): DeleteAllRequest;
}

export namespace DeleteAllRequest {
    export type AsObject = {
    }
}

export class DeleteAllResponse extends jspb.Message { 

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): DeleteAllResponse.AsObject;
    static toObject(includeInstance: boolean, msg: DeleteAllResponse): DeleteAllResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: DeleteAllResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): DeleteAllResponse;
    static deserializeBinaryFromReader(message: DeleteAllResponse, reader: jspb.BinaryReader): DeleteAllResponse;
}

export namespace DeleteAllResponse {
    export type AsObject = {
    }
}

export class HeartbeatRequest extends jspb.Message { 
    getId(): string;
    setId(value: string): HeartbeatRequest;
    getSecret(): string;
    setSecret(value: string): HeartbeatRequest;
    getVersion(): string;
    setVersion(value: string): HeartbeatRequest;
    getAuthEnabled(): boolean;
    setAuthEnabled(value: boolean): HeartbeatRequest;
    getClientId(): string;
    setClientId(value: string): HeartbeatRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): HeartbeatRequest.AsObject;
    static toObject(includeInstance: boolean, msg: HeartbeatRequest): HeartbeatRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: HeartbeatRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): HeartbeatRequest;
    static deserializeBinaryFromReader(message: HeartbeatRequest, reader: jspb.BinaryReader): HeartbeatRequest;
}

export namespace HeartbeatRequest {
    export type AsObject = {
        id: string,
        secret: string,
        version: string,
        authEnabled: boolean,
        clientId: string,
    }
}

export class HeartbeatResponse extends jspb.Message { 

    hasLicense(): boolean;
    clearLicense(): void;
    getLicense(): enterprise_enterprise_pb.LicenseRecord | undefined;
    setLicense(value?: enterprise_enterprise_pb.LicenseRecord): HeartbeatResponse;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): HeartbeatResponse.AsObject;
    static toObject(includeInstance: boolean, msg: HeartbeatResponse): HeartbeatResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: HeartbeatResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): HeartbeatResponse;
    static deserializeBinaryFromReader(message: HeartbeatResponse, reader: jspb.BinaryReader): HeartbeatResponse;
}

export namespace HeartbeatResponse {
    export type AsObject = {
        license?: enterprise_enterprise_pb.LicenseRecord.AsObject,
    }
}

export class UserClusterInfo extends jspb.Message { 
    getId(): string;
    setId(value: string): UserClusterInfo;
    getClusterDeploymentId(): string;
    setClusterDeploymentId(value: string): UserClusterInfo;
    getAddress(): string;
    setAddress(value: string): UserClusterInfo;
    getEnterpriseServer(): boolean;
    setEnterpriseServer(value: boolean): UserClusterInfo;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): UserClusterInfo.AsObject;
    static toObject(includeInstance: boolean, msg: UserClusterInfo): UserClusterInfo.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: UserClusterInfo, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): UserClusterInfo;
    static deserializeBinaryFromReader(message: UserClusterInfo, reader: jspb.BinaryReader): UserClusterInfo;
}

export namespace UserClusterInfo {
    export type AsObject = {
        id: string,
        clusterDeploymentId: string,
        address: string,
        enterpriseServer: boolean,
    }
}

export class ListUserClustersRequest extends jspb.Message { 

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ListUserClustersRequest.AsObject;
    static toObject(includeInstance: boolean, msg: ListUserClustersRequest): ListUserClustersRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ListUserClustersRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ListUserClustersRequest;
    static deserializeBinaryFromReader(message: ListUserClustersRequest, reader: jspb.BinaryReader): ListUserClustersRequest;
}

export namespace ListUserClustersRequest {
    export type AsObject = {
    }
}

export class ListUserClustersResponse extends jspb.Message { 
    clearClustersList(): void;
    getClustersList(): Array<UserClusterInfo>;
    setClustersList(value: Array<UserClusterInfo>): ListUserClustersResponse;
    addClusters(value?: UserClusterInfo, index?: number): UserClusterInfo;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ListUserClustersResponse.AsObject;
    static toObject(includeInstance: boolean, msg: ListUserClustersResponse): ListUserClustersResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ListUserClustersResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ListUserClustersResponse;
    static deserializeBinaryFromReader(message: ListUserClustersResponse, reader: jspb.BinaryReader): ListUserClustersResponse;
}

export namespace ListUserClustersResponse {
    export type AsObject = {
        clustersList: Array<UserClusterInfo.AsObject>,
    }
}
