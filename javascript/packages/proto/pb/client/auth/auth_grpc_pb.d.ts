// package: auth
// file: client/auth/auth.proto

/* tslint:disable */
/* eslint-disable */

import * as grpc from "@grpc/grpc-js";
import {handleClientStreamingCall} from "@grpc/grpc-js/build/src/server-call";
import * as client_auth_auth_pb from "../../client/auth/auth_pb";
import * as gogoproto_gogo_pb from "../../gogoproto/gogo_pb";
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
    getOneTimePassword: IAPIService_IGetOneTimePassword;
}

interface IAPIService_IActivate extends grpc.MethodDefinition<client_auth_auth_pb.ActivateRequest, client_auth_auth_pb.ActivateResponse> {
    path: "/auth.API/Activate";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<client_auth_auth_pb.ActivateRequest>;
    requestDeserialize: grpc.deserialize<client_auth_auth_pb.ActivateRequest>;
    responseSerialize: grpc.serialize<client_auth_auth_pb.ActivateResponse>;
    responseDeserialize: grpc.deserialize<client_auth_auth_pb.ActivateResponse>;
}
interface IAPIService_IDeactivate extends grpc.MethodDefinition<client_auth_auth_pb.DeactivateRequest, client_auth_auth_pb.DeactivateResponse> {
    path: "/auth.API/Deactivate";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<client_auth_auth_pb.DeactivateRequest>;
    requestDeserialize: grpc.deserialize<client_auth_auth_pb.DeactivateRequest>;
    responseSerialize: grpc.serialize<client_auth_auth_pb.DeactivateResponse>;
    responseDeserialize: grpc.deserialize<client_auth_auth_pb.DeactivateResponse>;
}
interface IAPIService_IGetConfiguration extends grpc.MethodDefinition<client_auth_auth_pb.GetConfigurationRequest, client_auth_auth_pb.GetConfigurationResponse> {
    path: "/auth.API/GetConfiguration";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<client_auth_auth_pb.GetConfigurationRequest>;
    requestDeserialize: grpc.deserialize<client_auth_auth_pb.GetConfigurationRequest>;
    responseSerialize: grpc.serialize<client_auth_auth_pb.GetConfigurationResponse>;
    responseDeserialize: grpc.deserialize<client_auth_auth_pb.GetConfigurationResponse>;
}
interface IAPIService_ISetConfiguration extends grpc.MethodDefinition<client_auth_auth_pb.SetConfigurationRequest, client_auth_auth_pb.SetConfigurationResponse> {
    path: "/auth.API/SetConfiguration";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<client_auth_auth_pb.SetConfigurationRequest>;
    requestDeserialize: grpc.deserialize<client_auth_auth_pb.SetConfigurationRequest>;
    responseSerialize: grpc.serialize<client_auth_auth_pb.SetConfigurationResponse>;
    responseDeserialize: grpc.deserialize<client_auth_auth_pb.SetConfigurationResponse>;
}
interface IAPIService_IGetAdmins extends grpc.MethodDefinition<client_auth_auth_pb.GetAdminsRequest, client_auth_auth_pb.GetAdminsResponse> {
    path: "/auth.API/GetAdmins";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<client_auth_auth_pb.GetAdminsRequest>;
    requestDeserialize: grpc.deserialize<client_auth_auth_pb.GetAdminsRequest>;
    responseSerialize: grpc.serialize<client_auth_auth_pb.GetAdminsResponse>;
    responseDeserialize: grpc.deserialize<client_auth_auth_pb.GetAdminsResponse>;
}
interface IAPIService_IModifyAdmins extends grpc.MethodDefinition<client_auth_auth_pb.ModifyAdminsRequest, client_auth_auth_pb.ModifyAdminsResponse> {
    path: "/auth.API/ModifyAdmins";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<client_auth_auth_pb.ModifyAdminsRequest>;
    requestDeserialize: grpc.deserialize<client_auth_auth_pb.ModifyAdminsRequest>;
    responseSerialize: grpc.serialize<client_auth_auth_pb.ModifyAdminsResponse>;
    responseDeserialize: grpc.deserialize<client_auth_auth_pb.ModifyAdminsResponse>;
}
interface IAPIService_IGetClusterRoleBindings extends grpc.MethodDefinition<client_auth_auth_pb.GetClusterRoleBindingsRequest, client_auth_auth_pb.GetClusterRoleBindingsResponse> {
    path: "/auth.API/GetClusterRoleBindings";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<client_auth_auth_pb.GetClusterRoleBindingsRequest>;
    requestDeserialize: grpc.deserialize<client_auth_auth_pb.GetClusterRoleBindingsRequest>;
    responseSerialize: grpc.serialize<client_auth_auth_pb.GetClusterRoleBindingsResponse>;
    responseDeserialize: grpc.deserialize<client_auth_auth_pb.GetClusterRoleBindingsResponse>;
}
interface IAPIService_IModifyClusterRoleBinding extends grpc.MethodDefinition<client_auth_auth_pb.ModifyClusterRoleBindingRequest, client_auth_auth_pb.ModifyClusterRoleBindingResponse> {
    path: "/auth.API/ModifyClusterRoleBinding";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<client_auth_auth_pb.ModifyClusterRoleBindingRequest>;
    requestDeserialize: grpc.deserialize<client_auth_auth_pb.ModifyClusterRoleBindingRequest>;
    responseSerialize: grpc.serialize<client_auth_auth_pb.ModifyClusterRoleBindingResponse>;
    responseDeserialize: grpc.deserialize<client_auth_auth_pb.ModifyClusterRoleBindingResponse>;
}
interface IAPIService_IAuthenticate extends grpc.MethodDefinition<client_auth_auth_pb.AuthenticateRequest, client_auth_auth_pb.AuthenticateResponse> {
    path: "/auth.API/Authenticate";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<client_auth_auth_pb.AuthenticateRequest>;
    requestDeserialize: grpc.deserialize<client_auth_auth_pb.AuthenticateRequest>;
    responseSerialize: grpc.serialize<client_auth_auth_pb.AuthenticateResponse>;
    responseDeserialize: grpc.deserialize<client_auth_auth_pb.AuthenticateResponse>;
}
interface IAPIService_IAuthorize extends grpc.MethodDefinition<client_auth_auth_pb.AuthorizeRequest, client_auth_auth_pb.AuthorizeResponse> {
    path: "/auth.API/Authorize";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<client_auth_auth_pb.AuthorizeRequest>;
    requestDeserialize: grpc.deserialize<client_auth_auth_pb.AuthorizeRequest>;
    responseSerialize: grpc.serialize<client_auth_auth_pb.AuthorizeResponse>;
    responseDeserialize: grpc.deserialize<client_auth_auth_pb.AuthorizeResponse>;
}
interface IAPIService_IWhoAmI extends grpc.MethodDefinition<client_auth_auth_pb.WhoAmIRequest, client_auth_auth_pb.WhoAmIResponse> {
    path: "/auth.API/WhoAmI";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<client_auth_auth_pb.WhoAmIRequest>;
    requestDeserialize: grpc.deserialize<client_auth_auth_pb.WhoAmIRequest>;
    responseSerialize: grpc.serialize<client_auth_auth_pb.WhoAmIResponse>;
    responseDeserialize: grpc.deserialize<client_auth_auth_pb.WhoAmIResponse>;
}
interface IAPIService_IGetScope extends grpc.MethodDefinition<client_auth_auth_pb.GetScopeRequest, client_auth_auth_pb.GetScopeResponse> {
    path: "/auth.API/GetScope";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<client_auth_auth_pb.GetScopeRequest>;
    requestDeserialize: grpc.deserialize<client_auth_auth_pb.GetScopeRequest>;
    responseSerialize: grpc.serialize<client_auth_auth_pb.GetScopeResponse>;
    responseDeserialize: grpc.deserialize<client_auth_auth_pb.GetScopeResponse>;
}
interface IAPIService_ISetScope extends grpc.MethodDefinition<client_auth_auth_pb.SetScopeRequest, client_auth_auth_pb.SetScopeResponse> {
    path: "/auth.API/SetScope";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<client_auth_auth_pb.SetScopeRequest>;
    requestDeserialize: grpc.deserialize<client_auth_auth_pb.SetScopeRequest>;
    responseSerialize: grpc.serialize<client_auth_auth_pb.SetScopeResponse>;
    responseDeserialize: grpc.deserialize<client_auth_auth_pb.SetScopeResponse>;
}
interface IAPIService_IGetACL extends grpc.MethodDefinition<client_auth_auth_pb.GetACLRequest, client_auth_auth_pb.GetACLResponse> {
    path: "/auth.API/GetACL";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<client_auth_auth_pb.GetACLRequest>;
    requestDeserialize: grpc.deserialize<client_auth_auth_pb.GetACLRequest>;
    responseSerialize: grpc.serialize<client_auth_auth_pb.GetACLResponse>;
    responseDeserialize: grpc.deserialize<client_auth_auth_pb.GetACLResponse>;
}
interface IAPIService_ISetACL extends grpc.MethodDefinition<client_auth_auth_pb.SetACLRequest, client_auth_auth_pb.SetACLResponse> {
    path: "/auth.API/SetACL";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<client_auth_auth_pb.SetACLRequest>;
    requestDeserialize: grpc.deserialize<client_auth_auth_pb.SetACLRequest>;
    responseSerialize: grpc.serialize<client_auth_auth_pb.SetACLResponse>;
    responseDeserialize: grpc.deserialize<client_auth_auth_pb.SetACLResponse>;
}
interface IAPIService_IGetOIDCLogin extends grpc.MethodDefinition<client_auth_auth_pb.GetOIDCLoginRequest, client_auth_auth_pb.GetOIDCLoginResponse> {
    path: "/auth.API/GetOIDCLogin";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<client_auth_auth_pb.GetOIDCLoginRequest>;
    requestDeserialize: grpc.deserialize<client_auth_auth_pb.GetOIDCLoginRequest>;
    responseSerialize: grpc.serialize<client_auth_auth_pb.GetOIDCLoginResponse>;
    responseDeserialize: grpc.deserialize<client_auth_auth_pb.GetOIDCLoginResponse>;
}
interface IAPIService_IGetAuthToken extends grpc.MethodDefinition<client_auth_auth_pb.GetAuthTokenRequest, client_auth_auth_pb.GetAuthTokenResponse> {
    path: "/auth.API/GetAuthToken";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<client_auth_auth_pb.GetAuthTokenRequest>;
    requestDeserialize: grpc.deserialize<client_auth_auth_pb.GetAuthTokenRequest>;
    responseSerialize: grpc.serialize<client_auth_auth_pb.GetAuthTokenResponse>;
    responseDeserialize: grpc.deserialize<client_auth_auth_pb.GetAuthTokenResponse>;
}
interface IAPIService_IExtendAuthToken extends grpc.MethodDefinition<client_auth_auth_pb.ExtendAuthTokenRequest, client_auth_auth_pb.ExtendAuthTokenResponse> {
    path: "/auth.API/ExtendAuthToken";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<client_auth_auth_pb.ExtendAuthTokenRequest>;
    requestDeserialize: grpc.deserialize<client_auth_auth_pb.ExtendAuthTokenRequest>;
    responseSerialize: grpc.serialize<client_auth_auth_pb.ExtendAuthTokenResponse>;
    responseDeserialize: grpc.deserialize<client_auth_auth_pb.ExtendAuthTokenResponse>;
}
interface IAPIService_IRevokeAuthToken extends grpc.MethodDefinition<client_auth_auth_pb.RevokeAuthTokenRequest, client_auth_auth_pb.RevokeAuthTokenResponse> {
    path: "/auth.API/RevokeAuthToken";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<client_auth_auth_pb.RevokeAuthTokenRequest>;
    requestDeserialize: grpc.deserialize<client_auth_auth_pb.RevokeAuthTokenRequest>;
    responseSerialize: grpc.serialize<client_auth_auth_pb.RevokeAuthTokenResponse>;
    responseDeserialize: grpc.deserialize<client_auth_auth_pb.RevokeAuthTokenResponse>;
}
interface IAPIService_ISetGroupsForUser extends grpc.MethodDefinition<client_auth_auth_pb.SetGroupsForUserRequest, client_auth_auth_pb.SetGroupsForUserResponse> {
    path: "/auth.API/SetGroupsForUser";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<client_auth_auth_pb.SetGroupsForUserRequest>;
    requestDeserialize: grpc.deserialize<client_auth_auth_pb.SetGroupsForUserRequest>;
    responseSerialize: grpc.serialize<client_auth_auth_pb.SetGroupsForUserResponse>;
    responseDeserialize: grpc.deserialize<client_auth_auth_pb.SetGroupsForUserResponse>;
}
interface IAPIService_IModifyMembers extends grpc.MethodDefinition<client_auth_auth_pb.ModifyMembersRequest, client_auth_auth_pb.ModifyMembersResponse> {
    path: "/auth.API/ModifyMembers";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<client_auth_auth_pb.ModifyMembersRequest>;
    requestDeserialize: grpc.deserialize<client_auth_auth_pb.ModifyMembersRequest>;
    responseSerialize: grpc.serialize<client_auth_auth_pb.ModifyMembersResponse>;
    responseDeserialize: grpc.deserialize<client_auth_auth_pb.ModifyMembersResponse>;
}
interface IAPIService_IGetGroups extends grpc.MethodDefinition<client_auth_auth_pb.GetGroupsRequest, client_auth_auth_pb.GetGroupsResponse> {
    path: "/auth.API/GetGroups";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<client_auth_auth_pb.GetGroupsRequest>;
    requestDeserialize: grpc.deserialize<client_auth_auth_pb.GetGroupsRequest>;
    responseSerialize: grpc.serialize<client_auth_auth_pb.GetGroupsResponse>;
    responseDeserialize: grpc.deserialize<client_auth_auth_pb.GetGroupsResponse>;
}
interface IAPIService_IGetUsers extends grpc.MethodDefinition<client_auth_auth_pb.GetUsersRequest, client_auth_auth_pb.GetUsersResponse> {
    path: "/auth.API/GetUsers";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<client_auth_auth_pb.GetUsersRequest>;
    requestDeserialize: grpc.deserialize<client_auth_auth_pb.GetUsersRequest>;
    responseSerialize: grpc.serialize<client_auth_auth_pb.GetUsersResponse>;
    responseDeserialize: grpc.deserialize<client_auth_auth_pb.GetUsersResponse>;
}
interface IAPIService_IGetOneTimePassword extends grpc.MethodDefinition<client_auth_auth_pb.GetOneTimePasswordRequest, client_auth_auth_pb.GetOneTimePasswordResponse> {
    path: "/auth.API/GetOneTimePassword";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<client_auth_auth_pb.GetOneTimePasswordRequest>;
    requestDeserialize: grpc.deserialize<client_auth_auth_pb.GetOneTimePasswordRequest>;
    responseSerialize: grpc.serialize<client_auth_auth_pb.GetOneTimePasswordResponse>;
    responseDeserialize: grpc.deserialize<client_auth_auth_pb.GetOneTimePasswordResponse>;
}

