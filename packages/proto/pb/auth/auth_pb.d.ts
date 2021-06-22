// package: auth_v2
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

export class RotateRootTokenRequest extends jspb.Message { 
    getRootToken(): string;
    setRootToken(value: string): RotateRootTokenRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): RotateRootTokenRequest.AsObject;
    static toObject(includeInstance: boolean, msg: RotateRootTokenRequest): RotateRootTokenRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: RotateRootTokenRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): RotateRootTokenRequest;
    static deserializeBinaryFromReader(message: RotateRootTokenRequest, reader: jspb.BinaryReader): RotateRootTokenRequest;
}

export namespace RotateRootTokenRequest {
    export type AsObject = {
        rootToken: string,
    }
}

export class RotateRootTokenResponse extends jspb.Message { 
    getRootToken(): string;
    setRootToken(value: string): RotateRootTokenResponse;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): RotateRootTokenResponse.AsObject;
    static toObject(includeInstance: boolean, msg: RotateRootTokenResponse): RotateRootTokenResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: RotateRootTokenResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): RotateRootTokenResponse;
    static deserializeBinaryFromReader(message: RotateRootTokenResponse, reader: jspb.BinaryReader): RotateRootTokenResponse;
}

export namespace RotateRootTokenResponse {
    export type AsObject = {
        rootToken: string,
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
    clearScopesList(): void;
    getScopesList(): Array<string>;
    setScopesList(value: Array<string>): OIDCConfig;
    addScopes(value: string, index?: number): string;
    getRequireEmailVerified(): boolean;
    setRequireEmailVerified(value: boolean): OIDCConfig;
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
        scopesList: Array<string>,
        requireEmailVerified: boolean,
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

export class TokenInfo extends jspb.Message { 
    getSubject(): string;
    setSubject(value: string): TokenInfo;

    hasExpiration(): boolean;
    clearExpiration(): void;
    getExpiration(): google_protobuf_timestamp_pb.Timestamp | undefined;
    setExpiration(value?: google_protobuf_timestamp_pb.Timestamp): TokenInfo;
    getHashedToken(): string;
    setHashedToken(value: string): TokenInfo;

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
        expiration?: google_protobuf_timestamp_pb.Timestamp.AsObject,
        hashedToken: string,
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

    hasExpiration(): boolean;
    clearExpiration(): void;
    getExpiration(): google_protobuf_timestamp_pb.Timestamp | undefined;
    setExpiration(value?: google_protobuf_timestamp_pb.Timestamp): WhoAmIResponse;

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
        expiration?: google_protobuf_timestamp_pb.Timestamp.AsObject,
    }
}

export class Roles extends jspb.Message { 

    getRolesMap(): jspb.Map<string, boolean>;
    clearRolesMap(): void;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Roles.AsObject;
    static toObject(includeInstance: boolean, msg: Roles): Roles.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: Roles, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Roles;
    static deserializeBinaryFromReader(message: Roles, reader: jspb.BinaryReader): Roles;
}

export namespace Roles {
    export type AsObject = {

        rolesMap: Array<[string, boolean]>,
    }
}

export class RoleBinding extends jspb.Message { 

    getEntriesMap(): jspb.Map<string, Roles>;
    clearEntriesMap(): void;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): RoleBinding.AsObject;
    static toObject(includeInstance: boolean, msg: RoleBinding): RoleBinding.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: RoleBinding, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): RoleBinding;
    static deserializeBinaryFromReader(message: RoleBinding, reader: jspb.BinaryReader): RoleBinding;
}

export namespace RoleBinding {
    export type AsObject = {

        entriesMap: Array<[string, Roles.AsObject]>,
    }
}

export class Resource extends jspb.Message { 
    getType(): ResourceType;
    setType(value: ResourceType): Resource;
    getName(): string;
    setName(value: string): Resource;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Resource.AsObject;
    static toObject(includeInstance: boolean, msg: Resource): Resource.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: Resource, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Resource;
    static deserializeBinaryFromReader(message: Resource, reader: jspb.BinaryReader): Resource;
}

