// package: admin_v2
// file: admin/admin.proto

/* tslint:disable */
/* eslint-disable */

import * as jspb from "google-protobuf";
import * as google_protobuf_empty_pb from "google-protobuf/google/protobuf/empty_pb";
import * as gogoproto_gogo_pb from "../gogoproto/gogo_pb";

export class ClusterInfo extends jspb.Message { 
    getId(): string;
    setId(value: string): ClusterInfo;
    getDeploymentId(): string;
    setDeploymentId(value: string): ClusterInfo;

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
    }
}
