/* eslint-disable */
// @ts-nocheck
/*
* This file is a generated Typescript file for GRPC Gateway, DO NOT MODIFY
*/

import * as fm from "../fetch.pb";
import * as GoogleProtobufTimestamp from "../google/protobuf/timestamp.pb";

export enum Permission {
    PERMISSION_UNKNOWN = "PERMISSION_UNKNOWN",
    CLUSTER_MODIFY_BINDINGS = "CLUSTER_MODIFY_BINDINGS",
    CLUSTER_GET_BINDINGS = "CLUSTER_GET_BINDINGS",
    CLUSTER_GET_PACHD_LOGS = "CLUSTER_GET_PACHD_LOGS",
    CLUSTER_GET_LOKI_LOGS = "CLUSTER_GET_LOKI_LOGS",
    CLUSTER_AUTH_ACTIVATE = "CLUSTER_AUTH_ACTIVATE",
    CLUSTER_AUTH_DEACTIVATE = "CLUSTER_AUTH_DEACTIVATE",
    CLUSTER_AUTH_GET_CONFIG = "CLUSTER_AUTH_GET_CONFIG",
    CLUSTER_AUTH_SET_CONFIG = "CLUSTER_AUTH_SET_CONFIG",
    CLUSTER_AUTH_GET_ROBOT_TOKEN = "CLUSTER_AUTH_GET_ROBOT_TOKEN",
    CLUSTER_AUTH_MODIFY_GROUP_MEMBERS = "CLUSTER_AUTH_MODIFY_GROUP_MEMBERS",
    CLUSTER_AUTH_GET_GROUPS = "CLUSTER_AUTH_GET_GROUPS",
    CLUSTER_AUTH_GET_GROUP_USERS = "CLUSTER_AUTH_GET_GROUP_USERS",
    CLUSTER_AUTH_EXTRACT_TOKENS = "CLUSTER_AUTH_EXTRACT_TOKENS",
    CLUSTER_AUTH_RESTORE_TOKEN = "CLUSTER_AUTH_RESTORE_TOKEN",
    CLUSTER_AUTH_GET_PERMISSIONS_FOR_PRINCIPAL = "CLUSTER_AUTH_GET_PERMISSIONS_FOR_PRINCIPAL",
    CLUSTER_AUTH_DELETE_EXPIRED_TOKENS = "CLUSTER_AUTH_DELETE_EXPIRED_TOKENS",
    CLUSTER_AUTH_REVOKE_USER_TOKENS = "CLUSTER_AUTH_REVOKE_USER_TOKENS",
    CLUSTER_AUTH_ROTATE_ROOT_TOKEN = "CLUSTER_AUTH_ROTATE_ROOT_TOKEN",
    CLUSTER_ENTERPRISE_ACTIVATE = "CLUSTER_ENTERPRISE_ACTIVATE",
    CLUSTER_ENTERPRISE_HEARTBEAT = "CLUSTER_ENTERPRISE_HEARTBEAT",
    CLUSTER_ENTERPRISE_GET_CODE = "CLUSTER_ENTERPRISE_GET_CODE",
    CLUSTER_ENTERPRISE_DEACTIVATE = "CLUSTER_ENTERPRISE_DEACTIVATE",
    CLUSTER_ENTERPRISE_PAUSE = "CLUSTER_ENTERPRISE_PAUSE",
    CLUSTER_IDENTITY_SET_CONFIG = "CLUSTER_IDENTITY_SET_CONFIG",
    CLUSTER_IDENTITY_GET_CONFIG = "CLUSTER_IDENTITY_GET_CONFIG",
    CLUSTER_IDENTITY_CREATE_IDP = "CLUSTER_IDENTITY_CREATE_IDP",
    CLUSTER_IDENTITY_UPDATE_IDP = "CLUSTER_IDENTITY_UPDATE_IDP",
    CLUSTER_IDENTITY_LIST_IDPS = "CLUSTER_IDENTITY_LIST_IDPS",
    CLUSTER_IDENTITY_GET_IDP = "CLUSTER_IDENTITY_GET_IDP",
    CLUSTER_IDENTITY_DELETE_IDP = "CLUSTER_IDENTITY_DELETE_IDP",
    CLUSTER_IDENTITY_CREATE_OIDC_CLIENT = "CLUSTER_IDENTITY_CREATE_OIDC_CLIENT",
    CLUSTER_IDENTITY_UPDATE_OIDC_CLIENT = "CLUSTER_IDENTITY_UPDATE_OIDC_CLIENT",
    CLUSTER_IDENTITY_LIST_OIDC_CLIENTS = "CLUSTER_IDENTITY_LIST_OIDC_CLIENTS",
    CLUSTER_IDENTITY_GET_OIDC_CLIENT = "CLUSTER_IDENTITY_GET_OIDC_CLIENT",
    CLUSTER_IDENTITY_DELETE_OIDC_CLIENT = "CLUSTER_IDENTITY_DELETE_OIDC_CLIENT",
    CLUSTER_DEBUG_DUMP = "CLUSTER_DEBUG_DUMP",
    CLUSTER_LICENSE_ACTIVATE = "CLUSTER_LICENSE_ACTIVATE",
    CLUSTER_LICENSE_GET_CODE = "CLUSTER_LICENSE_GET_CODE",
    CLUSTER_LICENSE_ADD_CLUSTER = "CLUSTER_LICENSE_ADD_CLUSTER",
    CLUSTER_LICENSE_UPDATE_CLUSTER = "CLUSTER_LICENSE_UPDATE_CLUSTER",
    CLUSTER_LICENSE_DELETE_CLUSTER = "CLUSTER_LICENSE_DELETE_CLUSTER",
    CLUSTER_LICENSE_LIST_CLUSTERS = "CLUSTER_LICENSE_LIST_CLUSTERS",
    CLUSTER_CREATE_SECRET = "CLUSTER_CREATE_SECRET",
    CLUSTER_LIST_SECRETS = "CLUSTER_LIST_SECRETS",
    SECRET_DELETE = "SECRET_DELETE",
    SECRET_INSPECT = "SECRET_INSPECT",
    CLUSTER_EDIT_CLUSTER_METADATA = "CLUSTER_EDIT_CLUSTER_METADATA",
    CLUSTER_DELETE_ALL = "CLUSTER_DELETE_ALL",
    REPO_READ = "REPO_READ",
    REPO_WRITE = "REPO_WRITE",
    REPO_MODIFY_BINDINGS = "REPO_MODIFY_BINDINGS",
    REPO_DELETE = "REPO_DELETE",
    REPO_INSPECT_COMMIT = "REPO_INSPECT_COMMIT",
    REPO_LIST_COMMIT = "REPO_LIST_COMMIT",
    REPO_DELETE_COMMIT = "REPO_DELETE_COMMIT",
    REPO_CREATE_BRANCH = "REPO_CREATE_BRANCH",
    REPO_LIST_BRANCH = "REPO_LIST_BRANCH",
    REPO_DELETE_BRANCH = "REPO_DELETE_BRANCH",
    REPO_INSPECT_FILE = "REPO_INSPECT_FILE",
    REPO_LIST_FILE = "REPO_LIST_FILE",
    REPO_ADD_PIPELINE_READER = "REPO_ADD_PIPELINE_READER",
    REPO_REMOVE_PIPELINE_READER = "REPO_REMOVE_PIPELINE_READER",
    REPO_ADD_PIPELINE_WRITER = "REPO_ADD_PIPELINE_WRITER",
    PIPELINE_LIST_JOB = "PIPELINE_LIST_JOB",
    CLUSTER_SET_DEFAULTS = "CLUSTER_SET_DEFAULTS",
    PROJECT_SET_DEFAULTS = "PROJECT_SET_DEFAULTS",
    PROJECT_CREATE = "PROJECT_CREATE",
    PROJECT_DELETE = "PROJECT_DELETE",
    PROJECT_LIST_REPO = "PROJECT_LIST_REPO",
    PROJECT_CREATE_REPO = "PROJECT_CREATE_REPO",
    PROJECT_MODIFY_BINDINGS = "PROJECT_MODIFY_BINDINGS",
}

