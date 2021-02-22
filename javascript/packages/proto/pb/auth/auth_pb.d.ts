// package: auth
// file: auth/auth.proto

/* tslint:disable */
/* eslint-disable */

import * as jspb from "google-protobuf";
import * as gogoproto_gogo_pb from "../gogoproto/gogo_pb";
import * as google_protobuf_timestamp_pb from "google-protobuf/google/protobuf/timestamp_pb";

export class ActivateRequest extends jspb.Message { 
    getRootToken(): string;
    setRootToken(value: string): ActivateRequest;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ActivateRequest.AsObject;
    static toObject(includeInstance: boolean, msg: ActivateRequest): ActivateRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ActivateRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ActivateRequest;
    static deserializeBinaryFromReader(message: ActivateRequest, reader: jspb.BinaryReader): ActivateRequest;
}

export namespace ActivateRequest {
    export type AsObject = {
        rootToken: string,
    }
}

export class ActivateResponse extends jspb.Message { 
    getPachToken(): string;
    setPachToken(value: string): ActivateResponse;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ActivateResponse.AsObject;
    static toObject(includeInstance: boolean, msg: ActivateResponse): ActivateResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ActivateResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ActivateResponse;
    static deserializeBinaryFromReader(message: ActivateResponse, reader: jspb.BinaryReader): ActivateResponse;
}

export namespace ActivateResponse {
    export type AsObject = {
        pachToken: string,
    }
}

export class DeactivateRequest extends jspb.Message { 

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): DeactivateRequest.AsObject;
    static toObject(includeInstance: boolean, msg: DeactivateRequest): DeactivateRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: DeactivateRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): DeactivateRequest;
    static deserializeBinaryFromReader(message: DeactivateRequest, reader: jspb.BinaryReader): DeactivateRequest;
}

export namespace DeactivateRequest {
    export type AsObject = {
    }
}

export class DeactivateResponse extends jspb.Message { 

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): DeactivateResponse.AsObject;
    static toObject(includeInstance: boolean, msg: DeactivateResponse): DeactivateResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: DeactivateResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): DeactivateResponse;
    static deserializeBinaryFromReader(message: DeactivateResponse, reader: jspb.BinaryReader): DeactivateResponse;
}

export namespace DeactivateResponse {
    export type AsObject = {
    }
}

export class OIDCConfig extends jspb.Message { 
    getIssuer(): string;
    setIssuer(value: string): OIDCConfig;

    getClientId(): string;
    setClientId(value: string): OIDCConfig;

    getClientSecret(): string;
    setClientSecret(value: string): OIDCConfig;

    getRedirectUri(): string;
    setRedirectUri(value: string): OIDCConfig;

    clearAdditionalScopesList(): void;
    getAdditionalScopesList(): Array<string>;
    setAdditionalScopesList(value: Array<string>): OIDCConfig;
    addAdditionalScopes(value: string, index?: number): string;

    getIgnoreEmailVerified(): boolean;
    setIgnoreEmailVerified(value: boolean): OIDCConfig;

    getLocalhostIssuer(): boolean;
    setLocalhostIssuer(value: boolean): OIDCConfig;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): OIDCConfig.AsObject;
    static toObject(includeInstance: boolean, msg: OIDCConfig): OIDCConfig.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: OIDCConfig, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): OIDCConfig;
    static deserializeBinaryFromReader(message: OIDCConfig, reader: jspb.BinaryReader): OIDCConfig;
}

export namespace OIDCConfig {
    export type AsObject = {
        issuer: string,
        clientId: string,
        clientSecret: string,
        redirectUri: string,
        additionalScopesList: Array<string>,
        ignoreEmailVerified: boolean,
        localhostIssuer: boolean,
    }
}

export class GetConfigurationRequest extends jspb.Message { 

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GetConfigurationRequest.AsObject;
    static toObject(includeInstance: boolean, msg: GetConfigurationRequest): GetConfigurationRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GetConfigurationRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GetConfigurationRequest;
    static deserializeBinaryFromReader(message: GetConfigurationRequest, reader: jspb.BinaryReader): GetConfigurationRequest;
}

export namespace GetConfigurationRequest {
    export type AsObject = {
    }
}

export class GetConfigurationResponse extends jspb.Message { 

    hasConfiguration(): boolean;
    clearConfiguration(): void;
    getConfiguration(): OIDCConfig | undefined;
    setConfiguration(value?: OIDCConfig): GetConfigurationResponse;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GetConfigurationResponse.AsObject;
    static toObject(includeInstance: boolean, msg: GetConfigurationResponse): GetConfigurationResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GetConfigurationResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GetConfigurationResponse;
    static deserializeBinaryFromReader(message: GetConfigurationResponse, reader: jspb.BinaryReader): GetConfigurationResponse;
}

export namespace GetConfigurationResponse {
    export type AsObject = {
        configuration?: OIDCConfig.AsObject,
    }
}

