package sourceignore

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
	"gotest.tools/assert"
)

func TestReadPatterns(t *testing.T) {
	tests := []struct {
		name       string
		ignore     string
		domain     []string
		matches    []string
		mismatches []string
	}{
		{
			name: "simple",
			ignore: `ignore-dir/*
!ignore-dir/include
`,
			matches:    []string{"ignore-dir/file.yaml"},
			mismatches: []string{"file.yaml", "ignore-dir/include"},
		},
		{
			name: "with comments",
			ignore: `ignore-dir/*
# !ignore-dir/include`,
			matches: []string{"ignore-dir/file.yaml", "ignore-dir/include"},
		},
		{
			name:       "domain scoped",
			domain:     []string{"domain", "scoped"},
			ignore:     "ignore-dir/*",
			matches:    []string{"domain/scoped/ignore-dir/file.yaml"},
			mismatches: []string{"ignore-dir/file.yaml"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.ignore)
			ps := ReadPatterns(reader, tt.domain)
			matcher := NewMatcher(ps)
			for _, m := range tt.matches {
				assert.Equal(t, matcher.Match(strings.Split(m, "/"), false), true, "expected %s to match", m)
			}
			for _, m := range tt.mismatches {
				assert.Equal(t, matcher.Match(strings.Split(m, "/"), false), false, "expected %s to not match", m)
			}
		})
	}
}

func TestReadIgnoreFile(t *testing.T) {
	f, err := os.CreateTemp("", IgnoreFilename)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	if _, err = f.Write([]byte(`# .sourceignore
ignore-this.txt`)); err != nil {
		t.Fatal(err)
	}
	f.Close()

	tests := []struct {
		name   string
		path   string
		domain []string
		want   []gitignore.Pattern
	}{
		{
			name: IgnoreFilename,
			path: f.Name(),
			want: []gitignore.Pattern{
				gitignore.ParsePattern("ignore-this.txt", nil),
			},
		},
		{
			name:   "with domain",
			path:   f.Name(),
			domain: strings.Split(filepath.Dir(f.Name()), string(filepath.Separator)),
			want: []gitignore.Pattern{
				gitignore.ParsePattern("ignore-this.txt", strings.Split(filepath.Dir(f.Name()), string(filepath.Separator))),
			},
		},
		{
			name: "non existing",
			path: "",
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReadIgnoreFile(tt.path, tt.domain)
			if err != nil {
				t.Error(err)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReadIgnoreFile() got = %d, want %#v", got, tt.want)
			}
		})
	}
}

func TestVCSPatterns(t *testing.T) {
	tests := []struct {
		name       string
		domain     []string
		patterns   []gitignore.Pattern
		matches    []string
		mismatches []string
	}{
		{
			name:       "simple matches",
			matches:    []string{".git/config", ".gitignore"},
			mismatches: []string{"workload.yaml", "workload.yml", "simple.txt"},
		},
		{
			name:       "domain scoped matches",
			domain:     []string{"directory"},
			matches:    []string{"directory/.git/config", "directory/.gitignore"},
			mismatches: []string{"other/.git/config"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matcher := NewDefaultMatcher(tt.patterns, tt.domain)
			for _, m := range tt.matches {
				assert.Equal(t, matcher.Match(strings.Split(m, "/"), false), true, "expected %s to match", m)
			}
			for _, m := range tt.mismatches {
				assert.Equal(t, matcher.Match(strings.Split(m, "/"), false), false, "expected %s to not match", m)
			}
		})
	}
}

func TestDefaultPatterns(t *testing.T) {
	tests := []struct {
		name       string
		domain     []string
		patterns   []gitignore.Pattern
		matches    []string
		mismatches []string
	}{
		{
			name:       "simple matches",
			matches:    []string{".github/workflows/workflow.yaml", "subdir/.flux.yaml", "subdir2/.sops.yaml"},
			mismatches: []string{"workload.yaml", "workload.yml", "simple.txt"},
		},
		{
			name:       "domain scoped matches",
			domain:     []string{"directory"},
			matches:    []string{"directory/.goreleaser.yml", "directory/.goreleaser.yml"},
			mismatches: []string{"other/.goreleaser.yml", "other/.goreleaser.yml"},
		},
		{
			name:       "patterns",
			patterns:   []gitignore.Pattern{gitignore.ParsePattern("!*.txt", nil)},
			mismatches: []string{"simple.txt"},
		},
		{
			name:       "domain scoped matches with custom patterns",
			domain:     []string{"directory"},
			patterns:   []gitignore.Pattern{gitignore.ParsePattern("**/workflow.yaml,**/simple.txt", []string{"directory"})},
			matches:    []string{"directory/.goreleaser.yml", "directory/.goreleaser.yml"},
			mismatches: []string{"other/.goreleaser.yml", "other/.goreleaser.yml"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matcher := NewDefaultMatcher(tt.patterns, tt.domain)
			for _, m := range tt.matches {
				assert.Equal(t, matcher.Match(strings.Split(m, "/"), false), true, "expected %s to match", m)
			}
			for _, m := range tt.mismatches {
				assert.Equal(t, matcher.Match(strings.Split(m, "/"), false), false, "expected %s to not match", m)
			}
		})
	}
}