export namespace Resource {
    export type AsObject = {
        type: ResourceType,
        name: string,
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

export class Role extends jspb.Message { 
    getName(): string;
    setName(value: string): Role;
    clearPermissionsList(): void;
    getPermissionsList(): Array<Permission>;
    setPermissionsList(value: Array<Permission>): Role;
    addPermissions(value: Permission, index?: number): Permission;
    clearResourceTypesList(): void;
    getResourceTypesList(): Array<ResourceType>;
    setResourceTypesList(value: Array<ResourceType>): Role;
    addResourceTypes(value: ResourceType, index?: number): ResourceType;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Role.AsObject;
    static toObject(includeInstance: boolean, msg: Role): Role.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: Role, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Role;
    static deserializeBinaryFromReader(message: Role, reader: jspb.BinaryReader): Role;
}

export namespace Role {
    export type AsObject = {
        name: string,
        permissionsList: Array<Permission>,
        resourceTypesList: Array<ResourceType>,
    }
}

export class AuthorizeRequest extends jspb.Message { 

    hasResource(): boolean;
    clearResource(): void;
    getResource(): Resource | undefined;
    setResource(value?: Resource): AuthorizeRequest;
    clearPermissionsList(): void;
    getPermissionsList(): Array<Permission>;
    setPermissionsList(value: Array<Permission>): AuthorizeRequest;
    addPermissions(value: Permission, index?: number): Permission;

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
        resource?: Resource.AsObject,
        permissionsList: Array<Permission>,
    }
}

export class AuthorizeResponse extends jspb.Message { 
    getAuthorized(): boolean;
    setAuthorized(value: boolean): AuthorizeResponse;
    clearSatisfiedList(): void;
    getSatisfiedList(): Array<Permission>;
    setSatisfiedList(value: Array<Permission>): AuthorizeResponse;
    addSatisfied(value: Permission, index?: number): Permission;
    clearMissingList(): void;
    getMissingList(): Array<Permission>;
    setMissingList(value: Array<Permission>): AuthorizeResponse;
    addMissing(value: Permission, index?: number): Permission;
    getPrincipal(): string;
    setPrincipal(value: string): AuthorizeResponse;

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
        satisfiedList: Array<Permission>,
        missingList: Array<Permission>,
        principal: string,
    }
}

export class GetPermissionsRequest extends jspb.Message { 

    hasResource(): boolean;
    clearResource(): void;
    getResource(): Resource | undefined;
    setResource(value?: Resource): GetPermissionsRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GetPermissionsRequest.AsObject;
    static toObject(includeInstance: boolean, msg: GetPermissionsRequest): GetPermissionsRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GetPermissionsRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GetPermissionsRequest;
    static deserializeBinaryFromReader(message: GetPermissionsRequest, reader: jspb.BinaryReader): GetPermissionsRequest;
}

export namespace GetPermissionsRequest {
    export type AsObject = {
        resource?: Resource.AsObject,
    }
}

export class GetPermissionsForPrincipalRequest extends jspb.Message { 

    hasResource(): boolean;
    clearResource(): void;
    getResource(): Resource | undefined;
    setResource(value?: Resource): GetPermissionsForPrincipalRequest;
    getPrincipal(): string;
    setPrincipal(value: string): GetPermissionsForPrincipalRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GetPermissionsForPrincipalRequest.AsObject;
    static toObject(includeInstance: boolean, msg: GetPermissionsForPrincipalRequest): GetPermissionsForPrincipalRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GetPermissionsForPrincipalRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GetPermissionsForPrincipalRequest;
    static deserializeBinaryFromReader(message: GetPermissionsForPrincipalRequest, reader: jspb.BinaryReader): GetPermissionsForPrincipalRequest;
}

export namespace GetPermissionsForPrincipalRequest {
    export type AsObject = {
        resource?: Resource.AsObject,
        principal: string,
    }
}

export class GetPermissionsResponse extends jspb.Message { 
    clearPermissionsList(): void;
    getPermissionsList(): Array<Permission>;
    setPermissionsList(value: Array<Permission>): GetPermissionsResponse;
    addPermissions(value: Permission, index?: number): Permission;
    clearRolesList(): void;
    getRolesList(): Array<string>;
    setRolesList(value: Array<string>): GetPermissionsResponse;
    addRoles(value: string, index?: number): string;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GetPermissionsResponse.AsObject;
    static toObject(includeInstance: boolean, msg: GetPermissionsResponse): GetPermissionsResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GetPermissionsResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GetPermissionsResponse;
    static deserializeBinaryFromReader(message: GetPermissionsResponse, reader: jspb.BinaryReader): GetPermissionsResponse;
}

