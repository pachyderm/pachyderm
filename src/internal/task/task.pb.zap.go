// Code generated by protoc-gen-zap (etc/proto/protoc-gen-zap). DO NOT EDIT.
//
// source: internal/task/task.proto

package task

import (
	zapcore "go.uber.org/zap/zapcore"
)

func (x *Group) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	return nil
}

func (x *Task) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	enc.AddString("id", x.ID)

	enc.AddString("state", x.State.String())

	if obj, ok := interface{}(x.Input).(zapcore.ObjectMarshaler); ok {
		enc.AddObject("input", obj)
	} else {
		enc.AddReflected("input", x.Input)
	}

	if obj, ok := interface{}(x.Output).(zapcore.ObjectMarshaler); ok {
		enc.AddObject("output", obj)
	} else {
		enc.AddReflected("output", x.Output)
	}

	enc.AddString("reason", x.Reason)

	enc.AddInt64("index", x.Index)

	return nil
}

func (x *Claim) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	return nil
}

func (x *TestTask) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	enc.AddString("id", x.ID)

	return nil
}