export const APIService: IAPIService;

export interface IAPIServer {
    activate: grpc.handleUnaryCall<client_auth_auth_pb.ActivateRequest, client_auth_auth_pb.ActivateResponse>;
    deactivate: grpc.handleUnaryCall<client_auth_auth_pb.DeactivateRequest, client_auth_auth_pb.DeactivateResponse>;
    getConfiguration: grpc.handleUnaryCall<client_auth_auth_pb.GetConfigurationRequest, client_auth_auth_pb.GetConfigurationResponse>;
    setConfiguration: grpc.handleUnaryCall<client_auth_auth_pb.SetConfigurationRequest, client_auth_auth_pb.SetConfigurationResponse>;
    getAdmins: grpc.handleUnaryCall<client_auth_auth_pb.GetAdminsRequest, client_auth_auth_pb.GetAdminsResponse>;
    modifyAdmins: grpc.handleUnaryCall<client_auth_auth_pb.ModifyAdminsRequest, client_auth_auth_pb.ModifyAdminsResponse>;
    getClusterRoleBindings: grpc.handleUnaryCall<client_auth_auth_pb.GetClusterRoleBindingsRequest, client_auth_auth_pb.GetClusterRoleBindingsResponse>;
    modifyClusterRoleBinding: grpc.handleUnaryCall<client_auth_auth_pb.ModifyClusterRoleBindingRequest, client_auth_auth_pb.ModifyClusterRoleBindingResponse>;
    authenticate: grpc.handleUnaryCall<client_auth_auth_pb.AuthenticateRequest, client_auth_auth_pb.AuthenticateResponse>;
    authorize: grpc.handleUnaryCall<client_auth_auth_pb.AuthorizeRequest, client_auth_auth_pb.AuthorizeResponse>;
    whoAmI: grpc.handleUnaryCall<client_auth_auth_pb.WhoAmIRequest, client_auth_auth_pb.WhoAmIResponse>;
    getScope: grpc.handleUnaryCall<client_auth_auth_pb.GetScopeRequest, client_auth_auth_pb.GetScopeResponse>;
    setScope: grpc.handleUnaryCall<client_auth_auth_pb.SetScopeRequest, client_auth_auth_pb.SetScopeResponse>;
    getACL: grpc.handleUnaryCall<client_auth_auth_pb.GetACLRequest, client_auth_auth_pb.GetACLResponse>;
    setACL: grpc.handleUnaryCall<client_auth_auth_pb.SetACLRequest, client_auth_auth_pb.SetACLResponse>;
    getOIDCLogin: grpc.handleUnaryCall<client_auth_auth_pb.GetOIDCLoginRequest, client_auth_auth_pb.GetOIDCLoginResponse>;
    getAuthToken: grpc.handleUnaryCall<client_auth_auth_pb.GetAuthTokenRequest, client_auth_auth_pb.GetAuthTokenResponse>;
    extendAuthToken: grpc.handleUnaryCall<client_auth_auth_pb.ExtendAuthTokenRequest, client_auth_auth_pb.ExtendAuthTokenResponse>;
    revokeAuthToken: grpc.handleUnaryCall<client_auth_auth_pb.RevokeAuthTokenRequest, client_auth_auth_pb.RevokeAuthTokenResponse>;
    setGroupsForUser: grpc.handleUnaryCall<client_auth_auth_pb.SetGroupsForUserRequest, client_auth_auth_pb.SetGroupsForUserResponse>;
    modifyMembers: grpc.handleUnaryCall<client_auth_auth_pb.ModifyMembersRequest, client_auth_auth_pb.ModifyMembersResponse>;
    getGroups: grpc.handleUnaryCall<client_auth_auth_pb.GetGroupsRequest, client_auth_auth_pb.GetGroupsResponse>;
    getUsers: grpc.handleUnaryCall<client_auth_auth_pb.GetUsersRequest, client_auth_auth_pb.GetUsersResponse>;
    getOneTimePassword: grpc.handleUnaryCall<client_auth_auth_pb.GetOneTimePasswordRequest, client_auth_auth_pb.GetOneTimePasswordResponse>;
}