export namespace GetPermissionsResponse {
    export type AsObject = {
        permissionsList: Array<Permission>,
        rolesList: Array<string>,
    }
}

export class ModifyRoleBindingRequest extends jspb.Message { 

    hasResource(): boolean;
    clearResource(): void;
    getResource(): Resource | undefined;
    setResource(value?: Resource): ModifyRoleBindingRequest;
    getPrincipal(): string;
    setPrincipal(value: string): ModifyRoleBindingRequest;
    clearRolesList(): void;
    getRolesList(): Array<string>;
    setRolesList(value: Array<string>): ModifyRoleBindingRequest;
    addRoles(value: string, index?: number): string;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ModifyRoleBindingRequest.AsObject;
    static toObject(includeInstance: boolean, msg: ModifyRoleBindingRequest): ModifyRoleBindingRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ModifyRoleBindingRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ModifyRoleBindingRequest;
    static deserializeBinaryFromReader(message: ModifyRoleBindingRequest, reader: jspb.BinaryReader): ModifyRoleBindingRequest;
}

export namespace ModifyRoleBindingRequest {
    export type AsObject = {
        resource?: Resource.AsObject,
        principal: string,
        rolesList: Array<string>,
    }
}

export class ModifyRoleBindingResponse extends jspb.Message { 

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ModifyRoleBindingResponse.AsObject;
    static toObject(includeInstance: boolean, msg: ModifyRoleBindingResponse): ModifyRoleBindingResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ModifyRoleBindingResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ModifyRoleBindingResponse;
    static deserializeBinaryFromReader(message: ModifyRoleBindingResponse, reader: jspb.BinaryReader): ModifyRoleBindingResponse;
}

export namespace ModifyRoleBindingResponse {
    export type AsObject = {
    }
}

export class GetRoleBindingRequest extends jspb.Message { 

    hasResource(): boolean;
    clearResource(): void;
    getResource(): Resource | undefined;
    setResource(value?: Resource): GetRoleBindingRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GetRoleBindingRequest.AsObject;
    static toObject(includeInstance: boolean, msg: GetRoleBindingRequest): GetRoleBindingRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GetRoleBindingRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GetRoleBindingRequest;
    static deserializeBinaryFromReader(message: GetRoleBindingRequest, reader: jspb.BinaryReader): GetRoleBindingRequest;
}

export namespace GetRoleBindingRequest {
    export type AsObject = {
        resource?: Resource.AsObject,
    }
}

export class GetRoleBindingResponse extends jspb.Message { 

    hasBinding(): boolean;
    clearBinding(): void;
    getBinding(): RoleBinding | undefined;
    setBinding(value?: RoleBinding): GetRoleBindingResponse;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GetRoleBindingResponse.AsObject;
    static toObject(includeInstance: boolean, msg: GetRoleBindingResponse): GetRoleBindingResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GetRoleBindingResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GetRoleBindingResponse;
    static deserializeBinaryFromReader(message: GetRoleBindingResponse, reader: jspb.BinaryReader): GetRoleBindingResponse;
}

export namespace GetRoleBindingResponse {
    export type AsObject = {
        binding?: RoleBinding.AsObject,
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

export class GetRobotTokenRequest extends jspb.Message { 
    getRobot(): string;
    setRobot(value: string): GetRobotTokenRequest;
    getTtl(): number;
    setTtl(value: number): GetRobotTokenRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GetRobotTokenRequest.AsObject;
    static toObject(includeInstance: boolean, msg: GetRobotTokenRequest): GetRobotTokenRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GetRobotTokenRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GetRobotTokenRequest;
    static deserializeBinaryFromReader(message: GetRobotTokenRequest, reader: jspb.BinaryReader): GetRobotTokenRequest;
}

export namespace GetRobotTokenRequest {
    export type AsObject = {
        robot: string,
        ttl: number,
    }
}

export class GetRobotTokenResponse extends jspb.Message { 
    getToken(): string;
    setToken(value: string): GetRobotTokenResponse;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GetRobotTokenResponse.AsObject;
    static toObject(includeInstance: boolean, msg: GetRobotTokenResponse): GetRobotTokenResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GetRobotTokenResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GetRobotTokenResponse;
    static deserializeBinaryFromReader(message: GetRobotTokenResponse, reader: jspb.BinaryReader): GetRobotTokenResponse;
}

export namespace GetRobotTokenResponse {
    export type AsObject = {
        token: string,
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
    }
}

export class GetGroupsForPrincipalRequest extends jspb.Message { 
    getPrincipal(): string;
    setPrincipal(value: string): GetGroupsForPrincipalRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): GetGroupsForPrincipalRequest.AsObject;
    static toObject(includeInstance: boolean, msg: GetGroupsForPrincipalRequest): GetGroupsForPrincipalRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: GetGroupsForPrincipalRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): GetGroupsForPrincipalRequest;
    static deserializeBinaryFromReader(message: GetGroupsForPrincipalRequest, reader: jspb.BinaryReader): GetGroupsForPrincipalRequest;
}

