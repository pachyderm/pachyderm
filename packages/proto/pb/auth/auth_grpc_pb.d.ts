// package: auth
// file: auth/auth.proto

/* tslint:disable */
/* eslint-disable */

import * as grpc from "@grpc/grpc-js";
import {handleClientStreamingCall} from "@grpc/grpc-js/build/src/server-call";
import * as auth_auth_pb from "../auth/auth_pb";
import * as gogoproto_gogo_pb from "../gogoproto/gogo_pb";
import * as google_protobuf_timestamp_pb from "google-protobuf/google/protobuf/timestamp_pb";

interface IAPIService extends grpc.ServiceDefinition<grpc.UntypedServiceImplementation> {
    activate: IAPIService_IActivate;
    deactivate: IAPIService_IDeactivate;
    getConfiguration: IAPIService_IGetConfiguration;
    setConfiguration: IAPIService_ISetConfiguration;
    authenticate: IAPIService_IAuthenticate;
    authorize: IAPIService_IAuthorize;
    getPermissions: IAPIService_IGetPermissions;
    getPermissionsForPrincipal: IAPIService_IGetPermissionsForPrincipal;
    whoAmI: IAPIService_IWhoAmI;
    modifyRoleBinding: IAPIService_IModifyRoleBinding;
    getRoleBinding: IAPIService_IGetRoleBinding;
    getOIDCLogin: IAPIService_IGetOIDCLogin;
    getRobotToken: IAPIService_IGetRobotToken;
    revokeAuthToken: IAPIService_IRevokeAuthToken;
    revokeAuthTokensForUser: IAPIService_IRevokeAuthTokensForUser;
    setGroupsForUser: IAPIService_ISetGroupsForUser;
    modifyMembers: IAPIService_IModifyMembers;
    getGroups: IAPIService_IGetGroups;
    getGroupsForPrincipal: IAPIService_IGetGroupsForPrincipal;
    getUsers: IAPIService_IGetUsers;
    extractAuthTokens: IAPIService_IExtractAuthTokens;
    restoreAuthToken: IAPIService_IRestoreAuthToken;
    deleteExpiredAuthTokens: IAPIService_IDeleteExpiredAuthTokens;
    rotateRootToken: IAPIService_IRotateRootToken;
}

