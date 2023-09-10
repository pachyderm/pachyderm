package remapfs

import (
	"fmt"
	"io"
	"io/fs"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/fsutil"
	"golang.org/x/exp/maps"
)

// Recipe controls how file attributes are remapped.
type Recipe struct {
	// Match matches a path in the underlying filesystem and indicates that this recipe should
	// be applied.  Only the first match of the regexp results in selection and renaming.  For
	// example, /abc(.)/ on abcdabce with template $1 results in the file d ending up in the
	// output, but not e.
	Match *regexp.Regexp
	// If non-empty, a matched file will have this new name.  Captures from the regexp are
	// available as $1, $2, etc.  No other template syntax is available (no \1, no ${1}, etc.).
	Rename string
	// If non-nil, a matched file will have these attributes; owner (UID), group (GID).
	Chown, Chgrp *uint32
	// Change the mode of a file.  Only bits set in ChmodMask will be modified.
	// <(original &^ mask) | (chmod & mask)>.  (This mask is not a umask, which is <original & umask>.)
	Chmod, ChmodMask fs.FileMode
}

type cacheRecord struct {
	recipes      []Recipe
	origName     string
	syntheticDir bool
	info         *remapFileInfo
}

type RemapFS struct {
	Recipes []Recipe
	fs      fs.FS
	cache   map[string]cacheRecord
}

var _ fs.FS = (*RemapFS)(nil)
var _ fs.ReadDirFS = (*RemapFS)(nil)
var _ fs.StatFS = (*RemapFS)(nil)

type remapFile struct {
	fs.File
	fs.FileInfo
	name   string
	fs     *RemapFS
	dirptr int
}

var _ fs.File = (*remapFile)(nil)
var _ fs.ReadDirFile = (*remapFile)(nil)

func (f *remapFile) Stat() (fs.FileInfo, error) {
	if f.FileInfo == nil {
		return nil, fs.ErrInvalid
	}
	return f.FileInfo, nil
}

func (f *remapFile) Read(p []byte) (int, error) {
	if f.File != nil {
		return f.File.Read(p)
	}
	return 0, fs.ErrInvalid
}

func (f *remapFile) Close() error {
	if f.File != nil {
		return f.File.Close()
	}
	return nil
}

func (f *remapFile) ReadDir(n int) (entries []fs.DirEntry, err error) {
	defer func() {
		f.dirptr += len(entries)
	}()
	entries, err = f.fs.ReadDir(f.name)
	switch {
	case err != nil:
		return nil, err
	case n < 0:
		return entries[f.dirptr:], nil
	case f.dirptr+n >= len(entries):
		return entries[f.dirptr:], io.EOF
	default:
		return entries[f.dirptr : f.dirptr+n], nil
	}
}

type remapFileInfo struct {
	name     string
	own, grp *uint32
	mode     fs.FileMode
	size     int64
	fs.FileInfo
}

var _ fs.FileInfo = (*remapFileInfo)(nil)
var _ fs.DirEntry = (*remapFileInfo)(nil)
var _ fsutil.HasUID = (*remapFileInfo)(nil)
var _ fsutil.HasGID = (*remapFileInfo)(nil)

