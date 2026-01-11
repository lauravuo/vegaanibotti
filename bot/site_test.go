package bot

import (
	"os"
	"testing"

	"github.com/lauravuo/vegaanibotti/blog/base"
)

func TestPostToSite_EscapesQuotesInTitle(t *testing.T) {
	site := InitSite()

	// Create a temporary directory for testing
	tempDir := t.TempDir()
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(originalDir)

	// Change to temp directory
	if err := os.Chdir(tempDir); err != nil {
		t.Fatal(err)
	}

	post := &base.Post{
		Title:        "Test \"quoted\" title",
		ThumbnailURL: "https://example.com/image.png",
		URL:          "https://example.com/post",
		Author:       "Test Author",
	}

	err = site.PostToSite(post)
	if err != nil {
		t.Fatalf("PostToSite failed: %v", err)
	}

	// Read the generated file
	// The file should be in ./site/content/YYYY/MM/YYYY-MM-DD.md
	// We need to find it
	entries, err := os.ReadDir("./site/content")
	if err != nil {
		t.Fatalf("Failed to read site/content: %v", err)
	}

	var content []byte
	for _, yearDir := range entries {
		if !yearDir.IsDir() {
			continue
		}
		monthEntries, err := os.ReadDir("./site/content/" + yearDir.Name())
		if err != nil {
			continue
		}
		for _, monthDir := range monthEntries {
			if !monthDir.IsDir() {
				continue
			}
			files, err := os.ReadDir("./site/content/" + yearDir.Name() + "/" + monthDir.Name())
			if err != nil {
				continue
			}
			for _, file := range files {
				if !file.IsDir() {
					content, err = os.ReadFile("./site/content/" + yearDir.Name() + "/" + monthDir.Name() + "/" + file.Name())
					if err != nil {
						t.Fatalf("Failed to read generated file: %v", err)
					}
					break
				}
			}
		}
	}

	if content == nil {
		t.Fatal("No file was generated")
	}

	contentStr := string(content)

	// Check that quotes are properly escaped
	if !contains(contentStr, `title: "Test \"quoted\" title"`) {
		t.Errorf("Expected escaped quotes in title, got: %s", contentStr)
	}
}

func TestEscapeYAMLString(t *testing.T) {
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
		result := escapeYAMLString(tt.input)
		if result != tt.expected {
			t.Errorf("escapeYAMLString(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
