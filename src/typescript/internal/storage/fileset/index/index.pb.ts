/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

import * as ChunkChunk from "../../chunk/chunk.pb"
export type Index = {
  path?: string
  range?: Range
  file?: File
  numFiles?: string
  sizeBytes?: string
}

export type Range = {
  offset?: string
  lastPath?: string
  chunkRef?: ChunkChunk.DataRef
}

export type File = {
  datum?: string
  dataRefs?: ChunkChunk.DataRef[]
}