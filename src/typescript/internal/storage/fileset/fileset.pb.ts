/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

import * as IndexIndex from "./index/index.pb"

type Absent<T, K extends keyof T> = { [k in Exclude<keyof T, K>]?: undefined };
type OneOf<T> =
  | { [k in keyof T]?: undefined }
  | (
    keyof T extends infer K ?
      (K extends string & keyof T ? { [k in K]: T[K] } & Absent<T, K>
        : never)
    : never);

type BaseMetadata = {
}

export type Metadata = BaseMetadata
  & OneOf<{ primitive: Primitive; composite: Composite }>

export type Composite = {
  layers?: string[]
}

export type Primitive = {
  deletive?: IndexIndex.Index
  additive?: IndexIndex.Index
  sizeBytes?: string
}

export type TestCacheValue = {
  fileSetId?: string
}