package bot_test

import (
	"context"
	"testing"

	"github.com/lauravuo/vegaanibotti/blog"
	"github.com/lauravuo/vegaanibotti/bot"
	"github.com/mattn/go-mastodon"
)

// MockMastodonClient is a mock implementation of the Mastodon client for testing purposes.
type MockMastodonClient struct{}

func (m *MockMastodonClient) PostStatus(_ context.Context, _ *mastodon.Toot) (*mastodon.Status, error) {
	return &mastodon.Status{}, nil
}

func TestPostToMastodon(t *testing.T) {
	t.Parallel()

	mastodonClient := &MockMastodonClient{}

	// Create a sample blog post
	post := &blog.Post{
		Title:       "Test Title",
		Description: "Test Description",
		URL:         "http://example.com",
		Hashtags:    []string{"tag1", "tag2"},
	}

	// Call the function to be tested
	m := &bot.Mastodon{Client: mastodonClient}
	err := m.PostToMastodon(post)
	// Assert that there is no error
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}
