/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

import * as fm from "../fetch.pb";
import * as GoogleProtobufEmpty from "../google/protobuf/empty.pb";
import * as Pps_v2Pps from "../pps/pps.pb";
export type CancelRequest = {
    __typename?: "CancelRequest";
    jobId?: string;
    dataFilters?: string[];
};

export type CancelResponse = {
    __typename?: "CancelResponse";
    success?: boolean;
};

export type NextDatumRequest = {
    __typename?: "NextDatumRequest";
    error?: string;
};

export type NextDatumResponse = {
    __typename?: "NextDatumResponse";
    env?: string[];
};

export class Worker {
    static Status(req: GoogleProtobufEmpty.Empty, initReq?: fm.InitReq): Promise<Pps_v2Pps.WorkerStatus> {
        return fm.fetchReq<GoogleProtobufEmpty.Empty, Pps_v2Pps.WorkerStatus>(`/pachyderm.worker.Worker/Status`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static Cancel(req: CancelRequest, initReq?: fm.InitReq): Promise<CancelResponse> {
        return fm.fetchReq<CancelRequest, CancelResponse>(`/pachyderm.worker.Worker/Cancel`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static NextDatum(req: NextDatumRequest, initReq?: fm.InitReq): Promise<NextDatumResponse> {
        return fm.fetchReq<NextDatumRequest, NextDatumResponse>(`/pachyderm.worker.Worker/NextDatum`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
}
