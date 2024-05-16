package index

import (
	"github.com/emicklei/dot"
	"strings"
	"sync"
)

// Option configures an index reader.
type Option func(r *Reader)

// WithRange sets a range filter for the read.
func WithRange(pathRange *PathRange) Option {
	return func(r *Reader) {
		r.filter = &pathFilter{pathRange: pathRange}
	}
}

// WithPrefix sets a prefix filter for the read.
func WithPrefix(prefix string) Option {
	return func(r *Reader) {
		r.filter = &pathFilter{prefix: prefix}
	}
}

// WithDatum adds a datum filter that matches a single datum.
func WithDatum(datum string) Option {
	return func(r *Reader) {
		r.datum = datum
	}
}

func WithPeek() Option {
	return func(r *Reader) {
		r.peek = true
	}
}

// WithShardConfig sets the sharding configuration.
func WithShardConfig(config *ShardConfig) Option {
	return func(r *Reader) {
		r.shardConfig = config
	}
}

// WithGraph adds graph nodes that are traversed by underlying index readers.
func WithGraph(g *dot.Graph) Option {
	mu := sync.Mutex{}
	return func(r *Reader) {
		r.mutex = &mu
		mu.Lock()
		name := make([]string, 0)
		if r.name != "" {
			name = strings.Split(r.name, "-")
		}
		s := g.Subgraph(r.name, dot.ClusterOption{})
		fileset, found := g.FindNodeWithLabel(name[0])
		if found {
			root := s.Node(name[1])
			root.Attrs("color", "black")
			fileset.Edge(root)
		}
		r.graph = s
		mu.Unlock()
	}
}

// WithName is used to sort subgraphs
func WithName(name string) Option {
	return func(r *Reader) {
		r.name = name
	}
}
