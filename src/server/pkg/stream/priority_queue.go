package stream

import (
	"io"
)

// Stream is the standard interface for a sorted stream that can be used in a PriorityQueue.
type Stream interface {
	Next() error
	Key() string
	Priority() int
}

// PriorityQueue implements a priority queue that operates on streams.
type PriorityQueue struct {
	queue []Stream
	size  int
	ss    []Stream
}

// NewPriorityQueue creates a new priority queue.
func NewPriorityQueue(ss []Stream) *PriorityQueue {
	return &PriorityQueue{
		queue: make([]Stream, len(ss)+1),
		ss:    ss,
	}
}

// Iterate iterates through the priority queue.
func (pq *PriorityQueue) Iterate(cb func([]Stream, ...string) error) error {
	for {
		ss, err := pq.Next()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		if err := cb(ss, pq.Peek()...); err != nil {
			return err
		}
	}
}

func (pq *PriorityQueue) isHigherPriority(i, j int) bool {
	si := pq.queue[i]
	sj := pq.queue[j]
	return si.Key() < sj.Key() || (si.Key() == sj.Key() && si.Priority() < sj.Priority())
}

func (pq *PriorityQueue) empty() bool {
	return len(pq.queue) == 1 || pq.queue[1] == nil
}

func (pq *PriorityQueue) insert(s Stream) error {
	// Get next in stream and insert it.
	if err := s.Next(); err != nil {
		if err == io.EOF {
			return nil
		}
		return err
	}
	pq.queue[pq.size+1] = s
	pq.size++
	// Propagate insert up the queue
	i := pq.size
	for i > 1 {
		if pq.isHigherPriority(i/2, i) {
			break
		}
		pq.swap(i/2, i)
		i /= 2
	}
	return nil
}

// Next gets the next set of streams in the priority queue.
func (pq *PriorityQueue) Next() ([]Stream, error) {
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
	ss := []Stream{pq.queue[1]}
	pq.fill()
	// Keep popping streams off the queue if they have the same key.
	for pq.queue[1] != nil && pq.queue[1].Key() == ss[0].Key() {
		ss = append(ss, pq.queue[1])
		pq.fill()
	}
	pq.ss = ss
	return ss, nil
}

// Peek peeks the next key in the priority queue.
func (pq *PriorityQueue) Peek() []string {
	if pq.empty() {
		return nil
	}
	return []string{pq.queue[1].Key()}
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
		} else if right > pq.size || pq.isHigherPriority(left, right) {
			next = left
		} else {
			next = right
		}
		if pq.isHigherPriority(i, next) {
			break
		}
		pq.swap(i, next)
		i = next
	}
}

func (pq *PriorityQueue) swap(i, j int) {
	pq.queue[i], pq.queue[j] = pq.queue[j], pq.queue[i]
}
