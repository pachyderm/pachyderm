/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

type Absent<T, K extends keyof T> = { [k in Exclude<keyof T, K>]?: undefined };
type OneOf<T> =
  | { [k in keyof T]?: undefined }
  | (
    keyof T extends infer K ?
      (K extends string & keyof T ? { [k in K]: T[K] } & Absent<T, K>
        : never)
    : never);

export enum HashAlgo {
  UNKNOWN_HASH = "UNKNOWN_HASH",
  BLAKE2b_256 = "BLAKE2b_256",
}

export enum CipherAlgo {
  UNKNOWN_CIPHER = "UNKNOWN_CIPHER",
  CHACHA20 = "CHACHA20",
}

export enum CompressAlgo {
  UNKNOWN_COMPRESS = "UNKNOWN_COMPRESS",
  GZIP = "GZIP",
}


type BaseRef = {
}

export type Ref = BaseRef
  & OneOf<{ http: HTTP; contentHash: ContentHash; sizeLimits: SizeLimits; cipher: Cipher; compress: Compress; slice: Slice; concat: Concat }>

export type HTTP = {
  url?: string
  headers?: {[key: string]: string}
}

export type ContentHash = {
  inner?: Ref
  algo?: HashAlgo
  hash?: Uint8Array
}

export type SizeLimits = {
  inner?: Ref
  min?: string
  max?: string
}

export type Cipher = {
  inner?: Ref
  algo?: CipherAlgo
  key?: Uint8Array
  nonce?: Uint8Array
}

export type Compress = {
  inner?: Ref
  algo?: CompressAlgo
}

export type Slice = {
  inner?: Ref
  start?: string
  end?: string
}

export type Concat = {
  refs?: Ref[]
}