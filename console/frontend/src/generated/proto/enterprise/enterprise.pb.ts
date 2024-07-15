/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

import * as fm from "../fetch.pb";
import * as GoogleProtobufTimestamp from "../google/protobuf/timestamp.pb";

export enum State {
    NONE = "NONE",
    ACTIVE = "ACTIVE",
    EXPIRED = "EXPIRED",
    HEARTBEAT_FAILED = "HEARTBEAT_FAILED",
}

export enum PauseStatusResponsePauseStatus {
    UNPAUSED = "UNPAUSED",
    PARTIALLY_PAUSED = "PARTIALLY_PAUSED",
    PAUSED = "PAUSED",
}

export type LicenseRecord = {
    __typename?: "LicenseRecord";
    activationCode?: string;
    expires?: GoogleProtobufTimestamp.Timestamp;
};

export type EnterpriseConfig = {
    __typename?: "EnterpriseConfig";
    licenseServer?: string;
    id?: string;
    secret?: string;
};

export type EnterpriseRecord = {
    __typename?: "EnterpriseRecord";
    license?: LicenseRecord;
    lastHeartbeat?: GoogleProtobufTimestamp.Timestamp;
    heartbeatFailed?: boolean;
};

export type TokenInfo = {
    __typename?: "TokenInfo";
    expires?: GoogleProtobufTimestamp.Timestamp;
};

export type ActivateRequest = {
    __typename?: "ActivateRequest";
    licenseServer?: string;
    id?: string;
    secret?: string;
};

export type ActivateResponse = {
    __typename?: "ActivateResponse";
};

export type GetStateRequest = {
    __typename?: "GetStateRequest";
};

export type GetStateResponse = {
    __typename?: "GetStateResponse";
    state?: State;
    info?: TokenInfo;
    activationCode?: string;
};

export type GetActivationCodeRequest = {
    __typename?: "GetActivationCodeRequest";
};

export type GetActivationCodeResponse = {
    __typename?: "GetActivationCodeResponse";
    state?: State;
    info?: TokenInfo;
    activationCode?: string;
};

export type HeartbeatRequest = {
    __typename?: "HeartbeatRequest";
};

export type HeartbeatResponse = {
    __typename?: "HeartbeatResponse";
};

export type DeactivateRequest = {
    __typename?: "DeactivateRequest";
};

export type DeactivateResponse = {
    __typename?: "DeactivateResponse";
};

export type PauseRequest = {
    __typename?: "PauseRequest";
};

export type PauseResponse = {
    __typename?: "PauseResponse";
};

export type UnpauseRequest = {
    __typename?: "UnpauseRequest";
};

export type UnpauseResponse = {
    __typename?: "UnpauseResponse";
};

export type PauseStatusRequest = {
    __typename?: "PauseStatusRequest";
};

export type PauseStatusResponse = {
    __typename?: "PauseStatusResponse";
    status?: PauseStatusResponsePauseStatus;
};

export class API {
    static Activate(req: ActivateRequest, initReq?: fm.InitReq): Promise<ActivateResponse> {
        return fm.fetchReq<ActivateRequest, ActivateResponse>(`/enterprise_v2.API/Activate`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static GetState(req: GetStateRequest, initReq?: fm.InitReq): Promise<GetStateResponse> {
        return fm.fetchReq<GetStateRequest, GetStateResponse>(`/enterprise_v2.API/GetState`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static GetActivationCode(req: GetActivationCodeRequest, initReq?: fm.InitReq): Promise<GetActivationCodeResponse> {
        return fm.fetchReq<GetActivationCodeRequest, GetActivationCodeResponse>(`/enterprise_v2.API/GetActivationCode`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static Heartbeat(req: HeartbeatRequest, initReq?: fm.InitReq): Promise<HeartbeatResponse> {
        return fm.fetchReq<HeartbeatRequest, HeartbeatResponse>(`/enterprise_v2.API/Heartbeat`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static Deactivate(req: DeactivateRequest, initReq?: fm.InitReq): Promise<DeactivateResponse> {
        return fm.fetchReq<DeactivateRequest, DeactivateResponse>(`/enterprise_v2.API/Deactivate`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static Pause(req: PauseRequest, initReq?: fm.InitReq): Promise<PauseResponse> {
        return fm.fetchReq<PauseRequest, PauseResponse>(`/enterprise_v2.API/Pause`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static Unpause(req: UnpauseRequest, initReq?: fm.InitReq): Promise<UnpauseResponse> {
        return fm.fetchReq<UnpauseRequest, UnpauseResponse>(`/enterprise_v2.API/Unpause`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static PauseStatus(req: PauseStatusRequest, initReq?: fm.InitReq): Promise<PauseStatusResponse> {
        return fm.fetchReq<PauseStatusRequest, PauseStatusResponse>(`/enterprise_v2.API/PauseStatus`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
}
