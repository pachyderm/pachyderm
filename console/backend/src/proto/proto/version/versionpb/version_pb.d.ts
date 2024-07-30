// package: versionpb_v2
// file: version/versionpb/version.proto

/* tslint:disable */
/* eslint-disable */

import * as jspb from "google-protobuf";
import * as google_protobuf_empty_pb from "google-protobuf/google/protobuf/empty_pb";

export class Version extends jspb.Message { 
    getMajor(): number;
    setMajor(value: number): Version;
    getMinor(): number;
    setMinor(value: number): Version;
    getMicro(): number;
    setMicro(value: number): Version;
    getAdditional(): string;
    setAdditional(value: string): Version;
    getGitCommit(): string;
    setGitCommit(value: string): Version;
    getGitTreeModified(): string;
    setGitTreeModified(value: string): Version;
    getBuildDate(): string;
    setBuildDate(value: string): Version;
    getGoVersion(): string;
    setGoVersion(value: string): Version;
    getPlatform(): string;
    setPlatform(value: string): Version;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Version.AsObject;
    static toObject(includeInstance: boolean, msg: Version): Version.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: Version, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Version;
    static deserializeBinaryFromReader(message: Version, reader: jspb.BinaryReader): Version;
}

export namespace Version {
    export type AsObject = {
        major: number,
        minor: number,
        micro: number,
        additional: string,
        gitCommit: string,
        gitTreeModified: string,
        buildDate: string,
        goVersion: string,
        platform: string,
    }
}
