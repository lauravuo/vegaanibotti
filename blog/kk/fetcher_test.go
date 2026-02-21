package kk_test

import (
	"errors"
	"os"
	"testing"

	"github.com/lainio/err2/try"
	"github.com/lauravuo/vegaanibotti/blog/kk"
)

const testDataPath = "./test_data/"

var errTest = errors.New("test error")

func getter(url, _ string) ([]byte, error) {
	if url == "https://www.kasviskapina.fi/" {
		return []byte(`<html><body><script src="/_next/static/test-hash/_buildManifest.js"></script></body></html>`), nil
	}

	data := try.To1(os.ReadFile("./test.json"))

	return data, nil
}

func getterNoManifest(_, _ string) ([]byte, error) {
	return []byte(`<html><body></body></html>`), nil
}

func getterNoPaaruoka(url, _ string) ([]byte, error) {
	if url == "https://www.kasviskapina.fi/" {
		return []byte(`<html><body><script src="/_next/static/test-hash/_buildManifest.js"></script></body></html>`), nil
	}

	return []byte(`{"pageProps":{"category":{"posts":[]}}}`), nil
}

func getterError(_, _ string) ([]byte, error) {
	return nil, errTest
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

		if len(post.Hashtags) == 0 {
			t.Errorf("Post missing Hashtags")
		}
	}
}

func TestFetchNewPostsFallback(t *testing.T) {
	t.Parallel()

	// getter returns an error - FetchNewPosts should fall back to cached posts (empty)
	recipes, err := kk.FetchNewPosts("./test_data/recipes_fallback.json", getterError, nil, false)
	if err != nil {
		t.Errorf("Expected success with fallback, got: %s", err)
	}

	if len(recipes.Posts) != 0 {
		t.Errorf("Expected 0 fallback posts, got %d", len(recipes.Posts))
	}
}

func TestFetchNewPostsNoManifest(t *testing.T) {
	t.Parallel()

	// getter returns HTML without build manifest - should fall back to cached posts (empty)
	recipes, err := kk.FetchNewPosts("./test_data/recipes_no_manifest.json", getterNoManifest, nil, false)
	if err != nil {
		t.Errorf("Expected success with fallback, got: %s", err)
	}

	if len(recipes.Posts) != 0 {
		t.Errorf("Expected 0 fallback posts, got %d", len(recipes.Posts))
	}
}

func TestFetchNewPostsNoPaaruoka(t *testing.T) {
	t.Parallel()

	// getter returns valid response but zero paaruoka posts - should fall back
	recipes, err := kk.FetchNewPosts("./test_data/recipes_no_paaruoka.json", getterNoPaaruoka, nil, false)
	if err != nil {
		t.Errorf("Expected success with fallback, got: %s", err)
	}

	if len(recipes.Posts) != 0 {
		t.Errorf("Expected 0 fallback posts, got %d", len(recipes.Posts))
	}
}

func TestFetchNewPostsPreviewOnly(t *testing.T) {
	t.Parallel()

	recipes, err := kk.FetchNewPosts("./test_data/recipes_preview.json", getter, nil, true)
	if err != nil {
		t.Errorf("Expected success, got: %s", err)
	}

	if len(recipes.Posts) == 0 {
		t.Errorf("Expected to find posts, got 0.")
	}

	if _, statErr := os.Stat("./test_data/recipes_preview.json"); !errors.Is(statErr, os.ErrNotExist) {
		t.Errorf("Expected recipes file to not be written in preview mode")
	}
}