export class SetConfigurationRequest extends jspb.Message { 

    hasConfiguration(): boolean;
    clearConfiguration(): void;
    getConfiguration(): OIDCConfig | undefined;
    setConfiguration(value?: OIDCConfig): SetConfigurationRequest;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): SetConfigurationRequest.AsObject;
    static toObject(includeInstance: boolean, msg: SetConfigurationRequest): SetConfigurationRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: SetConfigurationRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): SetConfigurationRequest;
    static deserializeBinaryFromReader(message: SetConfigurationRequest, reader: jspb.BinaryReader): SetConfigurationRequest;
}

export namespace SetConfigurationRequest {
    export type AsObject = {
        configuration?: OIDCConfig.AsObject,
    }
}

export class SetConfigurationResponse extends jspb.Message { 

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): SetConfigurationResponse.AsObject;
    static toObject(includeInstance: boolean, msg: SetConfigurationResponse): SetConfigurationResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: SetConfigurationResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): SetConfigurationResponse;
    static deserializeBinaryFromReader(message: SetConfigurationResponse, reader: jspb.BinaryReader): SetConfigurationResponse;
}

export namespace SetConfigurationResponse {
    export type AsObject = {
    }
}

export class ClusterRoles extends jspb.Message { 
    clearRolesList(): void;
    getRolesList(): Array<ClusterRole>;
    setRolesList(value: Array<ClusterRole>): ClusterRoles;
    addRoles(value: ClusterRole, index?: number): ClusterRole;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ClusterRoles.AsObject;
    static toObject(includeInstance: boolean, msg: ClusterRoles): ClusterRoles.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ClusterRoles, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ClusterRoles;
    static deserializeBinaryFromReader(message: ClusterRoles, reader: jspb.BinaryReader): ClusterRoles;
}

export namespace ClusterRoles {
    export type AsObject = {
        rolesList: Array<ClusterRole>,
    }
}

export class GetClusterRoleBindingsRequest extends jspb.Message { 

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GetClusterRoleBindingsRequest.AsObject;
    static toObject(includeInstance: boolean, msg: GetClusterRoleBindingsRequest): GetClusterRoleBindingsRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GetClusterRoleBindingsRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GetClusterRoleBindingsRequest;
    static deserializeBinaryFromReader(message: GetClusterRoleBindingsRequest, reader: jspb.BinaryReader): GetClusterRoleBindingsRequest;
}

export namespace GetClusterRoleBindingsRequest {
    export type AsObject = {
    }
}

export class GetClusterRoleBindingsResponse extends jspb.Message { 

    getBindingsMap(): jspb.Map<string, ClusterRoles>;
    clearBindingsMap(): void;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GetClusterRoleBindingsResponse.AsObject;
    static toObject(includeInstance: boolean, msg: GetClusterRoleBindingsResponse): GetClusterRoleBindingsResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GetClusterRoleBindingsResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GetClusterRoleBindingsResponse;
    static deserializeBinaryFromReader(message: GetClusterRoleBindingsResponse, reader: jspb.BinaryReader): GetClusterRoleBindingsResponse;
}

export namespace GetClusterRoleBindingsResponse {
    export type AsObject = {

        bindingsMap: Array<[string, ClusterRoles.AsObject]>,
    }
}

export class ModifyClusterRoleBindingRequest extends jspb.Message { 
    getPrincipal(): string;
    setPrincipal(value: string): ModifyClusterRoleBindingRequest;


    hasRoles(): boolean;
    clearRoles(): void;
    getRoles(): ClusterRoles | undefined;
    setRoles(value?: ClusterRoles): ModifyClusterRoleBindingRequest;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ModifyClusterRoleBindingRequest.AsObject;
    static toObject(includeInstance: boolean, msg: ModifyClusterRoleBindingRequest): ModifyClusterRoleBindingRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ModifyClusterRoleBindingRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ModifyClusterRoleBindingRequest;
    static deserializeBinaryFromReader(message: ModifyClusterRoleBindingRequest, reader: jspb.BinaryReader): ModifyClusterRoleBindingRequest;
}

export namespace ModifyClusterRoleBindingRequest {
    export type AsObject = {
        principal: string,
        roles?: ClusterRoles.AsObject,
    }
}

export class ModifyClusterRoleBindingResponse extends jspb.Message { 

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ModifyClusterRoleBindingResponse.AsObject;
    static toObject(includeInstance: boolean, msg: ModifyClusterRoleBindingResponse): ModifyClusterRoleBindingResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ModifyClusterRoleBindingResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ModifyClusterRoleBindingResponse;
    static deserializeBinaryFromReader(message: ModifyClusterRoleBindingResponse, reader: jspb.BinaryReader): ModifyClusterRoleBindingResponse;
}

export namespace ModifyClusterRoleBindingResponse {
    export type AsObject = {
    }
}