export interface IAPIClient {
    activate(request: client_auth_auth_pb.ActivateRequest, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.ActivateResponse) => void): grpc.ClientUnaryCall;
    activate(request: client_auth_auth_pb.ActivateRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.ActivateResponse) => void): grpc.ClientUnaryCall;
    activate(request: client_auth_auth_pb.ActivateRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.ActivateResponse) => void): grpc.ClientUnaryCall;
    deactivate(request: client_auth_auth_pb.DeactivateRequest, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.DeactivateResponse) => void): grpc.ClientUnaryCall;
    deactivate(request: client_auth_auth_pb.DeactivateRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.DeactivateResponse) => void): grpc.ClientUnaryCall;
    deactivate(request: client_auth_auth_pb.DeactivateRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.DeactivateResponse) => void): grpc.ClientUnaryCall;
    getConfiguration(request: client_auth_auth_pb.GetConfigurationRequest, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.GetConfigurationResponse) => void): grpc.ClientUnaryCall;
    getConfiguration(request: client_auth_auth_pb.GetConfigurationRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.GetConfigurationResponse) => void): grpc.ClientUnaryCall;
    getConfiguration(request: client_auth_auth_pb.GetConfigurationRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.GetConfigurationResponse) => void): grpc.ClientUnaryCall;
    setConfiguration(request: client_auth_auth_pb.SetConfigurationRequest, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.SetConfigurationResponse) => void): grpc.ClientUnaryCall;
    setConfiguration(request: client_auth_auth_pb.SetConfigurationRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.SetConfigurationResponse) => void): grpc.ClientUnaryCall;
    setConfiguration(request: client_auth_auth_pb.SetConfigurationRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.SetConfigurationResponse) => void): grpc.ClientUnaryCall;
    getAdmins(request: client_auth_auth_pb.GetAdminsRequest, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.GetAdminsResponse) => void): grpc.ClientUnaryCall;
    getAdmins(request: client_auth_auth_pb.GetAdminsRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.GetAdminsResponse) => void): grpc.ClientUnaryCall;
    getAdmins(request: client_auth_auth_pb.GetAdminsRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.GetAdminsResponse) => void): grpc.ClientUnaryCall;
    modifyAdmins(request: client_auth_auth_pb.ModifyAdminsRequest, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.ModifyAdminsResponse) => void): grpc.ClientUnaryCall;
    modifyAdmins(request: client_auth_auth_pb.ModifyAdminsRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.ModifyAdminsResponse) => void): grpc.ClientUnaryCall;
    modifyAdmins(request: client_auth_auth_pb.ModifyAdminsRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.ModifyAdminsResponse) => void): grpc.ClientUnaryCall;
    getClusterRoleBindings(request: client_auth_auth_pb.GetClusterRoleBindingsRequest, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.GetClusterRoleBindingsResponse) => void): grpc.ClientUnaryCall;
    getClusterRoleBindings(request: client_auth_auth_pb.GetClusterRoleBindingsRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.GetClusterRoleBindingsResponse) => void): grpc.ClientUnaryCall;
    getClusterRoleBindings(request: client_auth_auth_pb.GetClusterRoleBindingsRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.GetClusterRoleBindingsResponse) => void): grpc.ClientUnaryCall;
    modifyClusterRoleBinding(request: client_auth_auth_pb.ModifyClusterRoleBindingRequest, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.ModifyClusterRoleBindingResponse) => void): grpc.ClientUnaryCall;
    modifyClusterRoleBinding(request: client_auth_auth_pb.ModifyClusterRoleBindingRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.ModifyClusterRoleBindingResponse) => void): grpc.ClientUnaryCall;
    modifyClusterRoleBinding(request: client_auth_auth_pb.ModifyClusterRoleBindingRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.ModifyClusterRoleBindingResponse) => void): grpc.ClientUnaryCall;
    authenticate(request: client_auth_auth_pb.AuthenticateRequest, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.AuthenticateResponse) => void): grpc.ClientUnaryCall;
    authenticate(request: client_auth_auth_pb.AuthenticateRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.AuthenticateResponse) => void): grpc.ClientUnaryCall;
    authenticate(request: client_auth_auth_pb.AuthenticateRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.AuthenticateResponse) => void): grpc.ClientUnaryCall;
    authorize(request: client_auth_auth_pb.AuthorizeRequest, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.AuthorizeResponse) => void): grpc.ClientUnaryCall;
    authorize(request: client_auth_auth_pb.AuthorizeRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.AuthorizeResponse) => void): grpc.ClientUnaryCall;
    authorize(request: client_auth_auth_pb.AuthorizeRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.AuthorizeResponse) => void): grpc.ClientUnaryCall;
    whoAmI(request: client_auth_auth_pb.WhoAmIRequest, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.WhoAmIResponse) => void): grpc.ClientUnaryCall;
    whoAmI(request: client_auth_auth_pb.WhoAmIRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.WhoAmIResponse) => void): grpc.ClientUnaryCall;
    whoAmI(request: client_auth_auth_pb.WhoAmIRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.WhoAmIResponse) => void): grpc.ClientUnaryCall;
    getScope(request: client_auth_auth_pb.GetScopeRequest, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.GetScopeResponse) => void): grpc.ClientUnaryCall;
    getScope(request: client_auth_auth_pb.GetScopeRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.GetScopeResponse) => void): grpc.ClientUnaryCall;
    getScope(request: client_auth_auth_pb.GetScopeRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.GetScopeResponse) => void): grpc.ClientUnaryCall;
    setScope(request: client_auth_auth_pb.SetScopeRequest, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.SetScopeResponse) => void): grpc.ClientUnaryCall;
    setScope(request: client_auth_auth_pb.SetScopeRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.SetScopeResponse) => void): grpc.ClientUnaryCall;
    setScope(request: client_auth_auth_pb.SetScopeRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.SetScopeResponse) => void): grpc.ClientUnaryCall;
    getACL(request: client_auth_auth_pb.GetACLRequest, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.GetACLResponse) => void): grpc.ClientUnaryCall;
    getACL(request: client_auth_auth_pb.GetACLRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.GetACLResponse) => void): grpc.ClientUnaryCall;
    getACL(request: client_auth_auth_pb.GetACLRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.GetACLResponse) => void): grpc.ClientUnaryCall;
    setACL(request: client_auth_auth_pb.SetACLRequest, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.SetACLResponse) => void): grpc.ClientUnaryCall;
    setACL(request: client_auth_auth_pb.SetACLRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.SetACLResponse) => void): grpc.ClientUnaryCall;
    setACL(request: client_auth_auth_pb.SetACLRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.SetACLResponse) => void): grpc.ClientUnaryCall;
    getOIDCLogin(request: client_auth_auth_pb.GetOIDCLoginRequest, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.GetOIDCLoginResponse) => void): grpc.ClientUnaryCall;
    getOIDCLogin(request: client_auth_auth_pb.GetOIDCLoginRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.GetOIDCLoginResponse) => void): grpc.ClientUnaryCall;
    getOIDCLogin(request: client_auth_auth_pb.GetOIDCLoginRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.GetOIDCLoginResponse) => void): grpc.ClientUnaryCall;
    getAuthToken(request: client_auth_auth_pb.GetAuthTokenRequest, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.GetAuthTokenResponse) => void): grpc.ClientUnaryCall;
    getAuthToken(request: client_auth_auth_pb.GetAuthTokenRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.GetAuthTokenResponse) => void): grpc.ClientUnaryCall;
    getAuthToken(request: client_auth_auth_pb.GetAuthTokenRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.GetAuthTokenResponse) => void): grpc.ClientUnaryCall;
    extendAuthToken(request: client_auth_auth_pb.ExtendAuthTokenRequest, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.ExtendAuthTokenResponse) => void): grpc.ClientUnaryCall;
    extendAuthToken(request: client_auth_auth_pb.ExtendAuthTokenRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.ExtendAuthTokenResponse) => void): grpc.ClientUnaryCall;
    extendAuthToken(request: client_auth_auth_pb.ExtendAuthTokenRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.ExtendAuthTokenResponse) => void): grpc.ClientUnaryCall;
    revokeAuthToken(request: client_auth_auth_pb.RevokeAuthTokenRequest, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.RevokeAuthTokenResponse) => void): grpc.ClientUnaryCall;
    revokeAuthToken(request: client_auth_auth_pb.RevokeAuthTokenRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.RevokeAuthTokenResponse) => void): grpc.ClientUnaryCall;
    revokeAuthToken(request: client_auth_auth_pb.RevokeAuthTokenRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.RevokeAuthTokenResponse) => void): grpc.ClientUnaryCall;
    setGroupsForUser(request: client_auth_auth_pb.SetGroupsForUserRequest, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.SetGroupsForUserResponse) => void): grpc.ClientUnaryCall;
    setGroupsForUser(request: client_auth_auth_pb.SetGroupsForUserRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.SetGroupsForUserResponse) => void): grpc.ClientUnaryCall;
    setGroupsForUser(request: client_auth_auth_pb.SetGroupsForUserRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.SetGroupsForUserResponse) => void): grpc.ClientUnaryCall;
    modifyMembers(request: client_auth_auth_pb.ModifyMembersRequest, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.ModifyMembersResponse) => void): grpc.ClientUnaryCall;
    modifyMembers(request: client_auth_auth_pb.ModifyMembersRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.ModifyMembersResponse) => void): grpc.ClientUnaryCall;
    modifyMembers(request: client_auth_auth_pb.ModifyMembersRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.ModifyMembersResponse) => void): grpc.ClientUnaryCall;
    getGroups(request: client_auth_auth_pb.GetGroupsRequest, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.GetGroupsResponse) => void): grpc.ClientUnaryCall;
    getGroups(request: client_auth_auth_pb.GetGroupsRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.GetGroupsResponse) => void): grpc.ClientUnaryCall;
    getGroups(request: client_auth_auth_pb.GetGroupsRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.GetGroupsResponse) => void): grpc.ClientUnaryCall;
    getUsers(request: client_auth_auth_pb.GetUsersRequest, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.GetUsersResponse) => void): grpc.ClientUnaryCall;
    getUsers(request: client_auth_auth_pb.GetUsersRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.GetUsersResponse) => void): grpc.ClientUnaryCall;
    getUsers(request: client_auth_auth_pb.GetUsersRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.GetUsersResponse) => void): grpc.ClientUnaryCall;
    getOneTimePassword(request: client_auth_auth_pb.GetOneTimePasswordRequest, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.GetOneTimePasswordResponse) => void): grpc.ClientUnaryCall;
    getOneTimePassword(request: client_auth_auth_pb.GetOneTimePasswordRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.GetOneTimePasswordResponse) => void): grpc.ClientUnaryCall;
    getOneTimePassword(request: client_auth_auth_pb.GetOneTimePasswordRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.GetOneTimePasswordResponse) => void): grpc.ClientUnaryCall;
}

