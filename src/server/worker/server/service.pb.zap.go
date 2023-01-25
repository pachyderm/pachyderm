// Code generated by protoc-gen-zap (etc/proto/protoc-gen-zap). DO NOT EDIT.
//
// source: server/worker/server/service.proto

package server

import (
	zapcore "go.uber.org/zap/zapcore"
)

func (x *CancelRequest) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	enc.AddString("job_id", x.JobID)
	data_filtersArrMarshaller := func(enc zapcore.ArrayEncoder) error {
		for _, v := range x.DataFilters {
			enc.AppendString(v)
		}
		return nil
	}
	enc.AddArray("data_filters", zapcore.ArrayMarshalerFunc(data_filtersArrMarshaller))
	return nil
}

func (x *CancelResponse) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	enc.AddBool("success", x.Success)
	return nil
}
