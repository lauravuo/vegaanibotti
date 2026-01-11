package vv_test

import (
	"errors"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/lainio/err2/try"
	"github.com/lauravuo/vegaanibotti/blog/vv"
)

const testDataPath = "./test_data/"

var errNotFound = errors.New("not found")

func getter(targetURL, _ string) ([]byte, error) {
	if strings.Contains(targetURL, "1") {
		data := try.To1(os.ReadFile("./test_data.txt"))

		return data, nil
	}

	if strings.Contains(targetURL, "2") {
		data := try.To1(os.ReadFile("./test_data_2.txt"))

		return data, nil
	}

	return nil, errNotFound
}

func poster(_ string, _ url.Values, _ string) ([]byte, error) {
	return nil, errNotFound
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

//nolint:cyclop
func TestFetchNewPosts(t *testing.T) {
	t.Parallel()

	recipes, err := vv.FetchNewPosts("./test_data/recipes.json", getter, poster, false)
	if err != nil {
		t.Errorf("Expected success, got: %s", err)
	}

	if len(recipes.Posts) != 15 {
		t.Errorf("Expected to find 15 posts, got %d posts.", len(recipes.Posts))
	}

	post := recipes.Posts[0]
	if post.ID != 8610 {
		t.Errorf("Mismatch with post ID")
	}

	if post.Title != "Tomaatti-fetapasta" {
		t.Errorf("Mismatch with post title")
	}

	if post.URL != "https://vegeviettelys.fi/tomaatti-fetapasta/" {
		t.Errorf("Mismatch with post url %s", post.URL)
	}

	if post.Description != "" {
		t.Errorf("Mismatch with post desc")
	}

	if post.ThumbnailURL !=
		"" {
		t.Errorf("Mismatch with post thumbnail")
	}

	if post.ImageURL != post.ThumbnailURL {
		t.Errorf("Mismatch with post image: %s", post.ImageURL)
	}

	if len(post.Hashtags) != 3 {
		t.Errorf("Mismatch with post hashtags")
	}

	post2 := recipes.Posts[1]
	if post2.Title == "" || post2.Title == post.Title {
		t.Errorf("Invalid second post")
	}
}
