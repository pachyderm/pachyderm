package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"
)

type fakeGitHub struct {
	*http.ServeMux
	prs          []*prSpec
	pagesFetched int
}

func newFakeGitHub() *fakeGitHub {
	r := &fakeGitHub{
		ServeMux: http.NewServeMux(),
		prs:      make([]*prSpec, 0, 32),
	}
	fakePRs, err := os.Open("fake_github_prs.json")
	if err != nil {
		panic("could not open fake_github_prs.json:\n" + err.Error())
	}
	json.NewDecoder(fakePRs).Decode(&r.prs)

	r.HandleFunc("/repos/pachyderm/pachyderm/pulls", r.listHandler)
	r.HandleFunc("/", r.defaultHandler)
	return r
}

func (f *fakeGitHub) listHandler(w http.ResponseWriter, r *http.Request) {
	// Get params from request (mostly 'page', but check 'per_page')
	var (
		direction      = "desc"
		page           = 1
		resultsPerPage = 7 // real GitHub is 30 by default
		err            error
	)
	if r.URL.Query().Has("direction") {
		direction = r.URL.Query().Get("direction")
	}
	getIntParam := func(key string, defaultVal int) (int, error) {
		if r.URL.Query().Has(key) {
			result, err := strconv.Atoi(r.URL.Query().Get(key))
			if err != nil {
				return 0, fmt.Errorf("invalid parameter value for '%s': could not be parsed as an int (%q)", key, err)
			}
			return result, nil
		}
		return defaultVal, nil
	}
	if page, err = getIntParam("page", page); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	if resultsPerPage, err = getIntParam("per_page", resultsPerPage); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	// Prepare results -- mostly iterating through the PRs in the right order
	// numPages = ceil(results/resultsPerPage) = expression below.
	// See http://blog.pkh.me/p/36-figuring-out-round%2C-floor-and-ceil-with-integer-division.html
	numPages := (len(f.prs)-1)/resultsPerPage + 1
	if page > numPages {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("invalid parameter value %d for 'page': greater than the maximum page (%d)", page, numPages)))
		return
	}
	startIdx := (page - 1) * resultsPerPage
	endIdx := min(startIdx+resultsPerPage, len(f.prs))
	resp := f.prs[startIdx:endIdx]
	if direction == "desc" {
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
	// I got this strange header value by querying GitHub with curl and inspecting
	// the response (see README.md). This is how go-github seems to determine
	// pagination
	w.Header().Add("link", fmt.Sprintf("<https://api.github.com/repositories/23653453/pulls?page=%d&per_page=%d>; rel=\"next\", <https://api.github.com/repositories/23653453/pulls?page=%d&per_page=%d>; rel=\"last\"", page+1, resultsPerPage, numPages, resultsPerPage))

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
