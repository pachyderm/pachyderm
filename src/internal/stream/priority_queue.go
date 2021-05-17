package stream

import (
	"io"
	"sort"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

// Stream is the standard interface for a sorted stream that can be used in a PriorityQueue.
type Stream interface {
	Next() error
	Compare(Stream) int
}

type stream struct {
	stream   Stream
	priority int
}

// PriorityQueue implements a priority queue that operates on streams.
type PriorityQueue struct {
	queue []*stream
	size  int
	ss    []*stream
}

// NewPriorityQueue creates a new priority queue.
func NewPriorityQueue(ss []Stream) *PriorityQueue {
	var streams []*stream
	for _, s := range ss {
		streams = append(streams, &stream{
			stream:   s,
			priority: len(streams),
		})
	}
	return &PriorityQueue{
		queue: make([]*stream, len(ss)+1),
		ss:    streams,
	}
}

// Iterate iterates through the priority queue.
func (pq *PriorityQueue) Iterate(cb func([]Stream) error) error {
	for {
		ss, err := pq.next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		if err := cb(collectStreams(ss)); err != nil {
			return err
		}
	}
}

func collectStreams(ss []*stream) []Stream {
	var streams []Stream
	for _, s := range ss {
		streams = append(streams, s.stream)
	}
	return streams
}

func (pq *PriorityQueue) compare(i, j int) int {
	si := pq.queue[i]
	sj := pq.queue[j]
	return si.stream.Compare(sj.stream)
}

func (pq *PriorityQueue) empty() bool {
	return len(pq.queue) == 1 || pq.queue[1] == nil
}

func (pq *PriorityQueue) insert(s *stream) error {
	// Get next in stream and insert it.
	if err := s.stream.Next(); err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}
		return err
	}
	pq.queue[pq.size+1] = s
	pq.size++
	// Propagate insert up the queue
	i := pq.size
	for i > 1 {
		if pq.compare(i/2, i) < 0 {
			break
		}
		pq.swap(i/2, i)
		i /= 2
	}
	return nil
}

func (pq *PriorityQueue) next() ([]*stream, error) {
	// Re-insert streams
	if pq.ss != nil {
		for _, s := range pq.ss {
			if err := pq.insert(s); err != nil {
				return nil, err
			}
		}
	}
	if pq.empty() {
		return nil, io.EOF
	}
	ss := []*stream{pq.queue[1]}
	pq.fill()
	// Keep popping streams off the queue if they are equal.
	for pq.queue[1] != nil && pq.queue[1].stream.Compare(ss[0].stream) == 0 {
		ss = append(ss, pq.queue[1])
		pq.fill()
	}
	sort.SliceStable(ss, func(i, j int) bool {
		return ss[i].priority < ss[j].priority
	})
	pq.ss = ss
	return ss, nil
}

func (pq *PriorityQueue) fill() {
	// Replace first stream with last
	pq.queue[1] = pq.queue[pq.size]
	pq.queue[pq.size] = nil
	pq.size--
	// Propagate last stream down the queue
	i := 1
	var next int
	for {
		left, right := i*2, i*2+1
		if left > pq.size {
			break
		} else if right > pq.size || pq.compare(left, right) < 0 {
			next = left
		} else {
			next = right
		}
		if pq.compare(i, next) < 0 {
			break
		}
		pq.swap(i, next)
		i = next
	}
}

func (pq *PriorityQueue) swap(i, j int) {
	pq.queue[i], pq.queue[j] = pq.queue[j], pq.queue[i]
}
