package archive

import (
	"archive/zip"
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ZipDir compresses the directory at root into a zip archive.
// If the directory is a git repository, only tracked and untracked (non-ignored)
// files are included. Otherwise, it walks the tree skipping common non-source dirs.
// In either case, a .dinaignore file at root (if present) is applied as a final
// filter using .dockerignore-style patterns.
func ZipDir(root string) ([]byte, error) {
	matcher, err := loadIgnore(root)
	if err != nil {
		return nil, err
	}

	files, err := gitFiles(root)
	if err != nil {
		// Not a git repo or git not available — fall back to walk.
		files, err = walkFiles(root, matcher)
		if err != nil {
			return nil, err
		}
	}

	var buf bytes.Buffer
	w := zip.NewWriter(&buf)

	for _, rel := range files {
		if matcher.matches(rel) {
			continue
		}
		abs := filepath.Join(root, rel)
		info, err := os.Stat(abs)
		if err != nil {
			continue // file may have been deleted between listing and zipping
		}
		if info.IsDir() {
			continue
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return nil, err
		}
		header.Name = rel
		header.Method = zip.Deflate

		fw, err := w.CreateHeader(header)
		if err != nil {
			return nil, err
		}

		f, err := os.Open(abs)
		if err != nil {
			return nil, err
		}
		_, copyErr := io.Copy(fw, f)
		f.Close()
		if copyErr != nil {
			return nil, copyErr
		}
	}

	if err := w.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// gitFiles returns all tracked + untracked-but-not-ignored files via git.
func gitFiles(root string) ([]string, error) {
	// tracked files
	cmd := exec.Command("git", "ls-files")
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	// untracked, non-ignored files
	cmd2 := exec.Command("git", "ls-files", "--others", "--exclude-standard")
	cmd2.Dir = root
	out2, err := cmd2.Output()
	if err != nil {
		return nil, err
	}

	var files []string
	for _, line := range strings.Split(string(out), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			files = append(files, line)
		}
	}
	for _, line := range strings.Split(string(out2), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			files = append(files, line)
		}
	}
	return files, nil
}

// defaultIgnore lists directory names to skip in non-git fallback mode.
var defaultIgnore = map[string]bool{
	".git": true, "node_modules": true, "__pycache__": true,
	".venv": true, "vendor": true, "dist": true, ".next": true,
}

// walkFiles is the fallback when git is not available. The matcher is consulted
// per-directory so that excluded subtrees aren't traversed at all.
func walkFiles(root string, matcher *ignoreMatcher) ([]string, error) {
	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if rel == "." {
			return nil
		}
		if info.IsDir() {
			if defaultIgnore[info.Name()] {
				return filepath.SkipDir
			}
			if matcher.matches(rel) {
				return filepath.SkipDir
			}
			return nil
		}
		if matcher.matches(rel) {
			return nil
		}
		files = append(files, rel)
		return nil
	})
	return files, err
}
