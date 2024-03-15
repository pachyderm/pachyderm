/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

import * as fm from "../fetch.pb"
export type EditMetadataRequest = {
}

export type EditMetadataResponse = {
}

export class API {
  static EditMetadata(req: EditMetadataRequest, initReq?: fm.InitReq): Promise<EditMetadataResponse> {
    return fm.fetchReq<EditMetadataRequest, EditMetadataResponse>(`/metadata.API/EditMetadata`, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
}