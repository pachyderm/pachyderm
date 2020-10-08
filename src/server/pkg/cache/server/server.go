package server

import (
	"hash/adler32"
	"strings"
	"sync"
	"time"

	pb "github.com/golang/groupcache/groupcachepb"
	"github.com/pachyderm/pachyderm/src/client/pkg/shard"
	"github.com/pachyderm/pachyderm/src/server/pkg/cache/groupcachepb"
	"github.com/pachyderm/pachyderm/src/server/pkg/log"

	"github.com/golang/groupcache"
	"golang.org/x/net/context"
)

// CacheServer serves groupcache requests over grpc.
type CacheServer interface {
	groupcachepb.GroupCacheServer
	shard.Frontend
	shard.Server
}

// NewCacheServer creates a new CacheServer.
func NewCacheServer(router shard.Router, shards uint64) CacheServer {
	server := &groupCacheServer{
		Logger:      log.NewLogger("CacheServer"),
		router:      router,
		localShards: make(map[uint64]bool),
		shards:      shards,
	}
	groupcache.RegisterPeerPicker(func() groupcache.PeerPicker { return server })
	return server
}

type groupCacheServer struct {
	log.Logger
	router      shard.Router
	localShards map[uint64]bool
	mu          sync.Mutex
	version     int64
	shards      uint64
}

func (s *groupCacheServer) Get(
	ctx context.Context,
	request *groupcachepb.GetRequest,
) (response *groupcachepb.GetResponse, retErr error) {
	func() { s.Log(request, nil, nil, 0) }()
	// response is too big to log
	defer func(start time.Time) { s.Log(request, nil, retErr, time.Since(start)) }(time.Now())
	result := &groupcachepb.GetResponse{}
	if err := groupcache.GetGroup(*request.Group).Get(ctx, *request.Key, groupcache.AllocatingByteSliceSink(&result.Value)); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *groupCacheServer) PickPeer(key string) (groupcache.ProtoGetter, bool) {
	// We split the key on `-` characters and only consider the first result,
	// this allows users of the cache to have different keys hash to the same
	// server which can be very useful, checkout objBlockAPIServer for an
	// example of how.
	shard := uint64(adler32.Checksum([]byte(strings.Split(key, ".")[0]))) % s.shards
	if s.localShards[shard] {
		return nil, false
	}
	return &protoGetter{
		shard:   shard,
		version: s.version,
		router:  s.router,
	}, true
}

type protoGetter struct {
	shard   uint64
	version int64
	router  shard.Router
}

func (p *protoGetter) Get(ctx context.Context, in *pb.GetRequest, out *pb.GetResponse) error {
	conn, err := p.router.GetClientConn(p.shard, p.version)
	if err != nil {
		return err
	}
	// TODO this is pretty dumb, we're copying to an identical struct to fool
	// the type system to get this to go away we need to not symlink groupcachepb.
	request := &groupcachepb.GetRequest{
		Group:            in.Group,
		Key:              in.Key,
		XXX_unrecognized: in.XXX_unrecognized,
	}
	response, err := groupcachepb.NewGroupCacheClient(conn).Get(context.Background(), request)
	if err != nil {
		return err
	}
	out.Value = response.Value
	out.MinuteQps = response.MinuteQps
	out.XXX_unrecognized = response.XXX_unrecognized
	return nil
}

func (s *groupCacheServer) AddShard(shard uint64) error {
	s.mu.Lock()
	s.localShards[shard] = true
	s.mu.Unlock()
	return nil
}

func (s *groupCacheServer) DeleteShard(shard uint64) error {
	s.mu.Lock()
	delete(s.localShards, shard)
	s.mu.Unlock()
	return nil
}

func (s *groupCacheServer) Version(version int64) error {
	s.version = version
	return nil
}
