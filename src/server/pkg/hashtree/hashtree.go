package hashtree

import (
	"path"
	"strings"
)

func (h *HashTree) GlobFile(pattern string) ([]*Node, error) {
	// the pattern should always start with a slash because that's how our
	// paths are.
	if !strings.HasPrefix(pattern, "/") {
		pattern = "/" + pattern
	}

	var res []*Node
	for p, node := range h.Fs {
		matched, err := path.Match(pattern, p)
		if err != nil {
			if err == path.ErrBadPattern {
				return nil, MalformedGlobErr
			}
			return nil, err
		}
		if matched {
			res = append(res, node)
		}
	}
	return res, nil
}
