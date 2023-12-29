package cc_test

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/lainio/err2/try"
	"github.com/lauravuo/vegaanibotti/blog/cc"
)

const testDataPath = "./test_data/"

var errNotFound = errors.New("not found")

func getter(url, _ string) ([]byte, error) {
	if strings.Contains(url, "1") {
		data := try.To1(os.ReadFile("./test_data.txt"))

		return data, nil
	}

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

func TestFetchNewPosts(t *testing.T) {
	t.Parallel()

	posts, err := cc.FetchNewPosts("./test_data/recipes.json", getter, false)
	if err != nil {
		t.Errorf("Expected success, got: %s", err)
	}

	if len(posts) != 10 {
		t.Errorf("Expected to find 10 posts, got %d posts.", len(posts))
	}

	post := posts[0]
	if post.ID != 19236 {
		t.Errorf("Mismatch with post ID")
	}

	if post.Title != "Helppo vegemureke" {
		t.Errorf("Mismatch with post title")
	}

	if post.URL != "https://chocochili.net/2023/12/helppo-vegemureke/" {
		t.Errorf("Mismatch with post url")
	}

	if post.Description != "Vegaaninen mureke sopii myös joulupöytään!" {
		t.Errorf("Mismatch with post desc")
	}

	if len(post.Hashtags) != 6 {
		t.Errorf("Mismatch with post hashtags")
	}

	post2 := posts[1]
	if post2.Title == "" || post2.Title == post.Title {
		t.Errorf("Invalid second post")
	}
}