func (r *remapFileInfo) Name() string               { return r.name }
func (r *remapFileInfo) IsDir() bool                { return r.Mode()&fs.ModeDir != 0 }
func (r *remapFileInfo) Sys() any                   { return r }
func (r *remapFileInfo) Info() (fs.FileInfo, error) { return r, nil }
func (r *remapFileInfo) String() string {
	uid, _ := fsutil.FileUID(r)
	gid, _ := fsutil.FileGID(r)
	return fmt.Sprintf("uid:%d gid:%d %v", uid, gid, fs.FormatFileInfo(r))

}
func (r *remapFileInfo) ModTime() time.Time {
	if r.FileInfo != nil {
		return r.FileInfo.ModTime()
	}
	return time.Time{}
}
func (r *remapFileInfo) Size() int64 {
	if r.FileInfo != nil {
		return r.FileInfo.Size()
	}
	return r.size
}
func (r *remapFileInfo) Type() fs.FileMode {
	return r.Mode()
}
func (r *remapFileInfo) Mode() fs.FileMode {
	if r.mode == 0 && r.FileInfo != nil {
		return r.FileInfo.Mode()
	}
	return r.mode
}
func (r *remapFileInfo) GetUID() (uint32, bool) {
	if r.own != nil {
		return *r.own, true
	}
	if r.FileInfo != nil {
		return fsutil.FileUID(r.FileInfo)
	}
	return 0, false
}
func (r *remapFileInfo) GetGID() (uint32, bool) {
	if r.own != nil {
		return *r.grp, true
	}
	if r.FileInfo != nil {
		return fsutil.FileGID(r.FileInfo)
	}
	return 0, false
}

func isNum(r rune) bool {
	return r >= '0' && r <= '9'
}

func evalSubst(rx *regexp.Regexp, template, input string) (string, bool) {
	if rx == nil {
		return input, true
	}
	matches := rx.FindStringSubmatch(input)
	if len(matches) == 0 {
		return "", false
	}

	var result []rune
	var num []rune
	state := 0 // 0 -> reading; 1 -> reading numbers;
	doSubstitution := func() bool {
		cap, err := strconv.Atoi(string(num))
		if err != nil {
			return false
		}
		if cap >= len(matches) {
			return false
		}
		result = append(result, []rune(matches[cap])...)
		num = nil
		return true
	}
	for _, r := range template {
		switch state {
		case 0:
			if r == '$' {
				state = 1
				continue
			}
			result = append(result, r)
		case 1:
			if isNum(r) {
				num = append(num, r)
				continue
			} else {
				if ok := doSubstitution(); !ok {
					return "", false
				}
				state = 0
				result = append(result, r)
			}
		default:
			panic("invalid state " + strconv.Itoa(state))
		}
	}
	if state == 1 {
		if ok := doSubstitution(); !ok {
			return "", false
		}
	}
	return string(result), true
}

func (f *RemapFS) ensureCache() error {
	if f.cache != nil {
		return nil
	}
	var cacheOK bool
	defer func() {
		if !cacheOK {
			f.cache = nil
		}
	}()
	f.cache = make(map[string]cacheRecord)

	// Walk the underlying FS to figure out what the remappings are.  Unfortunately match ->
	// replace doesn't work in reverse when we want to ReadDir later on.
	err := fs.WalkDir(f.fs, ".", func(name string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if name == "." {
			return nil
		}
		name = path.Clean(name)
		for _, r := range f.Recipes {
			c := f.cache[name]
			if r.Match != nil && !r.Match.MatchString(name) {
				// Recipe does not apply.
				continue
			}
			if r.Rename != "" {
				// Do the rename if there is one.
				if newpath, ok := evalSubst(r.Match, r.Rename, name); ok {
					if c.origName != "" {
						return errors.Errorf("%v: more than one rule is renaming this file (was %v, now %v)", name, c.origName, newpath)
					}
					c.origName = name
					name = path.Clean(newpath) // Some sneak will rename a file to ../../../foo.
				}
			}
			if path.IsAbs(name) {
				return errors.Errorf("%v: path must not be absolute", name)
			}
			c.recipes = append(c.recipes, r)
			f.cache[name] = c
		}
		return nil
	})
	if err != nil {
		return err
	}

	// Synthesize missing directory entires.
	for name := range f.cache {
		parents := strings.Split(name, "/")
		for i := len(parents) - 1; i > 0; i-- {
			// Absolute paths have already been filtered, so parents[0] != "".
			parent := strings.Join(parents[:i], "/")
			if c, ok := f.cache[parent]; !ok {
				c.syntheticDir = true
				for _, r := range f.Recipes {
					if r.Match == nil {
						c.recipes = append(c.recipes, r)
					}
				}
				f.cache[parent] = c
			}
		}
	}

	// Apply recipe rules.
	for name, c := range f.cache {
		info := &remapFileInfo{
			name: path.Base(name),
			mode: 0o777 | fs.ModeDir,
			size: 4096, // The size of directories on my system.
		}
		if !c.syntheticDir {
			origInfo, err := fs.Stat(f.fs, c.origName)
			if err != nil {
				return errors.Wrapf(err, "stat %v (-> %v)", name, c.origName)
			}
			info = &remapFileInfo{
				name:     path.Base(name),
				FileInfo: origInfo,
			}
		}
		for _, r := range c.recipes {
			if r.Chgrp != nil {
				info.grp = r.Chgrp
			}
			if r.Chown != nil {
				info.own = r.Chown
			}
			info.mode = chmod(info.mode, r.Chmod, r.ChmodMask)
		}
		c.recipes = nil
		c.info = &remapFileInfo{FileInfo: info, name: path.Base(name)}
		f.cache[name] = c
	}

	// Done.
	cacheOK = true
	return nil
}