export enum ResourceType {
    RESOURCE_TYPE_UNKNOWN = "RESOURCE_TYPE_UNKNOWN",
    CLUSTER = "CLUSTER",
    REPO = "REPO",
    SPEC_REPO = "SPEC_REPO",
    PROJECT = "PROJECT",
}

export type ActivateRequest = {
    __typename?: "ActivateRequest";
    rootToken?: string;
};

export type ActivateResponse = {
    __typename?: "ActivateResponse";
    pachToken?: string;
};

export type DeactivateRequest = {
    __typename?: "DeactivateRequest";
};

export type DeactivateResponse = {
    __typename?: "DeactivateResponse";
};

export type RotateRootTokenRequest = {
    __typename?: "RotateRootTokenRequest";
    rootToken?: string;
};

export type RotateRootTokenResponse = {
    __typename?: "RotateRootTokenResponse";
    rootToken?: string;
};

export type OIDCConfig = {
    __typename?: "OIDCConfig";
    issuer?: string;
    clientId?: string;
    clientSecret?: string;
    redirectUri?: string;
    scopes?: string[];
    requireEmailVerified?: boolean;
    localhostIssuer?: boolean;
    userAccessibleIssuerHost?: string;
};

export type GetConfigurationRequest = {
    __typename?: "GetConfigurationRequest";
};

