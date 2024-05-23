/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

import * as GoogleProtobufDuration from "../google/protobuf/duration.pb"
import * as GoogleProtobufTimestamp from "../google/protobuf/timestamp.pb"

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
  message?: MessageRules
}

export type FieldRules = BaseFieldRules
  & OneOf<{ float: FloatRules; double: DoubleRules; int32: Int32Rules; int64: Int64Rules; uint32: UInt32Rules; uint64: UInt64Rules; sint32: SInt32Rules; sint64: SInt64Rules; fixed32: Fixed32Rules; fixed64: Fixed64Rules; sfixed32: SFixed32Rules; sfixed64: SFixed64Rules; bool: BoolRules; string: StringRules; bytes: BytesRules; enum: EnumRules; repeated: RepeatedRules; map: MapRules; any: AnyRules; duration: DurationRules; timestamp: TimestampRules }>

export type FloatRules = {
  const?: number
  lt?: number
  lte?: number
  gt?: number
  gte?: number
  in?: number[]
  notIn?: number[]
  ignoreEmpty?: boolean
}

export type DoubleRules = {
  const?: number
  lt?: number
  lte?: number
  gt?: number
  gte?: number
  in?: number[]
  notIn?: number[]
  ignoreEmpty?: boolean
}

export type Int32Rules = {
  const?: number
  lt?: number
  lte?: number
  gt?: number
  gte?: number
  in?: number[]
  notIn?: number[]
  ignoreEmpty?: boolean
}

export type Int64Rules = {
  const?: string
  lt?: string
  lte?: string
  gt?: string
  gte?: string
  in?: string[]
  notIn?: string[]
  ignoreEmpty?: boolean
}

export type UInt32Rules = {
  const?: number
  lt?: number
  lte?: number
  gt?: number
  gte?: number
  in?: number[]
  notIn?: number[]
  ignoreEmpty?: boolean
}

export type UInt64Rules = {
  const?: string
  lt?: string
  lte?: string
  gt?: string
  gte?: string
  in?: string[]
  notIn?: string[]
  ignoreEmpty?: boolean
}

export type SInt32Rules = {
  const?: number
  lt?: number
  lte?: number
  gt?: number
  gte?: number
  in?: number[]
  notIn?: number[]
  ignoreEmpty?: boolean
}

export type SInt64Rules = {
  const?: string
  lt?: string
  lte?: string
  gt?: string
  gte?: string
  in?: string[]
  notIn?: string[]
  ignoreEmpty?: boolean
}

export type Fixed32Rules = {
  const?: number
  lt?: number
  lte?: number
  gt?: number
  gte?: number
  in?: number[]
  notIn?: number[]
  ignoreEmpty?: boolean
}

export type Fixed64Rules = {
  const?: string
  lt?: string
  lte?: string
  gt?: string
  gte?: string
  in?: string[]
  notIn?: string[]
  ignoreEmpty?: boolean
}

export type SFixed32Rules = {
  const?: number
  lt?: number
  lte?: number
  gt?: number
  gte?: number
  in?: number[]
  notIn?: number[]
  ignoreEmpty?: boolean
}

export type SFixed64Rules = {
  const?: string
  lt?: string
  lte?: string
  gt?: string
  gte?: string
  in?: string[]
  notIn?: string[]
  ignoreEmpty?: boolean
}

export type BoolRules = {
  const?: boolean
}


type BaseStringRules = {
  const?: string
  len?: string
  minLen?: string
  maxLen?: string
  lenBytes?: string
  minBytes?: string
  maxBytes?: string
  pattern?: string
  prefix?: string
  suffix?: string
  contains?: string
  notContains?: string
  in?: string[]
  notIn?: string[]
  strict?: boolean
  ignoreEmpty?: boolean
}

export type StringRules = BaseStringRules
  & OneOf<{ email: boolean; hostname: boolean; ip: boolean; ipv4: boolean; ipv6: boolean; uri: boolean; uriRef: boolean; address: boolean; uuid: boolean; wellKnownRegex: KnownRegex }>


type BaseBytesRules = {
  const?: Uint8Array
  len?: string
  minLen?: string
  maxLen?: string
  pattern?: string
  prefix?: Uint8Array
  suffix?: Uint8Array
  contains?: Uint8Array
  in?: Uint8Array[]
  notIn?: Uint8Array[]
  ignoreEmpty?: boolean
}

export type BytesRules = BaseBytesRules
  & OneOf<{ ip: boolean; ipv4: boolean; ipv6: boolean }>

export type EnumRules = {
  const?: number
  definedOnly?: boolean
  in?: number[]
  notIn?: number[]
}

export type MessageRules = {
  skip?: boolean
  required?: boolean
}

export type RepeatedRules = {
  minItems?: string
  maxItems?: string
  unique?: boolean
  items?: FieldRules
  ignoreEmpty?: boolean
}

export type MapRules = {
  minPairs?: string
  maxPairs?: string
  noSparse?: boolean
  keys?: FieldRules
  values?: FieldRules
  ignoreEmpty?: boolean
}

export type AnyRules = {
  required?: boolean
  in?: string[]
  notIn?: string[]
}

export type DurationRules = {
  required?: boolean
  const?: GoogleProtobufDuration.Duration
  lt?: GoogleProtobufDuration.Duration
  lte?: GoogleProtobufDuration.Duration
  gt?: GoogleProtobufDuration.Duration
  gte?: GoogleProtobufDuration.Duration
  in?: GoogleProtobufDuration.Duration[]
  notIn?: GoogleProtobufDuration.Duration[]
}

export type TimestampRules = {
  required?: boolean
  const?: GoogleProtobufTimestamp.Timestamp
  lt?: GoogleProtobufTimestamp.Timestamp
  lte?: GoogleProtobufTimestamp.Timestamp
  gt?: GoogleProtobufTimestamp.Timestamp
  gte?: GoogleProtobufTimestamp.Timestamp
  ltNow?: boolean
  gtNow?: boolean
  within?: GoogleProtobufDuration.Duration
}