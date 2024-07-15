/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

import * as GoogleProtobufDuration from "../google/protobuf/duration.pb";
import * as GoogleProtobufTimestamp from "../google/protobuf/timestamp.pb";

type Absent<T, K extends keyof T> = { [k in Exclude<keyof T, K>]?: undefined };
type OneOf<T> =
    | { [k in keyof T]?: undefined }
    | (
        keyof T extends infer K ?
        (K extends string & keyof T ? { [k in K]: T[K] } & Absent<T, K>
            : never)
        : never);

export enum KnownRegex {
    UNKNOWN = "UNKNOWN",
    HTTP_HEADER_NAME = "HTTP_HEADER_NAME",
    HTTP_HEADER_VALUE = "HTTP_HEADER_VALUE",
}


type BaseFieldRules = {
    __typename?: "BaseFieldRules";
    message?: MessageRules;
};

export type FieldRules = BaseFieldRules
    & OneOf<{ float: FloatRules; double: DoubleRules; int32: Int32Rules; int64: Int64Rules; uint32: UInt32Rules; uint64: UInt64Rules; sint32: SInt32Rules; sint64: SInt64Rules; fixed32: Fixed32Rules; fixed64: Fixed64Rules; sfixed32: SFixed32Rules; sfixed64: SFixed64Rules; bool: BoolRules; string: StringRules; bytes: BytesRules; enum: EnumRules; repeated: RepeatedRules; map: MapRules; any: AnyRules; duration: DurationRules; timestamp: TimestampRules; }>;

export type FloatRules = {
    __typename?: "FloatRules";
    const?: number;
    lt?: number;
    lte?: number;
    gt?: number;
    gte?: number;
    in?: number[];
    notIn?: number[];
    ignoreEmpty?: boolean;
};

export type DoubleRules = {
    __typename?: "DoubleRules";
    const?: number;
    lt?: number;
    lte?: number;
    gt?: number;
    gte?: number;
    in?: number[];
    notIn?: number[];
    ignoreEmpty?: boolean;
};

export type Int32Rules = {
    __typename?: "Int32Rules";
    const?: number;
    lt?: number;
    lte?: number;
    gt?: number;
    gte?: number;
    in?: number[];
    notIn?: number[];
    ignoreEmpty?: boolean;
};

export type Int64Rules = {
    __typename?: "Int64Rules";
    const?: string;
    lt?: string;
    lte?: string;
    gt?: string;
    gte?: string;
    in?: string[];
    notIn?: string[];
    ignoreEmpty?: boolean;
};

export type UInt32Rules = {
    __typename?: "UInt32Rules";
    const?: number;
    lt?: number;
    lte?: number;
    gt?: number;
    gte?: number;
    in?: number[];
    notIn?: number[];
    ignoreEmpty?: boolean;
};

export type UInt64Rules = {
    __typename?: "UInt64Rules";
    const?: string;
    lt?: string;
    lte?: string;
    gt?: string;
    gte?: string;
    in?: string[];
    notIn?: string[];
    ignoreEmpty?: boolean;
};

export type SInt32Rules = {
    __typename?: "SInt32Rules";
    const?: number;
    lt?: number;
    lte?: number;
    gt?: number;
    gte?: number;
    in?: number[];
    notIn?: number[];
    ignoreEmpty?: boolean;
};

export type SInt64Rules = {
    __typename?: "SInt64Rules";
    const?: string;
    lt?: string;
    lte?: string;
    gt?: string;
    gte?: string;
    in?: string[];
    notIn?: string[];
    ignoreEmpty?: boolean;
};

export type Fixed32Rules = {
    __typename?: "Fixed32Rules";
    const?: number;
    lt?: number;
    lte?: number;
    gt?: number;
    gte?: number;
    in?: number[];
    notIn?: number[];
    ignoreEmpty?: boolean;
};

