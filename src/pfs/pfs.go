package pfs

import (
	"bytes"
	"encoding"
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

func (p *Project) Key() string {
	return p.GetName()
}

func (r *Repo) String() string {
	if r.GetType() == UserRepoType {
		if projectName := r.GetProject().String(); projectName != "" {
			return projectName + "/" + r.GetName()
		}
		return r.GetName()
	}
	if projectName := r.GetProject().String(); projectName != "" {
		return projectName + "/" + r.GetName() + "." + r.GetType()
	}
	return r.GetName() + "." + r.GetType()
}

func (r *Repo) Key() string {
	return r.GetProject().Key() + "/" + r.GetName() + "." + r.GetType()
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

func (c *Commit) Key() string {
	if c == nil {
		return ""
	}
	if c.Repo == nil {
		return "<nil repo>@" + c.Id
	}
	return c.Repo.Key() + "@" + c.Id
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

func (b *Branch) Key() string {
	return b.GetRepo().Key() + "@" + b.Name
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

// UnmarshalText implements encoding.TextUnmarshaler.
func (p *ProjectPicker) UnmarshalText(b []byte) error {
	p.Picker = &ProjectPicker_Name{
		Name: string(b),
	}
	if err := (&Project{Name: string(b)}).ValidateName(); err != nil {
		return err
	}
	if err := p.ValidateAll(); err != nil {
		return err
	}
	return nil
}

var _ encoding.TextUnmarshaler = (*ProjectPicker)(nil)

type repoName Repo // repoName is a repo object used only to store the name and type of a repo.

// UnmarshalText implements encoding.TextUnmarshaler.
func (n *repoName) UnmarshalText(b []byte) error {
	parts := bytes.SplitN(b, []byte{'.'}, 2)
	switch len(parts) {
	case 0:
		return errors.New("invalid repo name: empty")
	case 1:
		n.Name = string(b)
		n.Type = UserRepoType
	case 2:
		n.Name = string(parts[0])
		n.Type = string(parts[1])
	}
	return nil
}

var _ encoding.TextUnmarshaler = (*repoName)(nil)

// UnmarshalText implements encoding.TextUnmarshaler.
func (p *RepoPicker) UnmarshalText(b []byte) error {
	parts := bytes.SplitN(b, []byte{'/'}, 2)
	var project, repo []byte
	switch len(parts) {
	case 0:
		return errors.New("invalid repo picker: empty")
	case 1:
		project, repo = nil, parts[0]
	case 2:
		project, repo = parts[0], parts[1]
	default:
		return errors.New("invalid repo picker: too many slashes")
	}
	rnp := &RepoPicker_RepoName{
		Project: &ProjectPicker{},
	}
	p.Picker = &RepoPicker_Name{
		Name: rnp,
	}
	if project != nil {
		if err := rnp.Project.UnmarshalText(project); err != nil {
			return errors.Wrapf(err, "unmarshal project %s", project)
		}
	}
	var repoName repoName
	if err := repoName.UnmarshalText(repo); err != nil {
		return errors.Wrapf(err, "unmarshal repo name %s", repo)
	}
	rnp.Name = repoName.Name
	rnp.Type = repoName.Type
	if err := p.ValidateAll(); err != nil {
		return err
	}
	return nil
}

var _ encoding.TextUnmarshaler = (*RepoPicker)(nil)

func onlyHex(p []byte) bool {
	for _, b := range p {
		if (b >= 'a' && b <= 'f') || (b >= 'A' && b <= 'F') || (b >= '0' && b <= '9') {
			continue
		}
		return false
	}
	return true
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (p *CommitPicker) UnmarshalText(b []byte) error {
	parts := bytes.SplitN(b, []byte{'@'}, 2)
	var repo, id []byte
	switch len(parts) {
	case 0:
		return errors.New("invalid commit picker: empty")
	case 1:
		return errors.New("invalid commit picker: no @id")
	case 2:
		repo, id = parts[0], parts[1]
	default:
		return errors.New("invalid commit picker: too many @s")
	}
	var rp RepoPicker
	if err := rp.UnmarshalText(repo); err != nil {
		return errors.Wrapf(err, "unmarshal repo picker %s", repo)
	}
	// TODO(PFS-229): Implement the other parsers.
	if bytes.HasSuffix(id, []byte{'^'}) {
		// CommitPicker_AncestorOf
		return errors.New("ancestor of commit syntax is currently unimplemented (id^); specify the exact id instead")
	}
	if bytes.Contains(id, []byte{'.'}) {
		// CommitPicker_BranchRoot
		return errors.New("branch root commit syntax is currently unimplemented (id.N); specify the exact id instead")
	}

	// To distinguish global IDs and branch names, we look for a valid global ID first.  An ID
	// is a hex-encoded UUIDv4 without dashes.  UUIDv4s always have the 13th (1 indexed) byte
	// set to '4'.
	if len(id) == 32 && id[12] == '4' && onlyHex(id) {
		p.Picker = &CommitPicker_Id{
			Id: &CommitPicker_CommitByGlobalId{
				Id:   string(bytes.ToLower(id)),
				Repo: &rp,
			},
		}
	} else {
		// Anything that isn't a valid UUIDv4 is a branch name.
		p.Picker = &CommitPicker_BranchHead{
			BranchHead: &BranchPicker{
				Picker: &BranchPicker_Name{
					Name: &BranchPicker_BranchName{
						Repo: &rp,
						Name: string(id),
					},
				},
			},
		}
	}
	if err := p.ValidateAll(); err != nil {
		return err
	}
	return nil
}

var _ encoding.TextUnmarshaler = (*CommitPicker)(nil)
