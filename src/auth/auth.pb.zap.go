// Code generated by protoc-gen-zap (etc/proto/protoc-gen-zap). DO NOT EDIT.
//
// source: auth/auth.proto

package auth

import (
	fmt "fmt"
	protoextensions "github.com/pachyderm/pachyderm/v2/src/protoextensions"
	zapcore "go.uber.org/zap/zapcore"
)

func (x *ActivateRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	enc.AddString("root_token", "[MASKED]")
	return nil
}

func (x *ActivateResponse) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	enc.AddString("pach_token", "[MASKED]")
	return nil
}

func (x *DeactivateRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	return nil
}

func (x *DeactivateResponse) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	return nil
}

func (x *RotateRootTokenRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	enc.AddString("root_token", "[MASKED]")
	return nil
}

func (x *RotateRootTokenResponse) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	enc.AddString("root_token", "[MASKED]")
	return nil
}

func (x *OIDCConfig) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	enc.AddString("issuer", x.Issuer)
	enc.AddString("client_id", x.ClientID)
	enc.AddString("client_secret", "[MASKED]")
	enc.AddString("redirect_uri", x.RedirectURI)
	scopesArrMarshaller := func(enc zapcore.ArrayEncoder) error {
		for _, v := range x.Scopes {
			enc.AppendString(v)
		}
		return nil
	}
	enc.AddArray("scopes", zapcore.ArrayMarshalerFunc(scopesArrMarshaller))
	enc.AddBool("require_email_verified", x.RequireEmailVerified)
	enc.AddBool("localhost_issuer", x.LocalhostIssuer)
	enc.AddString("user_accessible_issuer_host", x.UserAccessibleIssuerHost)
	return nil
}

func (x *GetConfigurationRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	return nil
}

func (x *GetConfigurationResponse) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	if obj, ok := interface{}(x.Configuration).(zapcore.ObjectMarshaler); ok {
		enc.AddObject("configuration", obj)
	} else {
		enc.AddReflected("configuration", x.Configuration)
	}
	return nil
}

func (x *SetConfigurationRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	if obj, ok := interface{}(x.Configuration).(zapcore.ObjectMarshaler); ok {
		enc.AddObject("configuration", obj)
	} else {
		enc.AddReflected("configuration", x.Configuration)
	}
	return nil
}

func (x *SetConfigurationResponse) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	return nil
}

func (x *TokenInfo) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	enc.AddString("subject", x.Subject)
	if t := x.Expiration; t != nil {
		enc.AddTime("expiration", *t)
	}
	enc.AddString("hashed_token", x.HashedToken)
	return nil
}

func (x *AuthenticateRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	protoextensions.AddHalfString(enc, "oidc_state", x.OIDCState)
	protoextensions.AddHalfString(enc, "id_token", x.IdToken)
	return nil
}

func (x *AuthenticateResponse) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	enc.AddString("pach_token", "[MASKED]")
	return nil
}

func (x *WhoAmIRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	return nil
}

func (x *WhoAmIResponse) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	enc.AddString("username", x.Username)
	if t := x.Expiration; t != nil {
		enc.AddTime("expiration", *t)
	}
	return nil
}

func (x *GetRolesForPermissionRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	enc.AddString("permission", x.Permission.String())
	return nil
}

func (x *GetRolesForPermissionResponse) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	rolesArrMarshaller := func(enc zapcore.ArrayEncoder) error {
		for _, v := range x.Roles {
			if obj, ok := interface{}(v).(zapcore.ObjectMarshaler); ok {
				enc.AppendObject(obj)
			} else {
				enc.AppendReflected(v)
			}
		}
		return nil
	}
	enc.AddArray("roles", zapcore.ArrayMarshalerFunc(rolesArrMarshaller))
	return nil
}

func (x *Roles) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	enc.AddObject("roles", zapcore.ObjectMarshalerFunc(func(enc zapcore.ObjectEncoder) error {
		for k, v := range x.Roles {
			enc.AddBool(fmt.Sprintf("%v", k), v)
		}
		return nil
	}))
	return nil
}

func (x *RoleBinding) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	enc.AddObject("entries", zapcore.ObjectMarshalerFunc(func(enc zapcore.ObjectEncoder) error {
		for k, v := range x.Entries {
			if obj, ok := interface{}(v).(zapcore.ObjectMarshaler); ok {
				enc.AddObject(fmt.Sprintf("%v", k), obj)
			} else {
				enc.AddReflected(fmt.Sprintf("%v", k), v)
			}
		}
		return nil
	}))
	return nil
}

func (x *Resource) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	enc.AddString("type", x.Type.String())
	enc.AddString("name", x.Name)
	return nil
}

