package kk_test

import (
	"os"
	"testing"

	"github.com/lainio/err2/try"
	"github.com/lauravuo/vegaanibotti/blog/kk"
)

const testDataPath = "./test_data/"

func getter(url, _ string) ([]byte, error) {
	if url == "https://www.kasviskapina.fi/" {
		return []byte(`<html><body><script src="/_next/static/test-hash/_buildManifest.js"></script></body></html>`), nil
	}
	data := try.To1(os.ReadFile("./test.json"))

	return data, nil
}

func setup() {
	try.To(os.MkdirAll(testDataPath, 0o700))
}

func teardown() {
	os.RemoveAll(testDataPath)
}

func TestMain(m *testing.M) {
	setup()

	code := m.Run()

	teardown()

	os.Exit(code)
}

func TestFetchNewPosts(t *testing.T) {
	t.Parallel()

	recipes, err := kk.FetchNewPosts("./test_data/recipes.json", getter, nil, false)
	if err != nil {
		t.Errorf("Expected success, got: %s", err)
	}

	// Our test.json has 5 posts, but we filter for 'paaruoka' category.
	// In the curl test, we saw that many posts have 'paaruoka' in categories.
	if len(recipes.Posts) == 0 {
		t.Errorf("Expected to find posts, got 0.")
	}

	for _, post := range recipes.Posts {
		if post.ID == 0 {
			t.Errorf("Post missing ID")
		}
		if post.Title == "" {
			t.Errorf("Post missing Title")
		}
		if post.URL == "" {
			t.Errorf("Post missing URL")
		}
		if post.ThumbnailURL == "" {
			t.Errorf("Post missing ThumbnailURL")
		}
		if post.ImageURL == "" {
			t.Errorf("Post missing ImageURL")
		}
		if post.Hashtags == nil || len(post.Hashtags) == 0 {
			t.Errorf("Post missing Hashtags")
		}
	}
}
