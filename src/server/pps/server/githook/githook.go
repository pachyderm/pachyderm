package githook

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/pachyderm/pachyderm/src/client"
)

// GitHookPort specifies the port the server will listen on
const GitHookPort = 999
const apiVersion = "v1"

type flushWriter struct {
	f http.Flusher
	w io.Writer
}

func (fw *flushWriter) Write(p []byte) (n int, err error) {
	n, err = fw.w.Write(p)
	if fw.f != nil {
		fw.f.Flush()
	}
	return
}

// GitHookServer serves GetFile requests over HTTP
// e.g. http://localhost:30652/v1/pfs/repos/foo/commits/b7a1923be56744f6a3f1525ec222dc3b/files/ttt.log
type GitHookServer struct {
	*httprouter.Router
	client *client.APIClient
}

func NewGitHookServer(address string) (*GitHookServer, error) {
	router := httprouter.New()
	c, err := client.NewInCluster()
	if err != nil {
		return nil, err
	}
	s := &GitHookServer{
		router,
		c,
	}

	router.POST(fmt.Sprintf("/%v/handle/gitpush", apiVersion), s.gitPushHandler)
	router.NotFound = http.HandlerFunc(notFound)
	return s, nil
}

func (s *GitHookServer) gitPushHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fmt.Printf("GitHook got POST: %v\n", r)
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "couldn't read from body", http.StatusInternalServerError)
		return
	}
	fmt.Printf("Payload:\n%v\n", string(body))
	repos, err := s.client.ListRepo(nil)
	if err != nil {
		http.Error(w, "couldn't read from body", http.StatusInternalServerError)
		return
	}
	fmt.Printf("Repos:%v\n", repos)
	//func (c APIClient) PutFile(repoName string, commitID string, path string, reader io.Reader) (_ int, retErr error) {
	payload := json.Parse(body)
	s.client.PutFile("hook1", commitID, path, bytes.NewReader(body))

	fmt.Fprintf(w, "Received push payload:\n%v\n", string(body))
}

func notFound(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	http.Error(w, "route not found", http.StatusNotFound)
}
