package log

import (
	"bufio"
	"bytes"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestParseWorkerLines(t *testing.T) {
	buf := new(bytes.Buffer)
	enc := zapcore.NewJSONEncoder(workerEncoder)
	l := zap.New(zapcore.NewCore(enc, zapcore.AddSync(buf), zap.DebugLevel),
		zap.AddCaller(), zap.AddCallerSkip(1), zap.AddStacktrace(zapcore.DebugLevel),
		zap.Fields(
			pps.WorkerIDField("workerIDField"),
			pps.PipelineNameField("pipelineNameField"),
			pps.ProjectNameField("projectNameField"),
		),
	)
	l.Debug("debug", zap.Int("number", 42), zap.Complex128("complex", complex(42, 1.2345)))
	l.Info("info", zap.Strings("strings", []string{"many", "strings", "are", "here"}), pps.MasterField(true))
	l.Named("etcd-client").Warn("retrying unary invoker failed", zap.Strings("endpoints", []string{"etcd://1.2.3.4:1234"}))
	l.Error("error", zap.Error(errors.New("some error")), zap.Stack("someRandomName"), pps.UserField(true), pps.DataField([]*pps.InputFile{{Hash: []byte("abc"), Path: "/abc"}}), pps.JobIDField("jobIDField"), pps.DatumIDField("datumIDField"))
	if err := l.Sync(); err != nil {
		t.Fatalf("sync: %v", err)
	}

	m := &protojson.UnmarshalOptions{
		DiscardUnknown: true,
		AllowPartial:   true,
	}
	var got []*pps.LogMessage
	s := bufio.NewScanner(buf)
	var line int
	for s.Scan() {
		line++
		var msg pps.LogMessage
		text := s.Bytes()
		if err := m.Unmarshal(text, &msg); err != nil {
			t.Logf("line %d: %s", line, text)
			t.Errorf("unmarshal line %d: %v", line, err)
		}
		if msg.GetTs() == nil {
			t.Errorf("no time in line %d", line)
		}
		msg.Ts = nil
		got = append(got, &msg)
	}
	if err := s.Err(); err != nil {
		t.Fatalf("scanner: %v", err)
	}
	want := []*pps.LogMessage{
		{
			ProjectName:  "projectNameField",
			PipelineName: "pipelineNameField",
			WorkerId:     "workerIDField",
			Message:      "debug",
		},
		{
			ProjectName:  "projectNameField",
			PipelineName: "pipelineNameField",
			WorkerId:     "workerIDField",
			Message:      "info",
			Master:       true,
		},
		{
			ProjectName:  "projectNameField",
			PipelineName: "pipelineNameField",
			WorkerId:     "workerIDField",
			Message:      "retrying unary invoker failed",
		},
		{
			ProjectName:  "projectNameField",
			PipelineName: "pipelineNameField",
			WorkerId:     "workerIDField",
			Message:      "error",
			JobId:        "jobIDField",
			DatumId:      "datumIDField",
			User:         true,
			Data: []*pps.InputFile{
				{
					Hash: []byte("abc"),
					Path: "/abc",
				},
			},
		},
	}
	if diff := cmp.Diff(got, want, protocmp.Transform()); diff != "" {
		t.Errorf("diff (-got +want):\n%s", diff)
	}
}
