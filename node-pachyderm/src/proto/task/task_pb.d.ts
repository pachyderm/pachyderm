// package: taskapi
// file: task/task.proto

/* tslint:disable */
/* eslint-disable */

import * as jspb from "google-protobuf";
import * as gogoproto_gogo_pb from "../gogoproto/gogo_pb";

export class Group extends jspb.Message { 
    getNamespace(): string;
    setNamespace(value: string): Group;
    getGroup(): string;
    setGroup(value: string): Group;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Group.AsObject;
    static toObject(includeInstance: boolean, msg: Group): Group.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: Group, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Group;
    static deserializeBinaryFromReader(message: Group, reader: jspb.BinaryReader): Group;
}

export namespace Group {
    export type AsObject = {
        namespace: string,
        group: string,
    }
}

export class TaskInfo extends jspb.Message { 
    getId(): string;
    setId(value: string): TaskInfo;

    hasGroup(): boolean;
    clearGroup(): void;
    getGroup(): Group | undefined;
    setGroup(value?: Group): TaskInfo;
    getState(): State;
    setState(value: State): TaskInfo;
    getReason(): string;
    setReason(value: string): TaskInfo;
    getInputType(): string;
    setInputType(value: string): TaskInfo;
    getInputData(): string;
    setInputData(value: string): TaskInfo;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): TaskInfo.AsObject;
    static toObject(includeInstance: boolean, msg: TaskInfo): TaskInfo.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: TaskInfo, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): TaskInfo;
    static deserializeBinaryFromReader(message: TaskInfo, reader: jspb.BinaryReader): TaskInfo;
}

export namespace TaskInfo {
    export type AsObject = {
        id: string,
        group?: Group.AsObject,
        state: State,
        reason: string,
        inputType: string,
        inputData: string,
    }
}

export class ListTaskRequest extends jspb.Message { 

    hasGroup(): boolean;
    clearGroup(): void;
    getGroup(): Group | undefined;
    setGroup(value?: Group): ListTaskRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ListTaskRequest.AsObject;
    static toObject(includeInstance: boolean, msg: ListTaskRequest): ListTaskRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ListTaskRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ListTaskRequest;
    static deserializeBinaryFromReader(message: ListTaskRequest, reader: jspb.BinaryReader): ListTaskRequest;
}

export namespace ListTaskRequest {
    export type AsObject = {
        group?: Group.AsObject,
    }
}

export enum State {
    UNKNOWN = 0,
    RUNNING = 1,
    SUCCESS = 2,
    FAILURE = 3,
    CLAIMED = 4,
}
