package vmm_test

import (
	"errors"
	"net/url"
	"os"
	"testing"

	"github.com/lainio/err2/try"
	"github.com/lauravuo/vegaanibotti/blog/vmm"
)

const testDataPath = "./test_data/"

var errNotFound = errors.New("not found")

func getter(_, _ string) ([]byte, error) {
	return []byte("pcajaxamore_scroll = {\"nonce\":\"nonce\""), nil
}

func poster(_ string, params url.Values, _ string) ([]byte, error) {
	if params.Get("offset") == "0" {
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

//nolint:cyclop
func TestFetchNewPosts(t *testing.T) {
	t.Parallel()

	recipes, err := vmm.FetchNewPosts("./test_data/recipes.json", getter, poster, false)
	if err != nil {
		t.Errorf("Expected success, got: %s", err)
	}

	if len(recipes.Posts) != 6 {
		t.Errorf("Expected to find 6 posts, got %d posts.", len(recipes.Posts))
	}

	post := recipes.Posts[0]
	if post.ID != 16465 {
		t.Errorf("Mismatch with post ID")
	}

	if post.Title != "Haudutettu vegechili – chili sin carnella täytetyt enchiladat" {
		t.Errorf("Mismatch with post title")
	}

	if post.URL != "https://viimeistamuruamyoten.com/haudutettu-vegechili-chili-sin-carnella-taytetyt-enchiladat/" {
		t.Errorf("Mismatch with post url")
	}

	if post.Description != "" {
		t.Errorf("Mismatch with post desc")
	}

	if post.ThumbnailURL !=
		"https://viimeistamuruamyoten.com/wp-content/uploads/2021/10/vegechili-chili-sin-carne-enchiladat-vaaka-585x390.jpg" {
		t.Errorf("Mismatch with post thumbnail")
	}

	if post.ImageURL != post.ThumbnailURL {
		t.Errorf("Mismatch with post image: " + post.ImageURL)
	}

	if len(post.Hashtags) != 3 {
		t.Errorf("Mismatch with post hashtags")
	}

	post2 := recipes.Posts[1]
	if post2.Title == "" || post2.Title == post.Title {
		t.Errorf("Invalid second post")
	}
}
