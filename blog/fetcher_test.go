package blog_test

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/lainio/err2/try"
	"github.com/lauravuo/vegaanibotti/blog"
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

	posts, err := blog.FetchNewPosts("./test_data/recipes.json", getter)
	if err != nil {
		t.Errorf("Expected success, got: %s", err)
	}

	if len(posts) == 0 {
		t.Errorf("Expected to find posts, got 0 posts.")
	}
}
