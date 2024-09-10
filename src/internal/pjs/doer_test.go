package pjs_test

import (
	"context"
	"io"
	"sync"
	"testing"
	"testing/fstest"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachd"
	intpjs "github.com/pachyderm/pachyderm/v2/src/internal/pjs"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/pjs"
	"github.com/pachyderm/pachyderm/v2/src/storage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestDoer(t *testing.T) {
	var (
		c   = pachd.NewTestPachd(t, pachd.ActivateAuthOption(""))
		ctx = c.Ctx()
		fs  = fstest.MapFS{
			"name": &fstest.MapFile{Data: []byte("test")},
		}
		filesetID string
	)
	filesetID, err := c.FileSystemToFileset(ctx, fs)
	require.NoError(t, err, "adding program fileset")
	d := intpjs.Doer{
		Client:  c.PjsAPIClient,
		Program: filesetID,
		InputTranslator: func(ctx context.Context, msg *anypb.Any) ([]string, error) {
			// just write the message to a fileset
			b, err := proto.Marshal(msg)
			if err != nil {
				return nil, errors.Wrap(err, "marshal")
			}
			cc, err := c.FilesetClient.CreateFileset(ctx)
			if err != nil {
				return nil, errors.Wrap(err, "CreateFileSet")
			}
			if err := cc.Send(&storage.CreateFilesetRequest{
				Modification: &storage.CreateFilesetRequest_AppendFile{
					AppendFile: &storage.AppendFile{
						Path: "data",
						Data: &wrapperspb.BytesValue{Value: b},
					},
				},
			}); err != nil {
				return nil, errors.Wrap(err, "send")
			}
			resp, err := cc.CloseAndRecv()
			if err != nil {
				return nil, errors.Wrap(err, "close fileset")
			}
			return []string{resp.FilesetId}, nil
		},
		OutputTranslator: func(ctx context.Context, filesetIDs []string) (*anypb.Any, error) {
			if len(filesetIDs) != 1 {
				return nil, errors.Errorf("expected 1 fileset; got %d", len(filesetIDs))
			}
			cc, err := c.FilesetClient.ReadFileset(ctx, &storage.ReadFilesetRequest{
				FilesetId: filesetIDs[0],
			})
			if err != nil {
				return nil, errors.Wrap(err, "reading fileset")
			}
			for {
				r, err := cc.Recv()
				if err != nil {
					if err == io.EOF {
						break
					}
					return nil, errors.Wrap(err, "could not receive fileset")
				}
				if r.Path != "/sum" {
					continue
				}
				var a anypb.Any
				if err := proto.Unmarshal(r.Data.Value, &a); err != nil {
					return nil, errors.Wrap(err, "could not unmarshal result")
				}
				return &a, nil
			}
			return nil, errors.New("output?unimplemented")
		},
	}
	queueID, err := intpjs.HashFS(fs)
	require.NoError(t, err, "hashing program fileset")

	t.Run("NoJobs", func(t *testing.T) {
		workCh := make(chan *anypb.Any)
		close(workCh)
		err = d.Do(ctx, workCh, func(index int64, output *anypb.Any, err error) error {
			return nil
		})
		require.NoError(t, err, "doing")
	})

	t.Run("OneJob", func(t *testing.T) {
		var wg sync.WaitGroup
		var testErr error
		defer func() {
			if testErr != nil {
				t.Error(testErr)
			}
		}()
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()
		wg.Add(1)
		go func() {
			defer wg.Done()
			q, err := c.PjsAPIClient.ProcessQueue(ctx)
			if err != nil {
				errors.JoinInto(&testErr, errors.Wrap(err, "getting queue client"))
				return
			}
			if err := q.Send(&pjs.ProcessQueueRequest{
				Queue: &pjs.Queue{Id: queueID},
			}); err != nil {
				panic(err)
			}
			for {
				select {
				case <-ctx.Done():
					return
				default:
					resp, err := q.Recv()
					if err != nil {
						if err == io.EOF || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) || status.Code(err) == codes.Canceled {
							break
						}
						errors.JoinInto(&testErr, errors.Wrap(err, "receiving"))
						return
					}
					// read inputs, sum them, write output
					var acc int64
					for _, input := range resp.Input {
						cc, err := c.FilesetClient.ReadFileset(ctx, &storage.ReadFilesetRequest{FilesetId: input})
						if err != nil {
							panic(err)
						}
						for {
							resp, err := cc.Recv()
							if err != nil {
								if err == io.EOF {
									break
								}
								panic(err)
							}
							var (
								a   anypb.Any
								qid pjs.QueueInfoDetails
							)
							if err := proto.Unmarshal(resp.Data.Value, &a); err != nil {
								if err := q.Send(&pjs.ProcessQueueRequest{
									Result: &pjs.ProcessQueueRequest_Failed{
										Failed: true,
									},
								}); err != nil {
									panic(err)
								}
								continue
							}
							if err := a.UnmarshalTo(&qid); err != nil {
								if err := q.Send(&pjs.ProcessQueueRequest{
									Result: &pjs.ProcessQueueRequest_Failed{
										Failed: true,
									},
								}); err != nil {
									panic(err)
								}
								continue
							}
							acc += qid.Size
						}
					}
					cc, err := c.FilesetClient.CreateFileset(ctx)
					if err != nil {
						panic(err)
					}
					a, err := anypb.New(&pjs.QueueInfoDetails{Size: acc})
					if err != nil {
						panic(err)
					}
					b, err := proto.Marshal(a)
					if err := cc.Send(&storage.CreateFilesetRequest{
						Modification: &storage.CreateFilesetRequest_AppendFile{
							AppendFile: &storage.AppendFile{
								Path: "sum",
								Data: &wrapperspb.BytesValue{Value: b},
							}},
					}); err != nil {
						panic(err)
					}
					fsResp, err := cc.CloseAndRecv()
					if err != nil {
						panic(err)
					}
					if err := q.Send(&pjs.ProcessQueueRequest{
						Result: &pjs.ProcessQueueRequest_Success_{
							Success: &pjs.ProcessQueueRequest_Success{
								Output: []string{fsResp.FilesetId},
							},
						},
					}); err != nil {
						panic(err)
					}
				}
			}
		}()
		workCh := make(chan *anypb.Any)
		go func() {
			// smuggle an integer value through a random PJS message
			msg, err := anypb.New(&pjs.QueueInfoDetails{Size: 1})
			if err != nil {
				panic(err)
			}
			workCh <- msg
			close(workCh)
		}()
		err = d.Do(ctx, workCh, func(index int64, output *anypb.Any, err error) error {
			var qid pjs.QueueInfoDetails
			if err != nil {
				return errors.Wrap(err, "doing")
			}
			if err := output.UnmarshalTo(&qid); err != nil {
				return errors.Wrap(err, "could not unmarshal")
			}
			return nil
		})
		cancel()
		wg.Wait()
		require.NoError(t, err, "doing")
	})
}
