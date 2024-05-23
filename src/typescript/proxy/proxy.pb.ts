/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

import * as fm from "../fetch.pb"
export type ListenRequest = {
  channel?: string
}

export type ListenResponse = {
  extra?: string
}

export class API {
  static Listen(req: ListenRequest, entityNotifier?: fm.NotifyStreamEntityArrival<ListenResponse>, initReq?: fm.InitReq): Promise<void> {
    return fm.fetchStreamingRequest<ListenRequest, ListenResponse>(`/proxy.API/Listen`, entityNotifier, {...initReq, method: "POST", body: JSON.stringify(req, fm.replacer)})
  }
}