export class APIClient extends grpc.Client implements IAPIClient {
    constructor(address: string, credentials: grpc.ChannelCredentials, options?: Partial<grpc.ClientOptions>);
    public activate(request: client_auth_auth_pb.ActivateRequest, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.ActivateResponse) => void): grpc.ClientUnaryCall;
    public activate(request: client_auth_auth_pb.ActivateRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.ActivateResponse) => void): grpc.ClientUnaryCall;
    public activate(request: client_auth_auth_pb.ActivateRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.ActivateResponse) => void): grpc.ClientUnaryCall;
    public deactivate(request: client_auth_auth_pb.DeactivateRequest, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.DeactivateResponse) => void): grpc.ClientUnaryCall;
    public deactivate(request: client_auth_auth_pb.DeactivateRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.DeactivateResponse) => void): grpc.ClientUnaryCall;
    public deactivate(request: client_auth_auth_pb.DeactivateRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.DeactivateResponse) => void): grpc.ClientUnaryCall;
    public getConfiguration(request: client_auth_auth_pb.GetConfigurationRequest, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.GetConfigurationResponse) => void): grpc.ClientUnaryCall;
    public getConfiguration(request: client_auth_auth_pb.GetConfigurationRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.GetConfigurationResponse) => void): grpc.ClientUnaryCall;
    public getConfiguration(request: client_auth_auth_pb.GetConfigurationRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.GetConfigurationResponse) => void): grpc.ClientUnaryCall;
    public setConfiguration(request: client_auth_auth_pb.SetConfigurationRequest, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.SetConfigurationResponse) => void): grpc.ClientUnaryCall;
    public setConfiguration(request: client_auth_auth_pb.SetConfigurationRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.SetConfigurationResponse) => void): grpc.ClientUnaryCall;
    public setConfiguration(request: client_auth_auth_pb.SetConfigurationRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.SetConfigurationResponse) => void): grpc.ClientUnaryCall;
    public getAdmins(request: client_auth_auth_pb.GetAdminsRequest, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.GetAdminsResponse) => void): grpc.ClientUnaryCall;
    public getAdmins(request: client_auth_auth_pb.GetAdminsRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.GetAdminsResponse) => void): grpc.ClientUnaryCall;
    public getAdmins(request: client_auth_auth_pb.GetAdminsRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.GetAdminsResponse) => void): grpc.ClientUnaryCall;
    public modifyAdmins(request: client_auth_auth_pb.ModifyAdminsRequest, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.ModifyAdminsResponse) => void): grpc.ClientUnaryCall;
    public modifyAdmins(request: client_auth_auth_pb.ModifyAdminsRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.ModifyAdminsResponse) => void): grpc.ClientUnaryCall;
    public modifyAdmins(request: client_auth_auth_pb.ModifyAdminsRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.ModifyAdminsResponse) => void): grpc.ClientUnaryCall;
    public getClusterRoleBindings(request: client_auth_auth_pb.GetClusterRoleBindingsRequest, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.GetClusterRoleBindingsResponse) => void): grpc.ClientUnaryCall;
    public getClusterRoleBindings(request: client_auth_auth_pb.GetClusterRoleBindingsRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.GetClusterRoleBindingsResponse) => void): grpc.ClientUnaryCall;
    public getClusterRoleBindings(request: client_auth_auth_pb.GetClusterRoleBindingsRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.GetClusterRoleBindingsResponse) => void): grpc.ClientUnaryCall;
    public modifyClusterRoleBinding(request: client_auth_auth_pb.ModifyClusterRoleBindingRequest, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.ModifyClusterRoleBindingResponse) => void): grpc.ClientUnaryCall;
    public modifyClusterRoleBinding(request: client_auth_auth_pb.ModifyClusterRoleBindingRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.ModifyClusterRoleBindingResponse) => void): grpc.ClientUnaryCall;
    public modifyClusterRoleBinding(request: client_auth_auth_pb.ModifyClusterRoleBindingRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.ModifyClusterRoleBindingResponse) => void): grpc.ClientUnaryCall;
    public authenticate(request: client_auth_auth_pb.AuthenticateRequest, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.AuthenticateResponse) => void): grpc.ClientUnaryCall;
    public authenticate(request: client_auth_auth_pb.AuthenticateRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.AuthenticateResponse) => void): grpc.ClientUnaryCall;
    public authenticate(request: client_auth_auth_pb.AuthenticateRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.AuthenticateResponse) => void): grpc.ClientUnaryCall;
    public authorize(request: client_auth_auth_pb.AuthorizeRequest, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.AuthorizeResponse) => void): grpc.ClientUnaryCall;
    public authorize(request: client_auth_auth_pb.AuthorizeRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.AuthorizeResponse) => void): grpc.ClientUnaryCall;
    public authorize(request: client_auth_auth_pb.AuthorizeRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.AuthorizeResponse) => void): grpc.ClientUnaryCall;
    public whoAmI(request: client_auth_auth_pb.WhoAmIRequest, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.WhoAmIResponse) => void): grpc.ClientUnaryCall;
    public whoAmI(request: client_auth_auth_pb.WhoAmIRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.WhoAmIResponse) => void): grpc.ClientUnaryCall;
    public whoAmI(request: client_auth_auth_pb.WhoAmIRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.WhoAmIResponse) => void): grpc.ClientUnaryCall;
    public getScope(request: client_auth_auth_pb.GetScopeRequest, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.GetScopeResponse) => void): grpc.ClientUnaryCall;
    public getScope(request: client_auth_auth_pb.GetScopeRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.GetScopeResponse) => void): grpc.ClientUnaryCall;
    public getScope(request: client_auth_auth_pb.GetScopeRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.GetScopeResponse) => void): grpc.ClientUnaryCall;
    public setScope(request: client_auth_auth_pb.SetScopeRequest, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.SetScopeResponse) => void): grpc.ClientUnaryCall;
    public setScope(request: client_auth_auth_pb.SetScopeRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.SetScopeResponse) => void): grpc.ClientUnaryCall;
    public setScope(request: client_auth_auth_pb.SetScopeRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.SetScopeResponse) => void): grpc.ClientUnaryCall;
    public getACL(request: client_auth_auth_pb.GetACLRequest, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.GetACLResponse) => void): grpc.ClientUnaryCall;
    public getACL(request: client_auth_auth_pb.GetACLRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.GetACLResponse) => void): grpc.ClientUnaryCall;
    public getACL(request: client_auth_auth_pb.GetACLRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.GetACLResponse) => void): grpc.ClientUnaryCall;
    public setACL(request: client_auth_auth_pb.SetACLRequest, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.SetACLResponse) => void): grpc.ClientUnaryCall;
    public setACL(request: client_auth_auth_pb.SetACLRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.SetACLResponse) => void): grpc.ClientUnaryCall;
    public setACL(request: client_auth_auth_pb.SetACLRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.SetACLResponse) => void): grpc.ClientUnaryCall;
    public getOIDCLogin(request: client_auth_auth_pb.GetOIDCLoginRequest, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.GetOIDCLoginResponse) => void): grpc.ClientUnaryCall;
    public getOIDCLogin(request: client_auth_auth_pb.GetOIDCLoginRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.GetOIDCLoginResponse) => void): grpc.ClientUnaryCall;
    public getOIDCLogin(request: client_auth_auth_pb.GetOIDCLoginRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.GetOIDCLoginResponse) => void): grpc.ClientUnaryCall;
    public getAuthToken(request: client_auth_auth_pb.GetAuthTokenRequest, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.GetAuthTokenResponse) => void): grpc.ClientUnaryCall;
    public getAuthToken(request: client_auth_auth_pb.GetAuthTokenRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.GetAuthTokenResponse) => void): grpc.ClientUnaryCall;
    public getAuthToken(request: client_auth_auth_pb.GetAuthTokenRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.GetAuthTokenResponse) => void): grpc.ClientUnaryCall;
    public extendAuthToken(request: client_auth_auth_pb.ExtendAuthTokenRequest, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.ExtendAuthTokenResponse) => void): grpc.ClientUnaryCall;
    public extendAuthToken(request: client_auth_auth_pb.ExtendAuthTokenRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.ExtendAuthTokenResponse) => void): grpc.ClientUnaryCall;
    public extendAuthToken(request: client_auth_auth_pb.ExtendAuthTokenRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.ExtendAuthTokenResponse) => void): grpc.ClientUnaryCall;
    public revokeAuthToken(request: client_auth_auth_pb.RevokeAuthTokenRequest, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.RevokeAuthTokenResponse) => void): grpc.ClientUnaryCall;
    public revokeAuthToken(request: client_auth_auth_pb.RevokeAuthTokenRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.RevokeAuthTokenResponse) => void): grpc.ClientUnaryCall;
    public revokeAuthToken(request: client_auth_auth_pb.RevokeAuthTokenRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.RevokeAuthTokenResponse) => void): grpc.ClientUnaryCall;
    public setGroupsForUser(request: client_auth_auth_pb.SetGroupsForUserRequest, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.SetGroupsForUserResponse) => void): grpc.ClientUnaryCall;
    public setGroupsForUser(request: client_auth_auth_pb.SetGroupsForUserRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.SetGroupsForUserResponse) => void): grpc.ClientUnaryCall;
    public setGroupsForUser(request: client_auth_auth_pb.SetGroupsForUserRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.SetGroupsForUserResponse) => void): grpc.ClientUnaryCall;
    public modifyMembers(request: client_auth_auth_pb.ModifyMembersRequest, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.ModifyMembersResponse) => void): grpc.ClientUnaryCall;
    public modifyMembers(request: client_auth_auth_pb.ModifyMembersRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.ModifyMembersResponse) => void): grpc.ClientUnaryCall;
    public modifyMembers(request: client_auth_auth_pb.ModifyMembersRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.ModifyMembersResponse) => void): grpc.ClientUnaryCall;
    public getGroups(request: client_auth_auth_pb.GetGroupsRequest, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.GetGroupsResponse) => void): grpc.ClientUnaryCall;
    public getGroups(request: client_auth_auth_pb.GetGroupsRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.GetGroupsResponse) => void): grpc.ClientUnaryCall;
    public getGroups(request: client_auth_auth_pb.GetGroupsRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.GetGroupsResponse) => void): grpc.ClientUnaryCall;
    public getUsers(request: client_auth_auth_pb.GetUsersRequest, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.GetUsersResponse) => void): grpc.ClientUnaryCall;
    public getUsers(request: client_auth_auth_pb.GetUsersRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.GetUsersResponse) => void): grpc.ClientUnaryCall;
    public getUsers(request: client_auth_auth_pb.GetUsersRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.GetUsersResponse) => void): grpc.ClientUnaryCall;
    public getOneTimePassword(request: client_auth_auth_pb.GetOneTimePasswordRequest, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.GetOneTimePasswordResponse) => void): grpc.ClientUnaryCall;
    public getOneTimePassword(request: client_auth_auth_pb.GetOneTimePasswordRequest, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.GetOneTimePasswordResponse) => void): grpc.ClientUnaryCall;
    public getOneTimePassword(request: client_auth_auth_pb.GetOneTimePasswordRequest, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: client_auth_auth_pb.GetOneTimePasswordResponse) => void): grpc.ClientUnaryCall;
}