export class GetAdminsRequest extends jspb.Message { 

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GetAdminsRequest.AsObject;
    static toObject(includeInstance: boolean, msg: GetAdminsRequest): GetAdminsRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GetAdminsRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GetAdminsRequest;
    static deserializeBinaryFromReader(message: GetAdminsRequest, reader: jspb.BinaryReader): GetAdminsRequest;
}

export namespace GetAdminsRequest {
    export type AsObject = {
    }
}

export class GetAdminsResponse extends jspb.Message { 
    clearAdminsList(): void;
    getAdminsList(): Array<string>;
    setAdminsList(value: Array<string>): GetAdminsResponse;
    addAdmins(value: string, index?: number): string;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GetAdminsResponse.AsObject;
    static toObject(includeInstance: boolean, msg: GetAdminsResponse): GetAdminsResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GetAdminsResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GetAdminsResponse;
    static deserializeBinaryFromReader(message: GetAdminsResponse, reader: jspb.BinaryReader): GetAdminsResponse;
}

export namespace GetAdminsResponse {
    export type AsObject = {
        adminsList: Array<string>,
    }
}

export class ModifyAdminsRequest extends jspb.Message { 
    clearAddList(): void;
    getAddList(): Array<string>;
    setAddList(value: Array<string>): ModifyAdminsRequest;
    addAdd(value: string, index?: number): string;

    clearRemoveList(): void;
    getRemoveList(): Array<string>;
    setRemoveList(value: Array<string>): ModifyAdminsRequest;
    addRemove(value: string, index?: number): string;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ModifyAdminsRequest.AsObject;
    static toObject(includeInstance: boolean, msg: ModifyAdminsRequest): ModifyAdminsRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ModifyAdminsRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ModifyAdminsRequest;
    static deserializeBinaryFromReader(message: ModifyAdminsRequest, reader: jspb.BinaryReader): ModifyAdminsRequest;
}

export namespace ModifyAdminsRequest {
    export type AsObject = {
        addList: Array<string>,
        removeList: Array<string>,
    }
}

export class ModifyAdminsResponse extends jspb.Message { 

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ModifyAdminsResponse.AsObject;
    static toObject(includeInstance: boolean, msg: ModifyAdminsResponse): ModifyAdminsResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ModifyAdminsResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ModifyAdminsResponse;
    static deserializeBinaryFromReader(message: ModifyAdminsResponse, reader: jspb.BinaryReader): ModifyAdminsResponse;
}

export namespace ModifyAdminsResponse {
    export type AsObject = {
    }
}

export class TokenInfo extends jspb.Message { 
    getSubject(): string;
    setSubject(value: string): TokenInfo;

    getSource(): TokenInfo.TokenSource;
    setSource(value: TokenInfo.TokenSource): TokenInfo;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): TokenInfo.AsObject;
    static toObject(includeInstance: boolean, msg: TokenInfo): TokenInfo.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: TokenInfo, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): TokenInfo;
    static deserializeBinaryFromReader(message: TokenInfo, reader: jspb.BinaryReader): TokenInfo;
}

export namespace TokenInfo {
    export type AsObject = {
        subject: string,
        source: TokenInfo.TokenSource,
    }

    export enum TokenSource {
    INVALID = 0,
    AUTHENTICATE = 1,
    GET_TOKEN = 2,
    }

}

export class AuthenticateRequest extends jspb.Message { 
    getOidcState(): string;
    setOidcState(value: string): AuthenticateRequest;

    getIdToken(): string;
    setIdToken(value: string): AuthenticateRequest;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): AuthenticateRequest.AsObject;
    static toObject(includeInstance: boolean, msg: AuthenticateRequest): AuthenticateRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: AuthenticateRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): AuthenticateRequest;
    static deserializeBinaryFromReader(message: AuthenticateRequest, reader: jspb.BinaryReader): AuthenticateRequest;
}

export namespace AuthenticateRequest {
    export type AsObject = {
        oidcState: string,
        idToken: string,
    }
}

export class AuthenticateResponse extends jspb.Message { 
    getPachToken(): string;
    setPachToken(value: string): AuthenticateResponse;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): AuthenticateResponse.AsObject;
    static toObject(includeInstance: boolean, msg: AuthenticateResponse): AuthenticateResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: AuthenticateResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): AuthenticateResponse;
    static deserializeBinaryFromReader(message: AuthenticateResponse, reader: jspb.BinaryReader): AuthenticateResponse;
}

export namespace AuthenticateResponse {
    export type AsObject = {
        pachToken: string,
    }
}

export class WhoAmIRequest extends jspb.Message { 

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): WhoAmIRequest.AsObject;
    static toObject(includeInstance: boolean, msg: WhoAmIRequest): WhoAmIRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: WhoAmIRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): WhoAmIRequest;
    static deserializeBinaryFromReader(message: WhoAmIRequest, reader: jspb.BinaryReader): WhoAmIRequest;
}

export namespace WhoAmIRequest {
    export type AsObject = {
    }
}

