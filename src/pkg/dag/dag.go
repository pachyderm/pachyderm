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
		result = d.visit(id, seen, result, parent)
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
	return d.visit(id, seen, nil, parent)
}

func (d *DAG) Descendants(id string, to []string) []string {
	seen := make(map[string]bool)
	for _, toId := range to {
		seen[toId] = true
	}
	return d.visit(id, seen, nil, child)
}

type relation int

const (
	parent relation = iota
	child
)

func (d *DAG) visit(id string, seen map[string]bool, result []string, r relation) []string {
	if seen[id] {
		return result
	} else {
		seen[id] = true
	}
	if r == parent {
		for _, parentId := range d.parents[id] {
			result = d.visit(parentId, seen, result, r)
		}
		result = append(result, id)
	} else {
		result = append(result, id)
		for _, childId := range d.children[id] {
			result = d.visit(childId, seen, result, r)
		}
	}
	return result
}
