/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

import * as fm from "../fetch.pb";
import * as GoogleProtobufStruct from "../google/protobuf/struct.pb";
import * as GoogleProtobufTimestamp from "../google/protobuf/timestamp.pb";
export type User = {
    __typename?: "User";
    email?: string;
    lastAuthenticated?: GoogleProtobufTimestamp.Timestamp;
};

export type IdentityServerConfig = {
    __typename?: "IdentityServerConfig";
    issuer?: string;
    idTokenExpiry?: string;
    rotationTokenExpiry?: string;
};

export type SetIdentityServerConfigRequest = {
    __typename?: "SetIdentityServerConfigRequest";
    config?: IdentityServerConfig;
};

export type SetIdentityServerConfigResponse = {
    __typename?: "SetIdentityServerConfigResponse";
};

export type GetIdentityServerConfigRequest = {
    __typename?: "GetIdentityServerConfigRequest";
};

export type GetIdentityServerConfigResponse = {
    __typename?: "GetIdentityServerConfigResponse";
    config?: IdentityServerConfig;
};

export type IDPConnector = {
    __typename?: "IDPConnector";
    id?: string;
    name?: string;
    type?: string;
    configVersion?: string;
    jsonConfig?: string;
    config?: GoogleProtobufStruct.Struct;
};

export type CreateIDPConnectorRequest = {
    __typename?: "CreateIDPConnectorRequest";
    connector?: IDPConnector;
};

export type CreateIDPConnectorResponse = {
    __typename?: "CreateIDPConnectorResponse";
};

export type UpdateIDPConnectorRequest = {
    __typename?: "UpdateIDPConnectorRequest";
    connector?: IDPConnector;
};

export type UpdateIDPConnectorResponse = {
    __typename?: "UpdateIDPConnectorResponse";
};

export type ListIDPConnectorsRequest = {
    __typename?: "ListIDPConnectorsRequest";
};

export type ListIDPConnectorsResponse = {
    __typename?: "ListIDPConnectorsResponse";
    connectors?: IDPConnector[];
};

export type GetIDPConnectorRequest = {
    __typename?: "GetIDPConnectorRequest";
    id?: string;
};

export type GetIDPConnectorResponse = {
    __typename?: "GetIDPConnectorResponse";
    connector?: IDPConnector;
};

export type DeleteIDPConnectorRequest = {
    __typename?: "DeleteIDPConnectorRequest";
    id?: string;
};

export type DeleteIDPConnectorResponse = {
    __typename?: "DeleteIDPConnectorResponse";
};

export type OIDCClient = {
    __typename?: "OIDCClient";
    id?: string;
    redirectUris?: string[];
    trustedPeers?: string[];
    name?: string;
    secret?: string;
};

export type CreateOIDCClientRequest = {
    __typename?: "CreateOIDCClientRequest";
    client?: OIDCClient;
};

export type CreateOIDCClientResponse = {
    __typename?: "CreateOIDCClientResponse";
    client?: OIDCClient;
};

export type GetOIDCClientRequest = {
    __typename?: "GetOIDCClientRequest";
    id?: string;
};

export type GetOIDCClientResponse = {
    __typename?: "GetOIDCClientResponse";
    client?: OIDCClient;
};

export type ListOIDCClientsRequest = {
    __typename?: "ListOIDCClientsRequest";
};

export type ListOIDCClientsResponse = {
    __typename?: "ListOIDCClientsResponse";
    clients?: OIDCClient[];
};

export type UpdateOIDCClientRequest = {
    __typename?: "UpdateOIDCClientRequest";
    client?: OIDCClient;
};

export type UpdateOIDCClientResponse = {
    __typename?: "UpdateOIDCClientResponse";
};

export type DeleteOIDCClientRequest = {
    __typename?: "DeleteOIDCClientRequest";
    id?: string;
};

export type DeleteOIDCClientResponse = {
    __typename?: "DeleteOIDCClientResponse";
};

export type DeleteAllRequest = {
    __typename?: "DeleteAllRequest";
};

