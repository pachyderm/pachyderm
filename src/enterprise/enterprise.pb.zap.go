// Code generated by protoc-gen-zap (etc/proto/protoc-gen-zap). DO NOT EDIT.
//
// source: enterprise/enterprise.proto

package enterprise

import (
	protoextensions "github.com/pachyderm/pachyderm/v2/src/protoextensions"
	zapcore "go.uber.org/zap/zapcore"
)

func (x *LicenseRecord) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	protoextensions.AddHalfString(enc, "activation_code", x.ActivationCode)
	protoextensions.AddTimestamp(enc, "expires", x.Expires)
	return nil
}

func (x *EnterpriseConfig) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	enc.AddString("license_server", x.LicenseServer)
	enc.AddString("id", x.Id)
	enc.AddString("secret", x.Secret)
	return nil
}

func (x *EnterpriseRecord) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	enc.AddObject("license", x.License)
	protoextensions.AddTimestamp(enc, "last_heartbeat", x.LastHeartbeat)
	enc.AddBool("heartbeat_failed", x.HeartbeatFailed)
	return nil
}

func (x *TokenInfo) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	protoextensions.AddTimestamp(enc, "expires", x.Expires)
	return nil
}

func (x *ActivateRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	enc.AddString("license_server", x.LicenseServer)
	enc.AddString("id", x.Id)
	protoextensions.AddHalfString(enc, "secret", x.Secret)
	return nil
}

func (x *ActivateResponse) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	return nil
}

func (x *GetStateRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	return nil
}

func (x *GetStateResponse) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	enc.AddString("state", x.State.String())
	enc.AddObject("info", x.Info)
	protoextensions.AddHalfString(enc, "activation_code", x.ActivationCode)
	return nil
}

func (x *GetActivationCodeRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	return nil
}

func (x *GetActivationCodeResponse) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	enc.AddString("state", x.State.String())
	enc.AddObject("info", x.Info)
	protoextensions.AddHalfString(enc, "activation_code", x.ActivationCode)
	return nil
}

func (x *HeartbeatRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	return nil
}

func (x *HeartbeatResponse) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
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

func (x *PauseRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	return nil
}

func (x *PauseResponse) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	return nil
}

func (x *UnpauseRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	return nil
}

func (x *UnpauseResponse) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	return nil
}

func (x *PauseStatusRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	return nil
}

func (x *PauseStatusResponse) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}
	enc.AddString("status", x.Status.String())
	return nil
}
