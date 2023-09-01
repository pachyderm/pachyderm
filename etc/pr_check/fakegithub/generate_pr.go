package fakegithub

import (
	"bytes"
	"crypto/sha256"
	"embed"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"text/template"
	"time"
)

type User struct {
	Login string
	ID    int
	Type  string // "Organization" or "User"
}

type PR struct {
	Number  int       `json:"number"`
	Title   string    `json:"title"`
	Body    string    `json:"body"`
	Author  *User     `json:"author"`
	Created time.Time `json:"created"`
}

var (
	pachydermOrgUser = &User{
		Login: "pachyderm",
		ID:    10432478,
		Type:  "Organization",
	}

	// templateFiles contains the text of templates, loaded at compile time using
	// 'embed'. See https://pkg.go.dev/embed
	//go:embed templates/*.json.tmpl
	templateFiles embed.FS

	// t contains the parsed templates for the files in (templateFiles). It's
	// loaded in init(), rather than here, as explained in comments in init().
	tmpl *template.Template
)

func init() {
	var pachUser, pachRepo string
	// the functions "pachuser" and "pachrepo" refer to 't', so if you try to move
	// this block into a 'var' declaration, go yields the error "initialization
	// cycle for t". So define this in init() instead.
	tmpl = template.Must(template.New("").Funcs(template.FuncMap{
		"b64sha": func(l int, is ...interface{}) string {
			h := sha256.New()
			fmt.Fprint(h, is...)
			return base64.RawURLEncoding.EncodeToString(h.Sum(nil))[:l]
		},
		"hexsha": func(l int, is ...interface{}) string {
			h := sha256.New()
			fmt.Fprint(h, is...)
			return hex.EncodeToString(h.Sum(nil))[:l]
		},
		"intsha": func(l int, is ...interface{}) string {
			if l > 18 {
				panic("can't generate int sha >18 digits; better impl required")
			}
			h := sha256.New()
			fmt.Fprint(h, is...)
			var d uint64 = binary.LittleEndian.Uint64(h.Sum(nil)[:8])
			buf := make([]byte, l)
			for i := 0; i+1 < l; i++ {
				buf[l-i-1] = '0' + byte(d%10)
				d /= 10
			}
			buf[0] = '1' + byte(d%9) // mixed radix ensures no leading 0
			return string(buf)
		},
		"pachuser": func() string {
			return pachUser
		},
		"pachrepo": func() string {
			return pachRepo
		},
	}).ParseFS(templateFiles, "templates/*.json.tmpl"))

	pachUserBuilder := &strings.Builder{}
	if err := tmpl.ExecuteTemplate(pachUserBuilder, "user.json.tmpl", pachydermOrgUser); err != nil {
		panic("could not execute 'user' template: " + err.Error())
	}
	pachUser = pachUserBuilder.String()

	pachRepoBuilder := &strings.Builder{}
	if err := tmpl.ExecuteTemplate(pachRepoBuilder, "repo.json.tmpl", nil); err != nil {
		panic("could not execute 'repo' template: " + err.Error())
	}
	pachRepo = pachRepoBuilder.String()
}

// NewPRSpec creates a new PR spec. It sets the PR number, so users should call
// this instead of creating a PR directly, to ensure all generated PR numbers
// are unique.
//
// this is a bit redundant with fakeGitHub.AddPR, but there are tests that run
// checks on a single PR that use this broken-out version
func NewPRSpec(number int, title, body string, author *User, created time.Time) *PR {
	s := &PR{
		Number:  number,
		Title:   title,
		Body:    body,
		Author:  author,
		Created: created,
	}

	if s.Author == nil {
		s.Author = pachydermOrgUser
	}
	return s
}

// reformatJSON is a helper function used by the various json() methods
// to unmarshal and marshal template JSON. This helps by:
// 1. checking validity
// 2. sorting the keys
// 3. fixing formatting
func reformatJSON(buf *bytes.Buffer) ([]byte, error) {
	tmp := make(map[string]interface{})
	err := json.NewDecoder(buf).Decode(&tmp)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal internal template json: %w", err)
	}
	result, err := json.Marshal(tmp)
	if err != nil {
		return nil, fmt.Errorf("could not marshal json internally: %w", err)
	}
	return result, nil
}

func (u *User) MarshalJSON() ([]byte, error) {
	buf := &bytes.Buffer{}
	if err := tmpl.ExecuteTemplate(buf, "user.json.tmpl", u); err != nil {
		return nil, fmt.Errorf("could not execute 'user' template: %w", err)
	}
	return reformatJSON(buf)
}

func (p *PR) MarshalJSON() ([]byte, error) {
	buf := &bytes.Buffer{}
	if err := tmpl.ExecuteTemplate(buf, "pr.json.tmpl", p); err != nil {
		return nil, fmt.Errorf("could not execute 'pr' template: %w", err)
	}
	return reformatJSON(buf)
}