export class WhoAmIResponse extends jspb.Message { 
    getUsername(): string;
    setUsername(value: string): WhoAmIResponse;

    getIsAdmin(): boolean;
    setIsAdmin(value: boolean): WhoAmIResponse;

    getTtl(): number;
    setTtl(value: number): WhoAmIResponse;


    hasClusterRoles(): boolean;
    clearClusterRoles(): void;
    getClusterRoles(): ClusterRoles | undefined;
    setClusterRoles(value?: ClusterRoles): WhoAmIResponse;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): WhoAmIResponse.AsObject;
    static toObject(includeInstance: boolean, msg: WhoAmIResponse): WhoAmIResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: WhoAmIResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): WhoAmIResponse;
    static deserializeBinaryFromReader(message: WhoAmIResponse, reader: jspb.BinaryReader): WhoAmIResponse;
}

export namespace WhoAmIResponse {
    export type AsObject = {
        username: string,
        isAdmin: boolean,
        ttl: number,
        clusterRoles?: ClusterRoles.AsObject,
    }
}

export class ACL extends jspb.Message { 

    getEntriesMap(): jspb.Map<string, Scope>;
    clearEntriesMap(): void;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ACL.AsObject;
    static toObject(includeInstance: boolean, msg: ACL): ACL.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ACL, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ACL;
    static deserializeBinaryFromReader(message: ACL, reader: jspb.BinaryReader): ACL;
}

export namespace ACL {
    export type AsObject = {

        entriesMap: Array<[string, Scope]>,
    }
}

export class Users extends jspb.Message { 

    getUsernamesMap(): jspb.Map<string, boolean>;
    clearUsernamesMap(): void;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Users.AsObject;
    static toObject(includeInstance: boolean, msg: Users): Users.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: Users, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Users;
    static deserializeBinaryFromReader(message: Users, reader: jspb.BinaryReader): Users;
}

export namespace Users {
    export type AsObject = {

        usernamesMap: Array<[string, boolean]>,
    }
}

export class Groups extends jspb.Message { 

    getGroupsMap(): jspb.Map<string, boolean>;
    clearGroupsMap(): void;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Groups.AsObject;
    static toObject(includeInstance: boolean, msg: Groups): Groups.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: Groups, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Groups;
    static deserializeBinaryFromReader(message: Groups, reader: jspb.BinaryReader): Groups;
}

export namespace Groups {
    export type AsObject = {

        groupsMap: Array<[string, boolean]>,
    }
}

export class AuthorizeRequest extends jspb.Message { 
    getRepo(): string;
    setRepo(value: string): AuthorizeRequest;

    getScope(): Scope;
    setScope(value: Scope): AuthorizeRequest;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): AuthorizeRequest.AsObject;
    static toObject(includeInstance: boolean, msg: AuthorizeRequest): AuthorizeRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: AuthorizeRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): AuthorizeRequest;
    static deserializeBinaryFromReader(message: AuthorizeRequest, reader: jspb.BinaryReader): AuthorizeRequest;
}

export namespace AuthorizeRequest {
    export type AsObject = {
        repo: string,
        scope: Scope,
    }
}

export class AuthorizeResponse extends jspb.Message { 
    getAuthorized(): boolean;
    setAuthorized(value: boolean): AuthorizeResponse;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): AuthorizeResponse.AsObject;
    static toObject(includeInstance: boolean, msg: AuthorizeResponse): AuthorizeResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: AuthorizeResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): AuthorizeResponse;
    static deserializeBinaryFromReader(message: AuthorizeResponse, reader: jspb.BinaryReader): AuthorizeResponse;
}

export namespace AuthorizeResponse {
    export type AsObject = {
        authorized: boolean,
    }
}

export class GetScopeRequest extends jspb.Message { 
    getUsername(): string;
    setUsername(value: string): GetScopeRequest;

    clearReposList(): void;
    getReposList(): Array<string>;
    setReposList(value: Array<string>): GetScopeRequest;
    addRepos(value: string, index?: number): string;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GetScopeRequest.AsObject;
    static toObject(includeInstance: boolean, msg: GetScopeRequest): GetScopeRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GetScopeRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GetScopeRequest;
    static deserializeBinaryFromReader(message: GetScopeRequest, reader: jspb.BinaryReader): GetScopeRequest;
}

export namespace GetScopeRequest {
    export type AsObject = {
        username: string,
        reposList: Array<string>,
    }
}

export class GetScopeResponse extends jspb.Message { 
    clearScopesList(): void;
    getScopesList(): Array<Scope>;
    setScopesList(value: Array<Scope>): GetScopeResponse;
    addScopes(value: Scope, index?: number): Scope;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GetScopeResponse.AsObject;
    static toObject(includeInstance: boolean, msg: GetScopeResponse): GetScopeResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GetScopeResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GetScopeResponse;
    static deserializeBinaryFromReader(message: GetScopeResponse, reader: jspb.BinaryReader): GetScopeResponse;
}

