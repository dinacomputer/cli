package archive

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func TestGlobMatch(t *testing.T) {
	tests := []struct {
		pattern string
		name    string
		want    bool
	}{
		{"node_modules", "node_modules", true},
		{"node_modules", "node_modules/foo.js", false}, // direct match only; ancestor walking is matchOrAncestor's job
		{"*.log", "error.log", true},
		{"*.log", "logs/error.log", false},
		{"**/*.log", "error.log", true},
		{"**/*.log", "logs/error.log", true},
		{"**/*.log", "a/b/c/error.log", true},
		{"src/**/*.go", "src/foo.go", true},
		{"src/**/*.go", "src/a/b/foo.go", true},
		{"src/**/*.go", "other/foo.go", false},
		{"build/**", "build", true}, // ** matches zero components, so build/** matches build itself
		{"build/**", "build/output.bin", true},
		{"build/**", "build/a/b/c", true},
		{"foo?", "foo1", true},
		{"foo?", "fooab", false},
		{".env", ".env", true},
		{".env", ".env.local", false},
	}
	for _, tt := range tests {
		t.Run(tt.pattern+"_"+tt.name, func(t *testing.T) {
			if got := globMatch(tt.pattern, tt.name); got != tt.want {
				t.Errorf("globMatch(%q, %q) = %v, want %v", tt.pattern, tt.name, got, tt.want)
			}
		})
	}
}

func TestMatcher(t *testing.T) {
	dir := t.TempDir()
	contents := `# project ignores
node_modules
.nuxt
.output/
/dist
*.log

# include this one back
!important.log

src/**/generated
`
	if err := os.WriteFile(filepath.Join(dir, IgnoreFileName), []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}

	m, err := loadIgnore(dir)
	if err != nil {
		t.Fatalf("loadIgnore: %v", err)
	}
	if m == nil {
		t.Fatal("expected matcher, got nil")
	}

	tests := []struct {
		path string
		want bool
	}{
		{"node_modules", true},
		{"node_modules/foo/bar.js", true},
		{".nuxt/index.html", true},
		{".output/server.mjs", true},
		{"dist/app.js", true},
		{"server.log", true},
		{"important.log", false},
		{"logs/important.log", false}, // negation re-includes
		{"src/foo/generated/file.go", true},
		{"src/foo/generated", true},
		{"src/foo/main.go", false},
		{"main.go", false},
		{"README.md", false},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := m.matches(tt.path); got != tt.want {
				t.Errorf("matches(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestLoadIgnoreMissing(t *testing.T) {
	dir := t.TempDir()
	m, err := loadIgnore(dir)
	if err != nil {
		t.Fatalf("loadIgnore: %v", err)
	}
	if m != nil {
		t.Errorf("expected nil matcher when file is absent, got %#v", m)
	}
	// nil matcher must be safe to call.
	if m.matches("anything") {
		t.Errorf("nil matcher should not match")
	}
}

func TestLoadIgnoreOnlyComments(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, IgnoreFileName), []byte("# just a comment\n\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	m, err := loadIgnore(dir)
	if err != nil {
		t.Fatal(err)
	}
	if m != nil {
		t.Errorf("expected nil matcher for empty rule set")
	}
}

// TestZipDirHonorsDinaignore verifies the matcher is wired into ZipDir's walk
// fallback. We use a non-git directory so the walk path is exercised.
func TestZipDirHonorsDinaignore(t *testing.T) {
	dir := t.TempDir()
	files := map[string]string{
		"main.go":                   "package main",
		"README.md":                 "hi",
		"node_modules/lib/index.js": "// big",
		".output/server.mjs":        "// build",
		"error.log":                 "boom",
		"debug.log":                 "keep me",
		IgnoreFileName: `node_modules
.output
*.log
!debug.log
`,
	}
	for rel, body := range files {
		full := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	data, err := ZipDir(dir)
	if err != nil {
		t.Fatalf("ZipDir: %v", err)
	}

	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("zip.NewReader: %v", err)
	}
	var got []string
	for _, f := range r.File {
		got = append(got, f.Name)
	}
	sort.Strings(got)

	want := []string{
		IgnoreFileName,
		"README.md",
		"debug.log",
		"main.go",
	}
	sort.Strings(want)

	if len(got) != len(want) {
		t.Fatalf("zip contents = %v, want %v", got, want)
	}
	for i, name := range got {
		if name != want[i] {
			t.Fatalf("zip contents = %v, want %v", got, want)
		}
	}
}
