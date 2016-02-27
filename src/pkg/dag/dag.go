package dag

// DAG represents a directected acyclic graph
type DAG struct {
	parents  map[string][]string
	children map[string][]string
	leaves   map[string]bool
}

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

func (d *DAG) NewNode(id string, parents []string) {
	d.parents[id] = parents
	for _, parentId := range parents {
		d.children[parentId] = append(d.children[parentId], id)
		d.leaves[parentId] = false
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

func (d *DAG) Ancestors(id string, from []string) []string {
	seen := make(map[string]bool)
	for _, fromId := range from {
		seen[fromId] = true
	}
	return dfs(id, d.parents, seen)
}

func (d *DAG) Descendants(id string, to []string) []string {
	seen := make(map[string]bool)
	for _, toId := range to {
		seen[toId] = true
	}
	return bfs(id, d.children, seen)
}

func dfs(id string, edges map[string][]string, seen map[string]bool) []string {
	if seen[id] {
		return nil
	} else {
		seen[id] = true
	}
	var result []string
	for _, nId := range edges[id] {
		result = append(result, dfs(nId, edges, seen)...)
	}
	return append(result, id)
}

func bfs(id string, edges map[string][]string, seen map[string]bool) []string {
	if seen[id] {
		return nil
	} else {
		seen[id] = true
	}
	var result []string
	queue := []string{id}
	for len(queue) != 0 {
		result = append(result, queue[0])
		queue = queue[1:]
		for _, nId := range edges[result[len(result)-1]] {
			if !seen[nId] {
				seen[nId] = true
				queue = append(queue, nId)
			}
		}
	}
	return result
}
