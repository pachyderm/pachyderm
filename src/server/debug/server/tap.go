package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/docker/go-units"
	adminv3 "github.com/envoyproxy/go-control-plane/envoy/admin/v3"
	matcherv3 "github.com/envoyproxy/go-control-plane/envoy/config/common/matcher/v3"
	tapv3 "github.com/envoyproxy/go-control-plane/envoy/config/tap/v3"
	tapdatav3 "github.com/envoyproxy/go-control-plane/envoy/data/tap/v3"
	"github.com/gogo/protobuf/types"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcapgo"
	"github.com/pachyderm/pachyderm/v2/src/debug"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/promutil"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/wrapperspb"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const snapLen = units.MiB

var errTapEnd = errors.New("tap end request")

// Trace contains one chunk of a trace, with context information for generating a synthetic pcap
// entry.
type Trace struct {
	Source     string // The envoy instance this trace is from.
	TraceID    uint64 // That envoy's internal connection ID.
	Time       time.Time
	HTTP       *tapdatav3.HttpStreamedTraceSegment
	Connection *tapdatav3.Connection
	Event      *tapdatav3.SocketEvent
	Upstream   bool // If true, this is an envoy -> backend packet, not a client -> envoy packet.
	Sequence   int
}

func ipv4() *layers.IPv4 {
	return &layers.IPv4{
		Version:  4,
		Protocol: layers.IPProtocolTCP,
		TTL:      10,
		SrcIP:    net.IPv4(0, 0, 0, 0),
		DstIP:    net.IPv4(0, 0, 0, 0),
	}
}

func addresses(c *tapdatav3.Connection, upstream bool) (*layers.IPv4, *layers.TCP) {
	if c == nil {
		return ipv4(), &layers.TCP{}
	}
	src, dst := c.GetLocalAddress().GetSocketAddress(), c.GetRemoteAddress().GetSocketAddress()
	if upstream {
		src, dst = dst, src
	}
	ip := ipv4()
	ip.SrcIP = net.ParseIP(src.GetAddress())
	ip.DstIP = net.ParseIP(dst.GetAddress())
	return ip, &layers.TCP{
		SrcPort: layers.TCPPort(src.GetPortValue()),
		DstPort: layers.TCPPort(dst.GetPortValue()),
	}
}

func (t *Trace) asPacket() (int, int, []byte, error) {
	var truncated int
	log.Info(pctx.TODO(), "packet", zap.Any("packet", t))
	ip, tcp, payload := ipv4(), &layers.TCP{}, gopacket.Payload{}
	if t.Event == nil && t.Connection != nil {
		// Do 3 way handshake.
		return 0, 0, nil, nil
	}
	if h := t.HTTP; h != nil {
		p, err := protojson.Marshal(t.HTTP)
		if err != nil {
			return 0, 0, nil, errors.Wrap(err, "http: Marshal")
		}
		payload = gopacket.Payload(p)
	}
	if e := t.Event; e != nil {
		if e.GetClosed() != nil {
			// Do FIN/ACK/FIN/ACK
			return 0, 0, nil, nil
		}
		dir := t.Upstream
		if e.GetWrite() != nil {
			dir = !dir
		}
		ip, tcp = addresses(t.Connection, dir)
		if r := e.GetRead(); r != nil {
			payload = gopacket.Payload(r.GetData().GetAsBytes())
			if r.GetData().GetTruncated() {
				truncated = 1
			}
		} else if w := e.GetWrite(); w != nil {
			payload = gopacket.Payload(w.GetData().GetAsBytes())
			if w.GetData().GetTruncated() {
				truncated = 1
			}
		}
	}
	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true,
	}
	tcp.SetNetworkLayerForChecksum(ip)
	tcp.Seq = uint32(t.Sequence)
	tcp.Window = math.MaxUint16
	log.Info(pctx.TODO(), "layers", zap.Any("layers", []gopacket.SerializableLayer{ip, tcp, payload}))
	if err := gopacket.SerializeLayers(buf, opts, ip, tcp, payload); err != nil {
		return 0, 0, nil, errors.Wrap(err, "connection: SerializeLayers")
	}
	p := buf.Bytes()
	return len(p), len(p) + truncated, p, nil
}

type fnWriter func([]byte) (int, error)

func (f fnWriter) Write(p []byte) (int, error) {
	return f(p)
}