func chmod(in, chmod, mask fs.FileMode) fs.FileMode {
	in = (in &^ mask) | (chmod & mask)
	return in
}

func (f *RemapFS) Open(name string) (fs.File, error) {
	if name == "." {
		orig, err := f.fs.Open(".")
		if err != nil {
			return nil, errors.Wrap(err, "open underlying root dir")
		}
		info, err := fs.Stat(f.fs, ".")
		if err != nil {
			return nil, errors.Wrap(err, "stat underlying root dir")
		}
		result := &remapFile{
			name:     ".",
			File:     orig,
			FileInfo: info,
			fs:       f,
		}
		return result, nil
	}
	if err := f.ensureCache(); err != nil {
		return nil, errors.Wrap(err, "ensure cache")
	}
	c, ok := f.cache[name]
	if !ok {
		return nil, fs.ErrNotExist
	}
	if c.syntheticDir {
		return &remapFile{
			name:     name,
			FileInfo: c.info,
			fs:       f,
		}, nil
	}
	orig, err := f.fs.Open(c.origName)
	if err != nil {
		return nil, errors.Wrapf(err, "open underlying file %v", c.origName)
	}
	return &remapFile{
		name:     name,
		File:     orig,
		FileInfo: c.info,
		fs:       f,
	}, nil
}

func (f *RemapFS) Stat(name string) (fs.FileInfo, error) {
	if name == "." {
		orig, err := fs.Stat(f.fs, ".")
		if err != nil {
			return nil, err
		}
		return &remapFileInfo{name: ".", FileInfo: orig}, nil
	}
	if err := f.ensureCache(); err != nil {
		return nil, errors.Wrap(err, "ensure cache")
	}
	c, ok := f.cache[name]
	if !ok {
		return nil, fs.ErrNotExist
	}
	return c.info, nil
}

func (f *RemapFS) ReadDir(name string) ([]fs.DirEntry, error) {
	if err := f.ensureCache(); err != nil {
		return nil, errors.Wrap(err, "ensure cache")
	}
	if name == "." {
		name = ""
	}
	info := map[string]*remapFileInfo{}
	for k := range f.cache {
		switch {
		case name == "" || strings.HasPrefix(k, name+"/"):
			rel := k[len(name)+1:]
			if rel == "" {
				continue
			}
			if i := strings.IndexByte(rel, '/'); i != -1 {
				// Directory entry needed.
				continue
			}
			info[rel] = f.cache[k].info
		}
	}
	var result []fs.DirEntry
	names := maps.Keys(info)
	sort.Strings(names)
	for _, n := range names {
		result = append(result, fs.FileInfoToDirEntry(info[n]))
	}
	return result, nil
}
