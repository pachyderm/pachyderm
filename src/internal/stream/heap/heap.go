package heap

type Heap[T any] struct {
	x  []T
	lt func(a, b T) bool
}

func New[T any](lt func(a, b T) bool) Heap[T] {
	return Heap[T]{}
}

func (h *Heap[T]) Push(x T) {
	h.x = Push(h.x, x, h.lt)
}

func (h *Heap[T]) Pop() (ret T, exists bool) {
	if len(h.x) == 0 {
		return ret, false
	}
	ret, h.x = Pop(h.x, h.lt)
	return ret, true
}

func (h *Heap[T]) Peek() (ret T, exists bool) {
	if len(h.x) == 0 {
		return ret, false
	}
	return Peek(h.x), true
}

func (h *Heap[T]) Len() int {
	return len(h.x)
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