interface IAPIService_IActivate extends grpc.MethodDefinition<auth_auth_pb.ActivateRequest, auth_auth_pb.ActivateResponse> {
    path: "/auth.API/Activate";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<auth_auth_pb.ActivateRequest>;
    requestDeserialize: grpc.deserialize<auth_auth_pb.ActivateRequest>;
    responseSerialize: grpc.serialize<auth_auth_pb.ActivateResponse>;
    responseDeserialize: grpc.deserialize<auth_auth_pb.ActivateResponse>;
}
interface IAPIService_IDeactivate extends grpc.MethodDefinition<auth_auth_pb.DeactivateRequest, auth_auth_pb.DeactivateResponse> {
    path: "/auth.API/Deactivate";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<auth_auth_pb.DeactivateRequest>;
    requestDeserialize: grpc.deserialize<auth_auth_pb.DeactivateRequest>;
    responseSerialize: grpc.serialize<auth_auth_pb.DeactivateResponse>;
    responseDeserialize: grpc.deserialize<auth_auth_pb.DeactivateResponse>;
}
interface IAPIService_IGetConfiguration extends grpc.MethodDefinition<auth_auth_pb.GetConfigurationRequest, auth_auth_pb.GetConfigurationResponse> {
    path: "/auth.API/GetConfiguration";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<auth_auth_pb.GetConfigurationRequest>;
    requestDeserialize: grpc.deserialize<auth_auth_pb.GetConfigurationRequest>;
    responseSerialize: grpc.serialize<auth_auth_pb.GetConfigurationResponse>;
    responseDeserialize: grpc.deserialize<auth_auth_pb.GetConfigurationResponse>;
}
interface IAPIService_ISetConfiguration extends grpc.MethodDefinition<auth_auth_pb.SetConfigurationRequest, auth_auth_pb.SetConfigurationResponse> {
    path: "/auth.API/SetConfiguration";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<auth_auth_pb.SetConfigurationRequest>;
    requestDeserialize: grpc.deserialize<auth_auth_pb.SetConfigurationRequest>;
    responseSerialize: grpc.serialize<auth_auth_pb.SetConfigurationResponse>;
    responseDeserialize: grpc.deserialize<auth_auth_pb.SetConfigurationResponse>;
}
interface IAPIService_IAuthenticate extends grpc.MethodDefinition<auth_auth_pb.AuthenticateRequest, auth_auth_pb.AuthenticateResponse> {
    path: "/auth.API/Authenticate";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<auth_auth_pb.AuthenticateRequest>;
    requestDeserialize: grpc.deserialize<auth_auth_pb.AuthenticateRequest>;
    responseSerialize: grpc.serialize<auth_auth_pb.AuthenticateResponse>;
    responseDeserialize: grpc.deserialize<auth_auth_pb.AuthenticateResponse>;
}
interface IAPIService_IAuthorize extends grpc.MethodDefinition<auth_auth_pb.AuthorizeRequest, auth_auth_pb.AuthorizeResponse> {
    path: "/auth.API/Authorize";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<auth_auth_pb.AuthorizeRequest>;
    requestDeserialize: grpc.deserialize<auth_auth_pb.AuthorizeRequest>;
    responseSerialize: grpc.serialize<auth_auth_pb.AuthorizeResponse>;
    responseDeserialize: grpc.deserialize<auth_auth_pb.AuthorizeResponse>;
}
interface IAPIService_IGetPermissions extends grpc.MethodDefinition<auth_auth_pb.GetPermissionsRequest, auth_auth_pb.GetPermissionsResponse> {
    path: "/auth.API/GetPermissions";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<auth_auth_pb.GetPermissionsRequest>;
    requestDeserialize: grpc.deserialize<auth_auth_pb.GetPermissionsRequest>;
    responseSerialize: grpc.serialize<auth_auth_pb.GetPermissionsResponse>;
    responseDeserialize: grpc.deserialize<auth_auth_pb.GetPermissionsResponse>;
}
interface IAPIService_IGetPermissionsForPrincipal extends grpc.MethodDefinition<auth_auth_pb.GetPermissionsForPrincipalRequest, auth_auth_pb.GetPermissionsResponse> {
    path: "/auth.API/GetPermissionsForPrincipal";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<auth_auth_pb.GetPermissionsForPrincipalRequest>;
    requestDeserialize: grpc.deserialize<auth_auth_pb.GetPermissionsForPrincipalRequest>;
    responseSerialize: grpc.serialize<auth_auth_pb.GetPermissionsResponse>;
    responseDeserialize: grpc.deserialize<auth_auth_pb.GetPermissionsResponse>;
}
interface IAPIService_IWhoAmI extends grpc.MethodDefinition<auth_auth_pb.WhoAmIRequest, auth_auth_pb.WhoAmIResponse> {
    path: "/auth.API/WhoAmI";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<auth_auth_pb.WhoAmIRequest>;
    requestDeserialize: grpc.deserialize<auth_auth_pb.WhoAmIRequest>;
    responseSerialize: grpc.serialize<auth_auth_pb.WhoAmIResponse>;
    responseDeserialize: grpc.deserialize<auth_auth_pb.WhoAmIResponse>;
}
interface IAPIService_IModifyRoleBinding extends grpc.MethodDefinition<auth_auth_pb.ModifyRoleBindingRequest, auth_auth_pb.ModifyRoleBindingResponse> {
    path: "/auth.API/ModifyRoleBinding";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<auth_auth_pb.ModifyRoleBindingRequest>;
    requestDeserialize: grpc.deserialize<auth_auth_pb.ModifyRoleBindingRequest>;
    responseSerialize: grpc.serialize<auth_auth_pb.ModifyRoleBindingResponse>;
    responseDeserialize: grpc.deserialize<auth_auth_pb.ModifyRoleBindingResponse>;
}
interface IAPIService_IGetRoleBinding extends grpc.MethodDefinition<auth_auth_pb.GetRoleBindingRequest, auth_auth_pb.GetRoleBindingResponse> {
    path: "/auth.API/GetRoleBinding";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<auth_auth_pb.GetRoleBindingRequest>;
    requestDeserialize: grpc.deserialize<auth_auth_pb.GetRoleBindingRequest>;
    responseSerialize: grpc.serialize<auth_auth_pb.GetRoleBindingResponse>;
    responseDeserialize: grpc.deserialize<auth_auth_pb.GetRoleBindingResponse>;
}
interface IAPIService_IGetOIDCLogin extends grpc.MethodDefinition<auth_auth_pb.GetOIDCLoginRequest, auth_auth_pb.GetOIDCLoginResponse> {
    path: "/auth.API/GetOIDCLogin";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<auth_auth_pb.GetOIDCLoginRequest>;
    requestDeserialize: grpc.deserialize<auth_auth_pb.GetOIDCLoginRequest>;
    responseSerialize: grpc.serialize<auth_auth_pb.GetOIDCLoginResponse>;
    responseDeserialize: grpc.deserialize<auth_auth_pb.GetOIDCLoginResponse>;
}
interface IAPIService_IGetRobotToken extends grpc.MethodDefinition<auth_auth_pb.GetRobotTokenRequest, auth_auth_pb.GetRobotTokenResponse> {
    path: "/auth.API/GetRobotToken";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<auth_auth_pb.GetRobotTokenRequest>;
    requestDeserialize: grpc.deserialize<auth_auth_pb.GetRobotTokenRequest>;
    responseSerialize: grpc.serialize<auth_auth_pb.GetRobotTokenResponse>;
    responseDeserialize: grpc.deserialize<auth_auth_pb.GetRobotTokenResponse>;
}
interface IAPIService_IRevokeAuthToken extends grpc.MethodDefinition<auth_auth_pb.RevokeAuthTokenRequest, auth_auth_pb.RevokeAuthTokenResponse> {
    path: "/auth.API/RevokeAuthToken";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<auth_auth_pb.RevokeAuthTokenRequest>;
    requestDeserialize: grpc.deserialize<auth_auth_pb.RevokeAuthTokenRequest>;
    responseSerialize: grpc.serialize<auth_auth_pb.RevokeAuthTokenResponse>;
    responseDeserialize: grpc.deserialize<auth_auth_pb.RevokeAuthTokenResponse>;
}
interface IAPIService_IRevokeAuthTokensForUser extends grpc.MethodDefinition<auth_auth_pb.RevokeAuthTokensForUserRequest, auth_auth_pb.RevokeAuthTokensForUserResponse> {
    path: "/auth.API/RevokeAuthTokensForUser";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<auth_auth_pb.RevokeAuthTokensForUserRequest>;
    requestDeserialize: grpc.deserialize<auth_auth_pb.RevokeAuthTokensForUserRequest>;
    responseSerialize: grpc.serialize<auth_auth_pb.RevokeAuthTokensForUserResponse>;
    responseDeserialize: grpc.deserialize<auth_auth_pb.RevokeAuthTokensForUserResponse>;
}
interface IAPIService_ISetGroupsForUser extends grpc.MethodDefinition<auth_auth_pb.SetGroupsForUserRequest, auth_auth_pb.SetGroupsForUserResponse> {
    path: "/auth.API/SetGroupsForUser";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<auth_auth_pb.SetGroupsForUserRequest>;
    requestDeserialize: grpc.deserialize<auth_auth_pb.SetGroupsForUserRequest>;
    responseSerialize: grpc.serialize<auth_auth_pb.SetGroupsForUserResponse>;
    responseDeserialize: grpc.deserialize<auth_auth_pb.SetGroupsForUserResponse>;
}
interface IAPIService_IModifyMembers extends grpc.MethodDefinition<auth_auth_pb.ModifyMembersRequest, auth_auth_pb.ModifyMembersResponse> {
    path: "/auth.API/ModifyMembers";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<auth_auth_pb.ModifyMembersRequest>;
    requestDeserialize: grpc.deserialize<auth_auth_pb.ModifyMembersRequest>;
    responseSerialize: grpc.serialize<auth_auth_pb.ModifyMembersResponse>;
    responseDeserialize: grpc.deserialize<auth_auth_pb.ModifyMembersResponse>;
}
interface IAPIService_IGetGroups extends grpc.MethodDefinition<auth_auth_pb.GetGroupsRequest, auth_auth_pb.GetGroupsResponse> {
    path: "/auth.API/GetGroups";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<auth_auth_pb.GetGroupsRequest>;
    requestDeserialize: grpc.deserialize<auth_auth_pb.GetGroupsRequest>;
    responseSerialize: grpc.serialize<auth_auth_pb.GetGroupsResponse>;
    responseDeserialize: grpc.deserialize<auth_auth_pb.GetGroupsResponse>;
}
interface IAPIService_IGetGroupsForPrincipal extends grpc.MethodDefinition<auth_auth_pb.GetGroupsForPrincipalRequest, auth_auth_pb.GetGroupsResponse> {
    path: "/auth.API/GetGroupsForPrincipal";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<auth_auth_pb.GetGroupsForPrincipalRequest>;
    requestDeserialize: grpc.deserialize<auth_auth_pb.GetGroupsForPrincipalRequest>;
    responseSerialize: grpc.serialize<auth_auth_pb.GetGroupsResponse>;
    responseDeserialize: grpc.deserialize<auth_auth_pb.GetGroupsResponse>;
}
interface IAPIService_IGetUsers extends grpc.MethodDefinition<auth_auth_pb.GetUsersRequest, auth_auth_pb.GetUsersResponse> {
    path: "/auth.API/GetUsers";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<auth_auth_pb.GetUsersRequest>;
    requestDeserialize: grpc.deserialize<auth_auth_pb.GetUsersRequest>;
    responseSerialize: grpc.serialize<auth_auth_pb.GetUsersResponse>;
    responseDeserialize: grpc.deserialize<auth_auth_pb.GetUsersResponse>;
}
interface IAPIService_IExtractAuthTokens extends grpc.MethodDefinition<auth_auth_pb.ExtractAuthTokensRequest, auth_auth_pb.ExtractAuthTokensResponse> {
    path: "/auth.API/ExtractAuthTokens";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<auth_auth_pb.ExtractAuthTokensRequest>;
    requestDeserialize: grpc.deserialize<auth_auth_pb.ExtractAuthTokensRequest>;
    responseSerialize: grpc.serialize<auth_auth_pb.ExtractAuthTokensResponse>;
    responseDeserialize: grpc.deserialize<auth_auth_pb.ExtractAuthTokensResponse>;
}
interface IAPIService_IRestoreAuthToken extends grpc.MethodDefinition<auth_auth_pb.RestoreAuthTokenRequest, auth_auth_pb.RestoreAuthTokenResponse> {
    path: "/auth.API/RestoreAuthToken";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<auth_auth_pb.RestoreAuthTokenRequest>;
    requestDeserialize: grpc.deserialize<auth_auth_pb.RestoreAuthTokenRequest>;
    responseSerialize: grpc.serialize<auth_auth_pb.RestoreAuthTokenResponse>;
    responseDeserialize: grpc.deserialize<auth_auth_pb.RestoreAuthTokenResponse>;
}
interface IAPIService_IDeleteExpiredAuthTokens extends grpc.MethodDefinition<auth_auth_pb.DeleteExpiredAuthTokensRequest, auth_auth_pb.DeleteExpiredAuthTokensResponse> {
    path: "/auth.API/DeleteExpiredAuthTokens";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<auth_auth_pb.DeleteExpiredAuthTokensRequest>;
    requestDeserialize: grpc.deserialize<auth_auth_pb.DeleteExpiredAuthTokensRequest>;
    responseSerialize: grpc.serialize<auth_auth_pb.DeleteExpiredAuthTokensResponse>;
    responseDeserialize: grpc.deserialize<auth_auth_pb.DeleteExpiredAuthTokensResponse>;
}
interface IAPIService_IRotateRootToken extends grpc.MethodDefinition<auth_auth_pb.RotateRootTokenRequest, auth_auth_pb.RotateRootTokenResponse> {
    path: "/auth.API/RotateRootToken";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<auth_auth_pb.RotateRootTokenRequest>;
    requestDeserialize: grpc.deserialize<auth_auth_pb.RotateRootTokenRequest>;
    responseSerialize: grpc.serialize<auth_auth_pb.RotateRootTokenResponse>;
    responseDeserialize: grpc.deserialize<auth_auth_pb.RotateRootTokenResponse>;
}