export type GetConfigurationResponse = {
    __typename?: "GetConfigurationResponse";
    configuration?: OIDCConfig;
};

export type SetConfigurationRequest = {
    __typename?: "SetConfigurationRequest";
    configuration?: OIDCConfig;
};

export type SetConfigurationResponse = {
    __typename?: "SetConfigurationResponse";
};

export type TokenInfo = {
    __typename?: "TokenInfo";
    subject?: string;
    expiration?: GoogleProtobufTimestamp.Timestamp;
    hashedToken?: string;
};

export type AuthenticateRequest = {
    __typename?: "AuthenticateRequest";
    oidcState?: string;
    idToken?: string;
};

export type AuthenticateResponse = {
    __typename?: "AuthenticateResponse";
    pachToken?: string;
};

export type WhoAmIRequest = {
    __typename?: "WhoAmIRequest";
};

export type WhoAmIResponse = {
    __typename?: "WhoAmIResponse";
    username?: string;
    expiration?: GoogleProtobufTimestamp.Timestamp;
};

export type GetRolesForPermissionRequest = {
    __typename?: "GetRolesForPermissionRequest";
    permission?: Permission;
};

export type GetRolesForPermissionResponse = {
    __typename?: "GetRolesForPermissionResponse";
    roles?: Role[];
};

export type Roles = {
    __typename?: "Roles";
    roles?: { [key: string]: boolean; };
};

export type RoleBinding = {
    __typename?: "RoleBinding";
    entries?: { [key: string]: Roles; };
};

export type Resource = {
    __typename?: "Resource";
    type?: ResourceType;
    name?: string;
};

export type Users = {
    __typename?: "Users";
    usernames?: { [key: string]: boolean; };
};

export type Groups = {
    __typename?: "Groups";
    groups?: { [key: string]: boolean; };
};

export type Role = {
    __typename?: "Role";
    name?: string;
    permissions?: Permission[];
    canBeBoundTo?: ResourceType[];
    returnedFor?: ResourceType[];
};

export type AuthorizeRequest = {
    __typename?: "AuthorizeRequest";
    resource?: Resource;
    permissions?: Permission[];
};

export type AuthorizeResponse = {
    __typename?: "AuthorizeResponse";
    authorized?: boolean;
    satisfied?: Permission[];
    missing?: Permission[];
    principal?: string;
};

export type GetPermissionsRequest = {
    __typename?: "GetPermissionsRequest";
    resource?: Resource;
};

export type GetPermissionsForPrincipalRequest = {
    __typename?: "GetPermissionsForPrincipalRequest";
    resource?: Resource;
    principal?: string;
};

export type GetPermissionsResponse = {
    __typename?: "GetPermissionsResponse";
    permissions?: Permission[];
    roles?: string[];
};

export type ModifyRoleBindingRequest = {
    __typename?: "ModifyRoleBindingRequest";
    resource?: Resource;
    principal?: string;
    roles?: string[];
};

export type ModifyRoleBindingResponse = {
    __typename?: "ModifyRoleBindingResponse";
};

export type GetRoleBindingRequest = {
    __typename?: "GetRoleBindingRequest";
    resource?: Resource;
};

export type GetRoleBindingResponse = {
    __typename?: "GetRoleBindingResponse";
    binding?: RoleBinding;
};

export type SessionInfo = {
    __typename?: "SessionInfo";
    nonce?: string;
    email?: string;
    conversionErr?: boolean;
};

export type GetOIDCLoginRequest = {
    __typename?: "GetOIDCLoginRequest";
};

export type GetOIDCLoginResponse = {
    __typename?: "GetOIDCLoginResponse";
    loginUrl?: string;
    state?: string;
};

export type GetRobotTokenRequest = {
    __typename?: "GetRobotTokenRequest";
    robot?: string;
    ttl?: string;
};

export type GetRobotTokenResponse = {
    __typename?: "GetRobotTokenResponse";
    token?: string;
};

export type RevokeAuthTokenRequest = {
    __typename?: "RevokeAuthTokenRequest";
    token?: string;
};

export type RevokeAuthTokenResponse = {
    __typename?: "RevokeAuthTokenResponse";
    number?: string;
};

