// Code generated by protoc-gen-zap (etc/proto/protoc-gen-zap). DO NOT EDIT.
//
// source: server/worker/pipeline/transform/transform.proto

package transform

import (
	zapcore "go.uber.org/zap/zapcore"
)

func (x *CreateParallelDatumsTask) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	if obj, ok := interface{}(x.Job).(zapcore.ObjectMarshaler); ok {
		enc.AddObject("job", obj)
	} else {
		enc.AddReflected("job", x.Job)
	}
	enc.AddString("salt", x.Salt)
	enc.AddString("file_set_id", x.FileSetId)
	enc.AddString("base_file_set_id", x.BaseFileSetId)
	if obj, ok := interface{}(x.PathRange).(zapcore.ObjectMarshaler); ok {
		enc.AddObject("path_range", obj)
	} else {
		enc.AddReflected("path_range", x.PathRange)
	}
	return nil
}

func (x *CreateParallelDatumsTaskResult) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	enc.AddString("file_set_id", x.FileSetId)
	if obj, ok := interface{}(x.Stats).(zapcore.ObjectMarshaler); ok {
		enc.AddObject("stats", obj)
	} else {
		enc.AddReflected("stats", x.Stats)
	}
	return nil
}

func (x *CreateSerialDatumsTask) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	if obj, ok := interface{}(x.Job).(zapcore.ObjectMarshaler); ok {
		enc.AddObject("job", obj)
	} else {
		enc.AddReflected("job", x.Job)
	}
	enc.AddString("salt", x.Salt)
	enc.AddString("file_set_id", x.FileSetId)
	if obj, ok := interface{}(x.BaseMetaCommit).(zapcore.ObjectMarshaler); ok {
		enc.AddObject("base_meta_commit", obj)
	} else {
		enc.AddReflected("base_meta_commit", x.BaseMetaCommit)
	}
	enc.AddBool("no_skip", x.NoSkip)
	if obj, ok := interface{}(x.PathRange).(zapcore.ObjectMarshaler); ok {
		enc.AddObject("path_range", obj)
	} else {
		enc.AddReflected("path_range", x.PathRange)
	}
	return nil
}

func (x *CreateSerialDatumsTaskResult) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	enc.AddString("file_set_id", x.FileSetId)
	enc.AddString("output_delete_file_set_id", x.OutputDeleteFileSetId)
	enc.AddString("meta_delete_file_set_id", x.MetaDeleteFileSetId)
	if obj, ok := interface{}(x.Stats).(zapcore.ObjectMarshaler); ok {
		enc.AddObject("stats", obj)
	} else {
		enc.AddReflected("stats", x.Stats)
	}
	return nil
}

func (x *CreateDatumSetsTask) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	enc.AddString("file_set_id", x.FileSetId)
	if obj, ok := interface{}(x.PathRange).(zapcore.ObjectMarshaler); ok {
		enc.AddObject("path_range", obj)
	} else {
		enc.AddReflected("path_range", x.PathRange)
	}
	if obj, ok := interface{}(x.SetSpec).(zapcore.ObjectMarshaler); ok {
		enc.AddObject("set_spec", obj)
	} else {
		enc.AddReflected("set_spec", x.SetSpec)
	}
	return nil
}

func (x *CreateDatumSetsTaskResult) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	datum_setsArrMarshaller := func(enc zapcore.ArrayEncoder) error {
		for _, v := range x.DatumSets {
			if obj, ok := interface{}(v).(zapcore.ObjectMarshaler); ok {
				enc.AppendObject(obj)
			} else {
				enc.AppendReflected(v)
			}
		}
		return nil
	}
	enc.AddArray("datum_sets", zapcore.ArrayMarshalerFunc(datum_setsArrMarshaller))
	return nil
}

func (x *DatumSetTask) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	if obj, ok := interface{}(x.Job).(zapcore.ObjectMarshaler); ok {
		enc.AddObject("job", obj)
	} else {
		enc.AddReflected("job", x.Job)
	}
	enc.AddString("file_set_id", x.FileSetId)
	if obj, ok := interface{}(x.PathRange).(zapcore.ObjectMarshaler); ok {
		enc.AddObject("path_range", obj)
	} else {
		enc.AddReflected("path_range", x.PathRange)
	}
	if obj, ok := interface{}(x.OutputCommit).(zapcore.ObjectMarshaler); ok {
		enc.AddObject("output_commit", obj)
	} else {
		enc.AddReflected("output_commit", x.OutputCommit)
	}
	return nil
}

func (x *DatumSetTaskResult) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	if x == nil {
		return nil
	}

	enc.AddString("output_file_set_id", x.OutputFileSetId)
	enc.AddString("meta_file_set_id", x.MetaFileSetId)
	if obj, ok := interface{}(x.Stats).(zapcore.ObjectMarshaler); ok {
		enc.AddObject("stats", obj)
	} else {
		enc.AddReflected("stats", x.Stats)
	}
	return nil
}
