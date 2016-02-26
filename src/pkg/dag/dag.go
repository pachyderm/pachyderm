package dag

// DAG represents a directected acyclic graph
type DAG struct {
	parents  map[string][]string
	children map[string][]string
	leaves   map[string]bool
}

func (d DAG) NewNode(id string, parents []string) {
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
func (d DAG) Sorted() []string {
	seen := make(map[string]bool)
	var result []string
	for id := range d.parents {
		d.visit(id, seen, result, parent)
	}
	return result
}

func (d DAG) Leaves() []string {
	var result []string
	for leaf := range d.leaves {
		result = append(result, leaf)
	}
	return result
}

func (d DAG) Ancestors(id string, from []string) []string {
	seen := make(map[string]bool)
	for _, fromId := range from {
		seen[fromId] = true
	}
	var result []string
	d.visit(id, seen, result, parent)
	return result
}

func (d DAG) Descendants(id string, to []string) []string {
	seen := make(map[string]bool)
	for _, toId := range to {
		seen[toId] = true
	}
	var result []string
	d.visit(id, seen, result, child)
	return result
}

type relation int

const (
	parent relation = iota
	child
)

func (d DAG) visit(id string, seen map[string]bool, result []string, r relation) {
	if seen[id] {
		return
	} else {
		seen[id] = true
	}
	if r == parent {
		for _, parentId := range d.parents[id] {
			d.visit(parentId, seen, result, r)
		}
	} else {
		for _, childId := range d.children[id] {
			d.visit(childId, seen, result, r)
		}
	}
	result = append(result, id)
}
