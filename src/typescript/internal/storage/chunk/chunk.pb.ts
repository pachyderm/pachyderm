/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

export enum CompressionAlgo {
  NONE = "NONE",
  GZIP_BEST_SPEED = "GZIP_BEST_SPEED",
}

export enum EncryptionAlgo {
  ENCRYPTION_ALGO_UNKNOWN = "ENCRYPTION_ALGO_UNKNOWN",
  CHACHA20 = "CHACHA20",
}

export type DataRef = {
  ref?: Ref
  hash?: Uint8Array
  offsetBytes?: string
  sizeBytes?: string
}

export type Ref = {
  id?: Uint8Array
  sizeBytes?: string
  edge?: boolean
  dek?: Uint8Array
  encryptionAlgo?: EncryptionAlgo
  compressionAlgo?: CompressionAlgo
}