// Tap implements debug.DebugServer.
func (s *debugServer) Tap(rs debug.Debug_TapServer) error {
	log.Debug(rs.Context(), "waiting for tap start message")
	maybeStart, err := rs.Recv()
	if err != nil {
		return status.Errorf(codes.Aborted, "stream errored before start message: %v", err)
	}
	start := maybeStart.GetStart()
	if start == nil {
		return status.Errorf(codes.InvalidArgument, "expected non-nil start request")
	}
	envoys := start.EnvoyAddresses
	configID := start.ConfigId
	if len(envoys) == 0 {
		tctx, c := context.WithTimeout(rs.Context(), 30*time.Second)
		podList, err := s.env.GetKubeClient().CoreV1().Pods(s.env.Config().Namespace).List(
			tctx,
			metav1.ListOptions{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ListOptions",
					APIVersion: "v1",
				},
				LabelSelector: metav1.FormatLabelSelector(metav1.SetAsLabelSelector(
					map[string]string{
						"suite": "pachyderm",
						"app":   "pachyderm-proxy",
					},
				)),
			},
		)
		c()
		if err != nil {
			return status.Errorf(codes.Unavailable, "discover pachyderm-proxy pods: %v", err)
		}
		for _, p := range podList.Items {
			if p.Status.Phase == v1.PodRunning {
				envoys = append(envoys, p.Status.PodIP+":9901")
			}
		}
	}
	log.Debug(rs.Context(), "starting tap", zap.Strings("envoys", envoys), zap.String("config_id", configID))

	tapRequest := &adminv3.TapRequest{
		ConfigId: configID,
		TapConfig: &tapv3.TapConfig{
			Match: &matcherv3.MatchPredicate{
				Rule: &matcherv3.MatchPredicate_AnyMatch{
					AnyMatch: true,
				},
			},
			OutputConfig: &tapv3.OutputConfig{
				Sinks: []*tapv3.OutputSink{
					{
						// If you try to get raw protos, Envoy crashes with:
						// assert failure: sink_format_ ==
						// ProtoOutputSink::JSON_BODY_AS_BYTES ||
						// sink_format_ ==
						// ProtoOutputSink::JSON_BODY_AS_STRING. Details:
						// streaming admin output only supports JSON formats
						Format: tapv3.OutputSink_JSON_BODY_AS_BYTES,
						OutputSinkType: &tapv3.OutputSink_StreamingAdmin{
							StreamingAdmin: &tapv3.StreamingAdminSink{},
						},
					},
				},
				Streaming:          true,
				MaxBufferedRxBytes: wrapperspb.UInt32(snapLen),
				MaxBufferedTxBytes: wrapperspb.UInt32(snapLen),
			},
		},
	}
	tapRequestBytes, err := protojson.Marshal(tapRequest)
	if err != nil {
		return status.Errorf(codes.Unknown, "marshal TapRequest: %v", err)
	}

	traceCh := make(chan *Trace) // The captured traces.
	logCh := make(chan string)   // Log messages from the background captures.
	doneCh := make(chan string)  // When each instance finishes, it sends its name.
	client := &http.Client{
		Transport: promutil.InstrumentRoundTripper(fmt.Sprintf("tap.%v", configID), http.DefaultTransport),
	}
	ctx, endTaps := context.WithCancelCause(rs.Context())

	// Listen for End requests in the background to finish tapping.
	go func() {
		msg, err := rs.Recv()
		if err != nil {
			endTaps(errors.Wrap(err, "recv TapRequest"))
			return
		}
		if end := msg.GetEnd(); end != nil {
			endTaps(errTapEnd)
			return
		} else {
			log.Info(rs.Context(), "client sent invalid tap request; expecting End", log.Proto("request", msg))
			endTaps(errors.New("invalid tap request"))
			return
		}
	}()

	// Start a background goroutine to read Tap data from each Envoy instance.
	for _, addr := range envoys {
		go func(addr string) {
			defer func() { doneCh <- addr }()
			logf := func(format string, args ...any) {
				logCh <- fmt.Sprintf("proxy instance %v: %s", addr, fmt.Sprintf(format, args...))
			}
			sendTrace := func(t *Trace) {
				traceCh <- t
			}
			if err := startAndReadTap(ctx, configID, tapRequestBytes, addr, client, logf, sendTrace); err != nil {
				logf("%v", err)
			}
		}(addr)
	}

	out := fnWriter(func(b []byte) (int, error) {
		if err := rs.Send(&debug.TapResponse{
			Response: &debug.TapResponse_Capture{
				Capture: &types.BytesValue{
					Value: b,
				},
			},
		}); err != nil {
			if errors.Is(err, io.EOF) {
				return 0, io.EOF
			}
			return 0, errors.Wrap(err, "send tap capture")
		}
		return len(b), nil
	})
	w, err := pcapgo.NewNgWriterInterface(out, pcapgo.NgInterface{
		Name:                configID,
		OS:                  runtime.GOOS,
		SnapLength:          snapLen,
		TimestampResolution: 9,
		LinkType:            layers.LinkTypeIPv4,
	}, pcapgo.DefaultNgWriterOptions)
	if err != nil {
		log.Error(rs.Context(), "problem creating pcap writer; ending taps", zap.Error(err))
		endTaps(err)
	}
	// Process messages from the background taps.
	for nDone := 0; nDone < len(envoys); {
		select {
		case envoy := <-doneCh:
			// An instance is done.
			log.Debug(rs.Context(), "envoy instance finished", zap.String("envoy", envoy))
			nDone++
		case <-rs.Context().Done():
			// The root context is done; the client probably went away.
			log.Debug(rs.Context(), "root context done; ending taps")
			err := fmt.Errorf("root context done: %v", context.Cause(rs.Context()))
			endTaps(err)
		case l := <-logCh:
			// Send a log message to the client.
			log.Debug(rs.Context(), l)
			if err := rs.Send(&debug.TapResponse{
				Response: &debug.TapResponse_Log_{
					Log: &debug.TapResponse_Log{
						Msg: l,
					},
				},
			}); err != nil {
				log.Error(rs.Context(), "problem sending log message; ending taps", zap.Error(err), zap.String("msg", l))
				err := errors.Wrap(err, "send log message")
				endTaps(err) // Bail out when Send stops working.
			}
		case trace := <-traceCh:
			// Send capture data to the client.
			cl, l, p, err := trace.asPacket()
			if err != nil {
				log.Error(rs.Context(), "problem converting trace to pcap; skipping", zap.Error(err))
				continue
			}
			if l > 0 {
				if err := w.WritePacket(gopacket.CaptureInfo{
					Timestamp:      trace.Time,
					CaptureLength:  cl,
					Length:         l,
					InterfaceIndex: 0,
				}, p); err != nil {
					log.Error(rs.Context(), "problem sending capture bytes; ending taps", zap.Error(err))
					err := errors.Wrap(err, "send capture bytes")
					endTaps(err) // Bail out when Send stops working.
				}
			}
		}
	}

	// All background goroutines are gone.
	log.Debug(rs.Context(), "all taps are done")
	endTaps(errors.New("everything done")) // Should be a no-op.
	if err := w.Flush(); err != nil {
		return status.Errorf(codes.DataLoss, "flush: %v", err)
	}
	close(doneCh)
	close(logCh)
	close(traceCh)
	return nil
}

