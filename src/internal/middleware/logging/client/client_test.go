package client

import (
	"regexp"
	"strconv"
	"testing"

	grpc_testing "github.com/grpc-ecosystem/go-grpc-middleware/testing"
	testproto "github.com/grpc-ecosystem/go-grpc-middleware/testing/testproto"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

type InterceptorTestSuite struct {
	grpc_testing.InterceptorTestSuite
}

func (s *InterceptorTestSuite) matchesRegex(got []string, want []*regexp.Regexp, matchAnywhere ...*regexp.Regexp) {
	var errored bool
	defer func() {
		if errored {
			s.T().Log("got:")
			for i, x := range got {
				s.T().Logf("%v: %v", i, x)
			}
		}
	}()
	errored = true
	s.Require().Equal(len(want)+len(matchAnywhere), len(got), "len(got) should equal len(want)+len(matchAnywhere)")
	errored = false

	matchers := make(map[*regexp.Regexp]struct{})
	for _, rx := range matchAnywhere {
		matchers[rx] = struct{}{}
	}
	var skipped int
outer:
	for i, x := range got {
		for matcher := range matchers {
			if matcher.MatchString(x) {
				delete(matchers, matcher)
				skipped++
				continue outer
			}
		}
		if i > len(want) {
			s.Failf("out of regexps because of faulty non-matching matchAnywhere expression", "(line %d)", i)
			continue
		}
		rx := want[i-skipped]
		matches := rx.MatchString(x)
		s.Truef(matches, "/%s/ should match %s (line %d)", rx.String(), x, i)
		if !matches {
			errored = true
		}
	}

	var remainingMatchers []string
	for matcher := range matchers {
		remainingMatchers = append(remainingMatchers, matcher.String())
	}
	s.Equalf(0, len(matchers), "all matchAnywhere expressions should have been consumed, but still have %v", remainingMatchers)
}

func (s *InterceptorTestSuite) TestUnarySuccess() {
	ctx, h := log.TestWithCapture(s.T())
	client := s.NewClient(grpc.WithUnaryInterceptor(LogUnary))
	res, err := client.Ping(ctx, &testproto.PingRequest{Value: "test"})
	s.Require().NoError(err, "ping should succeed")
	s.Equal("test", res.GetValue(), "response should echo the request")

	var got []string
	for _, l := range h.Logs() {
		got = append(got, l.String())
	}
	s.matchesRegex(got, []*regexp.Regexp{
		regexp.MustCompile("Ping: span start"),
		regexp.MustCompile("request"),
		regexp.MustCompile("Ping: span finished ok"),
	})
}

func (s *InterceptorTestSuite) TestUnaryError() {
	ctx, h := log.TestWithCapture(s.T())
	client := s.NewClient(grpc.WithUnaryInterceptor(LogUnary))
	_, err := client.PingError(ctx, &testproto.PingRequest{ErrorCodeReturned: uint32(codes.DataLoss)})
	s.Require().Error(err, "ping should fail with DataLoss")

	var got []string
	for _, l := range h.Logs() {
		got = append(got, l.String())
	}
	s.matchesRegex(got, []*regexp.Regexp{
		regexp.MustCompile("PingError: span start"),
		regexp.MustCompile("request"),
		regexp.MustCompile("PingError: span failed.*DataLoss"),
	})
}

func (s *InterceptorTestSuite) TestBidirectionalStreamSuccess() {
	ctx, h := log.TestWithCapture(s.T())
	ctx = metadata.AppendToOutgoingContext(ctx, "foo", "bar")
	client := s.NewClient(grpc.WithStreamInterceptor(LogStream))
	c, err := client.PingStream(ctx)
	s.Require().NoError(err, "PingStream should start ok")

	for i := 0; i < 3; i++ {
		s.Require().NoErrorf(c.Send(&testproto.PingRequest{Value: strconv.Itoa(i)}), "send %d: should succeed", i)
		res, err := c.Recv()
		s.Require().NoErrorf(err, "recv %d: should succeed", i)
		s.Equal(strconv.Itoa(i), res.GetValue(), "response should echo the request")
	}
	s.Require().NoError(c.CloseSend())
	s.Require().Error(c.RecvMsg(nil))

	var got []string
	for _, l := range h.Logs() {
		got = append(got, l.String())
	}
	s.matchesRegex(got, []*regexp.Regexp{
		regexp.MustCompile(`PingStream: span start`),
		regexp.MustCompile(`stream started.*metadata=map\[foo:bar\]`),
		regexp.MustCompile(`value:"0"`), // the capitalization difference exists in the underlying proto
		regexp.MustCompile(`Value:"0"`),
		regexp.MustCompile(`value:"1"`),
		regexp.MustCompile(`Value:"1"`),
		regexp.MustCompile(`value:"2"`),
		regexp.MustCompile(`Value:"2"`),
		regexp.MustCompile(`send side of stream closed`),
		regexp.MustCompile(`PingStream: span finished ok`),
	}, regexp.MustCompile(`server headers headers=map\[`))
}

func TestInterceptors(t *testing.T) {
	// NOTE(jonathan): This test suite is missing a client streaming RPC, which we have a little
	// bit of special code to handle.
	suite.Run(t, new(InterceptorTestSuite))
}
