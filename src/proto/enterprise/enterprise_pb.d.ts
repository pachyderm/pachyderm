// package: enterprise_v2
// file: enterprise/enterprise.proto

/* tslint:disable */
/* eslint-disable */

import * as jspb from "google-protobuf";
import * as google_protobuf_timestamp_pb from "google-protobuf/google/protobuf/timestamp_pb";

export class LicenseRecord extends jspb.Message { 
    getActivationCode(): string;
    setActivationCode(value: string): LicenseRecord;

    hasExpires(): boolean;
    clearExpires(): void;
    getExpires(): google_protobuf_timestamp_pb.Timestamp | undefined;
    setExpires(value?: google_protobuf_timestamp_pb.Timestamp): LicenseRecord;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): LicenseRecord.AsObject;
    static toObject(includeInstance: boolean, msg: LicenseRecord): LicenseRecord.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: LicenseRecord, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): LicenseRecord;
    static deserializeBinaryFromReader(message: LicenseRecord, reader: jspb.BinaryReader): LicenseRecord;
}

export namespace LicenseRecord {
    export type AsObject = {
        activationCode: string,
        expires?: google_protobuf_timestamp_pb.Timestamp.AsObject,
    }
}

export class EnterpriseConfig extends jspb.Message { 
    getLicenseServer(): string;
    setLicenseServer(value: string): EnterpriseConfig;
    getId(): string;
    setId(value: string): EnterpriseConfig;
    getSecret(): string;
    setSecret(value: string): EnterpriseConfig;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): EnterpriseConfig.AsObject;
    static toObject(includeInstance: boolean, msg: EnterpriseConfig): EnterpriseConfig.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: EnterpriseConfig, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): EnterpriseConfig;
    static deserializeBinaryFromReader(message: EnterpriseConfig, reader: jspb.BinaryReader): EnterpriseConfig;
}

export namespace EnterpriseConfig {
    export type AsObject = {
        licenseServer: string,
        id: string,
        secret: string,
    }
}

export class EnterpriseRecord extends jspb.Message { 

    hasLicense(): boolean;
    clearLicense(): void;
    getLicense(): LicenseRecord | undefined;
    setLicense(value?: LicenseRecord): EnterpriseRecord;

    hasLastHeartbeat(): boolean;
    clearLastHeartbeat(): void;
    getLastHeartbeat(): google_protobuf_timestamp_pb.Timestamp | undefined;
    setLastHeartbeat(value?: google_protobuf_timestamp_pb.Timestamp): EnterpriseRecord;
    getHeartbeatFailed(): boolean;
    setHeartbeatFailed(value: boolean): EnterpriseRecord;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): EnterpriseRecord.AsObject;
    static toObject(includeInstance: boolean, msg: EnterpriseRecord): EnterpriseRecord.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: EnterpriseRecord, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): EnterpriseRecord;
    static deserializeBinaryFromReader(message: EnterpriseRecord, reader: jspb.BinaryReader): EnterpriseRecord;
}

export namespace EnterpriseRecord {
    export type AsObject = {
        license?: LicenseRecord.AsObject,
        lastHeartbeat?: google_protobuf_timestamp_pb.Timestamp.AsObject,
        heartbeatFailed: boolean,
    }
}

export class TokenInfo extends jspb.Message { 

    hasExpires(): boolean;
    clearExpires(): void;
    getExpires(): google_protobuf_timestamp_pb.Timestamp | undefined;
    setExpires(value?: google_protobuf_timestamp_pb.Timestamp): TokenInfo;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): TokenInfo.AsObject;
    static toObject(includeInstance: boolean, msg: TokenInfo): TokenInfo.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: TokenInfo, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): TokenInfo;
    static deserializeBinaryFromReader(message: TokenInfo, reader: jspb.BinaryReader): TokenInfo;
}

export namespace TokenInfo {
    export type AsObject = {
        expires?: google_protobuf_timestamp_pb.Timestamp.AsObject,
    }
}

export class ActivateRequest extends jspb.Message { 
    getLicenseServer(): string;
    setLicenseServer(value: string): ActivateRequest;
    getId(): string;
    setId(value: string): ActivateRequest;
    getSecret(): string;
    setSecret(value: string): ActivateRequest;

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
        licenseServer: string,
        id: string,
        secret: string,
    }
}

export class ActivateResponse extends jspb.Message { 

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
    }
}

export class GetStateRequest extends jspb.Message { 

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GetStateRequest.AsObject;
    static toObject(includeInstance: boolean, msg: GetStateRequest): GetStateRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GetStateRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GetStateRequest;
    static deserializeBinaryFromReader(message: GetStateRequest, reader: jspb.BinaryReader): GetStateRequest;
}

