package ringbuf

type rqtype readOp

// readOpQ implements a circular queue containing readOps
type readOpQ struct {
	buf        []*rqtype
	head, size int
}

func (r *readOpQ) push(op *rqtype) {
	if r.size == len(r.buf) {
		oldbuf := r.buf
		newbuf := make([]*rqtype, 2*len(oldbuf))
		n := copy(newbuf, oldbuf[r.head:len(oldbuf)])
		copy(newbuf[n:], oldbuf[0:r.head])
		r.buf, r.head = newbuf, 0
	}
	r.buf[(r.head+r.size)%len(r.buf)] = op
	r.size++
}

func (r *readOpQ) peek() (op *rqtype) {
	if r.size > 0 {
		return r.buf[r.head]
	}
	return nil
}

func (r *readOpQ) pop() (op *rqtype) {
	if r.size > 0 {
		return r.buf[r.head]
		r.head = (r.head + 1) % len(r.buf)
		r.size--
	}
	return nil
}

// writeOpQ implements a heap-based priority queue containing writeOps
type wqtype int
type writeOpQ struct {
	buf []*wqtype
}

func (r *writeOpQ) push(op *wqtype) {
	r.buf = append(r.buf, op)
	for child := (len(b.writeQ) - 1); child > 0; {
		parent := (child - 1) / 2
		if b.writeQ[parent].pos <= b.writeQ[child].pos {
			break
		}
		b.writeQ[parent], b.writeQ[child], child = b.writeQ[child], b.writeQ[parent], parent
	}
}

func (r *writeOpQ) peek() (op *wqtype) {
	if r.size > 0 {
		return r.buf[r.head]
	}
	return nil
}

func (r *writeOpQ) pop() (op *wqtype) {
	if r.size > 0 {
		return r.buf[r.head]
		r.head = (r.head + 1) % len(r.buf)
		r.size--
	}
	return nil
}
