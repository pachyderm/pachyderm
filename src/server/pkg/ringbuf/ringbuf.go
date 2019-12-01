package ringbuf

import (
	"fmt"
	"io"
	"sync"
)

// waiting represents a goroutine that's waiting to write data to a RingBufferAt
type waiting struct {
	sync.Mutex
	pos int64 // the first index (in the stream) to be written
}

func min(xs ...int) int64 {
	if len(xs) == 0 {
		return -1 << 63 // min int64
	}
	min := xs[0]
	for _, x := range xs[1:] {
		if x < min {
			min = x
		}
	}
	return min
}

// RingBufferAt implements a buffered pipe. Due to its ring-based model (data is
// written at a leading index and read from a trailing index) it never needs to
// reallocate. However, because it never reallocates, it must be sized
// appropriately on creation. Sizing your initial ringbuffer too small will
// yield poor performance, and sizing it too large will use a lot of memory
//
// The initial intended use case for this library is Amazon's
// s3manager.Downloader utility. The "stream" contained in this library is
// expected to correspond to a large file in S3 that is being downloaded in
// parts by the utility
//
// TODO(msteffen): maybe resize the buffer based on heuristics (e.g. too many
// waiters)
type RingBufferAt struct {
	sync.Mutex
	// Buf is the ring buffer, with a slightly unusual model: it's a window into a
	// much larger stream of bytes, to support the WriterAt interface needed by
	// the aws s3 client. 'fill' contains indices into the larger byte stream (so
	// that its elements are always sorted) and always contains at least two
	// elements, with the first index containing the trailing (i.e.  read-from)
	// end of 'buf' and the last index containing the leading (i.e.  write-to) end
	// of 'buf':
	// ---------------------------------------------------------------------------
	// buf    |......][XXXXXX.......XXXXXXX..........XX........][.................
	// stream |...XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX...
	// fill   | fill[0]^  [1]^   [2]^   [3]^      [4]^ ^[5](end)
	// ---------------------------------------------------------------------------
	// ...or, because 'buf' is a ring buffer, it may look more like:
	// ---------------------------------------------------------------------------
	// buf    |..][...............XXXXXX....][...XXXXXXX...............][.........
	// stream |...XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX...
	// fill   |            fill[0]^  [1]^     [2]^      ^[3](end)
	// ---------------------------------------------------------------------------
	buf []byte

	// fill tracks filled regions in 'buf'. See description for 'buf'.
	// Note: 'fill' satisfies the invariant
	//
	//   ∀ i ≠ 0 . fill[i] < fill[i+1], BUT fill[0] <= fill[1]
	//
	// ...because fill[0] is where new data must be written. [fill[0], fill[1])
	// is the only region that may be empty because the only way to advance
	// fill[0] is to write data into buf and then read it out.
	// Note that fill[0] == fill[1] does mean that there's no data available
	// to be read, but doesn't necessarily mean 'buf' is empty:
	// ------------------------------------------------------
	// buf    |   [.........|.....XXXXXXXXXXXXXXX...]
	// fill   |    fill[0,1]^  [2]^              ^[3](end)
	// ------------------------------------------------------
	// ...in this case, there is a gap at the beginning of buf waiting to be
	// filled by Write(p) or WriteAt(p, fill[0])
	fill []int64
	eof  bool // eof indicates that no more bytes will be written

	// readOps is a queue containing any in-progress read operations.
	readOps []readOp

	// writeQ is a priority queue (heap) of goroutines waiting to write, ordered
	// by the index to which they're planning to write (earlier indexes come
	// first).
	//
	// Like readWaitMu above, this works by taking advantage of the non-reentrancy
	// of go mutexes. Goroutines lock a mutex, add it to writeQ, then lock it
	// again. To unblock the goroutine, another thread reads from the RingBufferAt
	// (freeing up space) then unlocks the first writer, allowing the write
	// goroutine to proceed
	writeQ []writeOp
}

