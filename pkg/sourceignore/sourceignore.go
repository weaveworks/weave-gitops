package sourceignore

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
)

const (
	// Name of the ignore file
	IgnoreFilename = ".sourceignore"
	// Default ignore patterns for version control system files
	ExcludeVCS = ".git/,.gitignore,.gitmodules,.gitattributes"
	// Default ignore patterns for the CI/CD process
	ExcludeCI = ".github/,.circleci/,.travis.yml,.gitlab-ci.yml,appveyor.yml,.drone.yml,cloudbuild.yaml,codeship-services.yml,codeship-steps.yml"
	// Additional default ignore patterns
	ExcludeExtra = "**/.goreleaser.yml,**/.goreleaser.brew.yml,**/.sops.yaml,**/.flux.yaml,**/.golangci.yaml"
)

var (
	// ErrIgnoreFileExists is returned when the ignore file already exists when attempting to create it
	ErrIgnoreFileExists = fmt.Errorf("%s file  already exists", IgnoreFilename)
)

// IgnoreFilter ignores certain files based on a list of patterns and domain.
type IgnoreFilter func(p string, fi os.FileInfo) bool

// VCSPatterns returns a gitignore.Pattern slice with ExcludeVCS
// patterns.
func VCSPatterns(domain []string) []gitignore.Pattern {
	var ps []gitignore.Pattern
	for _, p := range strings.Split(ExcludeVCS, ",") {
		ps = append(ps, gitignore.ParsePattern(p, domain))
	}
	return ps
}

// DefaultPatterns returns a gitignore.Pattern slice with the default
// ExcludeCI, ExcludeExtra patterns.
func DefaultPatterns(domain []string) []gitignore.Pattern {
	all := strings.Join([]string{ExcludeCI, ExcludeExtra}, ",")
	var ps []gitignore.Pattern
	for _, p := range strings.Split(all, ",") {
		ps = append(ps, gitignore.ParsePattern(p, domain))
	}
	return ps
}

// NewMatcher returns a gitignore.Matcher for the given gitignore.Pattern
// slice. It mainly exists to compliment the API.
func NewMatcher(ps []gitignore.Pattern) gitignore.Matcher {
	return gitignore.NewMatcher(ps)
}

// NewDefaultMatcher returns a gitignore.Matcher with the DefaultPatterns
// as lowest priority patterns.
func NewDefaultMatcher(ps []gitignore.Pattern, domain []string) gitignore.Matcher {
	var defaultPs []gitignore.Pattern
	defaultPs = append(defaultPs, VCSPatterns(domain)...)
	defaultPs = append(defaultPs, DefaultPatterns(domain)...)
	ps = append(defaultPs, ps...)

	return gitignore.NewMatcher(ps)
}

// IgnoreFileFilter returns a IgnoreFilter
// that ignores certain files based on a list of patterns and domain.
func IgnoreFileFilter(ps []gitignore.Pattern, domain []string) IgnoreFilter {
	matcher := NewDefaultMatcher(ps, domain)
	return func(p string, fi os.FileInfo) bool {
		return matcher.Match(strings.Split(p, string(filepath.Separator)), fi.IsDir())
	}
}

// CreateIgnoreFile creates a new ignore file at the given directory
func CreateIgnoreFile(dir, ignoreFileName string, ignorePatternStrings []string) error {
	ignoreFilePath := filepath.Join(dir, ignoreFileName)

	// Check if the ignore file already exists
	if _, err := os.Stat(ignoreFilePath); err == nil {
		return ErrIgnoreFileExists
	}

	// Create the ignore file and write rules to it
	f, err := os.Create(ignoreFilePath)
	if err != nil {
		return fmt.Errorf("failed to create the %s file: %w", ignoreFileName, err)
	}
	defer f.Close()

	// Write the GitOps Run comment and ignore rules to the ignore file
	w := bufio.NewWriter(f)
	if _, err := fmt.Fprintln(w, "# Created by GitOps Run.\n# Please add ignore patterns to ignore specific YAML files or directories during validation below"); err != nil {
		return fmt.Errorf("failed to write comment to the %s file: %w", ignoreFileName, err)
	}

	for _, rule := range ignorePatternStrings {
		if _, err := fmt.Fprintln(w, strings.TrimSpace(rule)); err != nil {
			return fmt.Errorf("failed to write rule to the %s file: %w", ignoreFileName, err)
		}
	}

	if err := w.Flush(); err != nil {
		return fmt.Errorf("failed to flush buffer to the %s file: %w", ignoreFileName, err)
	}

	return nil
}

// ReadPatterns collects ignore patterns from the given reader and
// returns them as a gitignore.Pattern slice.
// If a domain is supplied, this is used as the scope of the read
// patterns.
func ReadPatterns(reader io.Reader, domain []string) []gitignore.Pattern {
	var ps []gitignore.Pattern
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		s := scanner.Text()
		if !strings.HasPrefix(s, "#") && len(strings.TrimSpace(s)) > 0 {
			ps = append(ps, gitignore.ParsePattern(s, domain))
		}
	}
	return ps
}

// ReadIgnoreFile attempts to read the file at the given path and
// returns the read patterns.
func ReadIgnoreFile(path string, domain []string) ([]gitignore.Pattern, error) {
	var ps []gitignore.Pattern
	if f, err := os.Open(path); err == nil {
		defer f.Close()
		ps = append(ps, ReadPatterns(f, domain)...)
	} else if !os.IsNotExist(err) {
		return nil, err
	}
	return ps, nil
}
