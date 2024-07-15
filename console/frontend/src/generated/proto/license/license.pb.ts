/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

import * as Enterprise_v2Enterprise from "../enterprise/enterprise.pb";
import * as fm from "../fetch.pb";
import * as GoogleProtobufTimestamp from "../google/protobuf/timestamp.pb";
export type ActivateRequest = {
    __typename?: "ActivateRequest";
    activationCode?: string;
    expires?: GoogleProtobufTimestamp.Timestamp;
};

export type ActivateResponse = {
    __typename?: "ActivateResponse";
    info?: Enterprise_v2Enterprise.TokenInfo;
};

export type GetActivationCodeRequest = {
    __typename?: "GetActivationCodeRequest";
};

export type GetActivationCodeResponse = {
    __typename?: "GetActivationCodeResponse";
    state?: Enterprise_v2Enterprise.State;
    info?: Enterprise_v2Enterprise.TokenInfo;
    activationCode?: string;
};

export type DeactivateRequest = {
    __typename?: "DeactivateRequest";
};

export type DeactivateResponse = {
    __typename?: "DeactivateResponse";
};

export type AddClusterRequest = {
    __typename?: "AddClusterRequest";
    id?: string;
    address?: string;
    secret?: string;
    userAddress?: string;
    clusterDeploymentId?: string;
    enterpriseServer?: boolean;
};

export type AddClusterResponse = {
    __typename?: "AddClusterResponse";
    secret?: string;
};

export type DeleteClusterRequest = {
    __typename?: "DeleteClusterRequest";
    id?: string;
};

export type DeleteClusterResponse = {
    __typename?: "DeleteClusterResponse";
};

export type ClusterStatus = {
    __typename?: "ClusterStatus";
    id?: string;
    address?: string;
    version?: string;
    authEnabled?: boolean;
    clientId?: string;
    lastHeartbeat?: GoogleProtobufTimestamp.Timestamp;
    createdAt?: GoogleProtobufTimestamp.Timestamp;
};

export type UpdateClusterRequest = {
    __typename?: "UpdateClusterRequest";
    id?: string;
    address?: string;
    userAddress?: string;
    clusterDeploymentId?: string;
    secret?: string;
};

export type UpdateClusterResponse = {
    __typename?: "UpdateClusterResponse";
};

export type ListClustersRequest = {
    __typename?: "ListClustersRequest";
};

export type ListClustersResponse = {
    __typename?: "ListClustersResponse";
    clusters?: ClusterStatus[];
};

export type DeleteAllRequest = {
    __typename?: "DeleteAllRequest";
};

export type DeleteAllResponse = {
    __typename?: "DeleteAllResponse";
};

export type HeartbeatRequest = {
    __typename?: "HeartbeatRequest";
    id?: string;
    secret?: string;
    version?: string;
    authEnabled?: boolean;
    clientId?: string;
};

export type HeartbeatResponse = {
    __typename?: "HeartbeatResponse";
    license?: Enterprise_v2Enterprise.LicenseRecord;
};

export type UserClusterInfo = {
    __typename?: "UserClusterInfo";
    id?: string;
    clusterDeploymentId?: string;
    address?: string;
    enterpriseServer?: boolean;
};

export type ListUserClustersRequest = {
    __typename?: "ListUserClustersRequest";
};

export type ListUserClustersResponse = {
    __typename?: "ListUserClustersResponse";
    clusters?: UserClusterInfo[];
};

export class API {
    static Activate(req: ActivateRequest, initReq?: fm.InitReq): Promise<ActivateResponse> {
        return fm.fetchReq<ActivateRequest, ActivateResponse>(`/license_v2.API/Activate`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static GetActivationCode(req: GetActivationCodeRequest, initReq?: fm.InitReq): Promise<GetActivationCodeResponse> {
        return fm.fetchReq<GetActivationCodeRequest, GetActivationCodeResponse>(`/license_v2.API/GetActivationCode`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static DeleteAll(req: DeleteAllRequest, initReq?: fm.InitReq): Promise<DeleteAllResponse> {
        return fm.fetchReq<DeleteAllRequest, DeleteAllResponse>(`/license_v2.API/DeleteAll`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static AddCluster(req: AddClusterRequest, initReq?: fm.InitReq): Promise<AddClusterResponse> {
        return fm.fetchReq<AddClusterRequest, AddClusterResponse>(`/license_v2.API/AddCluster`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static DeleteCluster(req: DeleteClusterRequest, initReq?: fm.InitReq): Promise<DeleteClusterResponse> {
        return fm.fetchReq<DeleteClusterRequest, DeleteClusterResponse>(`/license_v2.API/DeleteCluster`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static ListClusters(req: ListClustersRequest, initReq?: fm.InitReq): Promise<ListClustersResponse> {
        return fm.fetchReq<ListClustersRequest, ListClustersResponse>(`/license_v2.API/ListClusters`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static UpdateCluster(req: UpdateClusterRequest, initReq?: fm.InitReq): Promise<UpdateClusterResponse> {
        return fm.fetchReq<UpdateClusterRequest, UpdateClusterResponse>(`/license_v2.API/UpdateCluster`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static Heartbeat(req: HeartbeatRequest, initReq?: fm.InitReq): Promise<HeartbeatResponse> {
        return fm.fetchReq<HeartbeatRequest, HeartbeatResponse>(`/license_v2.API/Heartbeat`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static ListUserClusters(req: ListUserClustersRequest, initReq?: fm.InitReq): Promise<ListUserClustersResponse> {
        return fm.fetchReq<ListUserClustersRequest, ListUserClustersResponse>(`/license_v2.API/ListUserClusters`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
}