func startAndReadTap(ctx context.Context, configID string, tapRequestBytes []byte, addr string, client *http.Client, logf func(string, ...any), sendTrace func(*Trace)) error {
	isUpstream := strings.HasPrefix(configID, "tap-cluster-") // This is our own naming convention.
	req, err := http.NewRequestWithContext(ctx, "POST", "http://"+addr+"/tap", bytes.NewReader(tapRequestBytes))
	if err != nil {
		return errors.Wrap(err, "NewRequestWithContext")
	}
	res, err := client.Do(req)
	if err != nil {
		return errors.Wrap(err, "connecting to tap endpoint")
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		body, err := io.ReadAll(res.Body)
		if err != nil {
			body = append(body, []byte(fmt.Sprintf("\nread error: %v", err))...)
		}
		return errors.Errorf("non-200 response: %v (%s)", res.Status, body)
	}
	logf("started reading tap data")

	var seq int
	connections := make(map[uint64]*tapdatav3.Connection)
	r := json.NewDecoder(res.Body)
	for {
		var buf json.RawMessage
		if err := r.Decode(&buf); err != nil {
			if errors.Is(err, io.EOF) {
				return errors.Wrap(err, "unexpected disconnect")
			}
			if errors.Is(err, context.Canceled) {
				logf("finished reading")
				return nil
			}
			return errors.Wrap(err, "decode tap chunk")
		}
		seq++

		var trace tapdatav3.TraceWrapper
		if err := protojson.Unmarshal(buf, &trace); err != nil {
			logf("malformed TraceWrapper: %v", err)
			continue

		}
		//exhaustive:enforce
		switch trace.GetTrace().(type) {
		case *tapdatav3.TraceWrapper_HttpBufferedTrace:
			logf("unexpected HttpBufferedTrace")
		case *tapdatav3.TraceWrapper_SocketBufferedTrace:
			logf("unexpected SocketBufferedTrace")
		case *tapdatav3.TraceWrapper_HttpStreamedTraceSegment:
			sendTrace(&Trace{
				Source:   addr,
				Time:     time.Now(),
				TraceID:  trace.GetHttpStreamedTraceSegment().GetTraceId(),
				HTTP:     trace.GetHttpStreamedTraceSegment(),
				Sequence: seq,
			})
		case *tapdatav3.TraceWrapper_SocketStreamedTraceSegment:
			ssts := trace.GetSocketStreamedTraceSegment()
			tid := ssts.GetTraceId()
			// exhaustive:enforce
			switch ssts.GetMessagePiece().(type) {
			case *tapdatav3.SocketStreamedTraceSegment_Connection:
				c := ssts.GetConnection()
				connections[tid] = c
				sendTrace(&Trace{
					TraceID:    tid,
					Time:       time.Now(),
					Source:     addr,
					Connection: c,
					Upstream:   isUpstream,
					Sequence:   seq,
				})
			case *tapdatav3.SocketStreamedTraceSegment_Event:
				e := ssts.GetEvent()
				c := connections[tid]
				if _, ok := e.GetEventSelector().(*tapdatav3.SocketEvent_Closed_); ok {
					delete(connections, tid)
				}
				sendTrace(&Trace{
					TraceID:    tid,
					Time:       e.GetTimestamp().AsTime(),
					Source:     addr,
					Connection: c,
					Event:      e,
					Upstream:   isUpstream,
					Sequence:   seq,
				})
			}
		}
	}
}