export type SetGroupsForUserRequest = {
    __typename?: "SetGroupsForUserRequest";
    username?: string;
    groups?: string[];
};

export type SetGroupsForUserResponse = {
    __typename?: "SetGroupsForUserResponse";
};

export type ModifyMembersRequest = {
    __typename?: "ModifyMembersRequest";
    group?: string;
    add?: string[];
    remove?: string[];
};

export type ModifyMembersResponse = {
    __typename?: "ModifyMembersResponse";
};

export type GetGroupsRequest = {
    __typename?: "GetGroupsRequest";
};

export type GetGroupsForPrincipalRequest = {
    __typename?: "GetGroupsForPrincipalRequest";
    principal?: string;
};

export type GetGroupsResponse = {
    __typename?: "GetGroupsResponse";
    groups?: string[];
};

export type GetUsersRequest = {
    __typename?: "GetUsersRequest";
    group?: string;
};

export type GetUsersResponse = {
    __typename?: "GetUsersResponse";
    usernames?: string[];
};

export type ExtractAuthTokensRequest = {
    __typename?: "ExtractAuthTokensRequest";
};

export type ExtractAuthTokensResponse = {
    __typename?: "ExtractAuthTokensResponse";
    tokens?: TokenInfo[];
};

export type RestoreAuthTokenRequest = {
    __typename?: "RestoreAuthTokenRequest";
    token?: TokenInfo;
};

export type RestoreAuthTokenResponse = {
    __typename?: "RestoreAuthTokenResponse";
};

export type RevokeAuthTokensForUserRequest = {
    __typename?: "RevokeAuthTokensForUserRequest";
    username?: string;
};

export type RevokeAuthTokensForUserResponse = {
    __typename?: "RevokeAuthTokensForUserResponse";
    number?: string;
};

export type DeleteExpiredAuthTokensRequest = {
    __typename?: "DeleteExpiredAuthTokensRequest";
};

export type DeleteExpiredAuthTokensResponse = {
    __typename?: "DeleteExpiredAuthTokensResponse";
};