export const APIService: IAPIService;

export interface IAPIServer {
    activate: grpc.handleUnaryCall<auth_auth_pb.ActivateRequest, auth_auth_pb.ActivateResponse>;
    deactivate: grpc.handleUnaryCall<auth_auth_pb.DeactivateRequest, auth_auth_pb.DeactivateResponse>;
    getConfiguration: grpc.handleUnaryCall<auth_auth_pb.GetConfigurationRequest, auth_auth_pb.GetConfigurationResponse>;
    setConfiguration: grpc.handleUnaryCall<auth_auth_pb.SetConfigurationRequest, auth_auth_pb.SetConfigurationResponse>;
    authenticate: grpc.handleUnaryCall<auth_auth_pb.AuthenticateRequest, auth_auth_pb.AuthenticateResponse>;
    authorize: grpc.handleUnaryCall<auth_auth_pb.AuthorizeRequest, auth_auth_pb.AuthorizeResponse>;
    getPermissions: grpc.handleUnaryCall<auth_auth_pb.GetPermissionsRequest, auth_auth_pb.GetPermissionsResponse>;
    getPermissionsForPrincipal: grpc.handleUnaryCall<auth_auth_pb.GetPermissionsForPrincipalRequest, auth_auth_pb.GetPermissionsResponse>;
    whoAmI: grpc.handleUnaryCall<auth_auth_pb.WhoAmIRequest, auth_auth_pb.WhoAmIResponse>;
    modifyRoleBinding: grpc.handleUnaryCall<auth_auth_pb.ModifyRoleBindingRequest, auth_auth_pb.ModifyRoleBindingResponse>;
    getRoleBinding: grpc.handleUnaryCall<auth_auth_pb.GetRoleBindingRequest, auth_auth_pb.GetRoleBindingResponse>;
    getOIDCLogin: grpc.handleUnaryCall<auth_auth_pb.GetOIDCLoginRequest, auth_auth_pb.GetOIDCLoginResponse>;
    getRobotToken: grpc.handleUnaryCall<auth_auth_pb.GetRobotTokenRequest, auth_auth_pb.GetRobotTokenResponse>;
    revokeAuthToken: grpc.handleUnaryCall<auth_auth_pb.RevokeAuthTokenRequest, auth_auth_pb.RevokeAuthTokenResponse>;
    revokeAuthTokensForUser: grpc.handleUnaryCall<auth_auth_pb.RevokeAuthTokensForUserRequest, auth_auth_pb.RevokeAuthTokensForUserResponse>;
    setGroupsForUser: grpc.handleUnaryCall<auth_auth_pb.SetGroupsForUserRequest, auth_auth_pb.SetGroupsForUserResponse>;
    modifyMembers: grpc.handleUnaryCall<auth_auth_pb.ModifyMembersRequest, auth_auth_pb.ModifyMembersResponse>;
    getGroups: grpc.handleUnaryCall<auth_auth_pb.GetGroupsRequest, auth_auth_pb.GetGroupsResponse>;
    getGroupsForPrincipal: grpc.handleUnaryCall<auth_auth_pb.GetGroupsForPrincipalRequest, auth_auth_pb.GetGroupsResponse>;
    getUsers: grpc.handleUnaryCall<auth_auth_pb.GetUsersRequest, auth_auth_pb.GetUsersResponse>;
    extractAuthTokens: grpc.handleUnaryCall<auth_auth_pb.ExtractAuthTokensRequest, auth_auth_pb.ExtractAuthTokensResponse>;
    restoreAuthToken: grpc.handleUnaryCall<auth_auth_pb.RestoreAuthTokenRequest, auth_auth_pb.RestoreAuthTokenResponse>;
    deleteExpiredAuthTokens: grpc.handleUnaryCall<auth_auth_pb.DeleteExpiredAuthTokensRequest, auth_auth_pb.DeleteExpiredAuthTokensResponse>;
    rotateRootToken: grpc.handleUnaryCall<auth_auth_pb.RotateRootTokenRequest, auth_auth_pb.RotateRootTokenResponse>;
}

