package blog_test

import (
	"os"
	"testing"

	"github.com/lainio/err2/try"
	"github.com/lauravuo/vegaanibotti/blog"
	"github.com/lauravuo/vegaanibotti/blog/base"
)

const usedIDsPath = "./test_data/used.json"

func TestChooseNextPost(t *testing.T) {
	t.Parallel()

	// test when empty used ids
	posts := base.Collection{
		"cc": []base.Post{{
			ID:          1,
			Title:       "title",
			Description: "description",
			URL:         "https://example.com",
			Hashtags:    []string{"food"},
			Added:       true,
		}},
	}
	nextPost := blog.ChooseNextPost(posts, usedIDsPath)

	if nextPost.ID != posts["cc"][0].ID {
		t.Errorf("Mismatch with expected post id %d (%d)", nextPost.ID, posts["cc"][0].ID)
	}

	// test when one of the ids used
	posts = base.Collection{
		"cc": []base.Post{
			{ID: 1, Title: "title", Description: "description", URL: "https://example.com", Hashtags: []string{"food"}, Added: true},

			{ID: 2, Title: "title", Description: "description", URL: "https://example.com", Hashtags: []string{"food"}, Added: true},
		},
	}

	try.To(os.WriteFile(usedIDsPath, []byte("[1]"), base.WritePerm))

	nextPost = blog.ChooseNextPost(posts, usedIDsPath)

	if nextPost.ID != posts["cc"][1].ID {
		t.Errorf("Mismatch with expected post id %d (%d)", nextPost.ID, posts["cc"][0].ID)
	}

	// test when all of the ids used
	contents := try.To1(os.ReadFile(usedIDsPath))
	if string(contents) != "[1,2]" {
		t.Errorf("Mismatch with expected ids %s", string(contents))
	}

	nextPost = blog.ChooseNextPost(posts, usedIDsPath)
	expected := "[1]"

	if nextPost.ID == 2 {
		expected = "[2]"
	}

	contents = try.To1(os.ReadFile(usedIDsPath))

	if string(contents) != expected {
		t.Errorf("Mismatch with expected ids %s (%s)", string(contents), expected)
	}
}
