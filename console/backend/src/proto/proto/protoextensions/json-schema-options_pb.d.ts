// package: protoc.gen.jsonschema
// file: protoextensions/json-schema-options.proto

/* tslint:disable */
/* eslint-disable */

import * as jspb from "google-protobuf";
import * as google_protobuf_descriptor_pb from "google-protobuf/google/protobuf/descriptor_pb";

export class FieldOptions extends jspb.Message { 
    getIgnore(): boolean;
    setIgnore(value: boolean): FieldOptions;
    getRequired(): boolean;
    setRequired(value: boolean): FieldOptions;
    getMinLength(): number;
    setMinLength(value: number): FieldOptions;
    getMaxLength(): number;
    setMaxLength(value: number): FieldOptions;
    getPattern(): string;
    setPattern(value: string): FieldOptions;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): FieldOptions.AsObject;
    static toObject(includeInstance: boolean, msg: FieldOptions): FieldOptions.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: FieldOptions, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): FieldOptions;
    static deserializeBinaryFromReader(message: FieldOptions, reader: jspb.BinaryReader): FieldOptions;
}

export namespace FieldOptions {
    export type AsObject = {
        ignore: boolean,
        required: boolean,
        minLength: number,
        maxLength: number,
        pattern: string,
    }
}

export class FileOptions extends jspb.Message { 
    getIgnore(): boolean;
    setIgnore(value: boolean): FileOptions;
    getExtension$(): string;
    setExtension$(value: string): FileOptions;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): FileOptions.AsObject;
    static toObject(includeInstance: boolean, msg: FileOptions): FileOptions.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: FileOptions, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): FileOptions;
    static deserializeBinaryFromReader(message: FileOptions, reader: jspb.BinaryReader): FileOptions;
}

export namespace FileOptions {
    export type AsObject = {
        ignore: boolean,
        extension: string,
    }
}

export class MessageOptions extends jspb.Message { 
    getIgnore(): boolean;
    setIgnore(value: boolean): MessageOptions;
    getAllFieldsRequired(): boolean;
    setAllFieldsRequired(value: boolean): MessageOptions;
    getAllowNullValues(): boolean;
    setAllowNullValues(value: boolean): MessageOptions;
    getDisallowAdditionalProperties(): boolean;
    setDisallowAdditionalProperties(value: boolean): MessageOptions;
    getEnumsAsConstants(): boolean;
    setEnumsAsConstants(value: boolean): MessageOptions;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): MessageOptions.AsObject;
    static toObject(includeInstance: boolean, msg: MessageOptions): MessageOptions.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: MessageOptions, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): MessageOptions;
    static deserializeBinaryFromReader(message: MessageOptions, reader: jspb.BinaryReader): MessageOptions;
}

export namespace MessageOptions {
    export type AsObject = {
        ignore: boolean,
        allFieldsRequired: boolean,
        allowNullValues: boolean,
        disallowAdditionalProperties: boolean,
        enumsAsConstants: boolean,
    }
}

export class EnumOptions extends jspb.Message { 
    getEnumsAsConstants(): boolean;
    setEnumsAsConstants(value: boolean): EnumOptions;
    getEnumsAsStringsOnly(): boolean;
    setEnumsAsStringsOnly(value: boolean): EnumOptions;
    getEnumsTrimPrefix(): boolean;
    setEnumsTrimPrefix(value: boolean): EnumOptions;
    getIgnore(): boolean;
    setIgnore(value: boolean): EnumOptions;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): EnumOptions.AsObject;
    static toObject(includeInstance: boolean, msg: EnumOptions): EnumOptions.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: EnumOptions, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): EnumOptions;
    static deserializeBinaryFromReader(message: EnumOptions, reader: jspb.BinaryReader): EnumOptions;
}

export namespace EnumOptions {
    export type AsObject = {
        enumsAsConstants: boolean,
        enumsAsStringsOnly: boolean,
        enumsTrimPrefix: boolean,
        ignore: boolean,
    }
}

export const fieldOptions: jspb.ExtensionFieldInfo<FieldOptions>;

export const fileOptions: jspb.ExtensionFieldInfo<FileOptions>;

export const messageOptions: jspb.ExtensionFieldInfo<MessageOptions>;

export const enumOptions: jspb.ExtensionFieldInfo<EnumOptions>;