// b(ack) is a simple convenience function that returns the lowest index in
// b.fill (i.e. the index where data is read): [...back> XXXX...XXXX <front....]
func (b *RingBufferAt) b() int64 {
	return b.fill[0]
}

// f(ront) is a simple convenience function that returns the highest index in
// b.fill (i.e. the index where data is written): [..back> XXXX...XXXX <front..]
func (b *RingBufferAt) f() int64 {
	return b.fill[len(b.fill)-1]
}

// bfill() is a simple convenience function that maps b.fill[i] into b.buf (i.e.
// b.fill[i] % len(b.buf)). This does not do any bounds checks
func (b *RingBufferAt) bfill(i int) int {
	return int(b.fill[i] % int64(len(b.buf)))
}

// b(uf)b(ack) is a simple convenience function that returns b.b() mapped into
// b.buf (i.e. b.b()%len(buf)).
//
// Note that it's always true that b.b() < b.f(), but it may be that b.bb() >=
// b.bf(), if the buffer is cycling:
//      ..][...........back> XXXXXX..][...XXXXXXX <front........][.........
func (b *RingBufferAt) bb() int {
	return b.bfill(0)
}

// b(uf)f(ront) returns b.f() mapped into b.buf (i.e. b.b()%len(buf)). See note
// on b.bb(), that it may be that b.bf() <= b.bb()
func (b *RingBufferAt) bf() int {
	return b.bfill(len(b.fill) - 1)
}

// readOp represents an in-progress call to Read().
type readOp struct {
	sync.WaitGroup
	p []byte
}

func (r *readOp) Read() {
}

func (b *RingBufferAt) Read(p []byte) (n int, err error) {
	b.readMu.Lock()
	defer b.readMu.Lock()
	b.Lock()
	defer b.Unlock()
	// hasData :=
	for len(p) > 0 && (!b.eof || b.b() < b.f()) {
		var end int // we'll read [back, end), where back = b.back() % len(b.buf)
		// Figure out what part of b.buf we'd copy, ignoring fill for now
		if b.bf() > b.bb() {
			end = b.bf()
		} else {
			// Two cases to consider, to see that this code is correct
			// - b.bf() == b.bb(). This is true iff buf is completely full or empty
			// - b.bf() == 0.      Note that 0 == len(b.buf) % len(b.buf)
			end = len(b.buf)
		}

		// figure out if we need to copy less data due to hole in b.buf
		if dBack := int64(end - b.bb()); b.fill[1] < (b.b() + dBack) {
			end = b.bfill(1)
		}

		// Check if data is available right now, and wait in b.readCond if not
		if end <= b.bb() {
			if end < b.bb() {
				panic(fmt.Sprintf("end (%d) is less than back (%d)", end, b.bb()))
			}
			b.Unlock()
			b.readCond.Wait()
			// control has been returned -- data is available
			b.Lock()
			continue
		}

		// actually copy data from b.buf to 'p'
		dn := copy(p, b.buf[back:end])
		n += dn
		p = p[dn:]
		b.back() += int64(dn)
	}

	// return EOF if there is no more to read
	err = nil
	if b.back() >= b.front() && b.eof {
		err = io.EOF
	}

	// wake up any writers now that b.back() has advanced
	b.releaseNextWriter()
	return n, err
}

// waitToWriteAt adds the calling goroutine to b.writeQ
func (b *RingBufferAt) waitToWriteAt(pos int64) {
	w := &waiting{pos: pos}
	w.Lock()
	b.writeQ = append(b.writeQ, w)
	for child := (len(b.writeQ) - 1); child > 0; {
		parent := (child - 1) / 2
		if b.writeQ[parent].pos <= b.writeQ[child].pos {
			break
		}
		b.writeQ[parent], b.writeQ[child], child = b.writeQ[child], b.writeQ[parent], parent
	}

	// Yield control to the next goro. b.writeQ has already been modified, so no
	// further changes will be made to r until w is unlocked.
	b.Unlock()
	w.Lock() // sleep until more space is ready
	b.Lock() // resume work (will likely block on the releasing goro)
}

