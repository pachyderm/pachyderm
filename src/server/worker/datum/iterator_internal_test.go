package datum

import (
	"archive/tar"
	"bytes"
	"context"
	io "io"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/tarutil"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type fakeTar struct {
	pfs.API_GetFileTARClient
	recvd bool
	file  []byte
}

func (ft *fakeTar) Recv() (*wrapperspb.BytesValue, error) {
	if ft.recvd {
		return nil, io.EOF
	}
	var buf = new(bytes.Buffer)
	tw := tar.NewWriter(buf)
	if err := tw.WriteHeader(tarutil.NewHeader("/meta", int64(len(ft.file)))); err != nil {
		return nil, errors.Wrapf(err, "could not create header /meta %d bytes", len(ft.file))
	}
	if _, err := tw.Write(ft.file); err != nil {
		return nil, errors.Wrap(err, "could not write meta file")
	}
	if err := tw.Flush(); err != nil {
		return nil, errors.Wrap(err, "could not flush meta file")
	}
	ft.recvd = true
	return &wrapperspb.BytesValue{Value: buf.Bytes()}, nil
}

type tarGetter struct {
	metafile []byte
}

func (tg tarGetter) GetFileTAR(ctx context.Context, in *pfs.GetFileRequest, opts ...grpc.CallOption) (pfs.API_GetFileTARClient, error) {
	return &fakeTar{
		file: tg.metafile,
	}, nil
}

func TestIterateMeta(t *testing.T) {
	var testCases = map[string]struct {
		file    string
		meta    *Meta
		wantErr bool
	}{
		"missing data should err": {wantErr: true},
		"single meta message should succeed": {
			file: `{"job":{"pipeline":{"project":{"name":"default"}, "name":"first"}, "id":"ae78f69830044393ab2ce6bf8af4f26d"}}`,
			meta: &Meta{
				Job: &pps.Job{
					Pipeline: &pps.Pipeline{
						Project: &pfs.Project{
							Name: "default",
						},
						Name: "first",
					},
					Id: "ae78f69830044393ab2ce6bf8af4f26d",
				},
			},
		},
		"double meta message should succeed": {
			file: `{"job":{"pipeline":{"project":{"name":"default"}, "name":"first"}, "id":"ae78f69830044393ab2ce6bf8af4f26d"}}{"job":{"pipeline":{"project":{"name":"default"}, "name":"first"}, "id":"1238f69830044393ab2ce6bf8af4f26d"}}`,
			meta: &Meta{
				Job: &pps.Job{
					Pipeline: &pps.Pipeline{
						Project: &pfs.Project{
							Name: "default",
						},
						Name: "first",
					},
					Id: "ae78f69830044393ab2ce6bf8af4f26d",
				},
			},
		},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			var m *Meta
			if err := iterateMeta(pctx.TestContext(t), tarGetter{metafile: []byte(testCase.file)}, nil, nil, func(_ string, mm *Meta) error {
				m = mm
				return nil
			}); err != nil {
				if testCase.wantErr {
					return
				}
				t.Fatal(err)
			}
			if testCase.wantErr {
				t.Fatal("expected error")
			}
			if diff := cmp.Diff(m, testCase.meta, protocmp.Transform()); diff != "" {
				t.Fatalf("meta:\n %s", diff)
			}
		})
	}
}