export namespace GetScopeResponse {
    export type AsObject = {
        scopesList: Array<Scope>,
    }
}

export class SetScopeRequest extends jspb.Message { 
    getUsername(): string;
    setUsername(value: string): SetScopeRequest;

    getRepo(): string;
    setRepo(value: string): SetScopeRequest;

    getScope(): Scope;
    setScope(value: Scope): SetScopeRequest;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): SetScopeRequest.AsObject;
    static toObject(includeInstance: boolean, msg: SetScopeRequest): SetScopeRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: SetScopeRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): SetScopeRequest;
    static deserializeBinaryFromReader(message: SetScopeRequest, reader: jspb.BinaryReader): SetScopeRequest;
}

export namespace SetScopeRequest {
    export type AsObject = {
        username: string,
        repo: string,
        scope: Scope,
    }
}

export class SetScopeResponse extends jspb.Message { 

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): SetScopeResponse.AsObject;
    static toObject(includeInstance: boolean, msg: SetScopeResponse): SetScopeResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: SetScopeResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): SetScopeResponse;
    static deserializeBinaryFromReader(message: SetScopeResponse, reader: jspb.BinaryReader): SetScopeResponse;
}

export namespace SetScopeResponse {
    export type AsObject = {
    }
}

export class GetACLRequest extends jspb.Message { 
    getRepo(): string;
    setRepo(value: string): GetACLRequest;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GetACLRequest.AsObject;
    static toObject(includeInstance: boolean, msg: GetACLRequest): GetACLRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GetACLRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GetACLRequest;
    static deserializeBinaryFromReader(message: GetACLRequest, reader: jspb.BinaryReader): GetACLRequest;
}

export namespace GetACLRequest {
    export type AsObject = {
        repo: string,
    }
}

export class ACLEntry extends jspb.Message { 
    getUsername(): string;
    setUsername(value: string): ACLEntry;

    getScope(): Scope;
    setScope(value: Scope): ACLEntry;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ACLEntry.AsObject;
    static toObject(includeInstance: boolean, msg: ACLEntry): ACLEntry.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ACLEntry, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ACLEntry;
    static deserializeBinaryFromReader(message: ACLEntry, reader: jspb.BinaryReader): ACLEntry;
}

export namespace ACLEntry {
    export type AsObject = {
        username: string,
        scope: Scope,
    }
}

export class GetACLResponse extends jspb.Message { 
    clearEntriesList(): void;
    getEntriesList(): Array<ACLEntry>;
    setEntriesList(value: Array<ACLEntry>): GetACLResponse;
    addEntries(value?: ACLEntry, index?: number): ACLEntry;

    clearRobotEntriesList(): void;
    getRobotEntriesList(): Array<ACLEntry>;
    setRobotEntriesList(value: Array<ACLEntry>): GetACLResponse;
    addRobotEntries(value?: ACLEntry, index?: number): ACLEntry;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GetACLResponse.AsObject;
    static toObject(includeInstance: boolean, msg: GetACLResponse): GetACLResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GetACLResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GetACLResponse;
    static deserializeBinaryFromReader(message: GetACLResponse, reader: jspb.BinaryReader): GetACLResponse;
}

export namespace GetACLResponse {
    export type AsObject = {
        entriesList: Array<ACLEntry.AsObject>,
        robotEntriesList: Array<ACLEntry.AsObject>,
    }
}

export class SetACLRequest extends jspb.Message { 
    getRepo(): string;
    setRepo(value: string): SetACLRequest;

    clearEntriesList(): void;
    getEntriesList(): Array<ACLEntry>;
    setEntriesList(value: Array<ACLEntry>): SetACLRequest;
    addEntries(value?: ACLEntry, index?: number): ACLEntry;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): SetACLRequest.AsObject;
    static toObject(includeInstance: boolean, msg: SetACLRequest): SetACLRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: SetACLRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): SetACLRequest;
    static deserializeBinaryFromReader(message: SetACLRequest, reader: jspb.BinaryReader): SetACLRequest;
}

export namespace SetACLRequest {
    export type AsObject = {
        repo: string,
        entriesList: Array<ACLEntry.AsObject>,
    }
}

export class SetACLResponse extends jspb.Message { 

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): SetACLResponse.AsObject;
    static toObject(includeInstance: boolean, msg: SetACLResponse): SetACLResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: SetACLResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): SetACLResponse;
    static deserializeBinaryFromReader(message: SetACLResponse, reader: jspb.BinaryReader): SetACLResponse;
}

export namespace SetACLResponse {
    export type AsObject = {
    }
}

export class SessionInfo extends jspb.Message { 
    getNonce(): string;
    setNonce(value: string): SessionInfo;

    getEmail(): string;
    setEmail(value: string): SessionInfo;