export namespace GetStateRequest {
    export type AsObject = {
    }
}

export class GetStateResponse extends jspb.Message { 
    getState(): State;
    setState(value: State): GetStateResponse;

    hasInfo(): boolean;
    clearInfo(): void;
    getInfo(): TokenInfo | undefined;
    setInfo(value?: TokenInfo): GetStateResponse;
    getActivationCode(): string;
    setActivationCode(value: string): GetStateResponse;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GetStateResponse.AsObject;
    static toObject(includeInstance: boolean, msg: GetStateResponse): GetStateResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GetStateResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GetStateResponse;
    static deserializeBinaryFromReader(message: GetStateResponse, reader: jspb.BinaryReader): GetStateResponse;
}

export namespace GetStateResponse {
    export type AsObject = {
        state: State,
        info?: TokenInfo.AsObject,
        activationCode: string,
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
    getState(): State;
    setState(value: State): GetActivationCodeResponse;

    hasInfo(): boolean;
    clearInfo(): void;
    getInfo(): TokenInfo | undefined;
    setInfo(value?: TokenInfo): GetActivationCodeResponse;
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
        state: State,
        info?: TokenInfo.AsObject,
        activationCode: string,
    }
}

export class HeartbeatRequest extends jspb.Message { 

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
    }
}

export class HeartbeatResponse extends jspb.Message { 

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

export class PauseRequest extends jspb.Message { 

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): PauseRequest.AsObject;
    static toObject(includeInstance: boolean, msg: PauseRequest): PauseRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: PauseRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): PauseRequest;
    static deserializeBinaryFromReader(message: PauseRequest, reader: jspb.BinaryReader): PauseRequest;
}

export namespace PauseRequest {
    export type AsObject = {
    }
}

export class PauseResponse extends jspb.Message { 

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): PauseResponse.AsObject;
    static toObject(includeInstance: boolean, msg: PauseResponse): PauseResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: PauseResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): PauseResponse;
    static deserializeBinaryFromReader(message: PauseResponse, reader: jspb.BinaryReader): PauseResponse;
}

export namespace PauseResponse {
    export type AsObject = {
    }
}

export class UnpauseRequest extends jspb.Message { 

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): UnpauseRequest.AsObject;
    static toObject(includeInstance: boolean, msg: UnpauseRequest): UnpauseRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: UnpauseRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): UnpauseRequest;
    static deserializeBinaryFromReader(message: UnpauseRequest, reader: jspb.BinaryReader): UnpauseRequest;
}

export namespace UnpauseRequest {
    export type AsObject = {
    }
}

export class UnpauseResponse extends jspb.Message { 

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): UnpauseResponse.AsObject;
    static toObject(includeInstance: boolean, msg: UnpauseResponse): UnpauseResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: UnpauseResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): UnpauseResponse;
    static deserializeBinaryFromReader(message: UnpauseResponse, reader: jspb.BinaryReader): UnpauseResponse;
}

export namespace UnpauseResponse {
    export type AsObject = {
    }
}

export class PauseStatusRequest extends jspb.Message { 

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): PauseStatusRequest.AsObject;
    static toObject(includeInstance: boolean, msg: PauseStatusRequest): PauseStatusRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: PauseStatusRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): PauseStatusRequest;
    static deserializeBinaryFromReader(message: PauseStatusRequest, reader: jspb.BinaryReader): PauseStatusRequest;
}

export namespace PauseStatusRequest {
    export type AsObject = {
    }
}

export class PauseStatusResponse extends jspb.Message { 
    getStatus(): PauseStatusResponse.PauseStatus;
    setStatus(value: PauseStatusResponse.PauseStatus): PauseStatusResponse;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): PauseStatusResponse.AsObject;
    static toObject(includeInstance: boolean, msg: PauseStatusResponse): PauseStatusResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: PauseStatusResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): PauseStatusResponse;
    static deserializeBinaryFromReader(message: PauseStatusResponse, reader: jspb.BinaryReader): PauseStatusResponse;
}

export namespace PauseStatusResponse {
    export type AsObject = {
        status: PauseStatusResponse.PauseStatus,
    }

    export enum PauseStatus {
    UNPAUSED = 0,
    PARTIALLY_PAUSED = 1,
    PAUSED = 2,
    }

}

export enum State {
    NONE = 0,
    ACTIVE = 1,
    EXPIRED = 2,
    HEARTBEAT_FAILED = 3,
}
