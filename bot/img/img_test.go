package img_test

import (
	"errors"
	"os"
	"testing"

	"github.com/lainio/err2/try"
	"github.com/lauravuo/vegaanibotti/blog/base"
	"github.com/lauravuo/vegaanibotti/bot/img"
)

const (
	testDataPath = "./test_data"
)

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

func TestGenerateThumbnail(t *testing.T) {
	t.Parallel()

	post := base.Post{
		ID:          1,
		Title:       "title",
		Description: "description",
		URL:         "https://example.com",
		Hashtags:    []string{"food"},
		Added:       true,
	}

	img.GenerateThumbnail(&post, "./vegaanibotti.png", testDataPath+"/thumbnail")

	if _, err := os.Stat(testDataPath + "/thumbnail.png"); errors.Is(err, os.ErrNotExist) {
		t.Error("Thumbnail does not exist")
	}

	if _, err := os.Stat(testDataPath + "/thumbnail_small.png"); errors.Is(err, os.ErrNotExist) {
		t.Error("Small thumbnail does not exist")
	}

	entries := try.To1(os.ReadDir("../../site/content/2023/12"))
	for _, path := range entries {
		_ = string(try.To1(os.ReadFile("../../site/content/2023/12/" + path.Name())))
	}
}