    getConversionErr(): boolean;
    setConversionErr(value: boolean): SessionInfo;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): SessionInfo.AsObject;
    static toObject(includeInstance: boolean, msg: SessionInfo): SessionInfo.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: SessionInfo, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): SessionInfo;
    static deserializeBinaryFromReader(message: SessionInfo, reader: jspb.BinaryReader): SessionInfo;
}

export namespace SessionInfo {
    export type AsObject = {
        nonce: string,
        email: string,
        conversionErr: boolean,
    }
}

export class GetOIDCLoginRequest extends jspb.Message { 

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GetOIDCLoginRequest.AsObject;
    static toObject(includeInstance: boolean, msg: GetOIDCLoginRequest): GetOIDCLoginRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GetOIDCLoginRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GetOIDCLoginRequest;
    static deserializeBinaryFromReader(message: GetOIDCLoginRequest, reader: jspb.BinaryReader): GetOIDCLoginRequest;
}

export namespace GetOIDCLoginRequest {
    export type AsObject = {
    }
}

export class GetOIDCLoginResponse extends jspb.Message { 
    getLoginUrl(): string;
    setLoginUrl(value: string): GetOIDCLoginResponse;

    getState(): string;
    setState(value: string): GetOIDCLoginResponse;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GetOIDCLoginResponse.AsObject;
    static toObject(includeInstance: boolean, msg: GetOIDCLoginResponse): GetOIDCLoginResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GetOIDCLoginResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GetOIDCLoginResponse;
    static deserializeBinaryFromReader(message: GetOIDCLoginResponse, reader: jspb.BinaryReader): GetOIDCLoginResponse;
}

export namespace GetOIDCLoginResponse {
    export type AsObject = {
        loginUrl: string,
        state: string,
    }
}

export class GetAuthTokenRequest extends jspb.Message { 
    getSubject(): string;
    setSubject(value: string): GetAuthTokenRequest;

    getTtl(): number;
    setTtl(value: number): GetAuthTokenRequest;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GetAuthTokenRequest.AsObject;
    static toObject(includeInstance: boolean, msg: GetAuthTokenRequest): GetAuthTokenRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GetAuthTokenRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GetAuthTokenRequest;
    static deserializeBinaryFromReader(message: GetAuthTokenRequest, reader: jspb.BinaryReader): GetAuthTokenRequest;
}

export namespace GetAuthTokenRequest {
    export type AsObject = {
        subject: string,
        ttl: number,
    }
}

export class GetAuthTokenResponse extends jspb.Message { 
    getSubject(): string;
    setSubject(value: string): GetAuthTokenResponse;

    getToken(): string;
    setToken(value: string): GetAuthTokenResponse;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GetAuthTokenResponse.AsObject;
    static toObject(includeInstance: boolean, msg: GetAuthTokenResponse): GetAuthTokenResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GetAuthTokenResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GetAuthTokenResponse;
    static deserializeBinaryFromReader(message: GetAuthTokenResponse, reader: jspb.BinaryReader): GetAuthTokenResponse;
}

export namespace GetAuthTokenResponse {
    export type AsObject = {
        subject: string,
        token: string,
    }
}

export class ExtendAuthTokenRequest extends jspb.Message { 
    getToken(): string;
    setToken(value: string): ExtendAuthTokenRequest;

    getTtl(): number;
    setTtl(value: number): ExtendAuthTokenRequest;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ExtendAuthTokenRequest.AsObject;
    static toObject(includeInstance: boolean, msg: ExtendAuthTokenRequest): ExtendAuthTokenRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ExtendAuthTokenRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ExtendAuthTokenRequest;
    static deserializeBinaryFromReader(message: ExtendAuthTokenRequest, reader: jspb.BinaryReader): ExtendAuthTokenRequest;
}

export namespace ExtendAuthTokenRequest {
    export type AsObject = {
        token: string,
        ttl: number,
    }
}

export class ExtendAuthTokenResponse extends jspb.Message { 

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ExtendAuthTokenResponse.AsObject;
    static toObject(includeInstance: boolean, msg: ExtendAuthTokenResponse): ExtendAuthTokenResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ExtendAuthTokenResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ExtendAuthTokenResponse;
    static deserializeBinaryFromReader(message: ExtendAuthTokenResponse, reader: jspb.BinaryReader): ExtendAuthTokenResponse;
}

export namespace ExtendAuthTokenResponse {
    export type AsObject = {
    }
}

export class RevokeAuthTokenRequest extends jspb.Message { 
    getToken(): string;
    setToken(value: string): RevokeAuthTokenRequest;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): RevokeAuthTokenRequest.AsObject;
    static toObject(includeInstance: boolean, msg: RevokeAuthTokenRequest): RevokeAuthTokenRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: RevokeAuthTokenRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): RevokeAuthTokenRequest;
    static deserializeBinaryFromReader(message: RevokeAuthTokenRequest, reader: jspb.BinaryReader): RevokeAuthTokenRequest;
}

export namespace RevokeAuthTokenRequest {
    export type AsObject = {
        token: string,
    }
}