export type DeleteAllResponse = {
    __typename?: "DeleteAllResponse";
};

export class API {
    static SetIdentityServerConfig(req: SetIdentityServerConfigRequest, initReq?: fm.InitReq): Promise<SetIdentityServerConfigResponse> {
        return fm.fetchReq<SetIdentityServerConfigRequest, SetIdentityServerConfigResponse>(`/identity_v2.API/SetIdentityServerConfig`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static GetIdentityServerConfig(req: GetIdentityServerConfigRequest, initReq?: fm.InitReq): Promise<GetIdentityServerConfigResponse> {
        return fm.fetchReq<GetIdentityServerConfigRequest, GetIdentityServerConfigResponse>(`/identity_v2.API/GetIdentityServerConfig`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static CreateIDPConnector(req: CreateIDPConnectorRequest, initReq?: fm.InitReq): Promise<CreateIDPConnectorResponse> {
        return fm.fetchReq<CreateIDPConnectorRequest, CreateIDPConnectorResponse>(`/identity_v2.API/CreateIDPConnector`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static UpdateIDPConnector(req: UpdateIDPConnectorRequest, initReq?: fm.InitReq): Promise<UpdateIDPConnectorResponse> {
        return fm.fetchReq<UpdateIDPConnectorRequest, UpdateIDPConnectorResponse>(`/identity_v2.API/UpdateIDPConnector`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static ListIDPConnectors(req: ListIDPConnectorsRequest, initReq?: fm.InitReq): Promise<ListIDPConnectorsResponse> {
        return fm.fetchReq<ListIDPConnectorsRequest, ListIDPConnectorsResponse>(`/identity_v2.API/ListIDPConnectors`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static GetIDPConnector(req: GetIDPConnectorRequest, initReq?: fm.InitReq): Promise<GetIDPConnectorResponse> {
        return fm.fetchReq<GetIDPConnectorRequest, GetIDPConnectorResponse>(`/identity_v2.API/GetIDPConnector`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static DeleteIDPConnector(req: DeleteIDPConnectorRequest, initReq?: fm.InitReq): Promise<DeleteIDPConnectorResponse> {
        return fm.fetchReq<DeleteIDPConnectorRequest, DeleteIDPConnectorResponse>(`/identity_v2.API/DeleteIDPConnector`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static CreateOIDCClient(req: CreateOIDCClientRequest, initReq?: fm.InitReq): Promise<CreateOIDCClientResponse> {
        return fm.fetchReq<CreateOIDCClientRequest, CreateOIDCClientResponse>(`/identity_v2.API/CreateOIDCClient`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static UpdateOIDCClient(req: UpdateOIDCClientRequest, initReq?: fm.InitReq): Promise<UpdateOIDCClientResponse> {
        return fm.fetchReq<UpdateOIDCClientRequest, UpdateOIDCClientResponse>(`/identity_v2.API/UpdateOIDCClient`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static GetOIDCClient(req: GetOIDCClientRequest, initReq?: fm.InitReq): Promise<GetOIDCClientResponse> {
        return fm.fetchReq<GetOIDCClientRequest, GetOIDCClientResponse>(`/identity_v2.API/GetOIDCClient`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static ListOIDCClients(req: ListOIDCClientsRequest, initReq?: fm.InitReq): Promise<ListOIDCClientsResponse> {
        return fm.fetchReq<ListOIDCClientsRequest, ListOIDCClientsResponse>(`/identity_v2.API/ListOIDCClients`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static DeleteOIDCClient(req: DeleteOIDCClientRequest, initReq?: fm.InitReq): Promise<DeleteOIDCClientResponse> {
        return fm.fetchReq<DeleteOIDCClientRequest, DeleteOIDCClientResponse>(`/identity_v2.API/DeleteOIDCClient`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static DeleteAll(req: DeleteAllRequest, initReq?: fm.InitReq): Promise<DeleteAllResponse> {
        return fm.fetchReq<DeleteAllRequest, DeleteAllResponse>(`/identity_v2.API/DeleteAll`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
}