func (b *RingBufferAt) releaseNextWriter() {
	if len(b.writeQ) == 0 {
		return // nothing to do
	}
	end := len(b.writeQ) - 1
	// remove first element from heap
	first := b.writeQ[0]
	b.writeQ[0], b.writeQ = b.writeQ[end], b.writeQ[:end]
	// reheap last element
	for cur := 0; cur < len(b.writeQ); {
		l, r := (cur*2)+1, (cur+1)*2
		if b.writeQ[l].pos < b.writeQ[r].pos && b.writeQ[r].pos < b.writeQ[cur].pos {
			b.writeQ[cur], cur = b.writeQ[r], r
		} else if b.writeQ[r].pos <= b.writeQ[l].pos && b.writeQ[l].pos < b.writeQ[cur].pos {
			b.writeQ[cur], cur = b.writeQ[l], l
		} else {
			break
		}
	}
	// writer will still block until this returns and 'b' is unlocked
	first.Unlock()
}

func (b *RingBufferAt) Write(p []byte) (n int, err error) {
	b.Lock()
	defer b.Unlock()
	return b.write(p)
}

// write contains the implementation of b.Write, but doesn't lock 'b' so that it
// can be called by Write() and WriteAt()
func (b *RingBufferAt) write(p []byte) (n int, err error) {
	for len(p) > 0 {
		if int(b.front()-b.back()) >= len(b.buf) {
			if int(b.front()-b.back()) > len(b.buf) {
				panic(fmt.Sprintf("ringbuffer in invalid state: front-back is %d, but buf size is only %d", b.front()-b.back(), len(b.buf)))
			}
			b.waitToWriteAt(b.front()) // block until there's room to write
		}
		var end int // we'll write to [front, end), where front = b.front() % len(b.buf)

		// Figure out what part of b.buf we'd copy into. Since we're writing to the
		// end, fill don't matter
		back, front := int(b.back()%int64(len(b.buf))), int(b.front()%int64(len(b.buf)))
		if front > back {
			end = len(b.buf)
		} else {
			end = back
		}

		// actually copy data from b.buf to 'p'
		dn := copy(b.buf[front:end], p)
		n += dn
		p = p[dn:]
		b.front() += int64(dn)
	}
	return n, nil
}

// mingt binary searches through b.fill to find the least index that is greater
// than 'pos', or len(b.fill) if no such index exists
func (b *RingBufferAt) mingt(pos int64) int {
	if len(b.fill) == 0 {
		return 0
	}
	// find greatest value < pos (simpler to implement) and keep in 'm'
	l, r, m := 0, len(b.fill), 0
	for {
		switch m = (l + r) / 2; {
		case (r-l) == 1 || b.fill[m] == pos:
			break // done
		case b.fill[m] < pos:
			l = m
		case b.fill[m] > pos:
			r = m
		}
	}
	return m + 1
}

func (b *RingBufferAt) WriteAt(p []byte, pos int64) (n int, err error) {
	b.Lock()
	defer b.Unlock()
	back, front := int(b.back()%int64(len(b.buf))), int(b.front()%int64(len(b.buf)))
	if pos < b.front() {
		// fill in hole
	} else if pos == b.front() {
		return b.write(p)
	} else {
		i64Len := int64(len(b.buf))
		start, end := int(pos%i64Len), int(pos+int64(len(p))%i64Len)
		if end > len(b.buf) {
			end = len(b.buf)
		}
		if front < back && back < end {
			end = back
		}
		if end <= start {
			b.waitToWriteAt(pos)
		}

		// add new hole
		b.fill = append(b.fill, pos, pos+int64(end-start))
	}
}