export type Fixed64Rules = {
    __typename?: "Fixed64Rules";
    const?: string;
    lt?: string;
    lte?: string;
    gt?: string;
    gte?: string;
    in?: string[];
    notIn?: string[];
    ignoreEmpty?: boolean;
};

export type SFixed32Rules = {
    __typename?: "SFixed32Rules";
    const?: number;
    lt?: number;
    lte?: number;
    gt?: number;
    gte?: number;
    in?: number[];
    notIn?: number[];
    ignoreEmpty?: boolean;
};

export type SFixed64Rules = {
    __typename?: "SFixed64Rules";
    const?: string;
    lt?: string;
    lte?: string;
    gt?: string;
    gte?: string;
    in?: string[];
    notIn?: string[];
    ignoreEmpty?: boolean;
};

export type BoolRules = {
    __typename?: "BoolRules";
    const?: boolean;
};


type BaseStringRules = {
    __typename?: "BaseStringRules";
    const?: string;
    len?: string;
    minLen?: string;
    maxLen?: string;
    lenBytes?: string;
    minBytes?: string;
    maxBytes?: string;
    pattern?: string;
    prefix?: string;
    suffix?: string;
    contains?: string;
    notContains?: string;
    in?: string[];
    notIn?: string[];
    strict?: boolean;
    ignoreEmpty?: boolean;
};

export type StringRules = BaseStringRules
    & OneOf<{ email: boolean; hostname: boolean; ip: boolean; ipv4: boolean; ipv6: boolean; uri: boolean; uriRef: boolean; address: boolean; uuid: boolean; wellKnownRegex: KnownRegex; }>;


type BaseBytesRules = {
    __typename?: "BaseBytesRules";
    const?: Uint8Array;
    len?: string;
    minLen?: string;
    maxLen?: string;
    pattern?: string;
    prefix?: Uint8Array;
    suffix?: Uint8Array;
    contains?: Uint8Array;
    in?: Uint8Array[];
    notIn?: Uint8Array[];
    ignoreEmpty?: boolean;
};

export type BytesRules = BaseBytesRules
    & OneOf<{ ip: boolean; ipv4: boolean; ipv6: boolean; }>;

export type EnumRules = {
    __typename?: "EnumRules";
    const?: number;
    definedOnly?: boolean;
    in?: number[];
    notIn?: number[];
};

export type MessageRules = {
    __typename?: "MessageRules";
    skip?: boolean;
    required?: boolean;
};

export type RepeatedRules = {
    __typename?: "RepeatedRules";
    minItems?: string;
    maxItems?: string;
    unique?: boolean;
    items?: FieldRules;
    ignoreEmpty?: boolean;
};

export type MapRules = {
    __typename?: "MapRules";
    minPairs?: string;
    maxPairs?: string;
    noSparse?: boolean;
    keys?: FieldRules;
    values?: FieldRules;
    ignoreEmpty?: boolean;
};

export type AnyRules = {
    __typename?: "AnyRules";
    required?: boolean;
    in?: string[];
    notIn?: string[];
};

export type DurationRules = {
    __typename?: "DurationRules";
    required?: boolean;
    const?: GoogleProtobufDuration.Duration;
    lt?: GoogleProtobufDuration.Duration;
    lte?: GoogleProtobufDuration.Duration;
    gt?: GoogleProtobufDuration.Duration;
    gte?: GoogleProtobufDuration.Duration;
    in?: GoogleProtobufDuration.Duration[];
    notIn?: GoogleProtobufDuration.Duration[];
};

export type TimestampRules = {
    __typename?: "TimestampRules";
    required?: boolean;
    const?: GoogleProtobufTimestamp.Timestamp;
    lt?: GoogleProtobufTimestamp.Timestamp;
    lte?: GoogleProtobufTimestamp.Timestamp;
    gt?: GoogleProtobufTimestamp.Timestamp;
    gte?: GoogleProtobufTimestamp.Timestamp;
    ltNow?: boolean;
    gtNow?: boolean;
    within?: GoogleProtobufDuration.Duration;
};
