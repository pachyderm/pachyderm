package logs

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/pachyderm/pachyderm/v2/src/logs"
	"github.com/pachyderm/pachyderm/v2/src/pps"

	loki "github.com/pachyderm/pachyderm/v2/src/internal/lokiutil/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

type testPublisher struct {
	last *logs.GetLogsResponse
}

func (p *testPublisher) Publish(ctx context.Context, resp *logs.GetLogsResponse) error {
	p.last = resp
	return nil
}

func Test_adapter_publish(t *testing.T) {
	var (
		ctx   = pctx.TestContext(t)
		cases = map[string]struct {
			entry loki.Entry
			want  *logs.LogMessage
		}{
			"JSON timestamp": {
				entry: loki.Entry{
					Line:      `{"time": "2024-04-16T16:40:06.177970158Z"}`,
					Timestamp: time.Date(4096, 04, 16, 16, 40, 6, 177970158, time.UTC),
				},
				want: &logs.LogMessage{
					Verbatim: &logs.VerbatimLogMessage{
						Line:      []byte(`{"time": "2024-04-16T16:40:06.177970158Z"}`),
						Timestamp: timestamppb.New(time.Date(4096, 04, 16, 16, 40, 6, 177970158, time.UTC)),
					},
					Object: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"time": structpb.NewStringValue("2024-04-16T16:40:06.177970158Z"),
						},
					},
					PpsLogMessage:   new(pps.LogMessage),
					NativeTimestamp: timestamppb.New(time.Date(2024, 04, 16, 16, 40, 6, 177970158, time.UTC)),
				},
			},
			"PPS log message": {
				entry: loki.Entry{
					Line:      `{"severity":"info","ts":"2024-04-16T21:10:37.234186851Z","message":"/usr/local/lib/python3.4/dist-packages/matplotlib/font_manager.py:273: UserWarning: Matplotlib is building the font cache using fc-list. This may take a moment.","workerId":"default-edges-v1-8sx6n","projectName":"default","pipelineName":"edges","jobId":"c4cae897bc914bd4bdb6262db038ff15","data":[{"path":"/liberty.jpg","hash":"Vp/ZEXxcM96lYfLUnnuaFECJ1j4tuvla7TsY6XGF7qU="}],"datumId":"cef4a52be60465b328ea783037b6c46531ad7cc9f4e190f9c0e548f473cd1fd1","user":true,"stream":"stderr"}`,
					Timestamp: time.Date(1066, 10, 14, 9, 0, 0, 0, time.UTC),
				},
				want: &logs.LogMessage{
					Verbatim: &logs.VerbatimLogMessage{
						Line:      []byte(`{"severity":"info","ts":"2024-04-16T21:10:37.234186851Z","message":"/usr/local/lib/python3.4/dist-packages/matplotlib/font_manager.py:273: UserWarning: Matplotlib is building the font cache using fc-list. This may take a moment.","workerId":"default-edges-v1-8sx6n","projectName":"default","pipelineName":"edges","jobId":"c4cae897bc914bd4bdb6262db038ff15","data":[{"path":"/liberty.jpg","hash":"Vp/ZEXxcM96lYfLUnnuaFECJ1j4tuvla7TsY6XGF7qU="}],"datumId":"cef4a52be60465b328ea783037b6c46531ad7cc9f4e190f9c0e548f473cd1fd1","user":true,"stream":"stderr"}`),
						Timestamp: timestamppb.New(time.Date(1066, 10, 14, 9, 0, 0, 0, time.UTC)),
					},
					Object: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"severity":     structpb.NewStringValue("info"),
							"ts":           structpb.NewStringValue("2024-04-16T21:10:37.234186851Z"),
							"message":      structpb.NewStringValue("/usr/local/lib/python3.4/dist-packages/matplotlib/font_manager.py:273: UserWarning: Matplotlib is building the font cache using fc-list. This may take a moment."),
							"workerId":     structpb.NewStringValue("default-edges-v1-8sx6n"),
							"projectName":  structpb.NewStringValue("default"),
							"pipelineName": structpb.NewStringValue("edges"),
							"jobId":        structpb.NewStringValue("c4cae897bc914bd4bdb6262db038ff15"),
							"data": structpb.NewListValue(&structpb.ListValue{
								Values: []*structpb.Value{
									structpb.NewStructValue(&structpb.Struct{
										Fields: map[string]*structpb.Value{
											"path": structpb.NewStringValue("/liberty.jpg"),
											"hash": structpb.NewStringValue("Vp/ZEXxcM96lYfLUnnuaFECJ1j4tuvla7TsY6XGF7qU=")},
									})}}),
							"datumId": structpb.NewStringValue("cef4a52be60465b328ea783037b6c46531ad7cc9f4e190f9c0e548f473cd1fd1"),
							"user":    structpb.NewBoolValue(true),
							"stream":  structpb.NewStringValue("stderr"),
						},
					},
					PpsLogMessage: &pps.LogMessage{
						Data: []*pps.InputFile{
							{
								Path: "/liberty.jpg",
								Hash: []byte("V\x9f\xd9\x11|\\3ޥa\xf2Ԟ{\x9a\x14@\x89\xd6>-\xba\xf9Z\xed;\x18\xe9q\x85\xee\xa5"),
							},
						},
						DatumId:      "cef4a52be60465b328ea783037b6c46531ad7cc9f4e190f9c0e548f473cd1fd1",
						JobId:        "c4cae897bc914bd4bdb6262db038ff15",
						Message:      "/usr/local/lib/python3.4/dist-packages/matplotlib/font_manager.py:273: UserWarning: Matplotlib is building the font cache using fc-list. This may take a moment.",
						PipelineName: "edges",
						ProjectName:  "default",
						Ts:           timestamppb.New(time.Date(2024, 04, 16, 21, 10, 37, 234186851, time.UTC)),
						User:         true,
						WorkerId:     "default-edges-v1-8sx6n",
					},
					NativeTimestamp: timestamppb.New(time.Date(2024, 04, 16, 21, 10, 37, 234186851, time.UTC)),
				},
			},
			"blank": {
				entry: loki.Entry{},
				want: &logs.LogMessage{
					Verbatim: &logs.VerbatimLogMessage{
						Timestamp: timestamppb.New(time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC)),
					},
				},
			},
			"pgbouncer": {
				entry: loki.Entry{
					Line: `2024-04-16 21:33:01.362 UTC [1] LOG S-0x557bad51cc70: pachyderm/pachyderm@10.96.52.159:5432 closing because: server idle timeout (age=1350s)`,
				},
				want: &logs.LogMessage{
					Verbatim: &logs.VerbatimLogMessage{
						Line:      []byte(`2024-04-16 21:33:01.362 UTC [1] LOG S-0x557bad51cc70: pachyderm/pachyderm@10.96.52.159:5432 closing because: server idle timeout (age=1350s)`),
						Timestamp: timestamppb.New(time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC)),
					},
				},
			},
		}
	)

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			var (
				p testPublisher
				a = adapter{
					responsePublisher: &p,
				}
			)
			if err := a.publish(ctx, c.entry); err != nil {
				t.Fatalf("unexpected error %v", err)
			}
			require.NoDiff(t, c.want, p.last.GetLog(), []cmp.Option{protocmp.Transform()}, "published log should match expectation")
		})
	}
}