export interface IAPIClient {
    activate(request: auth_auth_pb.ActivateRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.ActivateResponse) => void): grpc.ClientUnaryCall;
    activate(request: auth_auth_pb.ActivateRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.ActivateResponse) => void): grpc.ClientUnaryCall;
    activate(request: auth_auth_pb.ActivateRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.ActivateResponse) => void): grpc.ClientUnaryCall;
    deactivate(request: auth_auth_pb.DeactivateRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.DeactivateResponse) => void): grpc.ClientUnaryCall;
    deactivate(request: auth_auth_pb.DeactivateRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.DeactivateResponse) => void): grpc.ClientUnaryCall;
    deactivate(request: auth_auth_pb.DeactivateRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.DeactivateResponse) => void): grpc.ClientUnaryCall;
    getConfiguration(request: auth_auth_pb.GetConfigurationRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetConfigurationResponse) => void): grpc.ClientUnaryCall;
    getConfiguration(request: auth_auth_pb.GetConfigurationRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetConfigurationResponse) => void): grpc.ClientUnaryCall;
    getConfiguration(request: auth_auth_pb.GetConfigurationRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetConfigurationResponse) => void): grpc.ClientUnaryCall;
    setConfiguration(request: auth_auth_pb.SetConfigurationRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.SetConfigurationResponse) => void): grpc.ClientUnaryCall;
    setConfiguration(request: auth_auth_pb.SetConfigurationRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.SetConfigurationResponse) => void): grpc.ClientUnaryCall;
    setConfiguration(request: auth_auth_pb.SetConfigurationRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.SetConfigurationResponse) => void): grpc.ClientUnaryCall;
    authenticate(request: auth_auth_pb.AuthenticateRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.AuthenticateResponse) => void): grpc.ClientUnaryCall;
    authenticate(request: auth_auth_pb.AuthenticateRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.AuthenticateResponse) => void): grpc.ClientUnaryCall;
    authenticate(request: auth_auth_pb.AuthenticateRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.AuthenticateResponse) => void): grpc.ClientUnaryCall;
    authorize(request: auth_auth_pb.AuthorizeRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.AuthorizeResponse) => void): grpc.ClientUnaryCall;
    authorize(request: auth_auth_pb.AuthorizeRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.AuthorizeResponse) => void): grpc.ClientUnaryCall;
    authorize(request: auth_auth_pb.AuthorizeRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.AuthorizeResponse) => void): grpc.ClientUnaryCall;
    getPermissions(request: auth_auth_pb.GetPermissionsRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetPermissionsResponse) => void): grpc.ClientUnaryCall;
    getPermissions(request: auth_auth_pb.GetPermissionsRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetPermissionsResponse) => void): grpc.ClientUnaryCall;
    getPermissions(request: auth_auth_pb.GetPermissionsRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetPermissionsResponse) => void): grpc.ClientUnaryCall;
    getPermissionsForPrincipal(request: auth_auth_pb.GetPermissionsForPrincipalRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetPermissionsResponse) => void): grpc.ClientUnaryCall;
    getPermissionsForPrincipal(request: auth_auth_pb.GetPermissionsForPrincipalRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetPermissionsResponse) => void): grpc.ClientUnaryCall;
    getPermissionsForPrincipal(request: auth_auth_pb.GetPermissionsForPrincipalRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetPermissionsResponse) => void): grpc.ClientUnaryCall;
    whoAmI(request: auth_auth_pb.WhoAmIRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.WhoAmIResponse) => void): grpc.ClientUnaryCall;
    whoAmI(request: auth_auth_pb.WhoAmIRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.WhoAmIResponse) => void): grpc.ClientUnaryCall;
    whoAmI(request: auth_auth_pb.WhoAmIRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.WhoAmIResponse) => void): grpc.ClientUnaryCall;
    modifyRoleBinding(request: auth_auth_pb.ModifyRoleBindingRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.ModifyRoleBindingResponse) => void): grpc.ClientUnaryCall;
    modifyRoleBinding(request: auth_auth_pb.ModifyRoleBindingRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.ModifyRoleBindingResponse) => void): grpc.ClientUnaryCall;
    modifyRoleBinding(request: auth_auth_pb.ModifyRoleBindingRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.ModifyRoleBindingResponse) => void): grpc.ClientUnaryCall;
    getRoleBinding(request: auth_auth_pb.GetRoleBindingRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetRoleBindingResponse) => void): grpc.ClientUnaryCall;
    getRoleBinding(request: auth_auth_pb.GetRoleBindingRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetRoleBindingResponse) => void): grpc.ClientUnaryCall;
    getRoleBinding(request: auth_auth_pb.GetRoleBindingRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetRoleBindingResponse) => void): grpc.ClientUnaryCall;
    getOIDCLogin(request: auth_auth_pb.GetOIDCLoginRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetOIDCLoginResponse) => void): grpc.ClientUnaryCall;
    getOIDCLogin(request: auth_auth_pb.GetOIDCLoginRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetOIDCLoginResponse) => void): grpc.ClientUnaryCall;
    getOIDCLogin(request: auth_auth_pb.GetOIDCLoginRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetOIDCLoginResponse) => void): grpc.ClientUnaryCall;
    getRobotToken(request: auth_auth_pb.GetRobotTokenRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetRobotTokenResponse) => void): grpc.ClientUnaryCall;
    getRobotToken(request: auth_auth_pb.GetRobotTokenRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetRobotTokenResponse) => void): grpc.ClientUnaryCall;
    getRobotToken(request: auth_auth_pb.GetRobotTokenRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetRobotTokenResponse) => void): grpc.ClientUnaryCall;
    revokeAuthToken(request: auth_auth_pb.RevokeAuthTokenRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.RevokeAuthTokenResponse) => void): grpc.ClientUnaryCall;
    revokeAuthToken(request: auth_auth_pb.RevokeAuthTokenRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.RevokeAuthTokenResponse) => void): grpc.ClientUnaryCall;
    revokeAuthToken(request: auth_auth_pb.RevokeAuthTokenRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.RevokeAuthTokenResponse) => void): grpc.ClientUnaryCall;
    revokeAuthTokensForUser(request: auth_auth_pb.RevokeAuthTokensForUserRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.RevokeAuthTokensForUserResponse) => void): grpc.ClientUnaryCall;
    revokeAuthTokensForUser(request: auth_auth_pb.RevokeAuthTokensForUserRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.RevokeAuthTokensForUserResponse) => void): grpc.ClientUnaryCall;
    revokeAuthTokensForUser(request: auth_auth_pb.RevokeAuthTokensForUserRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.RevokeAuthTokensForUserResponse) => void): grpc.ClientUnaryCall;
    setGroupsForUser(request: auth_auth_pb.SetGroupsForUserRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.SetGroupsForUserResponse) => void): grpc.ClientUnaryCall;
    setGroupsForUser(request: auth_auth_pb.SetGroupsForUserRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.SetGroupsForUserResponse) => void): grpc.ClientUnaryCall;
    setGroupsForUser(request: auth_auth_pb.SetGroupsForUserRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.SetGroupsForUserResponse) => void): grpc.ClientUnaryCall;
    modifyMembers(request: auth_auth_pb.ModifyMembersRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.ModifyMembersResponse) => void): grpc.ClientUnaryCall;
    modifyMembers(request: auth_auth_pb.ModifyMembersRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.ModifyMembersResponse) => void): grpc.ClientUnaryCall;
    modifyMembers(request: auth_auth_pb.ModifyMembersRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.ModifyMembersResponse) => void): grpc.ClientUnaryCall;
    getGroups(request: auth_auth_pb.GetGroupsRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetGroupsResponse) => void): grpc.ClientUnaryCall;
    getGroups(request: auth_auth_pb.GetGroupsRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetGroupsResponse) => void): grpc.ClientUnaryCall;
    getGroups(request: auth_auth_pb.GetGroupsRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetGroupsResponse) => void): grpc.ClientUnaryCall;
    getGroupsForPrincipal(request: auth_auth_pb.GetGroupsForPrincipalRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetGroupsResponse) => void): grpc.ClientUnaryCall;
    getGroupsForPrincipal(request: auth_auth_pb.GetGroupsForPrincipalRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetGroupsResponse) => void): grpc.ClientUnaryCall;
    getGroupsForPrincipal(request: auth_auth_pb.GetGroupsForPrincipalRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetGroupsResponse) => void): grpc.ClientUnaryCall;
    getUsers(request: auth_auth_pb.GetUsersRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetUsersResponse) => void): grpc.ClientUnaryCall;
    getUsers(request: auth_auth_pb.GetUsersRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetUsersResponse) => void): grpc.ClientUnaryCall;
    getUsers(request: auth_auth_pb.GetUsersRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetUsersResponse) => void): grpc.ClientUnaryCall;
    extractAuthTokens(request: auth_auth_pb.ExtractAuthTokensRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.ExtractAuthTokensResponse) => void): grpc.ClientUnaryCall;
    extractAuthTokens(request: auth_auth_pb.ExtractAuthTokensRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.ExtractAuthTokensResponse) => void): grpc.ClientUnaryCall;
    extractAuthTokens(request: auth_auth_pb.ExtractAuthTokensRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.ExtractAuthTokensResponse) => void): grpc.ClientUnaryCall;
    restoreAuthToken(request: auth_auth_pb.RestoreAuthTokenRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.RestoreAuthTokenResponse) => void): grpc.ClientUnaryCall;
    restoreAuthToken(request: auth_auth_pb.RestoreAuthTokenRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.RestoreAuthTokenResponse) => void): grpc.ClientUnaryCall;
    restoreAuthToken(request: auth_auth_pb.RestoreAuthTokenRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.RestoreAuthTokenResponse) => void): grpc.ClientUnaryCall;
    deleteExpiredAuthTokens(request: auth_auth_pb.DeleteExpiredAuthTokensRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.DeleteExpiredAuthTokensResponse) => void): grpc.ClientUnaryCall;
    deleteExpiredAuthTokens(request: auth_auth_pb.DeleteExpiredAuthTokensRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.DeleteExpiredAuthTokensResponse) => void): grpc.ClientUnaryCall;
    deleteExpiredAuthTokens(request: auth_auth_pb.DeleteExpiredAuthTokensRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.DeleteExpiredAuthTokensResponse) => void): grpc.ClientUnaryCall;
    rotateRootToken(request: auth_auth_pb.RotateRootTokenRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.RotateRootTokenResponse) => void): grpc.ClientUnaryCall;
    rotateRootToken(request: auth_auth_pb.RotateRootTokenRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.RotateRootTokenResponse) => void): grpc.ClientUnaryCall;
    rotateRootToken(request: auth_auth_pb.RotateRootTokenRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.RotateRootTokenResponse) => void): grpc.ClientUnaryCall;
}

