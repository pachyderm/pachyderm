package fakegithub

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"
)

type Server struct {
	*http.ServeMux
	prs          []*PR
	pagesFetched int
	nextPRNum    int
}

func New() *Server {
	r := &Server{
		ServeMux: http.NewServeMux(),
		prs:      make([]*PR, 0, 32),
	}
	r.HandleFunc("/repos/pachyderm/pachyderm/pulls", r.listHandler)
	r.HandleFunc("/", r.defaultHandler)
	return r
}

func (f *Server) AddPR(title, body string, author *User, created time.Time) {
	if created.IsZero() {
		panic("must set 'created' time when calling AddPR")
	}
	if len(f.prs) == 0 {
		f.nextPRNum = 100_000
	}

	pr := NewPRSpec(f.nextPRNum, title, body, author, created)
	f.nextPRNum++

	// maintain that f.prs is sorted by ascending creation time
	idx, _ := sort.Find(len(f.prs), func(i int) int {
		return pr.Created.Compare(f.prs[i].Created)
	})
	f.prs = append(f.prs, nil)
	if idx+1 < len(f.prs) {
		copy(f.prs[idx+1:], f.prs[idx:len(f.prs)-1])
	}
	f.prs[idx] = pr
}

func (f *Server) defaultHandler(w http.ResponseWriter, r *http.Request) {
	if !strings.Contains(r.URL.Path, "api") {
		// For some reason, go-github seems to spam /api and /apis, which AFAICT are
		// not valid paths of api.github.com. Everything works if I return 404 for
		// those calls (which is what GitHub returned when I curl'ed /api). Now I
		// don't even bother logging those requests.
		fmt.Printf("Request %s %q (sending 404)\n", r.Method, r.URL.Path)
	}
	http.Error(w, "404 Not Found", http.StatusNotFound)
}

func (f *Server) Start() {
	go http.ListenAndServe(":8080", f)
}