export class RevokeAuthTokenResponse extends jspb.Message { 

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): RevokeAuthTokenResponse.AsObject;
    static toObject(includeInstance: boolean, msg: RevokeAuthTokenResponse): RevokeAuthTokenResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: RevokeAuthTokenResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): RevokeAuthTokenResponse;
    static deserializeBinaryFromReader(message: RevokeAuthTokenResponse, reader: jspb.BinaryReader): RevokeAuthTokenResponse;
}

export namespace RevokeAuthTokenResponse {
    export type AsObject = {
    }
}

export class SetGroupsForUserRequest extends jspb.Message { 
    getUsername(): string;
    setUsername(value: string): SetGroupsForUserRequest;

    clearGroupsList(): void;
    getGroupsList(): Array<string>;
    setGroupsList(value: Array<string>): SetGroupsForUserRequest;
    addGroups(value: string, index?: number): string;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): SetGroupsForUserRequest.AsObject;
    static toObject(includeInstance: boolean, msg: SetGroupsForUserRequest): SetGroupsForUserRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: SetGroupsForUserRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): SetGroupsForUserRequest;
    static deserializeBinaryFromReader(message: SetGroupsForUserRequest, reader: jspb.BinaryReader): SetGroupsForUserRequest;
}

export namespace SetGroupsForUserRequest {
    export type AsObject = {
        username: string,
        groupsList: Array<string>,
    }
}

export class SetGroupsForUserResponse extends jspb.Message { 

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): SetGroupsForUserResponse.AsObject;
    static toObject(includeInstance: boolean, msg: SetGroupsForUserResponse): SetGroupsForUserResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: SetGroupsForUserResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): SetGroupsForUserResponse;
    static deserializeBinaryFromReader(message: SetGroupsForUserResponse, reader: jspb.BinaryReader): SetGroupsForUserResponse;
}

export namespace SetGroupsForUserResponse {
    export type AsObject = {
    }
}

export class ModifyMembersRequest extends jspb.Message { 
    getGroup(): string;
    setGroup(value: string): ModifyMembersRequest;

    clearAddList(): void;
    getAddList(): Array<string>;
    setAddList(value: Array<string>): ModifyMembersRequest;
    addAdd(value: string, index?: number): string;

    clearRemoveList(): void;
    getRemoveList(): Array<string>;
    setRemoveList(value: Array<string>): ModifyMembersRequest;
    addRemove(value: string, index?: number): string;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ModifyMembersRequest.AsObject;
    static toObject(includeInstance: boolean, msg: ModifyMembersRequest): ModifyMembersRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ModifyMembersRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ModifyMembersRequest;
    static deserializeBinaryFromReader(message: ModifyMembersRequest, reader: jspb.BinaryReader): ModifyMembersRequest;
}

export namespace ModifyMembersRequest {
    export type AsObject = {
        group: string,
        addList: Array<string>,
        removeList: Array<string>,
    }
}

export class ModifyMembersResponse extends jspb.Message { 

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ModifyMembersResponse.AsObject;
    static toObject(includeInstance: boolean, msg: ModifyMembersResponse): ModifyMembersResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ModifyMembersResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ModifyMembersResponse;
    static deserializeBinaryFromReader(message: ModifyMembersResponse, reader: jspb.BinaryReader): ModifyMembersResponse;
}

export namespace ModifyMembersResponse {
    export type AsObject = {
    }
}

export class GetGroupsRequest extends jspb.Message { 
    getUsername(): string;
    setUsername(value: string): GetGroupsRequest;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GetGroupsRequest.AsObject;
    static toObject(includeInstance: boolean, msg: GetGroupsRequest): GetGroupsRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GetGroupsRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GetGroupsRequest;
    static deserializeBinaryFromReader(message: GetGroupsRequest, reader: jspb.BinaryReader): GetGroupsRequest;
}

export namespace GetGroupsRequest {
    export type AsObject = {
        username: string,
    }
}

export class GetGroupsResponse extends jspb.Message { 
    clearGroupsList(): void;
    getGroupsList(): Array<string>;
    setGroupsList(value: Array<string>): GetGroupsResponse;
    addGroups(value: string, index?: number): string;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GetGroupsResponse.AsObject;
    static toObject(includeInstance: boolean, msg: GetGroupsResponse): GetGroupsResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GetGroupsResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GetGroupsResponse;
    static deserializeBinaryFromReader(message: GetGroupsResponse, reader: jspb.BinaryReader): GetGroupsResponse;
}

export namespace GetGroupsResponse {
    export type AsObject = {
        groupsList: Array<string>,
    }
}

export class GetUsersRequest extends jspb.Message { 
    getGroup(): string;
    setGroup(value: string): GetUsersRequest;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GetUsersRequest.AsObject;
    static toObject(includeInstance: boolean, msg: GetUsersRequest): GetUsersRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GetUsersRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GetUsersRequest;
    static deserializeBinaryFromReader(message: GetUsersRequest, reader: jspb.BinaryReader): GetUsersRequest;
}