export class APIClient extends grpc.Client implements IAPIClient {
    constructor(address: string, credentials: grpc.ChannelCredentials, options?: Partial<grpc.ClientOptions>);
    public activate(request: auth_auth_pb.ActivateRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.ActivateResponse) => void): grpc.ClientUnaryCall;
    public activate(request: auth_auth_pb.ActivateRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.ActivateResponse) => void): grpc.ClientUnaryCall;
    public activate(request: auth_auth_pb.ActivateRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.ActivateResponse) => void): grpc.ClientUnaryCall;
    public deactivate(request: auth_auth_pb.DeactivateRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.DeactivateResponse) => void): grpc.ClientUnaryCall;
    public deactivate(request: auth_auth_pb.DeactivateRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.DeactivateResponse) => void): grpc.ClientUnaryCall;
    public deactivate(request: auth_auth_pb.DeactivateRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.DeactivateResponse) => void): grpc.ClientUnaryCall;
    public getConfiguration(request: auth_auth_pb.GetConfigurationRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetConfigurationResponse) => void): grpc.ClientUnaryCall;
    public getConfiguration(request: auth_auth_pb.GetConfigurationRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetConfigurationResponse) => void): grpc.ClientUnaryCall;
    public getConfiguration(request: auth_auth_pb.GetConfigurationRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetConfigurationResponse) => void): grpc.ClientUnaryCall;
    public setConfiguration(request: auth_auth_pb.SetConfigurationRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.SetConfigurationResponse) => void): grpc.ClientUnaryCall;
    public setConfiguration(request: auth_auth_pb.SetConfigurationRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.SetConfigurationResponse) => void): grpc.ClientUnaryCall;
    public setConfiguration(request: auth_auth_pb.SetConfigurationRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.SetConfigurationResponse) => void): grpc.ClientUnaryCall;
    public authenticate(request: auth_auth_pb.AuthenticateRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.AuthenticateResponse) => void): grpc.ClientUnaryCall;
    public authenticate(request: auth_auth_pb.AuthenticateRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.AuthenticateResponse) => void): grpc.ClientUnaryCall;
    public authenticate(request: auth_auth_pb.AuthenticateRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.AuthenticateResponse) => void): grpc.ClientUnaryCall;
    public authorize(request: auth_auth_pb.AuthorizeRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.AuthorizeResponse) => void): grpc.ClientUnaryCall;
    public authorize(request: auth_auth_pb.AuthorizeRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.AuthorizeResponse) => void): grpc.ClientUnaryCall;
    public authorize(request: auth_auth_pb.AuthorizeRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.AuthorizeResponse) => void): grpc.ClientUnaryCall;
    public getPermissions(request: auth_auth_pb.GetPermissionsRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetPermissionsResponse) => void): grpc.ClientUnaryCall;
    public getPermissions(request: auth_auth_pb.GetPermissionsRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetPermissionsResponse) => void): grpc.ClientUnaryCall;
    public getPermissions(request: auth_auth_pb.GetPermissionsRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetPermissionsResponse) => void): grpc.ClientUnaryCall;
    public getPermissionsForPrincipal(request: auth_auth_pb.GetPermissionsForPrincipalRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetPermissionsResponse) => void): grpc.ClientUnaryCall;
    public getPermissionsForPrincipal(request: auth_auth_pb.GetPermissionsForPrincipalRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetPermissionsResponse) => void): grpc.ClientUnaryCall;
    public getPermissionsForPrincipal(request: auth_auth_pb.GetPermissionsForPrincipalRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetPermissionsResponse) => void): grpc.ClientUnaryCall;
    public whoAmI(request: auth_auth_pb.WhoAmIRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.WhoAmIResponse) => void): grpc.ClientUnaryCall;
    public whoAmI(request: auth_auth_pb.WhoAmIRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.WhoAmIResponse) => void): grpc.ClientUnaryCall;
    public whoAmI(request: auth_auth_pb.WhoAmIRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.WhoAmIResponse) => void): grpc.ClientUnaryCall;
    public modifyRoleBinding(request: auth_auth_pb.ModifyRoleBindingRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.ModifyRoleBindingResponse) => void): grpc.ClientUnaryCall;
    public modifyRoleBinding(request: auth_auth_pb.ModifyRoleBindingRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.ModifyRoleBindingResponse) => void): grpc.ClientUnaryCall;
    public modifyRoleBinding(request: auth_auth_pb.ModifyRoleBindingRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.ModifyRoleBindingResponse) => void): grpc.ClientUnaryCall;
    public getRoleBinding(request: auth_auth_pb.GetRoleBindingRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetRoleBindingResponse) => void): grpc.ClientUnaryCall;
    public getRoleBinding(request: auth_auth_pb.GetRoleBindingRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetRoleBindingResponse) => void): grpc.ClientUnaryCall;
    public getRoleBinding(request: auth_auth_pb.GetRoleBindingRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetRoleBindingResponse) => void): grpc.ClientUnaryCall;
    public getOIDCLogin(request: auth_auth_pb.GetOIDCLoginRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetOIDCLoginResponse) => void): grpc.ClientUnaryCall;
    public getOIDCLogin(request: auth_auth_pb.GetOIDCLoginRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetOIDCLoginResponse) => void): grpc.ClientUnaryCall;
    public getOIDCLogin(request: auth_auth_pb.GetOIDCLoginRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetOIDCLoginResponse) => void): grpc.ClientUnaryCall;
    public getRobotToken(request: auth_auth_pb.GetRobotTokenRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetRobotTokenResponse) => void): grpc.ClientUnaryCall;
    public getRobotToken(request: auth_auth_pb.GetRobotTokenRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetRobotTokenResponse) => void): grpc.ClientUnaryCall;
    public getRobotToken(request: auth_auth_pb.GetRobotTokenRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetRobotTokenResponse) => void): grpc.ClientUnaryCall;
    public revokeAuthToken(request: auth_auth_pb.RevokeAuthTokenRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.RevokeAuthTokenResponse) => void): grpc.ClientUnaryCall;
    public revokeAuthToken(request: auth_auth_pb.RevokeAuthTokenRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.RevokeAuthTokenResponse) => void): grpc.ClientUnaryCall;
    public revokeAuthToken(request: auth_auth_pb.RevokeAuthTokenRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.RevokeAuthTokenResponse) => void): grpc.ClientUnaryCall;
    public revokeAuthTokensForUser(request: auth_auth_pb.RevokeAuthTokensForUserRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.RevokeAuthTokensForUserResponse) => void): grpc.ClientUnaryCall;
    public revokeAuthTokensForUser(request: auth_auth_pb.RevokeAuthTokensForUserRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.RevokeAuthTokensForUserResponse) => void): grpc.ClientUnaryCall;
    public revokeAuthTokensForUser(request: auth_auth_pb.RevokeAuthTokensForUserRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.RevokeAuthTokensForUserResponse) => void): grpc.ClientUnaryCall;
    public setGroupsForUser(request: auth_auth_pb.SetGroupsForUserRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.SetGroupsForUserResponse) => void): grpc.ClientUnaryCall;
    public setGroupsForUser(request: auth_auth_pb.SetGroupsForUserRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.SetGroupsForUserResponse) => void): grpc.ClientUnaryCall;
    public setGroupsForUser(request: auth_auth_pb.SetGroupsForUserRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.SetGroupsForUserResponse) => void): grpc.ClientUnaryCall;
    public modifyMembers(request: auth_auth_pb.ModifyMembersRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.ModifyMembersResponse) => void): grpc.ClientUnaryCall;
    public modifyMembers(request: auth_auth_pb.ModifyMembersRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.ModifyMembersResponse) => void): grpc.ClientUnaryCall;
    public modifyMembers(request: auth_auth_pb.ModifyMembersRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.ModifyMembersResponse) => void): grpc.ClientUnaryCall;
    public getGroups(request: auth_auth_pb.GetGroupsRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetGroupsResponse) => void): grpc.ClientUnaryCall;
    public getGroups(request: auth_auth_pb.GetGroupsRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetGroupsResponse) => void): grpc.ClientUnaryCall;
    public getGroups(request: auth_auth_pb.GetGroupsRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetGroupsResponse) => void): grpc.ClientUnaryCall;
    public getGroupsForPrincipal(request: auth_auth_pb.GetGroupsForPrincipalRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetGroupsResponse) => void): grpc.ClientUnaryCall;
    public getGroupsForPrincipal(request: auth_auth_pb.GetGroupsForPrincipalRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetGroupsResponse) => void): grpc.ClientUnaryCall;
    public getGroupsForPrincipal(request: auth_auth_pb.GetGroupsForPrincipalRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetGroupsResponse) => void): grpc.ClientUnaryCall;
    public getUsers(request: auth_auth_pb.GetUsersRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetUsersResponse) => void): grpc.ClientUnaryCall;
    public getUsers(request: auth_auth_pb.GetUsersRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetUsersResponse) => void): grpc.ClientUnaryCall;
    public getUsers(request: auth_auth_pb.GetUsersRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetUsersResponse) => void): grpc.ClientUnaryCall;
    public extractAuthTokens(request: auth_auth_pb.ExtractAuthTokensRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.ExtractAuthTokensResponse) => void): grpc.ClientUnaryCall;
    public extractAuthTokens(request: auth_auth_pb.ExtractAuthTokensRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.ExtractAuthTokensResponse) => void): grpc.ClientUnaryCall;
    public extractAuthTokens(request: auth_auth_pb.ExtractAuthTokensRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.ExtractAuthTokensResponse) => void): grpc.ClientUnaryCall;
    public restoreAuthToken(request: auth_auth_pb.RestoreAuthTokenRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.RestoreAuthTokenResponse) => void): grpc.ClientUnaryCall;
    public restoreAuthToken(request: auth_auth_pb.RestoreAuthTokenRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.RestoreAuthTokenResponse) => void): grpc.ClientUnaryCall;
    public restoreAuthToken(request: auth_auth_pb.RestoreAuthTokenRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.RestoreAuthTokenResponse) => void): grpc.ClientUnaryCall;
    public deleteExpiredAuthTokens(request: auth_auth_pb.DeleteExpiredAuthTokensRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.DeleteExpiredAuthTokensResponse) => void): grpc.ClientUnaryCall;
    public deleteExpiredAuthTokens(request: auth_auth_pb.DeleteExpiredAuthTokensRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.DeleteExpiredAuthTokensResponse) => void): grpc.ClientUnaryCall;
    public deleteExpiredAuthTokens(request: auth_auth_pb.DeleteExpiredAuthTokensRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.DeleteExpiredAuthTokensResponse) => void): grpc.ClientUnaryCall;
    public rotateRootToken(request: auth_auth_pb.RotateRootTokenRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.RotateRootTokenResponse) => void): grpc.ClientUnaryCall;
    public rotateRootToken(request: auth_auth_pb.RotateRootTokenRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.RotateRootTokenResponse) => void): grpc.ClientUnaryCall;
    public rotateRootToken(request: auth_auth_pb.RotateRootTokenRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.RotateRootTokenResponse) => void): grpc.ClientUnaryCall;
}