func (x *Users) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	enc.AddObject("usernames", zapcore.ObjectMarshalerFunc(func(enc zapcore.ObjectEncoder) error {
		for k, v := range x.Usernames {
			enc.AddBool(fmt.Sprintf("%v", k), v)
		}
		return nil
	}))
	return nil
}

func (x *Groups) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	enc.AddObject("groups", zapcore.ObjectMarshalerFunc(func(enc zapcore.ObjectEncoder) error {
		for k, v := range x.Groups {
			enc.AddBool(fmt.Sprintf("%v", k), v)
		}
		return nil
	}))
	return nil
}

func (x *Role) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	enc.AddString("name", x.Name)
	permissionsArrMarshaller := func(enc zapcore.ArrayEncoder) error {
		for _, v := range x.Permissions {
			enc.AppendString(v.String())
		}
		return nil
	}
	enc.AddArray("permissions", zapcore.ArrayMarshalerFunc(permissionsArrMarshaller))
	resource_typesArrMarshaller := func(enc zapcore.ArrayEncoder) error {
		for _, v := range x.ResourceTypes {
			enc.AppendString(v.String())
		}
		return nil
	}
	enc.AddArray("resource_types", zapcore.ArrayMarshalerFunc(resource_typesArrMarshaller))
	return nil
}

func (x *AuthorizeRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	if obj, ok := interface{}(x.Resource).(zapcore.ObjectMarshaler); ok {
		enc.AddObject("resource", obj)
	} else {
		enc.AddReflected("resource", x.Resource)
	}
	permissionsArrMarshaller := func(enc zapcore.ArrayEncoder) error {
		for _, v := range x.Permissions {
			enc.AppendString(v.String())
		}
		return nil
	}
	enc.AddArray("permissions", zapcore.ArrayMarshalerFunc(permissionsArrMarshaller))
	return nil
}

func (x *AuthorizeResponse) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	enc.AddBool("authorized", x.Authorized)
	satisfiedArrMarshaller := func(enc zapcore.ArrayEncoder) error {
		for _, v := range x.Satisfied {
			enc.AppendString(v.String())
		}
		return nil
	}
	enc.AddArray("satisfied", zapcore.ArrayMarshalerFunc(satisfiedArrMarshaller))
	missingArrMarshaller := func(enc zapcore.ArrayEncoder) error {
		for _, v := range x.Missing {
			enc.AppendString(v.String())
		}
		return nil
	}
	enc.AddArray("missing", zapcore.ArrayMarshalerFunc(missingArrMarshaller))
	enc.AddString("principal", x.Principal)
	return nil
}

func (x *GetPermissionsRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	if obj, ok := interface{}(x.Resource).(zapcore.ObjectMarshaler); ok {
		enc.AddObject("resource", obj)
	} else {
		enc.AddReflected("resource", x.Resource)
	}
	return nil
}

func (x *GetPermissionsForPrincipalRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	if obj, ok := interface{}(x.Resource).(zapcore.ObjectMarshaler); ok {
		enc.AddObject("resource", obj)
	} else {
		enc.AddReflected("resource", x.Resource)
	}
	enc.AddString("principal", x.Principal)
	return nil
}

func (x *GetPermissionsResponse) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	permissionsArrMarshaller := func(enc zapcore.ArrayEncoder) error {
		for _, v := range x.Permissions {
			enc.AppendString(v.String())
		}
		return nil
	}
	enc.AddArray("permissions", zapcore.ArrayMarshalerFunc(permissionsArrMarshaller))
	rolesArrMarshaller := func(enc zapcore.ArrayEncoder) error {
		for _, v := range x.Roles {
			enc.AppendString(v)
		}
		return nil
	}
	enc.AddArray("roles", zapcore.ArrayMarshalerFunc(rolesArrMarshaller))
	return nil
}

func (x *ModifyRoleBindingRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	if obj, ok := interface{}(x.Resource).(zapcore.ObjectMarshaler); ok {
		enc.AddObject("resource", obj)
	} else {
		enc.AddReflected("resource", x.Resource)
	}
	enc.AddString("principal", x.Principal)
	rolesArrMarshaller := func(enc zapcore.ArrayEncoder) error {
		for _, v := range x.Roles {
			enc.AppendString(v)
		}
		return nil
	}
	enc.AddArray("roles", zapcore.ArrayMarshalerFunc(rolesArrMarshaller))
	return nil
}

func (x *ModifyRoleBindingResponse) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	return nil
}

func (x *GetRoleBindingRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	if obj, ok := interface{}(x.Resource).(zapcore.ObjectMarshaler); ok {
		enc.AddObject("resource", obj)
	} else {
		enc.AddReflected("resource", x.Resource)
	}
	return nil
}

