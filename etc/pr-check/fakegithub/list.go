package fakegithub

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func (f *Server) listHandler(w http.ResponseWriter, r *http.Request) {
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
		resp = make([]*PR, endIdx-startIdx)
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

// CheckAndResetPagesFetched is a testing function that shows how many pages of
// PRs have been served since the last time it was called. This is entirely used
// by pr_check tests to ensure that pr_check's scan() function doesn't fetch
// more data than it needs.
func (f *Server) CheckAndResetPagesFetched() (ret int) {
	ret = f.pagesFetched
	f.pagesFetched = 0
	return
}

// getListParams is a helper for listHandler that pulls all relevant params out
// of the request's query string & validates them.
func (f *Server) getListParams(query url.Values) (dir string, page, perPage int, err error) {
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

// linkHeader is a helper for listHandler that generates the value of the 'link'
// header included in GitHub responses. I got this strange header value by
// querying GitHub with curl and inspecting the response (see README.md). This
// is how go-github seems to determine pagination.
//
// This logic may seem more complicated than necessary (there are multiple
// checks for page > 1, for example), but the links are included in the same
// order as GitHub--prev, next, last, first--for fidelity.
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

// ceil calculates ⌈n/d⌉.
// See http://blog.pkh.me/p/36-figuring-out-round%2C-floor-and-ceil-with-integer-division.html
func ceil(n, d int) int {
	return (n-1)/d + 1
}
