package fileset

import "io"

type stream interface {
	next() error
	key() string
}

type priorityQueue struct {
	queue []stream
	size  int
}

func newPriorityQueue(ss []stream) *priorityQueue {
	q := make([]stream, len(ss)+1)
	// Insert streams.
	for _, s := range ss {
		if err := q.insert(s); err != nil {
			return err
		}
	}
	return &priorityQueue{queue: q}
}

func (pq *priorityQueue) key(i int) string {
	return pq.queue[i].key()
}

func (pq *priorityQueue) empty() bool {
	return pq.queue[1] == nil
}

func (pq *priorityQueue) insert(s stream) error {
	// Get next in stream and insert it.
	if err := s.next(); err != nil {
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
		if pq.key(i/2) <= pq.key(i) {
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
		for _, s := range ss {
			if err := pq.insert(s); err != nil {
				return err
			}
		}
	}
	if pq.empty() {
		return nil, io.EOF
	}
	ss := []stream{pq.queue[1]}
	pq.fill()
	// Keep popping streams off the queue if they have the same key.
	for pq.queue[1] != nil && pq.key(1) == ss[0].key() {
		ss = append(ss, pq.queue[1])
		pq.fill()
	}
	pq.ss = ss
	return ss
}

func (pq *priorityQueue) peek() string {
	return pq.queue[1].key()
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
		} else if right > pq.size || pq.key(left) <= pq.key(right) {
			next = left
		} else {
			next = right
		}
		if pq.key(i) <= pq.key(next) {
			break
		}
		pq.swap(i, next)
		i = next
	}
}

func (pq *priorityQueue) swap(i, j int) {
	pq.queue[i], pq.queue[j] = pq.queue[j], pq.queue[i]
}
