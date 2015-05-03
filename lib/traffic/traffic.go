// package traffic generates filesystem traffic useful for testing and
// benchmarking
package traffic

import ()

// RW indicates if the operation is a read or a write
type RW int

const (
	R RW = iota
	W    = iota
)

// Object enumerates the objects that exist in a filesystem
type Object int

const (
	File   Object = iota
	Commit        = iota
	Branch        = iota
)

// Op describes an operation on a filesystem
type Op struct {
	RW     RW
	Object Object
	Path   string
	Commit string
	Branch string
	Data   string
}

// Workload describe work to be done on the server.
// We generate workloads in specific ways such that read operations (o.RW == R)
// can be taken as facts about the system. For example a read operation on
// a file will indicate with its Data file what the file should look like right now
type Workload []Op

func (w Workload) FileValue(path, commit, branch string) string {
	for i := len(w) - 1; i >= 0; i-- { // iterate in reverse
		o := w[i]
		if o.RW == R {
			continue // do nothing for reads
		}
		if commit != "" && o.Object == Commit {
			branch = o.Branch
			commit = ""
			continue
		}
		if branch != "" && o.Object == Branch {
			commit = o.Commit
			branch = ""
			continue
		}
		if branch != "" && o.Object == File && branch == o.Branch {
			return o.Data
		}
	}
	return ""
}

func mAppend(m map[string][]string, key string, val string) {
	v, ok := m[key]
	if ok {
		m[key] = append(v, val)
	} else {
		m[key] = []string{val}
	}
}

// Computes Facts that can be derived from a workload.
// Facts are read Ops that should perform as true.
func (w Workload) Facts() Workload {
	res := make(Workload, 0)
	files := make(map[string]string)     // map from path to data
	members := make(map[string][]string) // map from commit to value

	for i := 0; i < len(w); i++ {
		o := w[i]
		if o.RW == R {
			continue // do nothing for reads
		}
		switch {
		case o.Object == File:
			files[o.Path] = o.Data
			mAppend(members, o.Branch, o.Path)
		case o.Object == Commit:
			members[o.Commit] = members[o.Branch]
		case o.Object == Branch:
			members[o.Branch] = members[o.Commit]
		}
	}

	for commit, names := range members {
		for _, name := range names {
			o := Op{
				RW:     R,
				Object: File,
				Path:   name,
				Commit: commit,
				Data:   files[name],
			}
			res = append(res, o)
		}
	}
	return res
}
