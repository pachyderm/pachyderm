package dag

// DAG represents a directected acyclic graph
type DAG struct {
	parents  map[string][]string
	children map[string][]string
	leaves   map[string]bool
}

// NewDAG creates a DAG and populates it with the given nodes.
func NewDAG(nodes map[string][]string) *DAG {
	result := &DAG{
		parents:  make(map[string][]string),
		children: make(map[string][]string),
		leaves:   make(map[string]bool),
	}
	for id, parents := range nodes {
		result.NewNode(id, parents)
	}
	return result
}

// NewNode adds a node to d.
func (d *DAG) NewNode(id string, parents []string) {
	d.parents[id] = parents
	for _, parentID := range parents {
		d.children[parentID] = append(d.children[parentID], id)
		d.leaves[parentID] = false
	}
	if _, ok := d.leaves[id]; !ok {
		d.leaves[id] = true
	}
}

// Sorted returns all nodes in a topologically sorted order
func (d *DAG) Sorted() []string {
	seen := make(map[string]bool)
	var result []string
	for id := range d.parents {
		result = append(result, dfs(id, d.parents, seen)...)
	}
	return result
}

// Leaves returns a slice containing all leaves in d.
func (d *DAG) Leaves() []string {
	var result []string
	for id, isLeaf := range d.leaves {
		// isLeaf might be false, explicit mark nodes as non leaves
		if isLeaf {
			result = append(result, id)
		}
	}
	return result
}

// Ancestors returns a slice containing all ancestors of a node, 'id',
// in d which are a descendant of at least one of the nodes in 'from'.
func (d *DAG) Ancestors(id string, from []string) []string {
	seen := make(map[string]bool)
	for _, fromID := range from {
		seen[fromID] = true
	}
	return dfs(id, d.parents, seen)
}

// Descendants returns a slice containing all descendants of a node, 'id',
// in d which are an ancestor of at least one of the nodes in 'to'.
func (d *DAG) Descendants(id string, to []string) []string {
	seen := make(map[string]bool)
	for _, toID := range to {
		seen[toID] = true
	}
	return bfs(id, d.children, seen)
}

// Ghosts returns nodes that were referenced as parents but never created.
func (d *DAG) Ghosts() []string {
	var result []string
	for id := range d.children {
		if _, ok := d.parents[id]; !ok {
			result = append(result, id)
		}
	}
	return result
}

func dfs(id string, edges map[string][]string, seen map[string]bool) []string {
	if seen[id] {
		return nil
	}

	seen[id] = true

	var result []string
	for _, nID := range edges[id] {
		result = append(result, dfs(nID, edges, seen)...)
	}
	return append(result, id)
}

func bfs(id string, edges map[string][]string, seen map[string]bool) []string {
	var result []string
	queue := []string{id}
	for len(queue) != 0 {
		result = append(result, queue[0])
		queue = queue[1:]
		for _, nID := range edges[result[len(result)-1]] {
			if !seen[nID] {
				seen[nID] = true
				queue = append(queue, nID)
			}
		}
	}
	return result
}
