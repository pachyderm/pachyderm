package pfs

import (
	"bytes"
	"encoding"
	"encoding/hex"
	"fmt"
	"hash"
	"regexp"
	"strconv"
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

func (c *Commit) NilBranch() {
	if c != nil {
		c.Branch = nil
	}
}

func (ci *CommitInfo) NilBranch() {
	if ci != nil {
		ci.Commit.NilBranch()
		ci.ParentCommit.NilBranch()
	}
}

func (f *File) NilBranch() {
	if f != nil {
		f.Commit.NilBranch()
	}
}

func (fi *FileInfo) NilBranch() {
	if fi != nil {
		fi.File.NilBranch()
	}
}

func (dfr *DiffFileResponse) NilBranch() {
	if dfr != nil {
		dfr.OldFile.NilBranch()
		dfr.NewFile.NilBranch()
	}
}

func (cf *CopyFile) NilBranch() {
	if cf != nil {
		cf.Src.NilBranch()
	}
}

func (bi *BranchInfo) NilBranch() {
	if bi != nil {
		bi.Head.NilBranch()
	}
}

func (csi *CommitSetInfo) NilBranch() {
	if csi != nil {
		for _, ci := range csi.Commits {
			ci.NilBranch()
		}
	}
}

func (fcr *FindCommitsResponse) NilBranch() {
	if fcr != nil && fcr.Result != nil {
		switch c := fcr.Result.(type) {
		case *FindCommitsResponse_LastSearchedCommit:
			c.LastSearchedCommit.NilBranch()
		case *FindCommitsResponse_FoundCommit:
			c.FoundCommit.NilBranch()
		}
	}
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

// counts the overall offset of branch root and ancestor of
// also removes everything of the ancestry reference
// this function also check following circumstances for ancestry reference:
// 1. ancestry reference should only contain "^" "." and digits
// 2. If there is branch root ".", it should be the first operator; It also should be the only "."
// 3. There should be a number right after "."
// 4. If branch root and ancestor of are used together, the offset of branch root should >= offset of ancestor of
// 5. There can be multiple "^"; but it should end when there is a number following a "^"; so "^^1^" is not valid
func countOffsetAndClean(b *[]byte, offset *uint32) error {
	atIndex := bytes.IndexAny(*b, "@")
	// first . ^ after @
	firstIndex := bytes.IndexAny((*b)[atIndex+1:], "^.") + atIndex + 1
	ancestryPart := (*b)[firstIndex:]
	re := regexp.MustCompile(`^(\.\d+)?((\^+)|(\^\d+))?$`)
	if !re.Match(ancestryPart) {
		return errors.New("invalid Ancestry format")
	}
	parts := re.FindSubmatch(ancestryPart)
	var branchRootOffset, ancestorOfOffset int
	if len(parts[1]) > 0 {
		num, err := strconv.Atoi(string(parts[1][1:]))
		if err != nil {
			return errors.New("invalid descendants format")
		}
		branchRootOffset = num
	}
	// part 4 is ancestorOf with number
	if len(parts[4]) > 0 {
		num, err := strconv.Atoi(string(parts[4][1:]))
		if err != nil {
			return errors.New("invalid ancestor format")
		}
		ancestorOfOffset = num
	} else {
		// part 2 is ancestorOf; it's zero if it's empty
		ancestorOfOffset = len(parts[2])
	}
	if ancestorOfOffset == 0 {
		*offset = uint32(branchRootOffset)
	}
	if branchRootOffset == 0 {
		*offset = uint32(ancestorOfOffset)
	}
	if branchRootOffset != 0 && ancestorOfOffset != 0 {
		if ancestorOfOffset >= branchRootOffset {
			return errors.New("there should be less ancestors than descendants")
		}
		*offset = uint32(branchRootOffset - ancestorOfOffset)
	}
	*b = (*b)[:firstIndex]
	return nil
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
	if bytes.Contains(id, []byte{'.'}) {
		var offset uint32
		if err := countOffsetAndClean(&b, &offset); err != nil {
			return errors.Wrapf(err, "calculate offset %s", b)
		}
		var bp BranchPicker
		if err := bp.UnmarshalText(b); err != nil {
			return errors.Wrapf(err, "unmarshal branch picker %s", b)
		}
		p.Picker = &CommitPicker_BranchRoot_{
			BranchRoot: &CommitPicker_BranchRoot{
				Branch: &bp,
				Offset: offset,
			},
		}
		return nil
	}
	if bytes.Contains(id, []byte{'^'}) {
		var offset uint32
		if err := countOffsetAndClean(&b, &offset); err != nil {
			return errors.Wrapf(err, "calculate offset %s", b)
		}
		var cp CommitPicker
		if err := cp.UnmarshalText(b); err != nil {
			return errors.Wrapf(err, "unmarshal commit picker %s", b)
		}
		p.Picker = &CommitPicker_Ancestor{
			Ancestor: &CommitPicker_AncestorOf{
				Start:  &cp,
				Offset: offset,
			},
		}
		return nil
	}
	// To distinguish global IDs and branch names, we look for a valid global ID first.  An ID
	// is a hex-encoded UUIDv4 without dashes.  UUIDv4s always have the 13th (1 indexed) byte
	// set to '4'.
	if len(id) == 32 && id[12] == '4' && onlyHex(id) {
		var rp RepoPicker
		if err := rp.UnmarshalText(repo); err != nil {
			return errors.Wrapf(err, "unmarshal repo picker %s", repo)
		}
		p.Picker = &CommitPicker_Id{
			Id: &CommitPicker_CommitByGlobalId{
				Id:   string(bytes.ToLower(id)),
				Repo: &rp,
			},
		}
	} else {
		// If the ID isn't a valid UUIDv4, the whole expression is treated as a branch name.
		var bp BranchPicker
		if err := bp.UnmarshalText(b); err != nil {
			return errors.Wrapf(err, "unmarshal branch picker %s", b)
		}
		p.Picker = &CommitPicker_BranchHead{
			BranchHead: &bp,
		}
	}
	if err := p.ValidateAll(); err != nil {
		return err
	}
	return nil
}

var _ encoding.TextUnmarshaler = (*CommitPicker)(nil)

// UnmarshalText implements encoding.TextUnmarshaler.
func (p *BranchPicker) UnmarshalText(b []byte) error {
	parts := bytes.SplitN(b, []byte{'@'}, 2)
	var repo, name []byte
	switch len(parts) {
	case 0:
		return errors.New("invalid branch picker: empty")
	case 1:
		return errors.New("invalid branch picker: no @id")
	case 2:
		repo, name = parts[0], parts[1]
	default:
		return errors.New("invalid branch picker: too many @s")
	}
	var rp RepoPicker
	if err := rp.UnmarshalText(repo); err != nil {
		return errors.Wrapf(err, "unmarshal repo picker %s", repo)
	}
	// If the branch name looks like a commit id, reject it.  Pachyderm does not allow branches
	// with this name.
	if len(name) == 32 && name[12] == '4' && onlyHex(name) {
		return errors.New("invalid branch picker: name refers to a commit, not a branch")
	}
	p.Picker = &BranchPicker_Name{
		Name: &BranchPicker_BranchName{
			Repo: &rp,
			Name: string(name),
		},
	}
	if err := p.ValidateAll(); err != nil {
		return err
	}
	return nil
}

var _ encoding.TextUnmarshaler = (*BranchPicker)(nil)
