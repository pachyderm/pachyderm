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
    getAdmins: IAPIService_IGetAdmins;
    modifyAdmins: IAPIService_IModifyAdmins;
    getClusterRoleBindings: IAPIService_IGetClusterRoleBindings;
    modifyClusterRoleBinding: IAPIService_IModifyClusterRoleBinding;
    authenticate: IAPIService_IAuthenticate;
    authorize: IAPIService_IAuthorize;
    whoAmI: IAPIService_IWhoAmI;
    getScope: IAPIService_IGetScope;
    setScope: IAPIService_ISetScope;
    getACL: IAPIService_IGetACL;
    setACL: IAPIService_ISetACL;
    getOIDCLogin: IAPIService_IGetOIDCLogin;
    getAuthToken: IAPIService_IGetAuthToken;
    extendAuthToken: IAPIService_IExtendAuthToken;
    revokeAuthToken: IAPIService_IRevokeAuthToken;
    setGroupsForUser: IAPIService_ISetGroupsForUser;
    modifyMembers: IAPIService_IModifyMembers;
    getGroups: IAPIService_IGetGroups;
    getUsers: IAPIService_IGetUsers;
    extractAuthTokens: IAPIService_IExtractAuthTokens;
    restoreAuthToken: IAPIService_IRestoreAuthToken;
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
interface IAPIService_IGetAdmins extends grpc.MethodDefinition<auth_auth_pb.GetAdminsRequest, auth_auth_pb.GetAdminsResponse> {
    path: "/auth.API/GetAdmins";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<auth_auth_pb.GetAdminsRequest>;
    requestDeserialize: grpc.deserialize<auth_auth_pb.GetAdminsRequest>;
    responseSerialize: grpc.serialize<auth_auth_pb.GetAdminsResponse>;
    responseDeserialize: grpc.deserialize<auth_auth_pb.GetAdminsResponse>;
}
interface IAPIService_IModifyAdmins extends grpc.MethodDefinition<auth_auth_pb.ModifyAdminsRequest, auth_auth_pb.ModifyAdminsResponse> {
    path: "/auth.API/ModifyAdmins";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<auth_auth_pb.ModifyAdminsRequest>;
    requestDeserialize: grpc.deserialize<auth_auth_pb.ModifyAdminsRequest>;
    responseSerialize: grpc.serialize<auth_auth_pb.ModifyAdminsResponse>;
    responseDeserialize: grpc.deserialize<auth_auth_pb.ModifyAdminsResponse>;
}
interface IAPIService_IGetClusterRoleBindings extends grpc.MethodDefinition<auth_auth_pb.GetClusterRoleBindingsRequest, auth_auth_pb.GetClusterRoleBindingsResponse> {
    path: "/auth.API/GetClusterRoleBindings";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<auth_auth_pb.GetClusterRoleBindingsRequest>;
    requestDeserialize: grpc.deserialize<auth_auth_pb.GetClusterRoleBindingsRequest>;
    responseSerialize: grpc.serialize<auth_auth_pb.GetClusterRoleBindingsResponse>;
    responseDeserialize: grpc.deserialize<auth_auth_pb.GetClusterRoleBindingsResponse>;
}
interface IAPIService_IModifyClusterRoleBinding extends grpc.MethodDefinition<auth_auth_pb.ModifyClusterRoleBindingRequest, auth_auth_pb.ModifyClusterRoleBindingResponse> {
    path: "/auth.API/ModifyClusterRoleBinding";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<auth_auth_pb.ModifyClusterRoleBindingRequest>;
    requestDeserialize: grpc.deserialize<auth_auth_pb.ModifyClusterRoleBindingRequest>;
    responseSerialize: grpc.serialize<auth_auth_pb.ModifyClusterRoleBindingResponse>;
    responseDeserialize: grpc.deserialize<auth_auth_pb.ModifyClusterRoleBindingResponse>;
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
interface IAPIService_IWhoAmI extends grpc.MethodDefinition<auth_auth_pb.WhoAmIRequest, auth_auth_pb.WhoAmIResponse> {
    path: "/auth.API/WhoAmI";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<auth_auth_pb.WhoAmIRequest>;
    requestDeserialize: grpc.deserialize<auth_auth_pb.WhoAmIRequest>;
    responseSerialize: grpc.serialize<auth_auth_pb.WhoAmIResponse>;
    responseDeserialize: grpc.deserialize<auth_auth_pb.WhoAmIResponse>;
}
interface IAPIService_IGetScope extends grpc.MethodDefinition<auth_auth_pb.GetScopeRequest, auth_auth_pb.GetScopeResponse> {
    path: "/auth.API/GetScope";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<auth_auth_pb.GetScopeRequest>;
    requestDeserialize: grpc.deserialize<auth_auth_pb.GetScopeRequest>;
    responseSerialize: grpc.serialize<auth_auth_pb.GetScopeResponse>;
    responseDeserialize: grpc.deserialize<auth_auth_pb.GetScopeResponse>;
}
interface IAPIService_ISetScope extends grpc.MethodDefinition<auth_auth_pb.SetScopeRequest, auth_auth_pb.SetScopeResponse> {
    path: "/auth.API/SetScope";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<auth_auth_pb.SetScopeRequest>;
    requestDeserialize: grpc.deserialize<auth_auth_pb.SetScopeRequest>;
    responseSerialize: grpc.serialize<auth_auth_pb.SetScopeResponse>;
    responseDeserialize: grpc.deserialize<auth_auth_pb.SetScopeResponse>;
}
interface IAPIService_IGetACL extends grpc.MethodDefinition<auth_auth_pb.GetACLRequest, auth_auth_pb.GetACLResponse> {
    path: "/auth.API/GetACL";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<auth_auth_pb.GetACLRequest>;
    requestDeserialize: grpc.deserialize<auth_auth_pb.GetACLRequest>;
    responseSerialize: grpc.serialize<auth_auth_pb.GetACLResponse>;
    responseDeserialize: grpc.deserialize<auth_auth_pb.GetACLResponse>;
}
interface IAPIService_ISetACL extends grpc.MethodDefinition<auth_auth_pb.SetACLRequest, auth_auth_pb.SetACLResponse> {
    path: "/auth.API/SetACL";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<auth_auth_pb.SetACLRequest>;
    requestDeserialize: grpc.deserialize<auth_auth_pb.SetACLRequest>;
    responseSerialize: grpc.serialize<auth_auth_pb.SetACLResponse>;
    responseDeserialize: grpc.deserialize<auth_auth_pb.SetACLResponse>;
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
interface IAPIService_IGetAuthToken extends grpc.MethodDefinition<auth_auth_pb.GetAuthTokenRequest, auth_auth_pb.GetAuthTokenResponse> {
    path: "/auth.API/GetAuthToken";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<auth_auth_pb.GetAuthTokenRequest>;
    requestDeserialize: grpc.deserialize<auth_auth_pb.GetAuthTokenRequest>;
    responseSerialize: grpc.serialize<auth_auth_pb.GetAuthTokenResponse>;
    responseDeserialize: grpc.deserialize<auth_auth_pb.GetAuthTokenResponse>;
}
interface IAPIService_IExtendAuthToken extends grpc.MethodDefinition<auth_auth_pb.ExtendAuthTokenRequest, auth_auth_pb.ExtendAuthTokenResponse> {
    path: "/auth.API/ExtendAuthToken";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<auth_auth_pb.ExtendAuthTokenRequest>;
    requestDeserialize: grpc.deserialize<auth_auth_pb.ExtendAuthTokenRequest>;
    responseSerialize: grpc.serialize<auth_auth_pb.ExtendAuthTokenResponse>;
    responseDeserialize: grpc.deserialize<auth_auth_pb.ExtendAuthTokenResponse>;
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

export const APIService: IAPIService;

export interface IAPIServer {
    activate: grpc.handleUnaryCall<auth_auth_pb.ActivateRequest, auth_auth_pb.ActivateResponse>;
    deactivate: grpc.handleUnaryCall<auth_auth_pb.DeactivateRequest, auth_auth_pb.DeactivateResponse>;
    getConfiguration: grpc.handleUnaryCall<auth_auth_pb.GetConfigurationRequest, auth_auth_pb.GetConfigurationResponse>;
    setConfiguration: grpc.handleUnaryCall<auth_auth_pb.SetConfigurationRequest, auth_auth_pb.SetConfigurationResponse>;
    getAdmins: grpc.handleUnaryCall<auth_auth_pb.GetAdminsRequest, auth_auth_pb.GetAdminsResponse>;
    modifyAdmins: grpc.handleUnaryCall<auth_auth_pb.ModifyAdminsRequest, auth_auth_pb.ModifyAdminsResponse>;
    getClusterRoleBindings: grpc.handleUnaryCall<auth_auth_pb.GetClusterRoleBindingsRequest, auth_auth_pb.GetClusterRoleBindingsResponse>;
    modifyClusterRoleBinding: grpc.handleUnaryCall<auth_auth_pb.ModifyClusterRoleBindingRequest, auth_auth_pb.ModifyClusterRoleBindingResponse>;
    authenticate: grpc.handleUnaryCall<auth_auth_pb.AuthenticateRequest, auth_auth_pb.AuthenticateResponse>;
    authorize: grpc.handleUnaryCall<auth_auth_pb.AuthorizeRequest, auth_auth_pb.AuthorizeResponse>;
    whoAmI: grpc.handleUnaryCall<auth_auth_pb.WhoAmIRequest, auth_auth_pb.WhoAmIResponse>;
    getScope: grpc.handleUnaryCall<auth_auth_pb.GetScopeRequest, auth_auth_pb.GetScopeResponse>;
    setScope: grpc.handleUnaryCall<auth_auth_pb.SetScopeRequest, auth_auth_pb.SetScopeResponse>;
    getACL: grpc.handleUnaryCall<auth_auth_pb.GetACLRequest, auth_auth_pb.GetACLResponse>;
    setACL: grpc.handleUnaryCall<auth_auth_pb.SetACLRequest, auth_auth_pb.SetACLResponse>;
    getOIDCLogin: grpc.handleUnaryCall<auth_auth_pb.GetOIDCLoginRequest, auth_auth_pb.GetOIDCLoginResponse>;
    getAuthToken: grpc.handleUnaryCall<auth_auth_pb.GetAuthTokenRequest, auth_auth_pb.GetAuthTokenResponse>;
    extendAuthToken: grpc.handleUnaryCall<auth_auth_pb.ExtendAuthTokenRequest, auth_auth_pb.ExtendAuthTokenResponse>;
    revokeAuthToken: grpc.handleUnaryCall<auth_auth_pb.RevokeAuthTokenRequest, auth_auth_pb.RevokeAuthTokenResponse>;
    setGroupsForUser: grpc.handleUnaryCall<auth_auth_pb.SetGroupsForUserRequest, auth_auth_pb.SetGroupsForUserResponse>;
    modifyMembers: grpc.handleUnaryCall<auth_auth_pb.ModifyMembersRequest, auth_auth_pb.ModifyMembersResponse>;
    getGroups: grpc.handleUnaryCall<auth_auth_pb.GetGroupsRequest, auth_auth_pb.GetGroupsResponse>;
    getUsers: grpc.handleUnaryCall<auth_auth_pb.GetUsersRequest, auth_auth_pb.GetUsersResponse>;
    extractAuthTokens: grpc.handleUnaryCall<auth_auth_pb.ExtractAuthTokensRequest, auth_auth_pb.ExtractAuthTokensResponse>;
    restoreAuthToken: grpc.handleUnaryCall<auth_auth_pb.RestoreAuthTokenRequest, auth_auth_pb.RestoreAuthTokenResponse>;
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
    getAdmins(request: auth_auth_pb.GetAdminsRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetAdminsResponse) => void): grpc.ClientUnaryCall;
    getAdmins(request: auth_auth_pb.GetAdminsRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetAdminsResponse) => void): grpc.ClientUnaryCall;
    getAdmins(request: auth_auth_pb.GetAdminsRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetAdminsResponse) => void): grpc.ClientUnaryCall;
    modifyAdmins(request: auth_auth_pb.ModifyAdminsRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.ModifyAdminsResponse) => void): grpc.ClientUnaryCall;
    modifyAdmins(request: auth_auth_pb.ModifyAdminsRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.ModifyAdminsResponse) => void): grpc.ClientUnaryCall;
    modifyAdmins(request: auth_auth_pb.ModifyAdminsRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.ModifyAdminsResponse) => void): grpc.ClientUnaryCall;
    getClusterRoleBindings(request: auth_auth_pb.GetClusterRoleBindingsRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetClusterRoleBindingsResponse) => void): grpc.ClientUnaryCall;
    getClusterRoleBindings(request: auth_auth_pb.GetClusterRoleBindingsRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetClusterRoleBindingsResponse) => void): grpc.ClientUnaryCall;
    getClusterRoleBindings(request: auth_auth_pb.GetClusterRoleBindingsRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetClusterRoleBindingsResponse) => void): grpc.ClientUnaryCall;
    modifyClusterRoleBinding(request: auth_auth_pb.ModifyClusterRoleBindingRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.ModifyClusterRoleBindingResponse) => void): grpc.ClientUnaryCall;
    modifyClusterRoleBinding(request: auth_auth_pb.ModifyClusterRoleBindingRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.ModifyClusterRoleBindingResponse) => void): grpc.ClientUnaryCall;
    modifyClusterRoleBinding(request: auth_auth_pb.ModifyClusterRoleBindingRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.ModifyClusterRoleBindingResponse) => void): grpc.ClientUnaryCall;
    authenticate(request: auth_auth_pb.AuthenticateRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.AuthenticateResponse) => void): grpc.ClientUnaryCall;
    authenticate(request: auth_auth_pb.AuthenticateRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.AuthenticateResponse) => void): grpc.ClientUnaryCall;
    authenticate(request: auth_auth_pb.AuthenticateRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.AuthenticateResponse) => void): grpc.ClientUnaryCall;
    authorize(request: auth_auth_pb.AuthorizeRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.AuthorizeResponse) => void): grpc.ClientUnaryCall;
    authorize(request: auth_auth_pb.AuthorizeRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.AuthorizeResponse) => void): grpc.ClientUnaryCall;
    authorize(request: auth_auth_pb.AuthorizeRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.AuthorizeResponse) => void): grpc.ClientUnaryCall;
    whoAmI(request: auth_auth_pb.WhoAmIRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.WhoAmIResponse) => void): grpc.ClientUnaryCall;
    whoAmI(request: auth_auth_pb.WhoAmIRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.WhoAmIResponse) => void): grpc.ClientUnaryCall;
    whoAmI(request: auth_auth_pb.WhoAmIRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.WhoAmIResponse) => void): grpc.ClientUnaryCall;
    getScope(request: auth_auth_pb.GetScopeRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetScopeResponse) => void): grpc.ClientUnaryCall;
    getScope(request: auth_auth_pb.GetScopeRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetScopeResponse) => void): grpc.ClientUnaryCall;
    getScope(request: auth_auth_pb.GetScopeRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetScopeResponse) => void): grpc.ClientUnaryCall;
    setScope(request: auth_auth_pb.SetScopeRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.SetScopeResponse) => void): grpc.ClientUnaryCall;
    setScope(request: auth_auth_pb.SetScopeRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.SetScopeResponse) => void): grpc.ClientUnaryCall;
    setScope(request: auth_auth_pb.SetScopeRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.SetScopeResponse) => void): grpc.ClientUnaryCall;
    getACL(request: auth_auth_pb.GetACLRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetACLResponse) => void): grpc.ClientUnaryCall;
    getACL(request: auth_auth_pb.GetACLRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetACLResponse) => void): grpc.ClientUnaryCall;
    getACL(request: auth_auth_pb.GetACLRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetACLResponse) => void): grpc.ClientUnaryCall;
    setACL(request: auth_auth_pb.SetACLRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.SetACLResponse) => void): grpc.ClientUnaryCall;
    setACL(request: auth_auth_pb.SetACLRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.SetACLResponse) => void): grpc.ClientUnaryCall;
    setACL(request: auth_auth_pb.SetACLRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.SetACLResponse) => void): grpc.ClientUnaryCall;
    getOIDCLogin(request: auth_auth_pb.GetOIDCLoginRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetOIDCLoginResponse) => void): grpc.ClientUnaryCall;
    getOIDCLogin(request: auth_auth_pb.GetOIDCLoginRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetOIDCLoginResponse) => void): grpc.ClientUnaryCall;
    getOIDCLogin(request: auth_auth_pb.GetOIDCLoginRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetOIDCLoginResponse) => void): grpc.ClientUnaryCall;
    getAuthToken(request: auth_auth_pb.GetAuthTokenRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetAuthTokenResponse) => void): grpc.ClientUnaryCall;
    getAuthToken(request: auth_auth_pb.GetAuthTokenRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetAuthTokenResponse) => void): grpc.ClientUnaryCall;
    getAuthToken(request: auth_auth_pb.GetAuthTokenRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetAuthTokenResponse) => void): grpc.ClientUnaryCall;
    extendAuthToken(request: auth_auth_pb.ExtendAuthTokenRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.ExtendAuthTokenResponse) => void): grpc.ClientUnaryCall;
    extendAuthToken(request: auth_auth_pb.ExtendAuthTokenRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.ExtendAuthTokenResponse) => void): grpc.ClientUnaryCall;
    extendAuthToken(request: auth_auth_pb.ExtendAuthTokenRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.ExtendAuthTokenResponse) => void): grpc.ClientUnaryCall;
    revokeAuthToken(request: auth_auth_pb.RevokeAuthTokenRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.RevokeAuthTokenResponse) => void): grpc.ClientUnaryCall;
    revokeAuthToken(request: auth_auth_pb.RevokeAuthTokenRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.RevokeAuthTokenResponse) => void): grpc.ClientUnaryCall;
    revokeAuthToken(request: auth_auth_pb.RevokeAuthTokenRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.RevokeAuthTokenResponse) => void): grpc.ClientUnaryCall;
    setGroupsForUser(request: auth_auth_pb.SetGroupsForUserRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.SetGroupsForUserResponse) => void): grpc.ClientUnaryCall;
    setGroupsForUser(request: auth_auth_pb.SetGroupsForUserRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.SetGroupsForUserResponse) => void): grpc.ClientUnaryCall;
    setGroupsForUser(request: auth_auth_pb.SetGroupsForUserRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.SetGroupsForUserResponse) => void): grpc.ClientUnaryCall;
    modifyMembers(request: auth_auth_pb.ModifyMembersRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.ModifyMembersResponse) => void): grpc.ClientUnaryCall;
    modifyMembers(request: auth_auth_pb.ModifyMembersRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.ModifyMembersResponse) => void): grpc.ClientUnaryCall;
    modifyMembers(request: auth_auth_pb.ModifyMembersRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.ModifyMembersResponse) => void): grpc.ClientUnaryCall;
    getGroups(request: auth_auth_pb.GetGroupsRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetGroupsResponse) => void): grpc.ClientUnaryCall;
    getGroups(request: auth_auth_pb.GetGroupsRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetGroupsResponse) => void): grpc.ClientUnaryCall;
    getGroups(request: auth_auth_pb.GetGroupsRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetGroupsResponse) => void): grpc.ClientUnaryCall;
    getUsers(request: auth_auth_pb.GetUsersRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetUsersResponse) => void): grpc.ClientUnaryCall;
    getUsers(request: auth_auth_pb.GetUsersRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetUsersResponse) => void): grpc.ClientUnaryCall;
    getUsers(request: auth_auth_pb.GetUsersRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetUsersResponse) => void): grpc.ClientUnaryCall;
    extractAuthTokens(request: auth_auth_pb.ExtractAuthTokensRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.ExtractAuthTokensResponse) => void): grpc.ClientUnaryCall;
    extractAuthTokens(request: auth_auth_pb.ExtractAuthTokensRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.ExtractAuthTokensResponse) => void): grpc.ClientUnaryCall;
    extractAuthTokens(request: auth_auth_pb.ExtractAuthTokensRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.ExtractAuthTokensResponse) => void): grpc.ClientUnaryCall;
    restoreAuthToken(request: auth_auth_pb.RestoreAuthTokenRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.RestoreAuthTokenResponse) => void): grpc.ClientUnaryCall;
    restoreAuthToken(request: auth_auth_pb.RestoreAuthTokenRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.RestoreAuthTokenResponse) => void): grpc.ClientUnaryCall;
    restoreAuthToken(request: auth_auth_pb.RestoreAuthTokenRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.RestoreAuthTokenResponse) => void): grpc.ClientUnaryCall;
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
    public getAdmins(request: auth_auth_pb.GetAdminsRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetAdminsResponse) => void): grpc.ClientUnaryCall;
    public getAdmins(request: auth_auth_pb.GetAdminsRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetAdminsResponse) => void): grpc.ClientUnaryCall;
    public getAdmins(request: auth_auth_pb.GetAdminsRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetAdminsResponse) => void): grpc.ClientUnaryCall;
    public modifyAdmins(request: auth_auth_pb.ModifyAdminsRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.ModifyAdminsResponse) => void): grpc.ClientUnaryCall;
    public modifyAdmins(request: auth_auth_pb.ModifyAdminsRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.ModifyAdminsResponse) => void): grpc.ClientUnaryCall;
    public modifyAdmins(request: auth_auth_pb.ModifyAdminsRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.ModifyAdminsResponse) => void): grpc.ClientUnaryCall;
    public getClusterRoleBindings(request: auth_auth_pb.GetClusterRoleBindingsRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetClusterRoleBindingsResponse) => void): grpc.ClientUnaryCall;
    public getClusterRoleBindings(request: auth_auth_pb.GetClusterRoleBindingsRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetClusterRoleBindingsResponse) => void): grpc.ClientUnaryCall;
    public getClusterRoleBindings(request: auth_auth_pb.GetClusterRoleBindingsRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetClusterRoleBindingsResponse) => void): grpc.ClientUnaryCall;
    public modifyClusterRoleBinding(request: auth_auth_pb.ModifyClusterRoleBindingRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.ModifyClusterRoleBindingResponse) => void): grpc.ClientUnaryCall;
    public modifyClusterRoleBinding(request: auth_auth_pb.ModifyClusterRoleBindingRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.ModifyClusterRoleBindingResponse) => void): grpc.ClientUnaryCall;
    public modifyClusterRoleBinding(request: auth_auth_pb.ModifyClusterRoleBindingRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.ModifyClusterRoleBindingResponse) => void): grpc.ClientUnaryCall;
    public authenticate(request: auth_auth_pb.AuthenticateRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.AuthenticateResponse) => void): grpc.ClientUnaryCall;
    public authenticate(request: auth_auth_pb.AuthenticateRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.AuthenticateResponse) => void): grpc.ClientUnaryCall;
    public authenticate(request: auth_auth_pb.AuthenticateRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.AuthenticateResponse) => void): grpc.ClientUnaryCall;
    public authorize(request: auth_auth_pb.AuthorizeRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.AuthorizeResponse) => void): grpc.ClientUnaryCall;
    public authorize(request: auth_auth_pb.AuthorizeRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.AuthorizeResponse) => void): grpc.ClientUnaryCall;
    public authorize(request: auth_auth_pb.AuthorizeRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.AuthorizeResponse) => void): grpc.ClientUnaryCall;
    public whoAmI(request: auth_auth_pb.WhoAmIRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.WhoAmIResponse) => void): grpc.ClientUnaryCall;
    public whoAmI(request: auth_auth_pb.WhoAmIRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.WhoAmIResponse) => void): grpc.ClientUnaryCall;
    public whoAmI(request: auth_auth_pb.WhoAmIRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.WhoAmIResponse) => void): grpc.ClientUnaryCall;
    public getScope(request: auth_auth_pb.GetScopeRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetScopeResponse) => void): grpc.ClientUnaryCall;
    public getScope(request: auth_auth_pb.GetScopeRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetScopeResponse) => void): grpc.ClientUnaryCall;
    public getScope(request: auth_auth_pb.GetScopeRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetScopeResponse) => void): grpc.ClientUnaryCall;
    public setScope(request: auth_auth_pb.SetScopeRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.SetScopeResponse) => void): grpc.ClientUnaryCall;
    public setScope(request: auth_auth_pb.SetScopeRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.SetScopeResponse) => void): grpc.ClientUnaryCall;
    public setScope(request: auth_auth_pb.SetScopeRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.SetScopeResponse) => void): grpc.ClientUnaryCall;
    public getACL(request: auth_auth_pb.GetACLRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetACLResponse) => void): grpc.ClientUnaryCall;
    public getACL(request: auth_auth_pb.GetACLRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetACLResponse) => void): grpc.ClientUnaryCall;
    public getACL(request: auth_auth_pb.GetACLRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetACLResponse) => void): grpc.ClientUnaryCall;
    public setACL(request: auth_auth_pb.SetACLRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.SetACLResponse) => void): grpc.ClientUnaryCall;
    public setACL(request: auth_auth_pb.SetACLRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.SetACLResponse) => void): grpc.ClientUnaryCall;
    public setACL(request: auth_auth_pb.SetACLRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.SetACLResponse) => void): grpc.ClientUnaryCall;
    public getOIDCLogin(request: auth_auth_pb.GetOIDCLoginRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetOIDCLoginResponse) => void): grpc.ClientUnaryCall;
    public getOIDCLogin(request: auth_auth_pb.GetOIDCLoginRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetOIDCLoginResponse) => void): grpc.ClientUnaryCall;
    public getOIDCLogin(request: auth_auth_pb.GetOIDCLoginRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetOIDCLoginResponse) => void): grpc.ClientUnaryCall;
    public getAuthToken(request: auth_auth_pb.GetAuthTokenRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetAuthTokenResponse) => void): grpc.ClientUnaryCall;
    public getAuthToken(request: auth_auth_pb.GetAuthTokenRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetAuthTokenResponse) => void): grpc.ClientUnaryCall;
    public getAuthToken(request: auth_auth_pb.GetAuthTokenRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetAuthTokenResponse) => void): grpc.ClientUnaryCall;
    public extendAuthToken(request: auth_auth_pb.ExtendAuthTokenRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.ExtendAuthTokenResponse) => void): grpc.ClientUnaryCall;
    public extendAuthToken(request: auth_auth_pb.ExtendAuthTokenRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.ExtendAuthTokenResponse) => void): grpc.ClientUnaryCall;
    public extendAuthToken(request: auth_auth_pb.ExtendAuthTokenRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.ExtendAuthTokenResponse) => void): grpc.ClientUnaryCall;
    public revokeAuthToken(request: auth_auth_pb.RevokeAuthTokenRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.RevokeAuthTokenResponse) => void): grpc.ClientUnaryCall;
    public revokeAuthToken(request: auth_auth_pb.RevokeAuthTokenRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.RevokeAuthTokenResponse) => void): grpc.ClientUnaryCall;
    public revokeAuthToken(request: auth_auth_pb.RevokeAuthTokenRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.RevokeAuthTokenResponse) => void): grpc.ClientUnaryCall;
    public setGroupsForUser(request: auth_auth_pb.SetGroupsForUserRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.SetGroupsForUserResponse) => void): grpc.ClientUnaryCall;
    public setGroupsForUser(request: auth_auth_pb.SetGroupsForUserRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.SetGroupsForUserResponse) => void): grpc.ClientUnaryCall;
    public setGroupsForUser(request: auth_auth_pb.SetGroupsForUserRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.SetGroupsForUserResponse) => void): grpc.ClientUnaryCall;
    public modifyMembers(request: auth_auth_pb.ModifyMembersRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.ModifyMembersResponse) => void): grpc.ClientUnaryCall;
    public modifyMembers(request: auth_auth_pb.ModifyMembersRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.ModifyMembersResponse) => void): grpc.ClientUnaryCall;
    public modifyMembers(request: auth_auth_pb.ModifyMembersRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.ModifyMembersResponse) => void): grpc.ClientUnaryCall;
    public getGroups(request: auth_auth_pb.GetGroupsRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetGroupsResponse) => void): grpc.ClientUnaryCall;
    public getGroups(request: auth_auth_pb.GetGroupsRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetGroupsResponse) => void): grpc.ClientUnaryCall;
    public getGroups(request: auth_auth_pb.GetGroupsRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetGroupsResponse) => void): grpc.ClientUnaryCall;
    public getUsers(request: auth_auth_pb.GetUsersRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetUsersResponse) => void): grpc.ClientUnaryCall;
    public getUsers(request: auth_auth_pb.GetUsersRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetUsersResponse) => void): grpc.ClientUnaryCall;
    public getUsers(request: auth_auth_pb.GetUsersRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.GetUsersResponse) => void): grpc.ClientUnaryCall;
    public extractAuthTokens(request: auth_auth_pb.ExtractAuthTokensRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.ExtractAuthTokensResponse) => void): grpc.ClientUnaryCall;
    public extractAuthTokens(request: auth_auth_pb.ExtractAuthTokensRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.ExtractAuthTokensResponse) => void): grpc.ClientUnaryCall;
    public extractAuthTokens(request: auth_auth_pb.ExtractAuthTokensRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.ExtractAuthTokensResponse) => void): grpc.ClientUnaryCall;
    public restoreAuthToken(request: auth_auth_pb.RestoreAuthTokenRequest, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.RestoreAuthTokenResponse) => void): grpc.ClientUnaryCall;
    public restoreAuthToken(request: auth_auth_pb.RestoreAuthTokenRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.RestoreAuthTokenResponse) => void): grpc.ClientUnaryCall;
    public restoreAuthToken(request: auth_auth_pb.RestoreAuthTokenRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: auth_auth_pb.RestoreAuthTokenResponse) => void): grpc.ClientUnaryCall;
}
