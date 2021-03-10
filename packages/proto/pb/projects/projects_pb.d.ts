// package: mock
// file: projects/projects.proto

/* tslint:disable */
/* eslint-disable */

import * as jspb from "google-protobuf";
import * as google_protobuf_empty_pb from "google-protobuf/google/protobuf/empty_pb";
import * as google_protobuf_timestamp_pb from "google-protobuf/google/protobuf/timestamp_pb";

export class Project extends jspb.Message { 
    getId(): string;
    setId(value: string): Project;

    getName(): string;
    setName(value: string): Project;

    getDescription(): string;
    setDescription(value: string): Project;


    hasCreatedat(): boolean;
    clearCreatedat(): void;
    getCreatedat(): google_protobuf_timestamp_pb.Timestamp | undefined;
    setCreatedat(value?: google_protobuf_timestamp_pb.Timestamp): Project;

    getStatus(): ProjectStatus;
    setStatus(value: ProjectStatus): Project;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Project.AsObject;
    static toObject(includeInstance: boolean, msg: Project): Project.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: Project, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Project;
    static deserializeBinaryFromReader(message: Project, reader: jspb.BinaryReader): Project;
}

export namespace Project {
    export type AsObject = {
        id: string,
        name: string,
        description: string,
        createdat?: google_protobuf_timestamp_pb.Timestamp.AsObject,
        status: ProjectStatus,
    }
}

export class Projects extends jspb.Message { 
    clearProjectInfoList(): void;
    getProjectInfoList(): Array<Project>;
    setProjectInfoList(value: Array<Project>): Projects;
    addProjectInfo(value?: Project, index?: number): Project;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Projects.AsObject;
    static toObject(includeInstance: boolean, msg: Projects): Projects.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: Projects, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Projects;
    static deserializeBinaryFromReader(message: Projects, reader: jspb.BinaryReader): Projects;
}

export namespace Projects {
    export type AsObject = {
        projectInfoList: Array<Project.AsObject>,
    }
}

export class ProjectRequest extends jspb.Message { 
    getProjectid(): string;
    setProjectid(value: string): ProjectRequest;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ProjectRequest.AsObject;
    static toObject(includeInstance: boolean, msg: ProjectRequest): ProjectRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ProjectRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ProjectRequest;
    static deserializeBinaryFromReader(message: ProjectRequest, reader: jspb.BinaryReader): ProjectRequest;
}

export namespace ProjectRequest {
    export type AsObject = {
        projectid: string,
    }
}

export enum ProjectStatus {
    HEALTHY = 0,
    UNHEALTHY = 1,
}