export namespace GetGroupsForPrincipalRequest {
    export type AsObject = {
        principal: string,
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
    getTokensList(): Array<TokenInfo>;
    setTokensList(value: Array<TokenInfo>): ExtractAuthTokensResponse;
    addTokens(value?: TokenInfo, index?: number): TokenInfo;

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
        tokensList: Array<TokenInfo.AsObject>,
    }
}

export class RestoreAuthTokenRequest extends jspb.Message { 

    hasToken(): boolean;
    clearToken(): void;
    getToken(): TokenInfo | undefined;
    setToken(value?: TokenInfo): RestoreAuthTokenRequest;

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
        token?: TokenInfo.AsObject,
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

export class RevokeAuthTokensForUserRequest extends jspb.Message { 
    getUsername(): string;
    setUsername(value: string): RevokeAuthTokensForUserRequest;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): RevokeAuthTokensForUserRequest.AsObject;
    static toObject(includeInstance: boolean, msg: RevokeAuthTokensForUserRequest): RevokeAuthTokensForUserRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: RevokeAuthTokensForUserRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): RevokeAuthTokensForUserRequest;
    static deserializeBinaryFromReader(message: RevokeAuthTokensForUserRequest, reader: jspb.BinaryReader): RevokeAuthTokensForUserRequest;
}

export namespace RevokeAuthTokensForUserRequest {
    export type AsObject = {
        username: string,
    }
}

export class RevokeAuthTokensForUserResponse extends jspb.Message { 

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): RevokeAuthTokensForUserResponse.AsObject;
    static toObject(includeInstance: boolean, msg: RevokeAuthTokensForUserResponse): RevokeAuthTokensForUserResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: RevokeAuthTokensForUserResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): RevokeAuthTokensForUserResponse;
    static deserializeBinaryFromReader(message: RevokeAuthTokensForUserResponse, reader: jspb.BinaryReader): RevokeAuthTokensForUserResponse;
}

export namespace RevokeAuthTokensForUserResponse {
    export type AsObject = {
    }
}

export class DeleteExpiredAuthTokensRequest extends jspb.Message { 

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): DeleteExpiredAuthTokensRequest.AsObject;
    static toObject(includeInstance: boolean, msg: DeleteExpiredAuthTokensRequest): DeleteExpiredAuthTokensRequest.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: DeleteExpiredAuthTokensRequest, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): DeleteExpiredAuthTokensRequest;
    static deserializeBinaryFromReader(message: DeleteExpiredAuthTokensRequest, reader: jspb.BinaryReader): DeleteExpiredAuthTokensRequest;
}

export namespace DeleteExpiredAuthTokensRequest {
    export type AsObject = {
    }
}

export class DeleteExpiredAuthTokensResponse extends jspb.Message { 

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): DeleteExpiredAuthTokensResponse.AsObject;
    static toObject(includeInstance: boolean, msg: DeleteExpiredAuthTokensResponse): DeleteExpiredAuthTokensResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: DeleteExpiredAuthTokensResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): DeleteExpiredAuthTokensResponse;
    static deserializeBinaryFromReader(message: DeleteExpiredAuthTokensResponse, reader: jspb.BinaryReader): DeleteExpiredAuthTokensResponse;
}

