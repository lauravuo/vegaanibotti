package bot_test

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lauravuo/vegaanibotti/blog/base"
	"github.com/lauravuo/vegaanibotti/bot"
)

//nolint:paralleltest // This test modifies global state (cwd) so it cannot run in parallel
func TestPostToSite_EscapesQuotesInTitle(t *testing.T) {
	site := bot.InitSite()

	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Change to temp directory
	t.Chdir(tempDir)

	post := &base.Post{
		Title:        "Test \"quoted\" title",
		ThumbnailURL: "https://example.com/image.png",
		URL:          "https://example.com/post",
		Author:       "Test Author",
	}

	err := site.PostToSite(post)
	if err != nil {
		t.Fatalf("PostToSite failed: %v", err)
	}

	content := findGeneratedFile(t, "./site/content")

	if content == nil {
		t.Fatal("No file was generated")
	}

	contentStr := string(content)

	// Check that quotes are properly escaped
	if !strings.Contains(contentStr, `title: "Test \"quoted\" title"`) {
		t.Errorf("Expected escaped quotes in title, got: %s", contentStr)
	}
}

func findGeneratedFile(t *testing.T, baseDir string) []byte {
	t.Helper()

	var foundContent []byte

	err := filepath.WalkDir(baseDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && strings.HasSuffix(d.Name(), ".md") {
			content, readErr := os.ReadFile(path)
			if readErr != nil {
				return fmt.Errorf("read file %s: %w", path, readErr)
			}

			foundContent = content

			return filepath.SkipAll // Stop searching after finding the first markdown file
		}

		return nil
	})
	if err != nil {
		t.Logf("Error walking directory: %v", err)

		return nil
	}

	return foundContent
}

func TestEscapeYAMLString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    `simple string`,
			expected: `simple string`,
		},
		{
			input:    `string with "quotes"`,
			expected: `string with \"quotes\"`,
		},
		{
			input:    `string with backslash \`,
			expected: `string with backslash \\`,
		},
		{
			input:    `complex "quoted" and \ backslash`,
			expected: `complex \"quoted\" and \\ backslash`,
		},
	}

	for _, tt := range tests {
		result := bot.EscapeYAMLString(tt.input)

		if result != tt.expected {
			t.Errorf("escapeYAMLString(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}
