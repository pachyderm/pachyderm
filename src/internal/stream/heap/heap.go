// Package heap implements a min-heap.
package heap

// Heap is a min-heap, implemented using a slice
type Heap[T any] struct {
	xs []T
	lt func(a, b T) bool
}

// New creates a new Heap and returns it.
// The heap will use lt for comparisons.
// lt must return true if a < b and false otherwise.
// lt should be a pure function of a and b, and should not retain either after it returns.
func New[T any](lt func(a, b T) bool) Heap[T] {
	return Heap[T]{
		lt: lt,
	}
}

func (h *Heap[T]) Push(x T) {
	h.xs = Push(h.xs, x, h.lt)
}

func (h *Heap[T]) Pop() (ret T, exists bool) {
	if len(h.xs) == 0 {
		return ret, false
	}
	ret, h.xs = Pop(h.xs, h.lt)
	return ret, true
}

func (h *Heap[T]) Peek() (ret T, exists bool) {
	if len(h.xs) == 0 {
		return ret, false
	}
	return Peek(h.xs), true
}

func (h *Heap[T]) Len() int {
	return len(h.xs)
}

// Push adds x to the heap h
func Push[E any, S ~[]E](h S, x E, lt func(a, b E) bool) S {
	h = append(h, x)
	up(h, len(h)-1, lt)
	return h
}

// Pop removes the minimum element from h and returns it and
// an updated S.
func Pop[E any, S ~[]E](h S, lt func(a, b E) bool) (E, S) {
	n := len(h) - 1
	swap(h, 0, n)
	down(h, 0, n, lt)
	return h[n], h[:n]
}

// Peek returns the lowest element of h.
// h is not modified
func Peek[E any, S ~[]E](h S) E {
	return h[0]
}

// The functions below are adapted from the standard library's heap implementation.

func up[E any, S ~[]E](x S, j int, lt func(a, b E) bool) {
	for {
		i := (j - 1) / 2 // parent
		if i == j || !lt(x[j], x[i]) {
			break
		}
		swap(x, i, j)
		j = i
	}
}

func down[E any, S ~[]E](h S, i0, n int, lt func(a, b E) bool) bool {
	i := i0
	for {
		j1 := 2*i + 1
		if j1 >= n || j1 < 0 { // j1 < 0 after int overflow
			break
		}
		j := j1 // left child
		if j2 := j1 + 1; j2 < n && lt(h[j2], h[j1]) {
			j = j2 // = 2*i + 2  // right child
		}
		if !lt(h[j], h[i]) {
			break
		}
		swap(h, i, j)
		i = j
	}
	return i > i0
}

func swap[E any, S ~[]E](x S, i, j int) {
	x[i], x[j] = x[j], x[i]
}
