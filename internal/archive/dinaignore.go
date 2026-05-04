package archive

import (
	"bufio"
	"errors"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// IgnoreFileName is the name of the per-project ignore file consulted by ZipDir.
const IgnoreFileName = ".dinaignore"

// ignoreMatcher decides whether a path should be excluded from the deploy archive.
// Patterns mirror .dockerignore semantics: shell-style globs, ** for recursive
// matches, # for comments, and a leading ! to re-include a previously excluded
// path. The result of matching is the verdict of the last pattern that matched.
type ignoreMatcher struct {
	patterns []ignorePattern
}

type ignorePattern struct {
	raw    string
	negate bool
}

// loadIgnore reads .dinaignore from root. Returns (nil, nil) if absent.
func loadIgnore(root string) (*ignoreMatcher, error) {
	f, err := os.Open(filepath.Join(root, IgnoreFileName))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	var patterns []ignorePattern
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		p := ignorePattern{}
		if strings.HasPrefix(line, "!") {
			p.negate = true
			line = strings.TrimPrefix(line, "!")
			line = strings.TrimSpace(line)
		}
		line = strings.TrimPrefix(line, "/")
		line = strings.TrimSuffix(line, "/")
		if line == "" {
			continue
		}
		p.raw = filepath.ToSlash(line)
		patterns = append(patterns, p)
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	if len(patterns) == 0 {
		return nil, nil
	}
	return &ignoreMatcher{patterns: patterns}, nil
}

// matches reports whether rel (a forward-slash path relative to the archive root)
// is excluded. A path is excluded when its last matching pattern is non-negated.
func (m *ignoreMatcher) matches(rel string) bool {
	if m == nil || rel == "" || rel == "." {
		return false
	}
	rel = filepath.ToSlash(rel)
	excluded := false
	for _, p := range m.patterns {
		if !matchOrAncestor(p.raw, rel) {
			continue
		}
		excluded = !p.negate
	}
	return excluded
}

// matchOrAncestor returns true if pattern matches rel or any of its ancestor
// directories. This makes "node_modules" exclude every file beneath it without
// requiring the user to write "node_modules/**".
func matchOrAncestor(pattern, rel string) bool {
	if globMatch(pattern, rel) {
		return true
	}
	cur := rel
	for {
		parent := path.Dir(cur)
		if parent == cur || parent == "." || parent == "/" {
			return false
		}
		if globMatch(pattern, parent) {
			return true
		}
		cur = parent
	}
}

// globMatch matches a pattern against name using path-component semantics.
// Supported wildcards: ?, *, ** (zero or more components), and character classes.
// Both pattern and name use forward slashes.
func globMatch(pattern, name string) bool {
	patParts := strings.Split(pattern, "/")
	nameParts := strings.Split(name, "/")
	return matchParts(patParts, nameParts)
}

func matchParts(pat, name []string) bool {
	for len(pat) > 0 {
		if pat[0] == "**" {
			// Trailing ** matches any remaining tail (including empty).
			if len(pat) == 1 {
				return true
			}
			rest := pat[1:]
			for i := 0; i <= len(name); i++ {
				if matchParts(rest, name[i:]) {
					return true
				}
			}
			return false
		}
		if len(name) == 0 {
			return false
		}
		ok, err := path.Match(pat[0], name[0])
		if err != nil || !ok {
			return false
		}
		pat = pat[1:]
		name = name[1:]
	}
	return len(name) == 0
}
