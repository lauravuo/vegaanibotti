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
	boldFontFile = "./font/Amatic_SC/AmaticSC-Bold.ttf"
	regFontFile  = "./font/Amatic_SC/AmaticSC-Regular.ttf"
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
		Author:      "A very very very long author", //nolint:dupword
	}

	path1, path2 := img.GenerateThumbnail(
		&post, "./vegaanibotti.png", testDataPath+"/thumbnail", boldFontFile, regFontFile)

	if path1 == "" {
		t.Error("Invalid image path")
	}

	if path2 == "" {
		t.Error("Invalid small image path")
	}

	if _, err := os.Stat(testDataPath + "/thumbnail.png"); errors.Is(err, os.ErrNotExist) {
		t.Error("Thumbnail does not exist")
	}

	if _, err := os.Stat(testDataPath + "/thumbnail_small.png"); errors.Is(err, os.ErrNotExist) {
		t.Error("Small thumbnail does not exist")
	}
}
