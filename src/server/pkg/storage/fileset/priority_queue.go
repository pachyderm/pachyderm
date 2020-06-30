package fileset

import (
	"io"

	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
)

type stream interface {
	next() error
	key() string
	streamPriority() int
}

type priorityQueue struct {
	queue []stream
	size  int
	ss    []stream
}

func newPriorityQueue(ss []stream) *priorityQueue {
	return &priorityQueue{
		queue: make([]stream, len(ss)+1),
		ss:    ss,
	}
}

func (pq *priorityQueue) iterate(f func([]stream, ...string) error) error {
	for {
		ss, err := pq.next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		if err := f(ss, pq.peek()...); err != nil {
			return err
		}
	}
}

func (pq *priorityQueue) isHigherPriority(i, j int) bool {
	si := pq.queue[i]
	sj := pq.queue[j]
	return si.key() < sj.key() || (si.key() == sj.key() && si.streamPriority() < sj.streamPriority())
}

func (pq *priorityQueue) empty() bool {
	return len(pq.queue) == 1 || pq.queue[1] == nil
}

func (pq *priorityQueue) insert(s stream) error {
	// Get next in stream and insert it.
	if err := s.next(); err != nil {
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
		if pq.isHigherPriority(i/2, i) {
			break
		}
		pq.swap(i/2, i)
		i /= 2
	}
	return nil
}

func (pq *priorityQueue) next() ([]stream, error) {
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
	ss := []stream{pq.queue[1]}
	pq.fill()
	// Keep popping streams off the queue if they have the same key.
	for pq.queue[1] != nil && pq.queue[1].key() == ss[0].key() {
		ss = append(ss, pq.queue[1])
		pq.fill()
	}
	pq.ss = ss
	return ss, nil
}

func (pq *priorityQueue) peek() []string {
	if pq.empty() {
		return nil
	}
	return []string{pq.queue[1].key()}
}

func (pq *priorityQueue) fill() {
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

func (pq *priorityQueue) swap(i, j int) {
	pq.queue[i], pq.queue[j] = pq.queue[j], pq.queue[i]
}