export class API {
    static Activate(req: ActivateRequest, initReq?: fm.InitReq): Promise<ActivateResponse> {
        return fm.fetchReq<ActivateRequest, ActivateResponse>(`/auth_v2.API/Activate`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static Deactivate(req: DeactivateRequest, initReq?: fm.InitReq): Promise<DeactivateResponse> {
        return fm.fetchReq<DeactivateRequest, DeactivateResponse>(`/auth_v2.API/Deactivate`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static GetConfiguration(req: GetConfigurationRequest, initReq?: fm.InitReq): Promise<GetConfigurationResponse> {
        return fm.fetchReq<GetConfigurationRequest, GetConfigurationResponse>(`/auth_v2.API/GetConfiguration`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static SetConfiguration(req: SetConfigurationRequest, initReq?: fm.InitReq): Promise<SetConfigurationResponse> {
        return fm.fetchReq<SetConfigurationRequest, SetConfigurationResponse>(`/auth_v2.API/SetConfiguration`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static Authenticate(req: AuthenticateRequest, initReq?: fm.InitReq): Promise<AuthenticateResponse> {
        return fm.fetchReq<AuthenticateRequest, AuthenticateResponse>(`/auth_v2.API/Authenticate`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static Authorize(req: AuthorizeRequest, initReq?: fm.InitReq): Promise<AuthorizeResponse> {
        return fm.fetchReq<AuthorizeRequest, AuthorizeResponse>(`/auth_v2.API/Authorize`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static GetPermissions(req: GetPermissionsRequest, initReq?: fm.InitReq): Promise<GetPermissionsResponse> {
        return fm.fetchReq<GetPermissionsRequest, GetPermissionsResponse>(`/auth_v2.API/GetPermissions`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static GetPermissionsForPrincipal(req: GetPermissionsForPrincipalRequest, initReq?: fm.InitReq): Promise<GetPermissionsResponse> {
        return fm.fetchReq<GetPermissionsForPrincipalRequest, GetPermissionsResponse>(`/auth_v2.API/GetPermissionsForPrincipal`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static WhoAmI(req: WhoAmIRequest, initReq?: fm.InitReq): Promise<WhoAmIResponse> {
        return fm.fetchReq<WhoAmIRequest, WhoAmIResponse>(`/auth_v2.API/WhoAmI`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static GetRolesForPermission(req: GetRolesForPermissionRequest, initReq?: fm.InitReq): Promise<GetRolesForPermissionResponse> {
        return fm.fetchReq<GetRolesForPermissionRequest, GetRolesForPermissionResponse>(`/auth_v2.API/GetRolesForPermission`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static ModifyRoleBinding(req: ModifyRoleBindingRequest, initReq?: fm.InitReq): Promise<ModifyRoleBindingResponse> {
        return fm.fetchReq<ModifyRoleBindingRequest, ModifyRoleBindingResponse>(`/auth_v2.API/ModifyRoleBinding`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static GetRoleBinding(req: GetRoleBindingRequest, initReq?: fm.InitReq): Promise<GetRoleBindingResponse> {
        return fm.fetchReq<GetRoleBindingRequest, GetRoleBindingResponse>(`/auth_v2.API/GetRoleBinding`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static GetOIDCLogin(req: GetOIDCLoginRequest, initReq?: fm.InitReq): Promise<GetOIDCLoginResponse> {
        return fm.fetchReq<GetOIDCLoginRequest, GetOIDCLoginResponse>(`/auth_v2.API/GetOIDCLogin`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static GetRobotToken(req: GetRobotTokenRequest, initReq?: fm.InitReq): Promise<GetRobotTokenResponse> {
        return fm.fetchReq<GetRobotTokenRequest, GetRobotTokenResponse>(`/auth_v2.API/GetRobotToken`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static RevokeAuthToken(req: RevokeAuthTokenRequest, initReq?: fm.InitReq): Promise<RevokeAuthTokenResponse> {
        return fm.fetchReq<RevokeAuthTokenRequest, RevokeAuthTokenResponse>(`/auth_v2.API/RevokeAuthToken`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static RevokeAuthTokensForUser(req: RevokeAuthTokensForUserRequest, initReq?: fm.InitReq): Promise<RevokeAuthTokensForUserResponse> {
        return fm.fetchReq<RevokeAuthTokensForUserRequest, RevokeAuthTokensForUserResponse>(`/auth_v2.API/RevokeAuthTokensForUser`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static SetGroupsForUser(req: SetGroupsForUserRequest, initReq?: fm.InitReq): Promise<SetGroupsForUserResponse> {
        return fm.fetchReq<SetGroupsForUserRequest, SetGroupsForUserResponse>(`/auth_v2.API/SetGroupsForUser`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static ModifyMembers(req: ModifyMembersRequest, initReq?: fm.InitReq): Promise<ModifyMembersResponse> {
        return fm.fetchReq<ModifyMembersRequest, ModifyMembersResponse>(`/auth_v2.API/ModifyMembers`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static GetGroups(req: GetGroupsRequest, initReq?: fm.InitReq): Promise<GetGroupsResponse> {
        return fm.fetchReq<GetGroupsRequest, GetGroupsResponse>(`/auth_v2.API/GetGroups`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static GetGroupsForPrincipal(req: GetGroupsForPrincipalRequest, initReq?: fm.InitReq): Promise<GetGroupsResponse> {
        return fm.fetchReq<GetGroupsForPrincipalRequest, GetGroupsResponse>(`/auth_v2.API/GetGroupsForPrincipal`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static GetUsers(req: GetUsersRequest, initReq?: fm.InitReq): Promise<GetUsersResponse> {
        return fm.fetchReq<GetUsersRequest, GetUsersResponse>(`/auth_v2.API/GetUsers`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static ExtractAuthTokens(req: ExtractAuthTokensRequest, initReq?: fm.InitReq): Promise<ExtractAuthTokensResponse> {
        return fm.fetchReq<ExtractAuthTokensRequest, ExtractAuthTokensResponse>(`/auth_v2.API/ExtractAuthTokens`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static RestoreAuthToken(req: RestoreAuthTokenRequest, initReq?: fm.InitReq): Promise<RestoreAuthTokenResponse> {
        return fm.fetchReq<RestoreAuthTokenRequest, RestoreAuthTokenResponse>(`/auth_v2.API/RestoreAuthToken`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static DeleteExpiredAuthTokens(req: DeleteExpiredAuthTokensRequest, initReq?: fm.InitReq): Promise<DeleteExpiredAuthTokensResponse> {
        return fm.fetchReq<DeleteExpiredAuthTokensRequest, DeleteExpiredAuthTokensResponse>(`/auth_v2.API/DeleteExpiredAuthTokens`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
    static RotateRootToken(req: RotateRootTokenRequest, initReq?: fm.InitReq): Promise<RotateRootTokenResponse> {
        return fm.fetchReq<RotateRootTokenRequest, RotateRootTokenResponse>(`/auth_v2.API/RotateRootToken`, { ...initReq, method: "POST", body: JSON.stringify(req, fm.replacer) });
    }
}
