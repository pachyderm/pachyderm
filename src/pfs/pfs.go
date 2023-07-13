package pfs

import (
	"encoding/hex"
	"fmt"
	"hash"
	"unicode"

	"google.golang.org/protobuf/proto"

	"github.com/pachyderm/pachyderm/v2/src/auth"

	"github.com/pachyderm/pachyderm/v2/src/internal/ancestry"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachhash"
)

const (
	// ChunkSize is the size of file chunks when resumable upload is used
	ChunkSize = int64(512 * 1024 * 1024) // 512 MB

	// default system repo types
	UserRepoType = "user"
	MetaRepoType = "meta"
	SpecRepoType = "spec"

	DefaultProjectName = "default"

	// projectNameLimit is the maximum length of a project name, determined
	// by the 63-character Kubernetes resource name limit and leaving eight
	// characters for pipeline names.
	//
	// TODO(CORE-1489): raise this to something more sensible, e.g. 63 or 64
	// or 256 characters.
	projectNameLimit = 63 - 8 - len("-v99")
)

// NewHash returns a hash that PFS uses internally to compute checksums.
func NewHash() hash.Hash {
	return pachhash.New()
}

// EncodeHash encodes a hash into a readable format.
func EncodeHash(bytes []byte) string {
	return hex.EncodeToString(bytes)
}

// DecodeHash decodes a hash into bytes.
func DecodeHash(hash string) ([]byte, error) {
	res, err := hex.DecodeString(hash)
	return res, errors.EnsureStack(err)
}

func (p *Project) String() string {
	return p.GetName()
}

func (r *Repo) String() string {
	if r.Type == UserRepoType {
		if projectName := r.Project.String(); projectName != "" {
			return projectName + "/" + r.Name
		}
		return r.Name
	}
	if projectName := r.Project.String(); projectName != "" {
		return projectName + "/" + r.Name + "." + r.Type
	}
	return r.Name + "." + r.Type
}

func (r *Repo) NewBranch(name string) *Branch {
	return &Branch{
		Repo: proto.Clone(r).(*Repo),
		Name: name,
	}
}

func (r *Repo) NewCommit(branch, id string) *Commit {
	return &Commit{
		Repo:   r,
		Id:     id,
		Branch: r.NewBranch(branch),
	}
}

func (c *Commit) NewFile(path string) *File {
	return &File{
		Commit: proto.Clone(c).(*Commit),
		Path:   path,
	}
}

func (c *Commit) String() string {
	return c.Repo.String() + "@" + c.Id
}

// TODO(provenance): there's a concern client code will unknowningly call GetRepo() when it shouldn't
func (c *Commit) AccessRepo() *Repo {
	if c.GetRepo() != nil {
		return c.GetRepo()
	}
	return c.GetBranch().GetRepo()
}

func (b *Branch) NewCommit(id string) *Commit {
	return &Commit{
		Branch: proto.Clone(b).(*Branch),
		Id:     id,
		Repo:   b.Repo,
	}
}

func (b *Branch) String() string {
	return b.Repo.String() + "@" + b.Name
}

// ValidateName returns an error if the project is nil or its name is an invalid
// project name.  DefaultProjectName is always valid; otherwise the ancestry
// package is used to validate the name.
func (p *Project) ValidateName() error {
	if p == nil {
		return errors.New("nil project")
	}
	if p.Name == DefaultProjectName {
		return nil
	}
	if len(p.Name) > projectNameLimit {
		return errors.Errorf("project name %q is %d characters longer than the %d max", p.Name, len(p.Name)-projectNameLimit, projectNameLimit)
	}
	if err := ancestry.ValidateName(p.Name); err != nil {
		return err
	}
	first := rune(p.Name[0])
	if !unicode.IsLetter(first) && !unicode.IsDigit(first) {
		return errors.Errorf("project names must start with an alphanumeric character")
	}
	return nil
}

// EnsureProject ensures that repo.Project is set.  It does nothing if repo is
// nil.
func (r *Repo) EnsureProject() {
	if r == nil {
		return
	}
	if r.Project.GetName() == "" {
		r.Project = &Project{Name: DefaultProjectName}
	}
}

// AuthResource returns the auth resource for a repo.  The resource name is the
// project name and repo name separated by a slash.  Notably, it does _not_
// include the repo type string.
func (r *Repo) AuthResource() *auth.Resource {
	var t auth.ResourceType
	if r.GetType() == SpecRepoType {
		t = auth.ResourceType_SPEC_REPO
	} else {
		t = auth.ResourceType_REPO
	}
	return &auth.Resource{
		Type: t,
		Name: fmt.Sprintf("%s/%s", r.GetProject().GetName(), r.GetName()),
	}
}

// AuthResource returns the auth resource for a project.
func (p *Project) AuthResource() *auth.Resource {
	return &auth.Resource{Type: auth.ResourceType_PROJECT, Name: p.GetName()}
}
