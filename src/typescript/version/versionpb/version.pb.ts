/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

import * as fm from "../../fetch.pb"
import * as GoogleProtobufEmpty from "../../google/protobuf/empty.pb"
export type Version = {
  major?: number
  minor?: number
  micro?: number
  additional?: string
  gitCommit?: string
  gitTreeModified?: string
  buildDate?: string
  goVersion?: string
  platform?: string
}

export class API {
  static GetVersion(req: GoogleProtobufEmpty.Empty, initReq?: fm.InitReq): Promise<Version> {
    return fm.fetchReq<GoogleProtobufEmpty.Empty, Version>(`/versionpb_v2.API/GetVersion`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
}