package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

type fakeGitHub struct {
	*http.ServeMux
	prs          []*prSpec
	pagesFetched int
	nextPRNum    int
}

func newFakeGitHub() *fakeGitHub {
	r := &fakeGitHub{
		ServeMux: http.NewServeMux(),
		prs:      make([]*prSpec, 0, 32),
	}
	r.HandleFunc("/repos/pachyderm/pachyderm/pulls", r.listHandler)
	r.HandleFunc("/", r.defaultHandler)
	return r
}

func (f *fakeGitHub) AddPR(title, body string, author *githubUser, created time.Time) {
	if created.IsZero() {
		if len(f.prs) == 0 {
			created = earliestPR
		} else {
			created = f.prs[0].Created.Add(time.Hour)
		}
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

func ceil(n, d int) int {
	// See http://blog.pkh.me/p/36-figuring-out-round%2C-floor-and-ceil-with-integer-division.html
	return (n-1)/d + 1
}

func (f *fakeGitHub) getListParams(query url.Values) (dir string, page, perPage int, err error) {
	// set default values
	dir = "desc"
	page = 1
	perPage = 7

	// check for parameter values
	if query.Has("direction") {
		dir = query.Get("direction")
	}
	if dir != "asc" && dir != "desc" {
		return "", 0, 0, fmt.Errorf("invalid parameter value %q for 'direction': must be \"asc\" or \"desc\"", dir)
	}

	if query.Has("page") {
		page, err = strconv.Atoi(query.Get("page"))
		if err != nil {
			return "", 0, 0, fmt.Errorf("invalid parameter value %q for 'page': could not be parsed as an int (%q)", query.Get("page"), err)
		}
	}

	if query.Has("per_page") {
		perPage, err = strconv.Atoi(query.Get("per_page"))
		if err != nil {
			return "", 0, 0, fmt.Errorf("invalid parameter value for '%s': could not be parsed as an int (%q)", query.Get("per_page"), err)
		}
	}

	numPages := ceil(len(f.prs), perPage)
	if page > numPages {
		return "", 0, 0, fmt.Errorf("invalid parameter value %d for 'page': greater than the maximum page (%d)", page, numPages)
	}

	return
}

// linkHeader generates the value of the 'link' header included in GitHub
// responses. I got this strange header value by querying GitHub with curl and
// inspecting the response (see README.md). This is how go-github seems to
// determine pagination.
//
// I moved this logic into a helper function because it's quite verbose, and
// listHandler is already long. This logic may seem more complicated than
// necessary (there are multiple checks for page > 1, for example), but the
// links are included in the same order as GitHub--prev, next, last, first--for
// fidelity.
func linkHeader(page, numPages, perPage int, includePerPage bool) string {
	links := make([]string, 0, 4)
	perPageParam := ""
	if includePerPage {
		perPageParam = fmt.Sprintf("&per_page=%d", perPage)
	}
	if page > 1 {
		links = append(links, fmt.Sprintf("<https://api.github.com/repositories/23653453/pulls?page=%d%s>; rel=\"prev\"", page-1, perPageParam))
	}
	if page < numPages {
		links = append(links, fmt.Sprintf("<https://api.github.com/repositories/23653453/pulls?page=%d%s>; rel=\"next\"", page+1, perPageParam))
		links = append(links, fmt.Sprintf("<https://api.github.com/repositories/23653453/pulls?page=%d%s>; rel=\"last\"", numPages, perPageParam))
	}
	if page > 1 {
		links = append(links, fmt.Sprintf("<https://api.github.com/repositories/23653453/pulls?page=1%s>; rel=\"first\"", perPageParam))
	}
	return strings.Join(links, ", ")
}

func (f *fakeGitHub) listHandler(w http.ResponseWriter, r *http.Request) {
	// Get params from request
	dir, page, perPage, err := f.getListParams(r.URL.Query())
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	// Prepare results -- mostly iterating through the PRs in the right order
	numPages := ceil(len(f.prs), perPage)
	startIdx := (page - 1) * perPage
	endIdx := min(startIdx+perPage, len(f.prs))
	resp := f.prs[startIdx:endIdx]
	if dir == "desc" {
		resp = make([]*prSpec, endIdx-startIdx)
		// like copy(), but iterate through f.prs backwards (startIdx is "distance
		// from the last item" in this case)
		for i := 0; i < len(resp); i++ {
			resp[i] = f.prs[len(f.prs)-1-startIdx-i]
		}
	}

	// Send response
	respBytes, err := json.Marshal(resp)
	if err != nil {
		panic("could not serialize response JSON: " + err.Error())
	}
	w.Header().Add("content-type", "application/json; charset=utf-8") // const
	w.Header().Add("x-github-media-type", "github.v3; format=json")   // const
	w.Header().Add("date", time.Now().Format(time.RFC1123))           // time.Now()
	w.Header().Add("content-length", strconv.Itoa(len(respBytes)))    // len(resp)
	w.Header().Add("link", linkHeader(page, numPages, perPage, r.URL.Query().Has("per_page")))

	fmt.Printf("Request %s %q (sending mock resp)\n", r.Method, r.URL.Path)
	n, err := w.Write(respBytes)
	if err != nil {
		panic("error sending /pulls response: " + err.Error())
	}
	if n < len(respBytes) {
		panic(fmt.Sprintf("intended to send %d bytes back, but only sent %d", len(respBytes), n))
	}
	f.pagesFetched++
}

func (f *fakeGitHub) defaultHandler(w http.ResponseWriter, r *http.Request) {
	if !strings.Contains(r.URL.Path, "api") {
		// For some reason, go-github seems to spam /api and /apis, which AFAICT are
		// not valid paths of api.github.com. Everything works if I return 404 for
		// those calls (which is what GitHub returned when I curl'ed /api). Now I
		// don't even bother logging those requests.
		fmt.Printf("Request %s %q (sending 404)\n", r.Method, r.URL.Path)
	}
	http.Error(w, "404 Not Found", http.StatusNotFound)
}

func (f *fakeGitHub) start() {
	go http.ListenAndServe(":8080", f)
}
