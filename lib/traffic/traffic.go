// package traffic generates filesystem traffic useful for testing and
// benchmarking
package traffic

import (
	"fmt"
	"math/rand"
	"reflect"
)

// RW indicates if the operation is a read or a write
type RW int

const (
	R RW = iota
	W    = iota
)

// Object enumerates the objects that exist in a filesystem
type Object int

const (
	File    Object = iota
	Commit         = iota
	Branch         = iota
	nObject        = iota
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

	for _, o := range w {
		if o.RW == R {
			continue // do nothing for reads
		}
		switch o.Object {
		case File:
			files[o.Path] = o.Data
			mAppend(members, o.Branch, o.Path)
		case Commit:
			members[o.Commit] = make([]string, len(members[o.Branch]))
			copy(members[o.Commit], members[o.Branch])
		case Branch:
			members[o.Branch] = make([]string, len(members[o.Commit]))
			copy(members[o.Branch], members[o.Commit])
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

func randObject(rand *rand.Rand) Object {
	roll := rand.Int() % 32
	switch {
	case roll == 0:
		return Branch
	case roll < 8:
		return Commit
	default:
		return File
	}
}

var letters = []rune("abcdefghijklmnopqrstuvwxyz")

// Generates a random sequence of letters. Useful for making filesystems that won't interfere with each other.
// This should be factored out to another file.
func RandWord(n int, rand *rand.Rand) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func (w Workload) Generate(rand *rand.Rand, size int) reflect.Value {
	res := make(Workload, 0)
	branches := []string{"master"}
	commits := []string{}
	var i int
	for i = 0; i < size*2; i++ {
		o := Op{RW: W, Object: randObject(rand)}
		switch o.Object {
		case File:
			o.Path = fmt.Sprintf("file%.10d", i)
			o.Branch = branches[rand.Int()%len(branches)]
			nWords := rand.Intn(400) + 100 // nWords in [100, 500)
			for i := 0; i < nWords; i++ {
				wordLength := rand.Intn(6) + 2 //wordLength in [2,8)
				o.Data += RandWord(wordLength, rand)
				o.Data += " "
			}
		case Commit:
			if len(branches) == 0 {
				continue
			}
			o.Commit = fmt.Sprintf("commit%.10d", i)
			o.Branch = branches[rand.Intn(len(branches))]
			commits = append(commits, o.Commit)
		case Branch:
			if len(commits) == 0 {
				continue
			}
			o.Branch = fmt.Sprintf("branch%.10d", i)
			o.Commit = commits[rand.Intn(len(commits))]
			branches = append(branches, o.Branch)
		}
		res = append(res, o)
	}
	// We add a commit for every branch so there are no dirty writes
	for _, b := range branches {
		res = append(res,
			Op{
				RW:     W,
				Object: Commit,
				Branch: b,
				Commit: fmt.Sprintf("commit%.10d", i),
			})
		i++
	}
	return reflect.ValueOf(res)
}