func (x *GetRoleBindingResponse) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	if obj, ok := interface{}(x.Binding).(zapcore.ObjectMarshaler); ok {
		enc.AddObject("binding", obj)
	} else {
		enc.AddReflected("binding", x.Binding)
	}
	return nil
}

func (x *SessionInfo) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	protoextensions.AddHalfString(enc, "nonce", x.Nonce)
	enc.AddString("email", x.Email)
	enc.AddBool("conversion_err", x.ConversionErr)
	return nil
}

func (x *GetOIDCLoginRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	return nil
}

func (x *GetOIDCLoginResponse) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	protoextensions.AddHalfString(enc, "login_url", x.LoginURL)
	protoextensions.AddHalfString(enc, "state", x.State)
	return nil
}

func (x *GetRobotTokenRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	enc.AddString("robot", x.Robot)
	enc.AddInt64("ttl", x.TTL)
	return nil
}

func (x *GetRobotTokenResponse) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	enc.AddString("token", "[MASKED]")
	return nil
}

func (x *RevokeAuthTokenRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	protoextensions.AddHalfString(enc, "token", x.Token)
	return nil
}

func (x *RevokeAuthTokenResponse) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	enc.AddInt64("number", x.Number)
	return nil
}

func (x *SetGroupsForUserRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	enc.AddString("username", x.Username)
	groupsArrMarshaller := func(enc zapcore.ArrayEncoder) error {
		for _, v := range x.Groups {
			enc.AppendString(v)
		}
		return nil
	}
	enc.AddArray("groups", zapcore.ArrayMarshalerFunc(groupsArrMarshaller))
	return nil
}

func (x *SetGroupsForUserResponse) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	return nil
}

func (x *ModifyMembersRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	enc.AddString("group", x.Group)
	addArrMarshaller := func(enc zapcore.ArrayEncoder) error {
		for _, v := range x.Add {
			enc.AppendString(v)
		}
		return nil
	}
	enc.AddArray("add", zapcore.ArrayMarshalerFunc(addArrMarshaller))
	removeArrMarshaller := func(enc zapcore.ArrayEncoder) error {
		for _, v := range x.Remove {
			enc.AppendString(v)
		}
		return nil
	}
	enc.AddArray("remove", zapcore.ArrayMarshalerFunc(removeArrMarshaller))
	return nil
}

func (x *ModifyMembersResponse) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	return nil
}

func (x *GetGroupsRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	return nil
}

func (x *GetGroupsForPrincipalRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	enc.AddString("principal", x.Principal)
	return nil
}

func (x *GetGroupsResponse) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	groupsArrMarshaller := func(enc zapcore.ArrayEncoder) error {
		for _, v := range x.Groups {
			enc.AppendString(v)
		}
		return nil
	}
	enc.AddArray("groups", zapcore.ArrayMarshalerFunc(groupsArrMarshaller))
	return nil
}

func (x *GetUsersRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	enc.AddString("group", x.Group)
	return nil
}

func (x *GetUsersResponse) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	usernamesArrMarshaller := func(enc zapcore.ArrayEncoder) error {
		for _, v := range x.Usernames {
			enc.AppendString(v)
		}
		return nil
	}
	enc.AddArray("usernames", zapcore.ArrayMarshalerFunc(usernamesArrMarshaller))
	return nil
}

func (x *ExtractAuthTokensRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	return nil
}

func (x *ExtractAuthTokensResponse) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	tokensArrMarshaller := func(enc zapcore.ArrayEncoder) error {
		for _, v := range x.Tokens {
			if obj, ok := interface{}(v).(zapcore.ObjectMarshaler); ok {
				enc.AppendObject(obj)
			} else {
				enc.AppendReflected(v)
			}
		}
		return nil
	}
	enc.AddArray("tokens", zapcore.ArrayMarshalerFunc(tokensArrMarshaller))
	return nil
}

func (x *RestoreAuthTokenRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	if obj, ok := interface{}(x.Token).(zapcore.ObjectMarshaler); ok {
		enc.AddObject("token", obj)
	} else {
		enc.AddReflected("token", x.Token)
	}
	return nil
}

func (x *RestoreAuthTokenResponse) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	return nil
}

func (x *RevokeAuthTokensForUserRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	enc.AddString("username", x.Username)
	return nil
}

func (x *RevokeAuthTokensForUserResponse) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	enc.AddInt64("number", x.Number)
	return nil
}

func (x *DeleteExpiredAuthTokensRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	return nil
}

func (x *DeleteExpiredAuthTokensResponse) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	return nil
}