export namespace DeleteExpiredAuthTokensResponse {
    export type AsObject = {
    }
}

export enum Permission {
    PERMISSION_UNKNOWN = 0,
    CLUSTER_MODIFY_BINDINGS = 100,
    CLUSTER_GET_BINDINGS = 101,
    CLUSTER_GET_PACHD_LOGS = 148,
    CLUSTER_AUTH_ACTIVATE = 102,
    CLUSTER_AUTH_DEACTIVATE = 103,
    CLUSTER_AUTH_GET_CONFIG = 104,
    CLUSTER_AUTH_SET_CONFIG = 105,
    CLUSTER_AUTH_GET_ROBOT_TOKEN = 139,
    CLUSTER_AUTH_MODIFY_GROUP_MEMBERS = 109,
    CLUSTER_AUTH_GET_GROUPS = 110,
    CLUSTER_AUTH_GET_GROUP_USERS = 111,
    CLUSTER_AUTH_EXTRACT_TOKENS = 112,
    CLUSTER_AUTH_RESTORE_TOKEN = 113,
    CLUSTER_AUTH_GET_PERMISSIONS_FOR_PRINCIPAL = 141,
    CLUSTER_AUTH_DELETE_EXPIRED_TOKENS = 140,
    CLUSTER_AUTH_REVOKE_USER_TOKENS = 142,
    CLUSTER_AUTH_ROTATE_ROOT_TOKEN = 147,
    CLUSTER_ENTERPRISE_ACTIVATE = 114,
    CLUSTER_ENTERPRISE_HEARTBEAT = 115,
    CLUSTER_ENTERPRISE_GET_CODE = 116,
    CLUSTER_ENTERPRISE_DEACTIVATE = 117,
    CLUSTER_IDENTITY_SET_CONFIG = 118,
    CLUSTER_IDENTITY_GET_CONFIG = 119,
    CLUSTER_IDENTITY_CREATE_IDP = 120,
    CLUSTER_IDENTITY_UPDATE_IDP = 121,
    CLUSTER_IDENTITY_LIST_IDPS = 122,
    CLUSTER_IDENTITY_GET_IDP = 123,
    CLUSTER_IDENTITY_DELETE_IDP = 124,
    CLUSTER_IDENTITY_CREATE_OIDC_CLIENT = 125,
    CLUSTER_IDENTITY_UPDATE_OIDC_CLIENT = 126,
    CLUSTER_IDENTITY_LIST_OIDC_CLIENTS = 127,
    CLUSTER_IDENTITY_GET_OIDC_CLIENT = 128,
    CLUSTER_IDENTITY_DELETE_OIDC_CLIENT = 129,
    CLUSTER_DEBUG_DUMP = 131,
    CLUSTER_LICENSE_ACTIVATE = 132,
    CLUSTER_LICENSE_GET_CODE = 133,
    CLUSTER_LICENSE_ADD_CLUSTER = 134,
    CLUSTER_LICENSE_UPDATE_CLUSTER = 135,
    CLUSTER_LICENSE_DELETE_CLUSTER = 136,
    CLUSTER_LICENSE_LIST_CLUSTERS = 137,
    CLUSTER_CREATE_SECRET = 143,
    CLUSTER_LIST_SECRETS = 144,
    SECRET_DELETE = 145,
    SECRET_INSPECT = 146,
    CLUSTER_DELETE_ALL = 138,
    REPO_READ = 200,
    REPO_WRITE = 201,
    REPO_MODIFY_BINDINGS = 202,
    REPO_DELETE = 203,
    REPO_INSPECT_COMMIT = 204,
    REPO_LIST_COMMIT = 205,
    REPO_DELETE_COMMIT = 206,
    REPO_CREATE_BRANCH = 207,
    REPO_LIST_BRANCH = 208,
    REPO_DELETE_BRANCH = 209,
    REPO_INSPECT_FILE = 210,
    REPO_LIST_FILE = 211,
    REPO_ADD_PIPELINE_READER = 212,
    REPO_REMOVE_PIPELINE_READER = 213,
    REPO_ADD_PIPELINE_WRITER = 214,
    PIPELINE_LIST_JOB = 301,
}

export enum ResourceType {
    RESOURCE_TYPE_UNKNOWN = 0,
    CLUSTER = 1,
    REPO = 2,
}
