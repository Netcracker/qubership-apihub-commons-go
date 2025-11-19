package scanner

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/qubership-apihub-commons-go/api-spec-exposer/config"
)

// Scanner handles directory scanning and file discovery
type Scanner struct {
	config          config.DiscoveryConfig
	identifierChain *IdentifierChain
}

// New creates a new scanner instance
func New(cfg config.DiscoveryConfig) *Scanner {
	return &Scanner{
		config: cfg,
		identifierChain: &IdentifierChain{
			identifiers: []Identifier{
				&RestIdentifier{},
				&GraphQLIdentifier{},
				&MarkdownIdentifier{},
				&BasicIdentifier{},
			},
		},
	}
}

// Scan scans the directory and returns discovered specs
func (s *Scanner) Scan() ([]config.SpecMetadata, []string, error) {
	var specs []config.SpecMetadata
	var warnings []string

	if s.config.ScanDirectory == "" {
		return nil, warnings, fmt.Errorf("scan directory is empty")
	}

	info, err := os.Stat(s.config.ScanDirectory)
	if err != nil {
		return nil, warnings, fmt.Errorf("cannot access scan directory: %w", err)
	}

	if !info.IsDir() {
		return nil, warnings, fmt.Errorf("scan directory is not a directory")
	}

	err = filepath.WalkDir(s.config.ScanDirectory, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("error accessing path %s: %v", path, err))
			return nil // Continue walking
		}

		if d.IsDir() {
			// Check if directory should be excluded
			if s.shouldExclude(path) {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if file should be excluded
		if s.shouldExclude(path) {
			return nil
		}

		content, err := s.readFile(path)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("cannot read file %s: %v", path, err))
			return nil
		}

		spec, err := s.identifierChain.Identify(path, content)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("detection failed for %s: %v", path, err))
			return nil
		}

		if spec != nil {
			specs = append(specs, *spec)
		}

		return nil
	})

	if err != nil {
		return specs, warnings, fmt.Errorf("error walking directory: %w", err)
	}

	return specs, warnings, nil
}

func (s *Scanner) shouldExclude(path string) bool {
	// Skip hidden files/directories (starting with .)
	base := filepath.Base(path)
	if strings.HasPrefix(base, ".") {
		return true
	}

	for _, pattern := range s.config.ExcludePatterns {
		matched, err := filepath.Match(pattern, base)
		if err == nil && matched {
			return true
		}

		// Also check full path matching
		relPath, err := filepath.Rel(s.config.ScanDirectory, path)
		if err == nil {
			matched, err := filepath.Match(pattern, relPath)
			if err == nil && matched {
				return true
			}

			if strings.Contains(relPath, strings.TrimSuffix(pattern, "/*")) {
				return true
			}
		}
	}

	return false
}

func (s *Scanner) readFile(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("cannot open file: %w", err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("cannot read file: %w", err)
	}

	return content, nil
}
