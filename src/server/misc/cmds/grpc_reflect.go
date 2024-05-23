package cmds

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

type rpcFunc func(context.Context, grpc.ClientConnInterface, io.Writer, ...string) error

func buildRPCs(protos ...protoreflect.FileDescriptor) map[string]rpcFunc {
	result := make(map[string]rpcFunc)
	for _, fd := range protos {
		svcs := fd.Services()
		for si := 0; si < svcs.Len(); si++ {
			sd := svcs.Get(si)
			methods := sd.Methods()
			for mi := 0; mi < methods.Len(); mi++ {
				md := methods.Get(mi)

				name := "/" + path.Join(string(sd.FullName()), string(md.Name()))
				in := md.Input()
				out := md.Output()
				switch {
				case md.IsStreamingClient() || md.IsStreamingServer():
					result[string(md.FullName())] = makeStreaming(name, in, out, md.IsStreamingServer(), md.IsStreamingClient())
				default:
					result[string(md.FullName())] = makeUnary(name, in, out)
				}
			}
		}
	}
	return result
}

func makeStreaming(name string, reqType, resType protoreflect.MessageDescriptor, isServerStreaming, isClientStreaming bool) rpcFunc {
	return func(ctx context.Context, cc grpc.ClientConnInterface, w io.Writer, req ...string) (retErr error) {
		if !isClientStreaming && len(req) > 1 {
			return errors.New("too many messages for a server streaming RPC")
		}
		if len(req) == 0 {
			req = append(req, `{}`)
		}

		ctx, killRPC := context.WithCancelCause(ctx)
		defer killRPC(nil)
		defer func() {
			select {
			case <-ctx.Done():
				errors.JoinInto(&retErr, errors.Wrap(context.Cause(ctx), "context error"))
			default:
			}
		}()

		sd := &grpc.StreamDesc{
			StreamName: name,
			// Note that this code doesn't care about whether this is bidirectional or
			// not; we treat all RPCs as bidirectional except for the input validation
			// above.
			ServerStreams: isServerStreaming,
			ClientStreams: isClientStreaming,
		}
		stream, err := cc.NewStream(ctx, sd, name)
		if err != nil {
			return errors.Wrap(handleGrpcError(err), "start stream")
		}

		go func() {
			defer func() {
				if err := stream.CloseSend(); err != nil {
					fmt.Fprintf(os.Stderr, "CloseSend: %v", handleGrpcError(err))
				}
			}()
			for i, msg := range req {
				req := dynamicpb.NewMessage(reqType)
				if err := protojson.Unmarshal([]byte(msg), req); err != nil {
					killRPC(errors.Wrapf(err, "unmarshal message %v of type %v", i, req.Descriptor().FullName()))
					return
				}
				if err := stream.SendMsg(req); err != nil {
					killRPC(errors.Wrapf(err, "send message %v", i))
					return
				}
			}
			if isClientStreaming {
				fmt.Fprintln(os.Stderr, "All messages sent")
			}
		}()

		for i := 0; ; i++ {
			msg := dynamicpb.NewMessage(resType)
			if err := stream.RecvMsg(msg); err != nil {
				if errors.Is(err, io.EOF) {
					return nil
				}
				return errors.Wrapf(handleGrpcError(err), "receive message %d", i)
			}
			js, err := protojson.Marshal(msg)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n    (problem marshaling to JSON: %v)", msg.String(), err)
			}
			fmt.Fprintf(w, "%s\n", js)
		}
	}
}

func makeUnary(name string, reqType, resType protoreflect.MessageDescriptor) rpcFunc {
	return func(ctx context.Context, cc grpc.ClientConnInterface, w io.Writer, req ...string) error {
		if len(req) > 1 {
			return errors.New("too many messages for a unary RPC")
		}
		if len(req) == 0 {
			req = append(req, `{}`)
		}
		in := dynamicpb.NewMessage(reqType)
		if err := protojson.Unmarshal([]byte(req[0]), in); err != nil {
			return errors.Wrapf(err, "unmarshal request into %v", in.Descriptor().FullName())
		}

		out := dynamicpb.NewMessage(resType)
		if err := cc.Invoke(ctx, name, in, out); err != nil {
			return errors.Wrap(handleGrpcError(err), "invoke")
		}
		js, err := protojson.Marshal(out)
		if err != nil {
			return errors.Wrap(err, "marshal response")
		}
		fmt.Fprintf(w, "%s\n", js)
		return nil
	}
}

func handleGrpcError(err error) error {
	s, ok := status.FromError(err)
	if ok {
		p := s.Proto()
		var details []string
		for _, d := range p.GetDetails() {
			if js, err := protojson.Marshal(d); err != nil {
				details = append(details, fmt.Sprintf("    %v (%v)", d.String(), err))
			} else {
				details = append(details, "    "+string(js))
			}
		}
		if len(details) > 0 {
			return errors.Errorf("%v\n  details:\n%s", err, strings.Join(details, "\n"))
		}
	}
	return err
}