export namespace GetUsersRequest {
    export type AsObject = {
        group: string,
    }
}

export class GetUsersResponse extends jspb.Message { 
    clearUsernamesList(): void;
    getUsernamesList(): Array<string>;
    setUsernamesList(value: Array<string>): GetUsersResponse;
    addUsernames(value: string, index?: number): string;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GetUsersResponse.AsObject;
    static toObject(includeInstance: boolean, msg: GetUsersResponse): GetUsersResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GetUsersResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GetUsersResponse;
    static deserializeBinaryFromReader(message: GetUsersResponse, reader: jspb.BinaryReader): GetUsersResponse;
}

export namespace GetUsersResponse {
    export type AsObject = {
        usernamesList: Array<string>,
    }
}

export class HashedAuthToken extends jspb.Message { 
    getHashedToken(): string;
    setHashedToken(value: string): HashedAuthToken;


    hasTokenInfo(): boolean;
    clearTokenInfo(): void;
    getTokenInfo(): TokenInfo | undefined;
    setTokenInfo(value?: TokenInfo): HashedAuthToken;


    hasExpiration(): boolean;
    clearExpiration(): void;
    getExpiration(): google_protobuf_timestamp_pb.Timestamp | undefined;
    setExpiration(value?: google_protobuf_timestamp_pb.Timestamp): HashedAuthToken;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): HashedAuthToken.AsObject;
    static toObject(includeInstance: boolean, msg: HashedAuthToken): HashedAuthToken.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: HashedAuthToken, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): HashedAuthToken;
    static deserializeBinaryFromReader(message: HashedAuthToken, reader: jspb.BinaryReader): HashedAuthToken;
}

export namespace HashedAuthToken {
    export type AsObject = {
        hashedToken: string,
        tokenInfo?: TokenInfo.AsObject,
        expiration?: google_protobuf_timestamp_pb.Timestamp.AsObject,
    }
}

export class ExtractAuthTokensRequest extends jspb.Message { 

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ExtractAuthTokensRequest.AsObject;
    static toObject(includeInstance: boolean, msg: ExtractAuthTokensRequest): ExtractAuthTokensRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ExtractAuthTokensRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ExtractAuthTokensRequest;
    static deserializeBinaryFromReader(message: ExtractAuthTokensRequest, reader: jspb.BinaryReader): ExtractAuthTokensRequest;
}

export namespace ExtractAuthTokensRequest {
    export type AsObject = {
    }
}

export class ExtractAuthTokensResponse extends jspb.Message { 
    clearTokensList(): void;
    getTokensList(): Array<HashedAuthToken>;
    setTokensList(value: Array<HashedAuthToken>): ExtractAuthTokensResponse;
    addTokens(value?: HashedAuthToken, index?: number): HashedAuthToken;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ExtractAuthTokensResponse.AsObject;
    static toObject(includeInstance: boolean, msg: ExtractAuthTokensResponse): ExtractAuthTokensResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ExtractAuthTokensResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ExtractAuthTokensResponse;
    static deserializeBinaryFromReader(message: ExtractAuthTokensResponse, reader: jspb.BinaryReader): ExtractAuthTokensResponse;
}

export namespace ExtractAuthTokensResponse {
    export type AsObject = {
        tokensList: Array<HashedAuthToken.AsObject>,
    }
}

export class RestoreAuthTokenRequest extends jspb.Message { 

    hasToken(): boolean;
    clearToken(): void;
    getToken(): HashedAuthToken | undefined;
    setToken(value?: HashedAuthToken): RestoreAuthTokenRequest;


    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): RestoreAuthTokenRequest.AsObject;
    static toObject(includeInstance: boolean, msg: RestoreAuthTokenRequest): RestoreAuthTokenRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: RestoreAuthTokenRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): RestoreAuthTokenRequest;
    static deserializeBinaryFromReader(message: RestoreAuthTokenRequest, reader: jspb.BinaryReader): RestoreAuthTokenRequest;
}

export namespace RestoreAuthTokenRequest {
    export type AsObject = {
        token?: HashedAuthToken.AsObject,
    }
}

export class RestoreAuthTokenResponse extends jspb.Message { 

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): RestoreAuthTokenResponse.AsObject;
    static toObject(includeInstance: boolean, msg: RestoreAuthTokenResponse): RestoreAuthTokenResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: RestoreAuthTokenResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): RestoreAuthTokenResponse;
    static deserializeBinaryFromReader(message: RestoreAuthTokenResponse, reader: jspb.BinaryReader): RestoreAuthTokenResponse;
}

export namespace RestoreAuthTokenResponse {
    export type AsObject = {
    }
}

export enum ClusterRole {
    UNDEFINED = 0,
    SUPER = 1,
    FS = 2,
}

export enum Scope {
    NONE = 0,
    READER = 1,
    WRITER = 2,
    OWNER = 3,
}
